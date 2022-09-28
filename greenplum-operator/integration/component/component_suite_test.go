package component_test

import (
	"os"
	"testing"

	"github.com/minio/minio-go/v6"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestComponent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Component Suite")
}

var (
	gpdbYamlFile string
	tempDir      string
	suiteFailed  bool

	minioClient *minio.Client
	pxfTempDir  string
	pxfYamlFile string
)

var _ = BeforeSuite(func() {
	suiteFailed = true // Mark as failed in case we don't get to the end. This is how we can detect a failure in BeforeSuite().
	ctrl.SetLogger(gplog.ForIntegration())
	Expect(kubewait.ForAPIService()).To(Succeed())
	CleanUpK8s(*OperatorImageTag)
	VerifyPVCsAreDeleted()

	// Setup GPDB without components (to test that components can be added later)
	manifestYaml := GetGreenplumManifestYaml(1, GpYamlOptions{})
	tempDir, gpdbYamlFile = CreateTempFile(manifestYaml)
	EnsureRegSecretIsCreated()
	operatorOptions := CurrentOperatorOptions()
	EnsureOperatorIsDeployed(&operatorOptions)
	if !IsSingleNode() {
		EnsureNodesAreLabeled()
	}
	EnsureGPDBIsDeployed(gpdbYamlFile, *GreenplumImageTag)
	Expect(kubewait.ForClusterReady(true)).To(Succeed())

	// Delete GPDB (TODO: remove this after we can add capability to add components to a running gpdb #173734658)
	DeleteGreenplumCluster()

	// Setup PXF
	accessKeyID := os.Getenv("ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")
	Expect(accessKeyID).NotTo(BeEmpty())
	Expect(secretAccessKey).NotTo(BeEmpty())
	var err error
	minioClient, err = minio.New("storage.googleapis.com", accessKeyID, secretAccessKey, true)
	Expect(err).NotTo(HaveOccurred())
	EnsurePXFTestFilesArePresent(minioClient, accessKeyID, secretAccessKey)
	pxfManifestYaml := GetPXFManifestYaml(2, PXFYamlOptions{UseConfBucket: true, AccessKeyID: accessKeyID, SecretAccessKey: secretAccessKey})
	pxfTempDir, pxfYamlFile = CreateTempFile(pxfManifestYaml)
	out, err := ApplyManifest(pxfYamlFile, "pxf")
	Expect(err).NotTo(HaveOccurred(), out)

	// Setup GPDB
	manifestYaml = GetGreenplumManifestYaml(1, GpYamlOptions{
		PXFService:    "my-greenplum-pxf",
	})
	tempDir, gpdbYamlFile = CreateTempFile(manifestYaml)
	EnsureOperatorIsDeployed(&operatorOptions)
	if !IsSingleNode() {
		EnsureNodesAreLabeled()
	}
	EnsureGPDBIsDeployed(gpdbYamlFile, *GreenplumImageTag)

	suiteFailed = false
})

var _ = AfterSuite(func() {
	if !suiteFailed {
		CleanupComponentService(pxfTempDir, pxfYamlFile, "my-greenplum-pxf")
		CleanUpK8s(*OperatorImageTag)
		VerifyPVCsAreDeleted()
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	}
})

var _ = BeforeEach(func() {
	Expect(kubewait.ForClusterReady(true)).To(Succeed())
	EnsureDataIsLoaded()
})

var _ = AfterEach(func() {
	if CurrentGinkgoTestDescription().Failed {
		suiteFailed = true
	}
})
