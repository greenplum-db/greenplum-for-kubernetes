package ssh

import (
	"fmt"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/multihost"
	cryptossh "golang.org/x/crypto/ssh"
	ctrl "sigs.k8s.io/controller-runtime"
)

type MultiHostExec struct {
	Command string
	Exec    ExecInterface
	Fs      vfs.Filesystem
}

var _ multihost.Operation = &MultiHostExec{}

func NewMultiHostExec(command string) *MultiHostExec {
	return &MultiHostExec{
		Command: command,
		Exec:    NewExec(),
		Fs:      vfs.OS(),
	}
}

func (m *MultiHostExec) Execute(host string) error {
	log := ctrl.Log.WithName("SSH exec").WithValues("host", host)
	keyBytes, err := vfs.ReadFile(m.Fs, "/home/gpadmin/.ssh/id_rsa")
	if err != nil {
		return err
	}
	clientPrivateKey, err := cryptossh.ParsePrivateKey(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	output, err := m.Exec.RunSSHCommand(host, m.Command, clientPrivateKey)
	if err != nil {
		log.Error(err, "SSH command failed", "command", m.Command, "output", string(output))
		return err
	}
	return nil
}
