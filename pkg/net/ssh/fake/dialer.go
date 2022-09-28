package fake

import (
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/client"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/dialer"
	"golang.org/x/crypto/ssh"
)

type FakeDialer struct {
	fakeErr       error
	hostPublicKey ssh.PublicKey
	fakeClient    *FakeSSHClient

	CallCount int
	Addr      string
}

var _ dialer.DialerInterface = &FakeDialer{}

func NewFakeDialer() *FakeDialer {
	return &FakeDialer{}
}

func (f *FakeDialer) WithError(err error) *FakeDialer {
	f.fakeErr = err
	return f
}

func (f *FakeDialer) WithHostPublicKey(publicKey string) *FakeDialer {
	var err error
	f.hostPublicKey, _, _, _, err = ssh.ParseAuthorizedKey([]byte(publicKey))
	Expect(err).NotTo(HaveOccurred())
	return f
}

func (f *FakeDialer) WithClient(fakeClient *FakeSSHClient) *FakeDialer {
	f.fakeClient = fakeClient
	return f
}

// DialerInterface implementation:

func (f *FakeDialer) Dial(network, hostAddr string, config *ssh.ClientConfig) (client.SSHClientInterface, error) {
	Expect(network).To(Equal("tcp"))
	Expect(f.fakeClient).NotTo(BeNil(), "client should be initialized in test setup")

	f.CallCount++
	f.Addr = hostAddr

	if f.fakeErr != nil {
		return nil, f.fakeErr
	}

	err := config.HostKeyCallback(hostAddr, nil, f.hostPublicKey)

	return f.fakeClient, err
}
