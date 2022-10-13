package admission

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"time"

	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/scheme"
	"k8s.io/api/certificates/v1beta1"
	certificates "k8s.io/api/certificates/v1beta1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8scsr "k8s.io/client-go/util/certificate/csr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	RSABits      = 2048
	Organization = "Pivotal"
	CSRName      = "greenplum-validating-webhook-csr"
)

type CertGenerator interface {
	GenerateX509CertificateSigningRequest(commonName string) (*rsa.PrivateKey, []byte, error)
	CreateCertificateSigningRequest(cert []byte) (*v1beta1.CertificateSigningRequest, error)
	ApproveCertificateSigningRequest(csr *v1beta1.CertificateSigningRequest) (*v1beta1.CertificateSigningRequest, error)
	WaitForSignedCertificate(csr *v1beta1.CertificateSigningRequest, timeout time.Duration) ([]byte, error)
	GetCertificate(cert []byte, key *rsa.PrivateKey) (tls.Certificate, error)
}

type CertificateGenerator struct {
	CtrlClient    client.Client
	KubeClientSet kubernetes.Interface
	Owner         metav1.Object
}

func (g *CertificateGenerator) GenerateX509CertificateSigningRequest(commonName string) (*rsa.PrivateKey, []byte, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, RSABits)
	if err != nil {
		return nil, nil, err
	}
	requestTemplate := x509.CertificateRequest{
		SignatureAlgorithm: x509.SHA256WithRSA,
		Subject: pkix.Name{
			Organization: []string{Organization},
			CommonName:   commonName,
		},
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &requestTemplate, rsaKey)
	if err != nil {
		return nil, nil, err
	}
	pemEncodedCSR := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csr})
	return rsaKey, pemEncodedCSR, err
}

func (g *CertificateGenerator) CreateCertificateSigningRequest(csrPEM []byte) (*v1beta1.CertificateSigningRequest, error) {
	csr := g.GenerateCertificateSigningRequest(csrPEM)
	if err := controllerutil.SetControllerReference(g.Owner, &csr, scheme.Scheme); err != nil {
		return nil, err
	}

	if err := g.CtrlClient.Delete(context.Background(), &csr); err != nil && !apierrs.IsNotFound(err) {
		return nil, err
	}

	if err := g.CtrlClient.Create(context.Background(), &csr); err != nil {
		return nil, err
	}

	Log.Info("CertificateSigningRequest: created")

	return &csr, nil
}

func (g *CertificateGenerator) ApproveCertificateSigningRequest(csr *v1beta1.CertificateSigningRequest) (*v1beta1.CertificateSigningRequest, error) {
	approvalCondition := certificates.CertificateSigningRequestCondition{
		Type:    certificates.CertificateApproved,
		Reason:  "AutoApproved",
		Message: "certificate approved by Greenplum Operator",
	}
	csr.Status.Conditions = append(csr.Status.Conditions, approvalCondition)
	return g.KubeClientSet.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(context.Background(), csr, metav1.UpdateOptions{})
}

func (g *CertificateGenerator) WaitForSignedCertificate(csr *v1beta1.CertificateSigningRequest, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return k8scsr.WaitForCertificate(ctx, g.KubeClientSet.CertificatesV1beta1().CertificateSigningRequests(), csr)
}

func (g *CertificateGenerator) GetCertificate(cert []byte, rsaKey *rsa.PrivateKey) (tls.Certificate, error) {
	rsaPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)})
	return tls.X509KeyPair(cert, rsaPem)
}

// helper functions:

func (g *CertificateGenerator) GenerateCertificateSigningRequest(cert []byte) certificates.CertificateSigningRequest {
	return certificates.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: CSRName,
			Labels: map[string]string{
				"app": "greenplum-operator",
			},
		},
		Spec: certificates.CertificateSigningRequestSpec{
			Groups: []string{"system:authenticated"},
			Usages: []certificates.KeyUsage{
				certificates.UsageDigitalSignature,
				certificates.UsageKeyEncipherment,
				certificates.UsageServerAuth,
			},
			Request: cert,
		},
	}
}
