package gpexpand

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

func TestRunGpexpand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RunGpexpand on Active Master Suite")
}

func TestHelperProcess(t *testing.T) {
	commandable.Command.HelperProcess()
}
