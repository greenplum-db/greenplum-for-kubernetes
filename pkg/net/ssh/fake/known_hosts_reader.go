package fake

import (
	"sync"

	"golang.org/x/crypto/ssh"
)

type KnownHostsReader struct {
	WasCalled  bool
	KnownHosts map[string]ssh.PublicKey
	Err        error
	mtx        sync.Mutex
}

func (r *KnownHostsReader) GetKnownHosts() (map[string]ssh.PublicKey, error) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.WasCalled = true
	return r.KnownHosts, r.Err
}
