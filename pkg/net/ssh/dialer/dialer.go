package dialer

import (
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/client"
	"golang.org/x/crypto/ssh"
)

type DialerInterface interface {
	Dial(network, addr string, config *ssh.ClientConfig) (client.SSHClientInterface, error)
}

type Dialer struct{}

func (r *Dialer) Dial(network, addr string, config *ssh.ClientConfig) (client.SSHClientInterface, error) {
	c, err := ssh.Dial(network, addr, config)
	return client.NewClient(c), err
}
