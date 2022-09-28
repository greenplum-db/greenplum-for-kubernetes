package keyscanner_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSshkeyscanner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sshkeyscanner Suite")
}
