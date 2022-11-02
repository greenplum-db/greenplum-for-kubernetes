package admission_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Validating Admission Webhook", func() {

	var (
		subject        admission.Webhook
		reactiveClient *reactive.Client
		fakeOwnerPod   *corev1.Pod
		fakeOwnerCRD   *apiextensionsv1.CustomResourceDefinition
		serviceName    string
		cg             *StubCertGenerator
		logBuf         *gbytes.Buffer
	)

	BeforeEach(func() {
		cg = &StubCertGenerator{}

		hashString := "-hash123-hash456"
		serviceName = admission.ServiceName + hashString

		fakeOwnerPod = &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-ns",
				Name:      "greenplum-operator" + hashString,
				UID:       "testUID",
			},
		}

		fakeOwnerCRD = &apiextensionsv1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "greenplumclusters.greenplum.pivotal.io",
				UID:  "testUID",
			},
		}

		reactiveClient = reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
		subject = admission.Webhook{
			KubeClient:      reactiveClient,
			Namespace:       "test-ns",
			ServiceOwner:    fakeOwnerPod,
			WebhookCfgOwner: fakeOwnerCRD,
			NameSuffix:      hashString,
			CertGenerator:   cg,
		}
		logBuf = gbytes.NewBuffer()
		admission.Log = gplog.ForTest(logBuf)
	})

	Describe("Run", func() {
		var mockServer *MockServer

		BeforeEach(func() {
			mockServer = &MockServer{}
			subject.Server = mockServer
			subject.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
				_, _ = writer.Write([]byte("I am the mock handler"))
			})
		})
		When("all is good", func() {
			It("logs startup and shutdown messages", func() {
				Expect(subject.Run(nil)).To(Succeed())
				Expect(logBuf).To(gbytes.Say("starting greenplum validating admission webhook server"))
				// shut down
				Expect(logBuf).To(gbytes.Say("shutting down greenplum validating admission webhook server"))
			})

			It("Creates validatingwebhookconfiguration", func() {
				Expect(subject.Run(nil)).To(Succeed())
				var webhookConfig admissionregistrationv1.ValidatingWebhookConfiguration
				webhookKey := types.NamespacedName{Name: admission.WebhookConfigName}
				Expect(reactiveClient.Get(nil, webhookKey, &webhookConfig)).To(Succeed())
				Expect(webhookConfig).NotTo(BeNil())
				Expect(webhookConfig.Webhooks[0].ClientConfig.CABundle).To(Equal(cg.waitStub.returnedCert))
			})

			It("starts a webhook server", func() {
				mockServer.started = make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				go subject.Run(ctx)

				Eventually(mockServer.started, 5*time.Second).Should(BeClosed())
				Expect(mockServer.cert).To(Equal(cg.getCertStub.returnedX509))
				Expect(mockServer.addr).To(Equal(":https"))
				resp := httptest.NewRecorder()
				mockServer.handler.ServeHTTP(resp, nil)
				Expect(resp.Body.String()).To(Equal("I am the mock handler"))

				Consistently(logBuf).ShouldNot(gbytes.Say("shutting down greenplum validating admission webhook server"))
				cancel()
				Eventually(logBuf).Should(gbytes.Say("shutting down greenplum validating admission webhook server"))
			})
		})

		When("GenerateAndSignTLSCertificate fails", func() {
			It("does not start a webhook server", func() {
				cg.getCertStub.err = errors.New("injected failure")
				err := subject.Run(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp(`getting certificate for webhook: [^"]*: injected failure`))
				Expect(logBuf).NotTo(gbytes.Say("shutting down greenplum validating admission webhook server"))
			})
		})

		When("ReconcileValidatingWebhookConfiguration fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("create", "validatingwebhookconfigurations", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("intentional failure")
				})
			})
			It("returns an error", func() {
				err := subject.Run(nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp(`creating ValidatingWebhookConfiguration: [^"]*: intentional failure`))
				Expect(logBuf).NotTo(gbytes.Say("shutting down greenplum validating admission webhook server"))
			})
		})

		When("Server.Start fails", func() {
			BeforeEach(func() {
				subject.Server = &MockServer{err: errors.New("intentional failure")}
			})
			It("returns an error", func() {
				Expect(subject.Run(nil)).To(MatchError("validating admission webhook server start failed: intentional failure"))
				Expect(logBuf).NotTo(gbytes.Say("shutting down greenplum validating admission webhook server"))
			})
		})
	})

	Describe("GenerateAndSignTLSCertificate", func() {
		When("all is good", func() {
			It("generates a certificate", func() {
				signedCertPEM, signedCertX509, err := subject.GenerateAndSignTLSCertificate()
				Expect(err).NotTo(HaveOccurred())

				Expect(cg.generateStub.receivedCommonName).To(Equal(serviceName + ".test-ns.svc"))
				Expect(cg.generateStub.returnedCSR).To(Equal(cg.createStub.receivedCert))
				Expect(cg.approveStub.receivedCSR).To(And(Not(BeNil()), Equal(cg.createStub.returnedCSR)))
				Expect(cg.waitStub.receivedCSR).To(And(Not(BeNil()), Equal(cg.approveStub.returnedCSR)))
				Expect(cg.waitStub.receivedTimeout).To(Equal(30 * time.Second))
				Expect(cg.getCertStub.receivedCert).To(And(Not(BeNil()), Equal(cg.waitStub.returnedCert)))
				Expect(cg.getCertStub.receivedKey).To(And(Not(BeNil()), Equal(cg.generateStub.returnedKey)))

				Expect(signedCertPEM).To(And(Not(BeNil()), Equal(cg.waitStub.returnedCert)))
				Expect(signedCertX509).To(Equal(&cg.getCertStub.returnedX509))
			})
		})

		ItReturnsAnError := func(expectedError string) {
			It("returns an error", func() {
				_, _, err := subject.GenerateAndSignTLSCertificate()
				Expect(err).To(MatchError(expectedError))
			})
		}

		When("GenerateX509CertificateSigningRequest fails", func() {
			BeforeEach(func() {
				cg.generateStub.err = errors.New("error")
			})
			ItReturnsAnError("failed to generate certificate signing request: error")
		})

		When("CreateCertificateSigningRequest fails", func() {
			BeforeEach(func() {
				cg.createStub.err = errors.New("error")
			})
			ItReturnsAnError("failed to create certificate signing request: error")
		})

		When("ApproveCertificateSigningRequest fails", func() {
			BeforeEach(func() {
				cg.approveStub.err = errors.New("error")
			})
			ItReturnsAnError("failed to approve certificate signing request: error")
		})

		When("WaitForSignedCertificate fails", func() {
			BeforeEach(func() {
				cg.waitStub.err = errors.New("error")
			})
			ItReturnsAnError("failure while waiting for approval: error")
		})

		When("GetCertificate fails", func() {
			BeforeEach(func() {
				cg.getCertStub.err = errors.New("error")
			})
			It("returns an error", func() {
				_, _, err := subject.GenerateAndSignTLSCertificate()
				Expect(err).To(MatchError("error loading keypair: error"))
			})
		})
	})

	Describe("ReconcileValidatingWebhookConfiguration", func() {
		var reactiveClient *reactive.Client
		BeforeEach(func() {
			reactiveClient = reactive.NewClient(fakeClient.NewFakeClientWithScheme(scheme.Scheme))
			subject.KubeClient = reactiveClient
		})
		When("all is good", func() {
			var err error
			BeforeEach(func() {
				err = subject.ReconcileValidatingWebhookConfiguration(nil, []byte("signed cert"))
			})

			It("succeeds", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates a ValidatingWebhookConfiguration", func() {
				var webhookConfig admissionregistrationv1.ValidatingWebhookConfiguration
				webhookKey := types.NamespacedName{Name: admission.WebhookConfigName}
				Expect(reactiveClient.Get(nil, webhookKey, &webhookConfig)).To(Succeed())
				Expect(webhookConfig.Webhooks[0].ClientConfig.CABundle).To(Equal([]byte("signed cert")))
				Expect(metav1.IsControlledBy(&webhookConfig, fakeOwnerCRD)).To(BeTrue())
			})

			It("creates a service for the webhook", func() {
				var service corev1.Service
				serviceKey := types.NamespacedName{Namespace: "test-ns", Name: serviceName}
				Expect(reactiveClient.Get(nil, serviceKey, &service)).To(Succeed())
				Expect(metav1.IsControlledBy(&service, fakeOwnerPod)).To(BeTrue())
			})
		})

		When("create ValidatingWebhookConfiguration fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("create", "validatingwebhookconfigurations", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("injected failure to create validatingwebhookconfiguration")
				})
			})
			It("returns an error", func() {
				Expect(subject.ReconcileValidatingWebhookConfiguration(nil, []byte("signed cert"))).To(
					MatchError("failed to create ValidatingWebhookConfiguration: injected failure to create validatingwebhookconfiguration"))
			})
		})

		When("create service fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("create", "services", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("injected failure to create service")
				})
			})
			It("returns an error", func() {
				Expect(subject.ReconcileValidatingWebhookConfiguration(nil, []byte("signed cert"))).To(
					MatchError("error creating Service for Webhook: injected failure to create service"))
			})
		})

		When("setting the validatingwebhookconfiguration owner reference fails", func() {
			BeforeEach(func() {
				subject.WebhookCfgOwner = nil
			})

			It("returns an error", func() {
				err := subject.ReconcileValidatingWebhookConfiguration(nil, []byte("signed cert"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("couldn't set OwnerReferences on ValidatingWebhookConfig.*"))
			})
		})

		When("setting the service owner reference fails", func() {
			BeforeEach(func() {
				subject.ServiceOwner = nil
			})

			It("returns an error", func() {
				err := subject.ReconcileValidatingWebhookConfiguration(nil, []byte("signed cert"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("couldn't set OwnerReferences on webhook Service.*"))
			})
		})

		When("the ValidatingWebhookConfiguration already exists", func() {
			var err error
			BeforeEach(func() {
				err = subject.ReconcileValidatingWebhookConfiguration(nil, []byte("old cert"))
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func() {
				By("creating a new operator pod")
				subject.NameSuffix = "-new"
				err = subject.ReconcileValidatingWebhookConfiguration(nil, []byte("new cert"))
			})
			It("succeeds", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			It("updates the caBundle", func() {
				var webhookConfig admissionregistrationv1.ValidatingWebhookConfiguration
				webhookKey := types.NamespacedName{Name: admission.WebhookConfigName}
				Expect(reactiveClient.Get(nil, webhookKey, &webhookConfig)).To(Succeed())
				Expect(webhookConfig).NotTo(BeNil())
				Expect(webhookConfig.Webhooks[0].ClientConfig.CABundle).To(Equal([]byte("new cert")))
			})

			It("updates the service name", func() {
				var webhookConfig admissionregistrationv1.ValidatingWebhookConfiguration
				webhookKey := types.NamespacedName{Name: admission.WebhookConfigName}
				Expect(reactiveClient.Get(nil, webhookKey, &webhookConfig)).To(Succeed())
				Expect(webhookConfig).NotTo(BeNil())
				Expect(webhookConfig.Webhooks[0].ClientConfig.Service.Name).To(Equal(admission.ServiceName + "-new"))
			})

			When("Update fails", func() {
				BeforeEach(func() {
					reactiveClient.PrependReactor("update", "validatingwebhookconfigurations", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("failure to update validatingwebhookconfiguration")
					})
				})
				It("returns the error", func() {
					Expect(err).To(MatchError("failed to create ValidatingWebhookConfiguration: failure to update validatingwebhookconfiguration"))
				})
			})
		})
	})

	Describe("ModifyWebhookConfiguration", func() {
		It("returns valid WebhookConfiguration", func() {
			certBytes := []byte("some cert bytes")
			validatingWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: admission.WebhookConfigName,
				},
			}
			subject.ModifyWebhookConfiguration(validatingWebhookConfig, certBytes)

			Expect(validatingWebhookConfig).NotTo(BeNil())
			Expect(validatingWebhookConfig.Name).To(Equal(admission.WebhookConfigName))
			Expect(validatingWebhookConfig.Namespace).To(Equal(""))
			Expect(validatingWebhookConfig.Labels).To(Equal(map[string]string{"app": "greenplum-operator"}))
			Expect(validatingWebhookConfig.Webhooks).To(HaveLen(1))
			validatingWebhook := validatingWebhookConfig.Webhooks[0]
			Expect(validatingWebhook.Name).To(Equal("greenplum.pivotal.io"))
			Expect(validatingWebhook.ClientConfig.Service.Name).To(Equal(serviceName))
			Expect(validatingWebhook.ClientConfig.Service.Namespace).To(Equal("test-ns"))
			Expect(*validatingWebhook.ClientConfig.Service.Path).To(Equal("/validate"))
			Expect(validatingWebhook.ClientConfig.CABundle).To(Equal(certBytes))
			Expect(validatingWebhook.Rules).To(HaveLen(2))
			Expect(validatingWebhook.Rules[0].Operations).To(Equal([]admissionregistrationv1.OperationType{"CREATE", "UPDATE"}))
			Expect(validatingWebhook.Rules[0].APIGroups[0]).To(Equal("greenplum.pivotal.io"))
			Expect(validatingWebhook.Rules[0].APIVersions[0]).To(Equal("v1"))
			Expect(validatingWebhook.Rules[0].Resources[0]).To(Equal("greenplumclusters"))
			Expect(validatingWebhook.Rules[1].Operations).To(Equal([]admissionregistrationv1.OperationType{"CREATE", "UPDATE"}))
			Expect(validatingWebhook.Rules[1].APIGroups[0]).To(Equal("greenplum.pivotal.io"))
			Expect(validatingWebhook.Rules[1].APIVersions[0]).To(Equal("v1beta1"))
			Expect(validatingWebhook.Rules[1].Resources[0]).To(Equal("greenplumpxfservices"))
			Expect(*validatingWebhook.FailurePolicy).To(Equal(admissionregistrationv1.Fail))
		})

		It("leaves existing labels when labels are not empty", func() {
			certBytes := []byte("some cert bytes")
			validatingWebhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: admission.WebhookConfigName,
					Labels: map[string]string{
						"app":       "not-a-greenplum-operator",
						"cool-tool": "soldering-iron", // An extra label we shouldn't touch
					},
				},
			}
			subject.ModifyWebhookConfiguration(validatingWebhookConfig, certBytes)

			Expect(validatingWebhookConfig.Labels).To(HaveKeyWithValue("app", "greenplum-operator"))
			Expect(validatingWebhookConfig.Labels).To(HaveKeyWithValue("cool-tool", "soldering-iron"))

		})
	})

	Describe("CreateSVCForValidatingWebhookConfiguration", func() {
		It("returns a valid svc configuration for validating webhook", func() {
			subject.Namespace = "another-namespace"
			svcConfig := subject.CreateSVCForValidatingWebhookConfiguration()
			Expect(svcConfig.Name).To(Equal(serviceName))
			Expect(svcConfig.Namespace).To(Equal("another-namespace"))
			Expect(svcConfig.APIVersion).To(Equal("v1"))
			Expect(svcConfig.Kind).To(Equal("Service"))
			Expect(svcConfig.Labels).To(Equal(map[string]string{"app": "greenplum-operator"}))
			Expect(svcConfig.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
			Expect(svcConfig.Spec.Ports[0].Name).To(Equal("webhook"))
			Expect(svcConfig.Spec.Ports[0].Port).To(Equal(int32(443)))
			Expect(svcConfig.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(443)))
			Expect(svcConfig.Spec.Selector).To(Equal(map[string]string{"app": "greenplum-operator"}))
		})
	})
})

