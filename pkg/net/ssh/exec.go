package ssh

import (
	"fmt"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/dialer"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	"golang.org/x/crypto/ssh"
)

type ExecInterface interface {
	RunSSHCommand(hostname string, command string, privateKey ssh.Signer) ([]byte, error)
}

type Exec struct {
	KnownHostsReader knownhosts.ReaderInterface
	Dialer           dialer.DialerInterface
}

func NewExec() *Exec {
	return &Exec{
		KnownHostsReader: knownhosts.NewReader(),
		Dialer:           &dialer.Dialer{},
	}
}

func (r *Exec) RunSSHCommand(hostname string, command string, clientPrivateKey ssh.Signer) ([]byte, error) {
	hostPublicKey, err := knownhosts.GetHostPublicKey(r.KnownHostsReader, hostname)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get host key for %s: %w", hostname, err)
	}

	clientconfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.FixedHostKey(hostPublicKey),
		User:            "gpadmin",
		// auth methods are not test in our mocks
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(clientPrivateKey),
		},
	}

	c, err := r.Dialer.Dial("tcp", hostname+":22", clientconfig)
	if err != nil {
		return []byte{}, err
	}
	defer c.Close()

	session, err := c.NewSession()
	if err != nil {
		return []byte{}, fmt.Errorf("could not create ssh session: %w", err)
	}

	return session.CombinedOutput(command)
}
