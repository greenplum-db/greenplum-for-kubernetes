package admission_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/structmatcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/testing/reactive"
	"k8s.io/api/certificates/v1beta1"
	certificates "k8s.io/api/certificates/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CertificateGenerator", func() {

	var (
		ctx             context.Context
		subject         *admission.CertificateGenerator
		simpleClientSet *testclient.Clientset
		reactiveClient  *reactive.Client
		fakeOwner       *apiextensionsv1.CustomResourceDefinition
	)

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		simpleClientSet = testclient.NewSimpleClientset()
		reactiveClient = reactive.NewClient(fake.NewFakeClientWithScheme(scheme.Scheme))

		fakeOwner = &apiextensionsv1.CustomResourceDefinition{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CustomResourceDefinition",
				APIVersion: "apiextensions.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "greenplumclusters.greenplum.pivotal.io",
				UID:  "testUID",
			},
		}

		subject = &admission.CertificateGenerator{
			CtrlClient:    reactiveClient,
			KubeClientSet: simpleClientSet,
			Owner:         fakeOwner,
		}
	})

	Describe("GenerateX509CertificateSigningRequest", func() {
		It("generates a valid certificate signing request", func() {
			key, cert, err := subject.GenerateX509CertificateSigningRequest("webhook.svc")
			Expect(err).NotTo(HaveOccurred())

			Expect(key.Validate()).To(Succeed())
			pemBlock, _ := pem.Decode(cert)
			Expect(pemBlock).NotTo(BeNil())
			x509CSR, err := x509.ParseCertificateRequest(pemBlock.Bytes)
			Expect(err).NotTo(HaveOccurred())
			Expect(x509CSR.PublicKey).To(Equal(&key.PublicKey))
			Expect(x509CSR.Subject.Organization).To(Equal([]string{admission.Organization}))
			Expect(x509CSR.Subject.CommonName).To(Equal("webhook.svc"))
		})
	})

	Describe("CreateCertificateSigningRequest", func() {
		var csrPEM []byte

		BeforeEach(func() {
			var err error
			_, csrPEM, err = subject.GenerateX509CertificateSigningRequest("webhook.svc")
			Expect(err).NotTo(HaveOccurred())
		})

		When("there is no previously-existing CertificateSigningRequest", func() {
			It("creates a CertificateSigningRequest", func() {
				csr, err := subject.CreateCertificateSigningRequest(csrPEM)
				Expect(err).NotTo(HaveOccurred())

				var csrResult certificates.CertificateSigningRequest
				Expect(reactiveClient.Get(nil, types.NamespacedName{Name: admission.CSRName}, &csrResult)).To(Succeed())
				Expect(csr).To(structmatcher.MatchStruct(csrResult).ExcludingFields("ObjectMeta.ResourceVersion", "TypeMeta"))
				Expect(metav1.IsControlledBy(&csrResult, fakeOwner)).To(BeTrue())
			})
		})

		When("our CSR already exists (from a previously-existing operator pod)", func() {
			var deleteCalled bool
			BeforeEach(func() {
				_, err := subject.CreateCertificateSigningRequest(csrPEM)
				Expect(err).NotTo(HaveOccurred())
				By("changing the certificate PEM")
				_, csrPEM, err = subject.GenerateX509CertificateSigningRequest("webhook.svc")
				Expect(err).NotTo(HaveOccurred())
				By("ensuring that a Delete request is made")
				reactiveClient.PrependReactor("delete", "certificatesigningrequests", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					deleteCalled = true
					return false, nil, nil
				})
			})
			It("deletes and re-creates the CertificateSigningRequest", func() {
				csr, err := subject.CreateCertificateSigningRequest(csrPEM)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleteCalled).To(BeTrue(), "a delete request should have been made")
				var csrResult certificates.CertificateSigningRequest
				Expect(reactiveClient.Get(nil, types.NamespacedName{Name: admission.CSRName}, &csrResult)).To(Succeed())
				Expect(csr).To(structmatcher.MatchStruct(csrResult).ExcludingFields("ObjectMeta.ResourceVersion", "TypeMeta"))
				Expect(metav1.IsControlledBy(&csrResult, fakeOwner)).To(BeTrue())
			})
		})

		When("deleting the CSR fails", func() {
			BeforeEach(func() {
				_, err := subject.CreateCertificateSigningRequest(csrPEM)
				Expect(err).NotTo(HaveOccurred())
				By("injecting an error when the Delete request is made")
				reactiveClient.PrependReactor("delete", "certificatesigningrequests", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("delete failed")
				})
			})
			It("deletes and re-creates the CertificateSigningRequest", func() {
				_, err := subject.CreateCertificateSigningRequest(csrPEM)
				Expect(err).To(MatchError("delete failed"))
			})
		})

		When("creating the CSR fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("create", "certificatesigningrequests", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("create failed")
				})
			})
			It("returns an error", func() {
				_, err := subject.CreateCertificateSigningRequest(csrPEM)
				Expect(err).To(MatchError("create failed"))
			})
		})

		When("it fails to add an owner reference to the CSR", func() {
			BeforeEach(func() {
				subject.Owner = nil
			})
			It("returns an error", func() {
				_, err := subject.CreateCertificateSigningRequest(csrPEM)
				// We do not control this error message so we don't want to test the exact string
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ApproveCertificateSigningRequest", func() {
		var csr *v1beta1.CertificateSigningRequest

		BeforeEach(func() {
			csr = &certificates.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: admission.CSRName,
				},
				Spec: certificates.CertificateSigningRequestSpec{
					Groups:  []string{"system:authenticated"},
					Request: []byte("cert"),
					Usages: []certificates.KeyUsage{
						certificates.UsageDigitalSignature,
						certificates.UsageKeyEncipherment,
						certificates.UsageServerAuth},
				},
			}
			_, err := simpleClientSet.CertificatesV1beta1().CertificateSigningRequests().Create(ctx, csr, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		When("all is good", func() {
			It("approves the certificate signing request", func() {
				csrWithCert, err := subject.ApproveCertificateSigningRequest(csr)
				Expect(err).NotTo(HaveOccurred())

				csrWithCertResult, err := subject.KubeClientSet.CertificatesV1beta1().CertificateSigningRequests().Get(ctx, admission.CSRName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())
				Expect(csrWithCert).To(structmatcher.MatchStruct(csrWithCertResult))

				approvalCondition := certificates.CertificateSigningRequestCondition{
					Type:    certificates.CertificateApproved,
					Reason:  "AutoApproved",
					Message: "certificate approved by Greenplum Operator",
				}
				Expect(csrWithCertResult.Status.Conditions).To(ContainElement(approvalCondition))
			})
		})

		When("all is not good", func() {
			BeforeEach(func() {
				simpleClientSet.PrependReactor("update", "certificatesigningrequests", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New("failed")
				})
			})

			It("returns an error", func() {
				_, err := subject.ApproveCertificateSigningRequest(csr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("failed"))
			})
		})
	})

	Describe("WaitForSignedCertificate", func() {
		var csr *v1beta1.CertificateSigningRequest

		BeforeEach(func() {
			csr = &certificates.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: admission.CSRName,
				},
				Spec: certificates.CertificateSigningRequestSpec{
					Groups:  []string{"system:authenticated"},
					Request: []byte("cert"),
					Usages:  []certificates.KeyUsage{"digital signature", "key encipherment", "server auth"},
				},
				Status: certificates.CertificateSigningRequestStatus{
					Conditions: []certificates.CertificateSigningRequestCondition{{
						Type:    certificates.CertificateApproved,
						Reason:  "AutoApproved",
						Message: "approved by test",
					}},
				},
			}
			_, err := simpleClientSet.CertificatesV1beta1().CertificateSigningRequests().Create(ctx, csr, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		})

		When("certificate is approved", func() {
			BeforeEach(func() {
				// this step is necessary because the simple client set does not fill in the certificate upon approval
				// the approval controller is responsible for filling in certs when csr's are marked approved
				csr.Status.Certificate = []byte("certificate")
				_, err := simpleClientSet.CertificatesV1beta1().CertificateSigningRequests().Update(ctx, csr, metav1.UpdateOptions{})
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns client CA bytes", func() {
				sCert, err := subject.WaitForSignedCertificate(csr, 1*time.Second)
				Expect(err).NotTo(HaveOccurred())
				Expect(sCert).To(Equal([]byte("certificate")))
			})
		})

		When("status.certificate is not populated", func() {
			It("returns an error", func() {
				_, err := subject.WaitForSignedCertificate(csr, 1*time.Millisecond)
				Expect(err).To(MatchError("timed out waiting for the condition"))
			})
		})
	})

	Describe("GetCertificate", func() {

		var (
			rsaPem    *rsa.PrivateKey
			certBytes []byte
			certPem   []byte
		)

		BeforeEach(func() {
			var err error
			rsaPem, err = rsa.GenerateKey(rand.Reader, admission.RSABits)
			Expect(err).NotTo(HaveOccurred())
			notBefore := time.Now()
			notAfter := time.Now().Add(24 * time.Hour)
			serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
			serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
			Expect(err).NotTo(HaveOccurred())
			template := x509.Certificate{
				SerialNumber: serialNumber,
				Subject: pkix.Name{
					Organization: []string{"Test Company"},
				},

				NotBefore: notBefore,
				NotAfter:  notAfter,

				KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
				BasicConstraintsValid: true,
				IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
				IsCA:                  true,
			}
			certBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, &rsaPem.PublicKey, rsaPem)
			Expect(err).NotTo(HaveOccurred())
			certPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
		})

		When("given proper rsaKey and certPem", func() {
			It("returns a signed certificate", func() {
				sCert, err := subject.GetCertificate(certPem, rsaPem)
				Expect(err).NotTo(HaveOccurred())
				Expect(sCert.PrivateKey).To(Equal(rsaPem))
				Expect(sCert.Certificate).To(Equal([][]uint8{certBytes}))
			})
		})

		When("given an invalid rsaKey", func() {
			It("returns an error", func() {
				rsaPem.D.Rem(big.NewInt(100), big.NewInt(10))
				_, err := subject.GetCertificate(certPem, rsaPem)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("tls: failed to parse private key"))
			})
		})

		When("given an invalid certPem", func() {
			It("returns an error", func() {
				_, err := subject.GetCertificate(nil, rsaPem)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("tls: failed to find any PEM data in certificate input"))
			})
		})
	})

	Describe("GenerateCertificateSigningRequest", func() {
		It("returns a valid CertificateSigningRequest", func() {
			certByte := []byte("some random certificate")
			csr := subject.GenerateCertificateSigningRequest(certByte)
			Expect(csr.Name).To(Equal(admission.CSRName))
			Expect(csr.Labels).To(Equal(map[string]string{"app": "greenplum-operator"}))
			Expect(csr.Spec.Groups).To(Equal([]string{"system:authenticated"}))
			Expect(csr.Spec.Request).To(Equal(certByte))
			Expect(csr.Spec.Usages).To(Equal([]v1beta1.KeyUsage{
				certificates.UsageDigitalSignature,
				certificates.UsageKeyEncipherment,
				certificates.UsageServerAuth,
			}))
		})
	})
})
