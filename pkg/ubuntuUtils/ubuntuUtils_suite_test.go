package ubuntuUtils

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCreateGpdbUser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Create GPDB User Test Suite")
}
