package startContainerUtils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/commandable"
)

func TestStartContainerUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StartContainerUtils Suite")
}

func TestHelperProcess(t *testing.T) {
	commandable.Command.HelperProcess()
}
