package fake

import (
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/client"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/session"
)

type FakeSSHClient struct {
	fakeSSHSession *FakeSSHSession
	fakeError      error
	WasClosed      bool
}

var _ client.SSHClientInterface = &FakeSSHClient{}

func NewFakeSSHClient() *FakeSSHClient {
	return &FakeSSHClient{}
}

func (c *FakeSSHClient) WithSession(fakeSSHSession *FakeSSHSession) *FakeSSHClient {
	c.fakeSSHSession = fakeSSHSession
	return c
}

func (c *FakeSSHClient) WithError(err error) *FakeSSHClient {
	c.fakeError = err
	return c
}

// SSHClientInterface implementation:

func (c *FakeSSHClient) NewSession() (session.SSHSessionInterface, error) {
	Expect(c.fakeSSHSession).NotTo(BeNil(), "session should be initialized in test setup")
	return c.fakeSSHSession, c.fakeError
}

func (c *FakeSSHClient) Close() error {
	c.WasClosed = true
	return nil
}
