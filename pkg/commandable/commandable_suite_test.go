package commandable_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommandable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commandable Suite")
}
