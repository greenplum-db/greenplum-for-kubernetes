package multihost_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestResolve(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multihost Suite")
}
