package gpexpandjob

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGpexpandjob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gpexpandjob Suite")
}
