package testing

import (
	"sync"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
)

// Technically a fake (has a working implementation) and a spy (records how it was called).
type FakeOperation struct {
	FakeErrors  map[string]error
	HostRecords []string
	mtx         sync.Mutex
}

var _ multihost.Operation = &FakeOperation{}

func (r *FakeOperation) Execute(addr string) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.HostRecords = append(r.HostRecords, addr)
	if fakeError, ok := r.FakeErrors[addr]; ok && fakeError != nil {
		return fakeError
	}
	return nil
}
