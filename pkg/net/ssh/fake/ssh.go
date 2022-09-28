package fake

import (
	"sync"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh"
	cryptossh "golang.org/x/crypto/ssh"
)

const (
	ExamplePrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDCR6dpOemsu0bToIJpHktl+4KpqUOqNM8fvtFmP3SYJPKHs9Xe
JtyiZKYYIwfR5mpabdsutgRUDnLnTsRsF3yv8tBEVUyfu8zOqqKbiDqxGfu9BRSB
VSC/A4kvdD0TgIhug7HbjXgSwTPYPf54gnCeaAm+FDDVAMdKt4Z/ocbmrQIDAQAB
AoGAConef+u/TDpgbixfxpn5FxAcl11yKTJyJdOxAi3hAjvG2CueJ03OXBS/mcGU
tAMes8cPw6nl9DVQcFGqf/6KKdz65bzrJ5to6Qpx2jbD22PjVA+KX27tyvsD1TuH
Cr7T4JNVvDP7950R/7cSPtZcpsC2BU9TNER7f05GmbTO04ECQQDlFSsdpzRnYV9d
4b55bOicJbDVWsQev+Vk1X5Y/ti1/6id8QzwWWojiuR5ohjEfF3uHEuewR9K50k5
27pGGUBNAkEA2RudJ71KBRkX7LKzrcyxGzNzLAWRn7bo+N+Z0iHD+2KT+nhTNOg/
QrzXv+YjWlPNcr4iW3mtPFZDeDNzF4Rv4QJAE5I2Z8ckK/zep+ekXT1XthdmPyQN
A0+DqpSuwa2sGAhqgGvaniIVdknkcRvPH+I8KB6Uu1BmewC9ecry5BA+NQJAZyR+
SeXcp4VfX10ajaQkM7cCrVRL9aOxFKMt8a2G7QPNJ35IkWcQvsT2fr136C7N+Qgp
TGoHChY1YYKX2AFcIQJAWa+6i+ya+dNg/HyzEjiPOWuCpb90A9f8hWKmHTwMMlDw
JF/Patfrdki7X+1w3yiHbY/Z+Fnl8AjaQkFgrGKdJA==
-----END RSA PRIVATE KEY-----`
	ExamplePublicKey            = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC0spg+4tb/tfSSoroLo43QzqeqAXVsOJMAaF/nNsxZAzSjNtIEryOPjm+kUJSvN6fKBgIoBDBFjIH7KkhdaJC64M3r6mmA8lzk/K7eq1POvy94elJXCZ4Q0vA/B6wgK96yF2D6L6anS9nuqxnThOu8tFta6uwvxdiZpjRJbUPcKipBqfRpjIcp86p+nLDJADSXqs9ZQT68AGf+kIk5z2xynCZ3a+rrTviY3K3qMi25PW7mB7PLqP0RtyueWkbVuKRiNBkXuWv8j2eSh9uhLOGiKIWheGxDBPG8UuSVFuXzZ5Lm9exW09diD8kJG638PgqpLWcU1Y87y1IEssUSdl++/3Iqh2h6mjZ8gAaS2+SZfCbjgVj8k4peo8XhfBdFWmBwMs9qc0O/oxBiSzxhFn7nqe5nRrf3xMjgFzmLeNBt5mTWEEb1bKJB2u3lBYyuWatcdgIoUCVSh/okGMY0WnGKQs6La5pGQgZO61jWECDk3KYqr6ncXNdveUjxSqJdDqmbTgkLu7KohjJVLwAr51vOdYc2Ly/+pBDB/ropqTIfF/TXqvcIyUEpoZYyEEGQv9c1JcXn0ynw1DXrlJHLrrsIN4P0j8ihfXAUEgls3Pcnxa1g92D74CiuMozTeEYtk09mc9RhVEHjRzliIZ1xlJQAWdCPG+GKu8SlhzzJZTbHmQ==`
	ExamplePublicKeyForMismatch = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC0spg+4tb/tfSSoroLo43QzqeqAXVsOJMAaF/nNsxZAzSjNtIEryOPjm+kUJSvN6fKBgIoBDBFjIH7KkhdaJC64A3r6mmA8lzk/K7eq1POvy94elJXCZ4Q0vA/B6wgK96yF2D6L6anS9nuqxnThOu8tFta6uwvxdiZpjRJbUPcKipBqfRpjIcp86p+nLDJADSXqs9ZQT68AGf+kIk5z2xynCZ3a+rrTviY3K3qMi25PW7mB7PLqP0RtyueWkbVuKRiNBkXuWv8j2eSh9uhLOGiKIWheGxDBPG8UuSVFuXzZ5Lm9exW09diD8kJG638PgqpLWcU1Y87y1IEssUSdl++/3Iqh2h6mjZ8gAaS2+SZfCbjgVj8k4peo8XhfBdFWmBwMs9qc0O/oxBiSzxhFn7nqe5nRrf3xMjgFzmLeNBt5mTWEEb1bKJB2u3lBYyuWatcdgIoUCVSh/okGMY0WnGKQs6La5pGQgZO61jWECDk3KYqr6ncXNdveUjxSqJdDqmbTgkLu7KohjJVLwAr51vOdYc2Ly/+pBDB/ropqTIfF/TXqvcIyUEpoZYyEEGQv9c1JcXn0ynw1DXrlJHLrrsIN4P0j8ihfXAUEgls3Pcnxa1g92D74CiuMozTeEYtk09mc9RhVEHjRzliIZ1xlJQAWdCPG+GKu8SlhzzJZTbHmQ==`
)

func GenerateSSHFake() (*FakeDialer, *FakeSSHClient, *FakeSSHSession) {
	fakeSession := NewFakeSSHSession()
	fakeClient := NewFakeSSHClient().WithSession(fakeSession)
	fakeDialer := NewFakeDialer().WithClient(fakeClient)
	return fakeDialer, fakeClient, fakeSession
}

var _ ssh.ExecInterface = &FakeExec{}

type FakeExec struct {
	CalledHostnames   []string
	CalledCommands    []string
	CalledPrivateKeys []cryptossh.Signer
	FakeError         error
	FakeOutput        []byte
	mtx               sync.Mutex
}

func NewExec() *FakeExec {
	return &FakeExec{}
}

func (f *FakeExec) WithError(err error) *FakeExec {
	f.FakeError = err
	return f
}

func (f *FakeExec) WithOutput(output string) *FakeExec {
	f.FakeOutput = []byte(output)
	return f
}

func (f *FakeExec) RunSSHCommand(hostname string, command string, privateKey cryptossh.Signer) ([]byte, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	f.CalledHostnames = append(f.CalledHostnames, hostname)
	f.CalledCommands = append(f.CalledCommands, command)
	f.CalledPrivateKeys = append(f.CalledPrivateKeys, privateKey)
	return f.FakeOutput, f.FakeError
}
