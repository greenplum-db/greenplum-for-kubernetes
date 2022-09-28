package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

func TestStartPXF(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Start PXF Suite")
}

func TestHelperProcess(t *testing.T) {
	commandable.Command.HelperProcess()
}
