package hostpod_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHostpod(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hostpod Suite")
}
