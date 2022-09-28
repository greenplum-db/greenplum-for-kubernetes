package instanceconfig_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInstanceconfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instanceconfig Suite")
}
