package ssh_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	realssh "golang.org/x/crypto/ssh"
)

var _ = Describe("Exec", func() {
	var (
		subject              *ssh.Exec
		fakeKnownHostsReader *fakessh.KnownHostsReader
		fakeDialer           *fakessh.FakeDialer
		fakeSSHClient        *fakessh.FakeSSHClient
		fakeSSHSession       *fakessh.FakeSSHSession
		fakeClientPrivateKey realssh.Signer

		sshResult []byte
		sshErr    error
	)
	BeforeEach(func() {
		var err error
		fakeClientPrivateKey, err = realssh.ParsePrivateKey([]byte(fakessh.ExamplePrivateKey))
		Expect(err).NotTo(HaveOccurred())
		fakeHostPublicKey, _, _, _, err := realssh.ParseAuthorizedKey([]byte(fakessh.ExamplePublicKey))
		Expect(err).NotTo(HaveOccurred())

		fakeKnownHostsReader = &fakessh.KnownHostsReader{}
		fakeKnownHostsReader.KnownHosts = map[string]realssh.PublicKey{
			"test-host-1": fakeHostPublicKey,
		}

		fakeDialer, fakeSSHClient, fakeSSHSession = fakessh.GenerateSSHFake()
		fakeDialer = fakeDialer.WithHostPublicKey(fakessh.ExamplePublicKey)
		fakeSSHSession = fakeSSHSession.WithOutput([]byte("fake command output"))

		subject = &ssh.Exec{
			KnownHostsReader: fakeKnownHostsReader,
			Dialer:           fakeDialer,
		}
	})
	JustBeforeEach(func() {
		sshResult, sshErr = subject.RunSSHCommand("test-host-1", "cat show-me-the-output.txt", fakeClientPrivateKey)
	})
	It("returns ssh session combined output", func() {
		Expect(sshErr).NotTo(HaveOccurred())
		Expect(string(sshResult)).To(Equal("fake command output"))
	})
	It("closes the client connection", func() {
		Expect(fakeSSHClient.WasClosed).To(BeTrue())
	})

	When("GetHostPublicKey fails", func() {
		BeforeEach(func() {
			fakeKnownHostsReader.Err = errors.New("could not get host key")
		})
		It("returns an error", func() {
			Expect(fakeKnownHostsReader.WasCalled).To(BeTrue())
			Expect(sshErr).To(MatchError("failed to get host key for test-host-1: could not get host key"))
		})
	})

	When("Dial fails", func() {
		BeforeEach(func() {
			fakeDialer.WithError(errors.New("could not dial host"))
		})
		It("returns an error", func() {
			Expect(sshErr).To(MatchError("could not dial host"))
		})
	})

	When("hostkey callback fails", func() {
		BeforeEach(func() {
			fakeDialer.WithHostPublicKey(fakessh.ExamplePublicKeyForMismatch)
		})
		It("dialer returns an error", func() {
			Expect(sshErr).To(MatchError("ssh: host key mismatch"))
		})
	})

	When("NewSession fails", func() {
		BeforeEach(func() {
			fakeSSHClient.WithError(errors.New("session failure"))
		})
		It("returns an error", func() {
			Expect(sshErr).To(MatchError("could not create ssh session: session failure"))
		})
		It("closes the connection", func() {
			Expect(fakeSSHClient.WasClosed).To(BeTrue())
		})
	})

	When("CombinedOutput fails", func() {
		BeforeEach(func() {
			fakeSSHSession.WithOutput([]byte{}).WithError(errors.New("execution error"))
		})
		It("returns an error", func() {
			Expect(sshErr).To(MatchError("execution error"))
		})
		It("closes the connection", func() {
			Expect(fakeSSHClient.WasClosed).To(BeTrue())
		})
	})

})
