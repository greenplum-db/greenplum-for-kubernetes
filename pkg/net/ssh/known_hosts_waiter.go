package ssh

import (
	"time"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/poll"
	apiwait "k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

type KnownHostsWaiter struct {
	PollWait         poll.PollFunc
	KnownHostsReader knownhosts.ReaderInterface
}

var _ multihost.Operation = &KnownHostsWaiter{}

func NewKnownHostsWaiter() *KnownHostsWaiter {
	return &KnownHostsWaiter{
		PollWait:         apiwait.PollImmediate,
		KnownHostsReader: knownhosts.NewReader(),
	}
}

func (r *KnownHostsWaiter) Execute(host string) error {
	log := ctrl.Log.WithName("known_hosts waiter").WithValues("host", host)
	log.V(1).Info("waiting for known_hosts entry")
	err := r.PollWait(1*time.Second, 30*time.Second, func() (bool, error) {
		_, err := knownhosts.GetHostPublicKey(r.KnownHostsReader, host)
		if err != nil {
			if _, ok := err.(*knownhosts.HostNotFound); ok {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
	if err != nil {
		log.Error(err, "failed waiting for known_hosts entry")
	} else {
		log.V(1).Info("found known_hosts entry")
	}
	return err
}
