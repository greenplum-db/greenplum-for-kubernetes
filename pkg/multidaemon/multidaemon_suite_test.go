package multidaemon_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMultidaemon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multidaemon Suite")
}
