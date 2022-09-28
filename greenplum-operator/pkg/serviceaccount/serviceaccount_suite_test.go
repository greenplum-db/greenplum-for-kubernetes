package serviceaccount_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	AppName       = "greenplum"
	NamespaceName = "test-ns"
)

func TestServiceaccount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Serviceaccount Suite")
}
