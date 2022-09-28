package sshkeygen

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
)

const RSAKeySize = 4096

type SSHKeyGeneration interface {
	NewSSHPublicKey(key interface{}) (ssh.PublicKey, error)
	MarshalPKCS1PrivateKey(key *rsa.PrivateKey) []byte
	GeneratePrivateKey(random io.Reader, bits int) (*rsa.PrivateKey, error)
}

type SSHSecretCreator interface {
	GenerateKey() (map[string][]byte, error)
}

type SSHSecret struct {
	KeyGen SSHKeyGeneration
}

func New() *SSHSecret {
	return &SSHSecret{
		KeyGen: sshKeyGenerator{},
	}
}

func ModifySecret(clusterName string, secret *corev1.Secret, keyData map[string][]byte) {
	if keyData != nil {
		secret.Data = keyData
	}

	labels := map[string]string{
		"app":               greenplumv1.AppName,
		"greenplum-cluster": clusterName,
	}
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	for key, value := range labels {
		secret.Labels[key] = value
	}
	secret.Type = corev1.SecretTypeOpaque
}

func (s SSHSecret) GenerateKey() (map[string][]byte, error) {
	privateKey, err := s.KeyGen.GeneratePrivateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return nil, err
	}

	privateKeyBytes := s.KeyGen.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privateKeyBytes,
	}
	encodedPrivateKey := pem.EncodeToMemory(&privBlock)

	publicRsaKey, err := s.KeyGen.NewSSHPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}
	encodedPublicKey := ssh.MarshalAuthorizedKey(publicRsaKey)

	return map[string][]byte{
		"id_rsa.pub": encodedPublicKey,
		"id_rsa":     encodedPrivateKey,
	}, nil
}

type sshKeyGenerator struct{}

func (r sshKeyGenerator) GeneratePrivateKey(random io.Reader, bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(random, bits)
}

func (r sshKeyGenerator) MarshalPKCS1PrivateKey(key *rsa.PrivateKey) []byte {
	return x509.MarshalPKCS1PrivateKey(key)
}

func (r sshKeyGenerator) NewSSHPublicKey(key interface{}) (ssh.PublicKey, error) {
	return ssh.NewPublicKey(key)
}
