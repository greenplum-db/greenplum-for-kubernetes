package keyscanner_test

import (
	"errors"
	"time"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	fakessh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"golang.org/x/crypto/ssh"
	apiwait "k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("SSHKeyScanner", func() {
	var (
		memoryfs vfs.Filesystem
	)

	BeforeEach(func() {
		memoryfs = memfs.Create()
		Expect(vfs.MkdirAll(memoryfs, "/home/gpadmin/.ssh/", 755)).To(Succeed())
		Expect(vfs.WriteFile(memoryfs, "/home/gpadmin/.ssh/id_rsa", []byte(fakessh.ExamplePrivateKey), 644)).To(Succeed())
	})

	When("SSHKeyScan is called and there are no errors", func() {
		var (
			app        *keyscanner.SSHKeyScanner
			fakeDialer *fakessh.FakeDialer
			fakeClient *fakessh.FakeSSHClient
		)
		BeforeEach(func() {
			fakeDialer, fakeClient, _ = fakessh.GenerateSSHFake()
			fakeDialer = fakeDialer.WithHostPublicKey(fakessh.ExamplePublicKey)
			poller := &mockAPIWait{}
			app = &keyscanner.SSHKeyScanner{Dialer: fakeDialer, PollWait: poller.Poll, Fs: memoryfs}
		})
		It("returns appropriate public key", func() {
			expectedHostKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(fakessh.ExamplePublicKey))
			Expect(err).NotTo(HaveOccurred())
			outputHostKey := app.SSHKeyScan("127.0.0.1", 5*time.Minute)
			Expect(outputHostKey.Err).To(BeNil())
			Expect(outputHostKey.Hostname).To(Equal("127.0.0.1"))
			Expect(outputHostKey.PublicKey).To(Equal(expectedHostKey))
		})
		It("appends :22 to host name", func() {
			app.SSHKeyScan("127.0.0.1", 5*time.Minute)
			Expect(fakeDialer.Addr).To(Equal("127.0.0.1:22"))
		})
		It("closes the ssh connection", func() {
			app.SSHKeyScan("127.0.0.1", 5*time.Minute)
			Expect(fakeClient.WasClosed).To(BeTrue())
		})
	})

	When("SSHKeyScan is called with an invalid host", func() {
		var (
			fakeDialer *fakessh.FakeDialer
			app        *keyscanner.SSHKeyScanner
			poller     *mockAPIWait
		)
		BeforeEach(func() {
			fakeDialer, _, _ = fakessh.GenerateSSHFake()
			fakeDialer = fakeDialer.WithError(errors.New("could not connect to host"))
			poller = &mockAPIWait{}
			app = &keyscanner.SSHKeyScanner{Dialer: fakeDialer, PollWait: poller.Poll, Fs: memoryfs}
		})
		It("tells the poller to retry", func() {
			app.SSHKeyScan("127.0.0.1", 5*time.Minute)
			Expect(poller.called).To(Equal(1))
			Expect(poller.conditionResult.isFinished).To(BeFalse(), "should retry")
			Expect(poller.conditionResult.err).To(BeNil())
		})
	})

	When("SSHKeyScan times out", func() {
		var (
			fakeDialer *fakessh.FakeDialer
			fakeClient *fakessh.FakeSSHClient
			app        *keyscanner.SSHKeyScanner
		)
		BeforeEach(func() {
			fakeDialer, fakeClient, _ = fakessh.GenerateSSHFake()
			fakeDialer = fakeDialer.WithHostPublicKey(fakessh.ExamplePublicKey)
			poller := &mockAPIWait{returns: errors.New("timed out Poll")}
			app = &keyscanner.SSHKeyScanner{Dialer: fakeDialer, PollWait: poller.Poll, Fs: memoryfs}
		})
		It("returns an error", func() {
			outputHostKey := app.SSHKeyScan("127.0.0.1", 5*time.Minute)
			Expect(outputHostKey.Hostname).To(Equal("127.0.0.1"))
			Expect(outputHostKey.Err).To(MatchError("timed out waiting for keyscan on 127.0.0.1:22"))
		})
		It("closes the ssh connection", func() {
			app.SSHKeyScan("127.0.0.1", 5*time.Minute)
			Expect(fakeClient.WasClosed).To(BeTrue())
		})
	})

	When("GetGpadminPrivateKey() is called", func() {
		It("returns a valid private key", func() {
			signer, err := keyscanner.GetGpadminPrivateKey(memoryfs)
			Expect(err).NotTo(HaveOccurred())
			Expect(signer).NotTo(BeNil())
			Expect(signer.PublicKey().Marshal()).NotTo(BeNil())
			keyBytes, err := vfs.ReadFile(memoryfs, "/home/gpadmin/.ssh/id_rsa")
			Expect(err).NotTo(HaveOccurred())
			keySigner, err := ssh.ParsePrivateKey(keyBytes)
			Expect(err).NotTo(HaveOccurred())
			Expect(signer.PublicKey().Marshal()).To(Equal(keySigner.PublicKey().Marshal()))
		})
		It("returns an error  on non exiting id_rsa file", func() {
			memoryfs = memfs.Create()
			signer, err := keyscanner.GetGpadminPrivateKey(memoryfs)
			Expect(err).To(HaveOccurred())
			Expect(signer).To(BeNil())
			Expect(err.Error()).To(Equal("failed to read host rsa private publicKey file: open /home/gpadmin/.ssh/id_rsa: file does not exist"))
		})
		It("returns an error on bad private key", func() {
			Expect(vfs.WriteFile(memoryfs, "/home/gpadmin/.ssh/id_rsa", []byte("bad private key"), 644)).To(Succeed())
			signer, err := keyscanner.GetGpadminPrivateKey(memoryfs)
			Expect(err).To(HaveOccurred())
			Expect(signer).To(BeNil())
			Expect(err.Error()).To(Equal("failed to get signer from private publicKey bytes: ssh: no key found"))
		})
	})
})

type mockAPIWait struct {
	called          int
	conditionResult struct {
		isFinished bool
		err        error
	}
	returns error
}

func (f *mockAPIWait) Poll(_, _ time.Duration, condition apiwait.ConditionFunc) error {
	f.called++
	f.conditionResult.isFinished, f.conditionResult.err = condition()
	return f.returns
}
