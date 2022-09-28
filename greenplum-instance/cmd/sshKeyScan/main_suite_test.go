package main

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

func TestKeyScan(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KeyScan Test Suite")
}

func TestHelperProcess(t *testing.T) {
	commandable.Command.HelperProcess()
}
