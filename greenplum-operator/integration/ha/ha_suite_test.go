package ha_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration HA Suite")
}

var (
	gpdbYamlFile string
	tempDir      string
	suiteFailed  bool
)

var _ = BeforeSuite(func() {
	suiteFailed = true // Mark as failed in case we don't get to the end. This is how we can detect a failure in BeforeSuite().
	ctrl.SetLogger(gplog.ForIntegration())
	Expect(kubewait.ForAPIService()).To(Succeed())
	CleanUpK8s(*OperatorImageTag)
	VerifyPVCsAreDeleted()

	manifestYaml := GetGreenplumManifestYaml(2, GpYamlOptions{Standby: "yes", Mirrors: "yes"})
	tempDir, gpdbYamlFile = CreateTempFile(manifestYaml)

	suiteFailed = false
})

var _ = AfterSuite(func() {
	if !suiteFailed {
		CleanUpK8s(*OperatorImageTag)
		VerifyPVCsAreDeleted()
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	}
})

var _ = BeforeEach(func() {
	EnsureRegSecretIsCreated()
	operatorOptions := CurrentOperatorOptions()
	EnsureOperatorIsDeployed(&operatorOptions)
	if !IsSingleNode() {
		EnsureNodesAreLabeled()
	}
	EnsureGPDBIsDeployed(gpdbYamlFile, *GreenplumImageTag)
	Expect(kubewait.ForClusterReady(true)).To(Succeed())
	EnsureDataIsLoaded()
})

var _ = AfterEach(func() {
	if CurrentGinkgoTestDescription().Failed {
		suiteFailed = true
	}
})
