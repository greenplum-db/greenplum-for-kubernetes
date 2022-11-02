package admission

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	WebhookConfigName = "greenplum-validating-webhook-config"
	ServiceName       = "greenplum-validating-webhook-service"
)

type ValidatingWebhook interface {
	Run(ctx context.Context) error
}

type Webhook struct {
	KubeClient      client.Client
	Namespace       string
	ServiceOwner    metav1.Object
	WebhookCfgOwner metav1.Object
	NameSuffix      string
	Server          Server
	Handler         http.Handler
	CertGenerator   CertGenerator
}

var _ ValidatingWebhook = &Webhook{}

func (w *Webhook) Run(ctx context.Context) error {
	Log.Info("starting greenplum validating admission webhook server")

	signedCertPEM, signedCertX509, err := w.GenerateAndSignTLSCertificate()
	if err != nil {
		return fmt.Errorf("getting certificate for webhook: %w", err)
	}

	err = w.ReconcileValidatingWebhookConfiguration(context.Background(), signedCertPEM)
	if err != nil {
		Log.Error(err, "Error creating ValidatingWebhookConfiguration")
		return fmt.Errorf("creating ValidatingWebhookConfiguration: %w", err)
	}

	err = w.Server.Start(ctx.Done(), *signedCertX509, ":https", w.Handler)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("validating admission webhook server start failed: %w", err)
	}
	Log.Info("shutting down greenplum validating admission webhook server")
	return nil
}

func (w *Webhook) GenerateAndSignTLSCertificate() ([]byte, *tls.Certificate, error) {
	svcCommonName := fmt.Sprintf("%s.%s.svc", ServiceName+w.NameSuffix, w.Namespace)
	rsaKey, csrPEM, err := w.CertGenerator.GenerateX509CertificateSigningRequest(svcCommonName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate certificate signing request")
	}

	csr, err := w.CertGenerator.CreateCertificateSigningRequest(csrPEM)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create certificate signing request")
	}

	approvedCSR, err := w.CertGenerator.ApproveCertificateSigningRequest(csr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to approve certificate signing request")
	}

	signedCertPEM, err := w.CertGenerator.WaitForSignedCertificate(approvedCSR, 30*time.Second)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failure while waiting for approval")
	}

	signedCertX509, err := w.CertGenerator.GetCertificate(signedCertPEM, rsaKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error loading keypair")
	}

	return signedCertPEM, &signedCertX509, nil
}

func (w *Webhook) ReconcileValidatingWebhookConfiguration(ctx context.Context, signedCert []byte) error {
	webhookConfig := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: WebhookConfigName,
		},
	}
	result, err := controllerutil.CreateOrUpdate(ctx, w.KubeClient, webhookConfig, func() error {
		w.ModifyWebhookConfiguration(webhookConfig, signedCert)
		if err := controllerutil.SetControllerReference(w.WebhookCfgOwner, webhookConfig, scheme.Scheme); err != nil {
			return errors.Wrap(err, "couldn't set OwnerReferences on ValidatingWebhookConfig")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to create ValidatingWebhookConfiguration")
	}
	if result != controllerutil.OperationResultNone {
		Log.Info("ValidatingWebhookConfiguration: " + string(result))
	}

	webhookService := w.CreateSVCForValidatingWebhookConfiguration()
	err = controllerutil.SetControllerReference(w.ServiceOwner, webhookService, scheme.Scheme)
	if err != nil {
		return errors.Wrap(err, "couldn't set OwnerReferences on webhook Service")
	}
	err = w.KubeClient.Create(ctx, webhookService)
	if err != nil {
		return errors.Wrap(err, "error creating Service for Webhook")
	}
	return nil
}

func (w *Webhook) ModifyWebhookConfiguration(webhookConfig *admissionregistrationv1.ValidatingWebhookConfiguration, signedCertBundle []byte) {
	fail := admissionregistrationv1.Fail
	sideEffectClassNone := admissionregistrationv1.SideEffectClassNone

	if webhookConfig.Labels == nil {
		webhookConfig.Labels = make(map[string]string)
	}
	webhookConfig.Labels["app"] = "greenplum-operator"
	webhookConfig.Webhooks = []admissionregistrationv1.ValidatingWebhook{
		{
			Name: "greenplum.pivotal.io",
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				Service: &admissionregistrationv1.ServiceReference{
					Namespace: w.Namespace,
					Name:      ServiceName + w.NameSuffix,
					Path:      heapvalue.NewString("/validate"),
				},
				CABundle: signedCertBundle,
			},
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"greenplum.pivotal.io"},
						APIVersions: []string{"v1"},
						Resources:   []string{"greenplumclusters"},
					},
				},
				{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"greenplum.pivotal.io"},
						APIVersions: []string{"v1beta1"},
						Resources:   []string{"greenplumpxfservices"},
					},
				},
			},
			FailurePolicy:           &fail,
			SideEffects:             &sideEffectClassNone,
			AdmissionReviewVersions: []string{"v1beta1"},
		},
	}
}

func (w *Webhook) CreateSVCForValidatingWebhookConfiguration() *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName + w.NameSuffix,
			Namespace: w.Namespace,
			Labels:    map[string]string{"app": "greenplum-operator"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "webhook",
					Port:       443,
					TargetPort: intstr.FromInt(443),
					NodePort:   0,
				},
			},
			Selector: map[string]string{"app": "greenplum-operator"},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}
