package dns

import (
	"context"
	"errors"
	"net"
	"time"

	"code.cloudfoundry.org/clock"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Resolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
}

type ConsistentDNSResolver struct {
	Resolver Resolver
	Clock    clock.Clock
}

var _ multihost.Operation = &ConsistentDNSResolver{}

func NewConsistentResolver() *ConsistentDNSResolver {
	return &ConsistentDNSResolver{
		Resolver: net.DefaultResolver,
		Clock:    clock.NewClock(),
	}
}

func (cr *ConsistentDNSResolver) Execute(addr string) error {
	log := ctrl.Log.WithName("DNS resolver").WithValues("host", addr)
	log.V(1).Info("attempting to resolve DNS entry")
	pollFunc := func() bool {
		_, err := cr.Resolver.LookupHost(context.Background(), addr)
		return err == nil
	}
	err := cr.PollUntilConsistent(pollFunc)
	if err != nil {
		log.Error(err, "failed to resolve DNS entry")
	} else {
		log.V(1).Info("resolved DNS entry")
	}
	return err
}

func (cr *ConsistentDNSResolver) PollUntilConsistent(poll func() bool) error {
	deadline := cr.Clock.Now().Add(5 * time.Minute)
	for cr.Clock.Now().Before(deadline) {
		if poll() {
			cr.Clock.Sleep(1 * time.Second)
			isSuccessful := true
			for i := 0; i < 15; i++ {
				if !poll() {
					isSuccessful = false
					break
				}
			}
			if isSuccessful {
				return nil
			}
		}
		cr.Clock.Sleep(1 * time.Second)
	}
	return errors.New("DNS lookup timed out")
}
