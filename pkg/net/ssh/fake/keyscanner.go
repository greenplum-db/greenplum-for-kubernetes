package fake

import (
	"sync"
	"time"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/keyscanner"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type KeyScanner struct {
	WasCalled bool
	Hostnames []string
	Err       error
	mtx       sync.Mutex
}

func (k *KeyScanner) SSHKeyScan(hostname string, _ time.Duration) keyscanner.HostKey {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	k.WasCalled = true
	k.Hostnames = append(k.Hostnames, hostname)
	return keyscanner.HostKey{
		Hostname:  hostname,
		Err:       k.Err,
		PublicKey: KeyForHost(hostname),
	}
}

func KeyForHost(host string) ssh.PublicKey {
	return &agent.Key{Blob: []byte(host), Format: "FakeKey"}
}
