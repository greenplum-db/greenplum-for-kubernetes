package keyscanner

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/blang/vfs"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/dialer"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/poll"
	"golang.org/x/crypto/ssh"
	apiwait "k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var Log = ctrl.Log.WithName("keyscanner")

type SSHKeyScannerInterface interface {
	SSHKeyScan(host string, timeout time.Duration) HostKey
}

type SSHKeyScanner struct {
	Dialer   dialer.DialerInterface
	PollWait poll.PollFunc
	Fs       vfs.Filesystem
}

type HostKey struct {
	Hostname  string
	Err       error
	PublicKey ssh.PublicKey
}

func NewSSHKeyScanner() *SSHKeyScanner {
	return &SSHKeyScanner{
		Dialer:   &dialer.Dialer{},
		PollWait: apiwait.PollImmediate,
		Fs:       vfs.OS(),
	}
}

func (s *SSHKeyScanner) SSHKeyScan(host string, timeout time.Duration) HostKey {
	var publicKey ssh.PublicKey
	var timeoutErr error

	addr := host + ":22"
	err := s.PollWait(1*time.Second, timeout, func() (done bool, err error) {
		allKeysWriterCh, lastKeyReaderCh := NewCompressingChan()
		c, dialError := s.Dialer.Dial("tcp", addr, getSSHConfig(s.Fs, allKeysWriterCh))
		if dialError != nil {
			Log.V(1).Info("ssh dial", "host", host, "error", dialError)
			return false, nil
		}
		defer c.Close()
		close(allKeysWriterCh)
		publicKey = <-lastKeyReaderCh
		return true, nil
	})
	if err != nil {
		timeoutErr = errors.New("timed out waiting for keyscan on " + addr)
	}

	return HostKey{
		Hostname:  host,
		Err:       timeoutErr,
		PublicKey: publicKey,
	}
}

// NewCompressingChan reads multiple values from allCh until it is closed, and discards all but the last. It then writes
// that final value to lastCh.
func NewCompressingChan() (chan<- ssh.PublicKey, <-chan ssh.PublicKey) {
	allCh := make(chan ssh.PublicKey)
	lastCh := make(chan ssh.PublicKey)
	go func() {
		var lastVal ssh.PublicKey
		for lastVal = range allCh {
			// nop
		}
		lastCh <- lastVal
	}()
	return allCh, lastCh
}

func getSSHConfig(filesystem vfs.Filesystem, keyCh chan<- ssh.PublicKey) *ssh.ClientConfig {
	hostPrivateKey, _ := GetGpadminPrivateKey(filesystem)
	return &ssh.ClientConfig{
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			keyCh <- key
			return nil
		},
		User: "gpadmin",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(hostPrivateKey),
		},
	}
}

func GetGpadminPrivateKey(fs vfs.Filesystem) (ssh.Signer, error) {
	privateBytes, err := vfs.ReadFile(fs, "/home/gpadmin/.ssh/id_rsa")
	if err != nil {
		return nil, fmt.Errorf("failed to read host rsa private publicKey file: %w", err)
	}

	privateKey, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to get signer from private publicKey bytes: %w", err)
	}

	return privateKey, nil
}
