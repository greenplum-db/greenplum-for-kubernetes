package main

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

func TestRunGpexpand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RunGpexpand Suite")
}

func TestHelperProcess(t *testing.T) {
	commandable.Command.HelperProcess()
}
