package ssh_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	netssh "github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/fake"
	cryptossh "golang.org/x/crypto/ssh"
	apiwait "k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("Execute", func() {
	var (
		subject              *netssh.KnownHostsWaiter
		fakeKnownHostsReader *fake.KnownHostsReader
	)
	BeforeEach(func() {
		fakeKnownHostsReader = &fake.KnownHostsReader{}
		subject = &netssh.KnownHostsWaiter{
			PollWait:         apiwait.PollImmediate,
			KnownHostsReader: fakeKnownHostsReader,
		}
	})
	When("a host is present in knownHosts", func() {
		BeforeEach(func() {
			fakeKnownHostsReader.KnownHosts = map[string]cryptossh.PublicKey{
				"master-0":    fake.KeyForHost("master-0"),
				"segment-a-0": fake.KeyForHost("segment-a-0"),
			}
		})
		It("succeeds", func() {
			Expect(subject.Execute("master-0")).To(Succeed())
		})
	})

	When("a host is not present in knownHosts", func() {
		var (
			poller = struct {
				called       bool
				finished     bool
				conditionErr error
			}{}
		)
		BeforeEach(func() {
			pollFunc := func(interval, timeout time.Duration, condition apiwait.ConditionFunc) error {
				poller.called = true
				poller.finished, poller.conditionErr = condition()
				// short circuiting the timeout
				return errors.New("timed out")
			}
			fakeKnownHostsReader.KnownHosts = map[string]cryptossh.PublicKey{
				"master-0": fake.KeyForHost("master-0"),
			}
			subject.PollWait = pollFunc
		})
		It("times out", func() {
			Expect(subject.Execute("invalid-host")).To(MatchError("timed out"))
			Expect(poller.called).To(BeTrue(), "should call Poll")
			Expect(poller.finished).To(BeFalse(), "should retry")
			Expect(poller.conditionErr).NotTo(HaveOccurred())
		})
	})

	When("an error occurs reading known_hosts", func() {
		BeforeEach(func() {
			fakeKnownHostsReader.Err = errors.New("catastrophic failure")
		})
		It("returns an error", func() {
			Expect(subject.Execute("some-host")).To(MatchError("catastrophic failure"))
		})
	})
})
