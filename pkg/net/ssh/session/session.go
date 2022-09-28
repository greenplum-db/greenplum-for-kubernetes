package session

type SSHSessionInterface interface {
	CombinedOutput(cmd string) ([]byte, error)
}
