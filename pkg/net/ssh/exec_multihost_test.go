package ssh_test

import (
	"errors"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	cryptossh "golang.org/x/crypto/ssh"
)

var _ = Describe("MultiHostExec", func() {
	var (
		subject  ssh.MultiHostExec
		fakeExec *fakessh.FakeExec
		fakeFs   vfs.Filesystem
	)
	BeforeEach(func() {
		fakeExec = &fakessh.FakeExec{}
		fakeFs = memfs.Create()
		Expect(vfs.MkdirAll(fakeFs, "/home/gpadmin/.ssh", 700)).To(Succeed())
		Expect(vfs.WriteFile(fakeFs, "/home/gpadmin/.ssh/id_rsa",
			[]byte(fakessh.ExamplePrivateKey), 0600)).To(Succeed())
		subject = ssh.MultiHostExec{
			Command: "/tools/waitForKnownHosts --newPrimarySegmentCount 2",
			Exec:    fakeExec,
			Fs:      fakeFs,
		}
	})

	It("runs the command on the host", func() {
		Expect(subject.Execute("master-0")).To(Succeed())
		Expect(fakeExec.CalledHostnames).To(ConsistOf([]string{"master-0"}))
		Expect(fakeExec.CalledCommands).To(ConsistOf([]string{"/tools/waitForKnownHosts --newPrimarySegmentCount 2"}))
	})

	It("uses the private key in /home/gpadmin/.ssh/id_rsa", func() {
		Expect(subject.Execute("master-0")).To(Succeed())
		expectedKey, err := cryptossh.ParsePrivateKey([]byte(fakessh.ExamplePrivateKey))
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeExec.CalledPrivateKeys).To(ConsistOf([]cryptossh.Signer{expectedKey}))
	})

	When("reading ssh key file fails", func() {
		BeforeEach(func() {
			fakeFs = memfs.Create()
			subject.Fs = fakeFs
		})
		It("returns an error", func() {
			Expect(subject.Execute("master-0")).To(MatchError("open /home/gpadmin/.ssh/id_rsa: file does not exist"))
		})
	})
	When("parsing private key fails", func() {
		BeforeEach(func() {
			fakeFs = memfs.Create()
			Expect(vfs.MkdirAll(fakeFs, "/home/gpadmin/.ssh", 700)).To(Succeed())
			Expect(vfs.WriteFile(fakeFs, "/home/gpadmin/.ssh/id_rsa",
				[]byte("invalid key"), 0600)).To(Succeed())
			subject.Fs = fakeFs
		})
		It("returns an error", func() {
			Expect(subject.Execute("master-0")).To(MatchError("failed to parse private key: ssh: no key found"))
		})
	})
	When("RunSSHCommand fails", func() {
		BeforeEach(func() {
			fakeExec.WithError(errors.New("fake error")).WithOutput("fake output")
		})
		It("returns an error containing the output of the failed command", func() {
			Expect(subject.Execute("master-0")).To(MatchError("fake error"))
		})
	})
})
