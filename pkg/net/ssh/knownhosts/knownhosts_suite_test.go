package knownhosts_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKnownhosts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Knownhosts Suite")
}
