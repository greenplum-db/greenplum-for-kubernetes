package fake

import "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/session"

type FakeSSHSession struct {
	fakeError  error
	fakeOutput []byte
}

var _ session.SSHSessionInterface = &FakeSSHSession{}

func NewFakeSSHSession() *FakeSSHSession {
	return &FakeSSHSession{}
}

func (s *FakeSSHSession) WithError(fakeError error) *FakeSSHSession {
	s.fakeError = fakeError
	return s
}

func (s *FakeSSHSession) WithOutput(fakeOutput []byte) *FakeSSHSession {
	s.fakeOutput = fakeOutput
	return s
}

// SSHSessionInterface implementation:

func (s *FakeSSHSession) CombinedOutput(cmd string) ([]byte, error) {
	return s.fakeOutput, s.fakeError
}
