package keyscanner

import (
	"bytes"
	"errors"
	"strings"
	"time"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/knownhosts"
	"golang.org/x/crypto/ssh"
)

func ScanHostKeys(keyScanner SSHKeyScannerInterface, knownHostsReader knownhosts.ReaderInterface, hostnameList []string) (string, error) {
	var hostKeys strings.Builder

	done := make(chan HostKey)
	for _, host := range hostnameList {
		go func(hostname string) {
			Log.Info("starting keyscan", "host", hostname)
			done <- keyScanner.SSHKeyScan(hostname, 5*time.Minute)
		}(host)
	}

	knownHosts, err := knownHostsReader.GetKnownHosts()
	if err != nil {
		return "", err
	}

	for range hostnameList {
		hostKey := <-done
		if hostKey.Err != nil {
			Log.Error(hostKey.Err, "keyscan failed", "host", hostKey.Hostname)
			err = hostKey.Err
		} else {
			Log.Info("keyscan successful", "host", hostKey.Hostname)
			if knownKey, ok := knownHosts[hostKey.Hostname]; !ok {
				hostKeys.WriteString(hostKey.Hostname)
				hostKeys.WriteRune(' ')
				// MarshalAuthorizedKey ends in a newline
				hostKeys.Write(ssh.MarshalAuthorizedKey(hostKey.PublicKey))
			} else if !bytes.Equal(hostKey.PublicKey.Marshal(), knownKey.Marshal()) {
				err = errors.New("scanned key does not match known key")
				Log.Error(err, "keyscan failed", "host", hostKey.Hostname)
			}
		}
	}

	if err != nil {
		return "", err
	}
	return hostKeys.String(), nil
}
