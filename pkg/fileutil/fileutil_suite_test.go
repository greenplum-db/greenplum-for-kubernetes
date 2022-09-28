package fileutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFileutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fileutil Suite")
}
