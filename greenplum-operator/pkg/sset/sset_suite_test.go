package sset_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StatefulSet Suite")
}