type MockServer struct {
	started chan struct{}
	cert    tls.Certificate
	addr    string
	handler http.Handler
	err     error
}

var _ admission.Server = &MockServer{}

func (s *MockServer) Start(stopCh <-chan struct{}, cert tls.Certificate, addr string, handler http.Handler) error {
	s.cert = cert
	s.addr = addr
	s.handler = handler
	if s.started != nil {
		close(s.started)
	}
	if stopCh != nil {
		<-stopCh
	}
	if s.err == nil {
		return http.ErrServerClosed
	}
	return s.err
}

func (*MockServer) Shutdown() error {
	panic("implement me")
}

type StubCertGenerator struct {
	generateStub struct {
		receivedCommonName string
		returnedKey        *rsa.PrivateKey
		returnedCSR        []byte
		err                error
	}
	createStub struct {
		receivedCert []byte
		returnedCSR  *v1beta1.CertificateSigningRequest
		err          error
	}
	approveStub struct {
		receivedCSR *v1beta1.CertificateSigningRequest
		returnedCSR *v1beta1.CertificateSigningRequest
		err         error
	}
	waitStub struct {
		receivedCSR     *v1beta1.CertificateSigningRequest
		receivedTimeout time.Duration
		returnedCert    []byte
		err             error
	}
	getCertStub struct {
		receivedCert []byte
		receivedKey  *rsa.PrivateKey
		returnedX509 tls.Certificate
		err          error
	}
}

