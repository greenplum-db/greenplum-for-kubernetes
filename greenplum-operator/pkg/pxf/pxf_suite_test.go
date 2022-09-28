package pxf_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPxf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pxf Suite")
}
