package dns_test

import (
	"context"
	"errors"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/dns"
)

var _ = Describe("Execute", func() {
	var (
		subject      *dns.ConsistentDNSResolver
		startTime    time.Time
		fakeClock    *fakeclock.FakeClock
		fakeResolver *FakeResolver
		successCh    chan error
	)
	BeforeEach(func() {
		startTime = time.Date(2019, 11, 16, 0, 0, 0, 0, time.UTC)
		fakeClock = fakeclock.NewFakeClock(startTime)
		fakeResolver = &FakeResolver{}
		subject = &dns.ConsistentDNSResolver{
			Resolver: fakeResolver,
			Clock:    fakeClock,
		}
		successCh = make(chan error)
	})
	When("a valid address is provided", func() {
		BeforeEach(func() {
			fakeResolver.addresses = []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"}
			fakeResolver.err = nil
		})
		It("should succeed", func() {
			go func() {
				successCh <- subject.Execute("whatever.com")
			}()
			fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
			Eventually(successCh).Should(Receive(BeNil()))
			Expect(fakeResolver.called > 0).To(BeTrue())
		})
	})
	When("an invalid address is provided", func() {
		BeforeEach(func() {
			fakeResolver.addresses = []string{}
			fakeResolver.err = errors.New("cannot resolve")
		})
		It("should timeout", func() {
			go func() {
				successCh <- subject.Execute("whatever.com")
			}()
			for i := 0; i < 300; i++ {
				fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
			}
			Eventually(successCh).Should(Receive(MatchError("DNS lookup timed out")))
			Expect(fakeResolver.called > 0).To(BeTrue())
		})
	})
})

var _ = Describe("PollUntilConsistent", func() {
	var (
		startTime time.Time
		fakeClock *fakeclock.FakeClock
		subject   *dns.ConsistentDNSResolver
		errCh     chan error
	)
	BeforeEach(func() {
		startTime = time.Date(2019, 11, 16, 0, 0, 0, 0, time.UTC)
		fakeClock = fakeclock.NewFakeClock(startTime)
		subject = &dns.ConsistentDNSResolver{
			Clock: fakeClock,
		}
		errCh = make(chan error)
	})
	It("succeeds if poll succeeds", func() {
		var numPolls int
		poll := func() bool {
			numPolls++
			return true
		}
		go func() {
			errCh <- subject.PollUntilConsistent(poll)
		}()
		fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
		Eventually(errCh).Should(Receive(nil))
		// 16 successes (1 second wait)
		Expect(numPolls).To(Equal(16))
	})
	It("times out if 'poll' fails", func() {
		var numPolls int
		poll := func() bool {
			numPolls++
			return false
		}
		go func() {
			errCh <- subject.PollUntilConsistent(poll)
		}()
		for i := 0; i < 300; i++ {
			fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
		}
		Eventually(errCh).Should(Receive(MatchError("DNS lookup timed out")))
		// 300 failures (300 second wait)
		Expect(numPolls).To(Equal(300))
	})
	It("polls immediately", func() {
		var numPolls int
		var immediatePoll bool
		poll := func() bool {
			if fakeClock.Now() == startTime {
				immediatePoll = true
			}
			numPolls++
			return true
		}
		go func() {
			errCh <- subject.PollUntilConsistent(poll)
		}()
		fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
		Eventually(errCh).Should(Receive(nil))
		Expect(immediatePoll).To(BeTrue())
		// 16 successes (1 second wait)
		Expect(numPolls).To(Equal(16))
	})
	It("succeeds if 'poll' succeeds after 5 seconds", func() {
		var numPolls int
		poll := func() bool {
			numPolls++
			if numPolls > 5 {
				return true
			}
			return false
		}
		go func() {
			errCh <- subject.PollUntilConsistent(poll)
		}()
		for i := 0; i < 6; i++ {
			fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
		}
		Eventually(errCh).Should(Receive(nil))
		// 5 failures (5 second wait), 16 successes (1 second wait)
		Expect(numPolls).To(Equal(21))
	})
	It("times out if poll succeeds for a bit, but then fails", func() {
		var numPolls int
		poll := func() bool {
			numPolls++
			if numPolls < 16 {
				return true
			}
			return false
		}
		go func() {
			errCh <- subject.PollUntilConsistent(poll)
		}()
		for i := 0; i < 300; i++ {
			fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
		}
		Eventually(errCh).Should(Receive(MatchError("DNS lookup timed out")))
		// 15 successes (1 second wait), 299 failures (299 second wait)
		Expect(numPolls).To(Equal(314))
	})
	It("succeeds after a brief failure between successes", func() {
		var numPolls int
		poll := func() bool {
			numPolls++
			if numPolls < 16 || numPolls >= 32 {
				return true
			}
			return false
		}
		go func() {
			errCh <- subject.PollUntilConsistent(poll)
		}()
		for i := 0; i < 18; i++ {
			fakeClock.WaitForWatcherAndIncrement(1 * time.Second)
		}
		Eventually(errCh).Should(Receive(nil))
		// 15 successes (1 second wait), 16 failures (16 second wait), 16 successes (1 second wait)
		Expect(numPolls).To(Equal(47))
	})
})

type FakeResolver struct {
	addresses []string
	err       error
	called    int
}

func (f *FakeResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	f.called++
	return f.addresses, f.err
}