func (fcg *StubCertGenerator) GenerateX509CertificateSigningRequest(commonName string) (*rsa.PrivateKey, []byte, error) {
	fcg.generateStub.receivedCommonName = commonName
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	fcg.generateStub.returnedKey = rsaKey
	fcg.generateStub.returnedCSR = []byte("CERTIFICATE REQUEST")
	return fcg.generateStub.returnedKey, fcg.generateStub.returnedCSR, fcg.generateStub.err
}

func (fcg *StubCertGenerator) CreateCertificateSigningRequest(cert []byte) (*v1beta1.CertificateSigningRequest, error) {
	fcg.createStub.receivedCert = cert
	fcg.createStub.returnedCSR = &v1beta1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a-fake-csr",
		},
	}
	return fcg.createStub.returnedCSR, fcg.createStub.err
}

func (fcg *StubCertGenerator) ApproveCertificateSigningRequest(csr *v1beta1.CertificateSigningRequest) (*v1beta1.CertificateSigningRequest, error) {
	fcg.approveStub.receivedCSR = csr
	fcg.approveStub.returnedCSR = &v1beta1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: "approved-csr",
		},
	}
	return fcg.approveStub.returnedCSR, fcg.approveStub.err
}

func (fcg *StubCertGenerator) WaitForSignedCertificate(csr *v1beta1.CertificateSigningRequest, timeout time.Duration) ([]byte, error) {
	fcg.waitStub.receivedCSR = csr
	fcg.waitStub.receivedTimeout = timeout
	fcg.waitStub.returnedCert = []byte("signed cert PEM")
	return fcg.waitStub.returnedCert, fcg.waitStub.err
}

func (fcg *StubCertGenerator) GetCertificate(cert []byte, key *rsa.PrivateKey) (tls.Certificate, error) {
	fcg.getCertStub.receivedCert = cert
	fcg.getCertStub.receivedKey = key
	fcg.getCertStub.returnedX509 = tls.Certificate{Certificate: [][]byte{[]byte("cert PEM")}}
	return fcg.getCertStub.returnedX509, fcg.getCertStub.err
}
