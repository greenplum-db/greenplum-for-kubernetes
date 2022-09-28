package gpexpandconfig

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

func TestGenerateGpexpandConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GpexpandConfig Suite")
}

func TestHelperProcess(t *testing.T) {
	commandable.Command.HelperProcess()
}
