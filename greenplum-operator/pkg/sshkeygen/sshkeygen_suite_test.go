package sshkeygen_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSshkeygen(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sshkeygen Suite")
}
