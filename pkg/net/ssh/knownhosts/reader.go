package knownhosts

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/blang/vfs"
	"golang.org/x/crypto/ssh"
)

type ReaderInterface interface {
	GetKnownHosts() (map[string]ssh.PublicKey, error)
}

type Reader struct {
	Fs vfs.Filesystem
}

type HostNotFound struct {
	Host string
}

func (h *HostNotFound) Error() string {
	return fmt.Sprintf("host %s was not found in the known_hosts file", h.Host)
}

func NewReader() *Reader {
	return &Reader{Fs: vfs.OS()}
}

func (r *Reader) GetKnownHosts() (map[string]ssh.PublicKey, error) {
	knownHostMap := make(map[string]ssh.PublicKey)
	knownHostsFilename := "/home/gpadmin/.ssh/known_hosts"

	hostsBytes, err := vfs.ReadFile(r.Fs, knownHostsFilename)
	if err != nil {
		if os.IsNotExist(err) { // treat missing file as empty
			return knownHostMap, nil
		}
		return nil, fmt.Errorf("could not read %s: %w", knownHostsFilename, err)
	}
	for {
		marker, hosts, pubkey, _, bytesRemaining, parseError := ssh.ParseKnownHosts(hostsBytes)
		if parseError != nil {
			if parseError == io.EOF {
				break
			}
			return nil, fmt.Errorf("could not parse %s: %w", knownHostsFilename, parseError)
		}
		if marker != "" {
			return nil, errors.New("known_hosts markers are not currently supported")
		}

		for _, host := range hosts {
			knownHostMap[host] = pubkey
		}

		hostsBytes = bytesRemaining
	}
	return knownHostMap, nil
}

func GetHostPublicKey(reader ReaderInterface, hostname string) (ssh.PublicKey, error) {
	knownHostMap, err := reader.GetKnownHosts()
	if err != nil {
		return nil, err
	}
	key, ok := knownHostMap[hostname]
	if !ok {
		return nil, &HostNotFound{Host: hostname}
	}
	return key, nil
}
