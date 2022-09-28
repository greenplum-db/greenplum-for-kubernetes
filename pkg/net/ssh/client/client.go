package client

import (
	"github.com/pivotal/greenplum-for-kubernetes/pkg/net/ssh/session"
	"golang.org/x/crypto/ssh"
)

type SSHClientInterface interface {
	Close() error
	NewSession() (session.SSHSessionInterface, error)
}

type Client struct {
	clientImpl *ssh.Client
}

func NewClient(c *ssh.Client) *Client {
	return &Client{clientImpl: c}
}

func (c *Client) Close() error {
	return c.clientImpl.Close()
}

func (c *Client) NewSession() (session.SSHSessionInterface, error) {
	return c.clientImpl.NewSession()
}
