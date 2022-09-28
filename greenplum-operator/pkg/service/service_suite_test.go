package service_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	AppName       = "greenplum"
	NamespaceName = "test"
	ClusterName   = "my-greenplum"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Greenplum Service Suite")
}
