package sshkeygen_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"io"
	"math/big"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/sshkeygen"
	"golang.org/x/crypto/ssh"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("SSHSecret.ModifySecret()", func() {
	var secret *corev1.Secret
	BeforeEach(func() {
		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      "ssh-secrets",
				Namespace: "testns",
			},
		}
	})
	When("secret data is passed in", func() {
		secretData := map[string][]byte{
			"id_rsa":     []byte("some-key-value"),
			"id_rsa.pub": []byte("some-pub-key-value"),
		}
		It("fills in a K8s Secret", func() {
			sshkeygen.ModifySecret("my-greenplum", secret, secretData)
			Expect(secret.Type).To(Equal(corev1.SecretTypeOpaque))
			Expect(secret.ObjectMeta.Name).To(Equal("ssh-secrets"))
			Expect(secret.ObjectMeta.Namespace).To(Equal("testns"))
			Expect(secret.Data).To(HaveKey("id_rsa"))
			Expect(secret.Data["id_rsa"]).NotTo(BeNil())
			Expect(secret.Data).To(HaveKey("id_rsa.pub"))
			Expect(secret.Data["id_rsa.pub"]).NotTo(BeNil())
			Expect(secret.ObjectMeta.Labels["app"]).To(Equal("greenplum"))
			Expect(secret.ObjectMeta.Labels["greenplum-cluster"]).To(Equal("my-greenplum"))
		})
	})
	When("secret data is nil", func() {
		It("does not set the key", func() {
			sshkeygen.ModifySecret("my-greenplum", secret, nil)
			Expect(secret.Type).To(Equal(corev1.SecretTypeOpaque))
			Expect(secret.ObjectMeta.Name).To(Equal("ssh-secrets"))
			Expect(secret.ObjectMeta.Namespace).To(Equal("testns"))
			Expect(secret.Data).To(BeNil())
			Expect(secret.ObjectMeta.Labels["app"]).To(Equal("greenplum"))
			Expect(secret.ObjectMeta.Labels["greenplum-cluster"]).To(Equal("my-greenplum"))
		})
	})
})

var _ = Describe("SSHSecret.GenerateKey()", func() {
	var (
		sshSecret  sshkeygen.SSHSecret
		fakeKeyGen *fakeSSHKeyGen

		result map[string][]byte
		err    error
	)

	BeforeEach(func() {
		fakeKeyGen = &fakeSSHKeyGen{}
		fakeKeyGen.NewSSHPublickKeyMock.base64PublicKey = examplePublicKey
		fakeKeyGen.GeneratePrivateKeyMock.result = &rsa.PrivateKey{
			PublicKey: rsa.PublicKey{N: big.NewInt(31), E: 2000},
			D:         big.NewInt(1000),
		}

		sshSecret = sshkeygen.SSHSecret{
			KeyGen: fakeKeyGen,
		}
	})

	JustBeforeEach(func() {
		result, err = sshSecret.GenerateKey()
	})

	It("calls GeneratePrivatekey() with correct args and gets a private keyArgReceived", func() {
		Expect(fakeKeyGen.GeneratePrivateKeyMock.bits).To(Equal(4096))
		Expect(fakeKeyGen.GeneratePrivateKeyMock.reader).To(Equal(rand.Reader))
	})
	It("passes generated RSAKey to PkcMarshaller", func() {
		Expect(fakeKeyGen.MarshalPKCS1PrivateKeyStub.key).To(Equal(fakeKeyGen.GeneratePrivateKeyMock.result))
	})
	It("passes the generated public key to NewSSHPublicKey", func() {
		Expect(fakeKeyGen.NewSSHPublickKeyMock.publicKey).To(Equal(&fakeKeyGen.GeneratePrivateKeyMock.result.PublicKey))
	})

	It("produces a PEM encoded private key", func() {
		Expect(string(result["id_rsa"])).To(Equal(
			"-----BEGIN RSA PRIVATE KEY-----\n" + // Hey cred-alert-cli, this is "fake"!
				base64.StdEncoding.EncodeToString([]byte("hello world")) + "\n" +
				"-----END RSA PRIVATE KEY-----\n"))
	})

	It("produces a public key", func() {
		Expect(string(result["id_rsa.pub"])).To(Equal("ssh-rsa " + examplePublicKey + "\n"))
	})

	When("GeneratePrivateKey() has an error", func() {
		BeforeEach(func() {
			fakeKeyGen.GeneratePrivateKeyMock.err = errors.New("an error from GeneratePrivateKey")
		})
		It("fails", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("an error from GeneratePrivateKey"))
		})
	})

	When("NewSSHPublicKey() has an error", func() {
		BeforeEach(func() {
			fakeKeyGen.NewSSHPublickKeyMock.base64PublicKey = invalidPublicKey
		})
		It("fails", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("ssh: short read"))
		})
	})
})

type fakeSSHKeyGen struct {
	NewSSHPublickKeyMock struct {
		base64PublicKey string
		publicKey       *rsa.PublicKey
	}
	MarshalPKCS1PrivateKeyStub struct {
		key *rsa.PrivateKey
	}
	GeneratePrivateKeyMock struct {
		wasCalled bool
		bits      int
		reader    io.Reader
		result    *rsa.PrivateKey
		err       error
	}
}

const (
	// Generated with: ssh-keygen -t rsa -b 1024 -N "" -f /tmp/key; cat /tmp/key.pub
	examplePublicKey = "AAAAB3NzaC1yc2EAAAADAQABAAAAgQCtisiapt6lKmB/0cVg1i+tIJdtUtGucCziVvPWCJg0mb3h29thhoLlVRjtRpLCza0TBx4y9iBPl72EUB1c20WNm7cRSq8hs88VDoZCovbXY7seIPCvwuMb5v7W4eLk8Vi7wO53hzyIVZ3svZmogHCaif1GScSFd8BrHq34c5mZSQ=="
	invalidPublicKey = "this is not even base64!"
)

func (gen *fakeSSHKeyGen) NewSSHPublicKey(key interface{}) (ssh.PublicKey, error) {
	Expect(key).To(BeAssignableToTypeOf(gen.NewSSHPublickKeyMock.publicKey))
	gen.NewSSHPublickKeyMock.publicKey = key.(*rsa.PublicKey)

	base64PubKey := gen.NewSSHPublickKeyMock.base64PublicKey
	pubkey, err := base64.StdEncoding.DecodeString(base64PubKey)
	if base64PubKey != invalidPublicKey {
		Expect(err).NotTo(HaveOccurred(), "sanity check")
	}
	// TODO: Try to use ssh.ParseAuthorizedKey here instead. The "authorized key" format is "ssh-rsa <base64-key> <comment>", as in id_rsa.pub and authorized_keys.
	return ssh.ParsePublicKey(pubkey)
}

func (gen *fakeSSHKeyGen) MarshalPKCS1PrivateKey(key *rsa.PrivateKey) []byte {
	gen.MarshalPKCS1PrivateKeyStub.key = key
	return []byte("hello world")
}

func (gen *fakeSSHKeyGen) GeneratePrivateKey(random io.Reader, bits int) (*rsa.PrivateKey, error) {
	gen.GeneratePrivateKeyMock.wasCalled = true
	gen.GeneratePrivateKeyMock.bits = bits
	gen.GeneratePrivateKeyMock.reader = random
	return gen.GeneratePrivateKeyMock.result, gen.GeneratePrivateKeyMock.err
}
