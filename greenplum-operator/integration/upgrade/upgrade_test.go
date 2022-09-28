package upgrade_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/admission"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("upgrade")

var _ = Describe("Upgrade scenarios "+OldOperatorVersion+" -> latest", func() {
	var (
		gpdbYamlFile string
		tempDir      string
		pxfYamlFile  string
		pxfTempDir   string
		suiteFailed  bool
	)
	BeforeSuite(func() {
		suiteFailed = true // Mark as failed in case we don't get to the end. This is how we can detect a failure in BeforeSuite().
		ctrl.SetLogger(gplog.ForIntegration())
		CleanUpK8s(*OperatorImageTag)
		Expect(kubewait.ForAPIService()).To(Succeed())
		// old operators' regsecret will not overwrite this secret
		EnsureRegSecretIsCreated()
		// setup old operator
		SetupOperator(OldOperatorOptions())
		VerifyPVCsAreDeleted()

		// setup old PXF Service
		pxfManifestYaml := GetPXFManifestYaml(2, PXFYamlOptions{})
		pxfTempDir, pxfYamlFile = CreateTempFile(pxfManifestYaml)
		out, err := ApplyManifest(pxfYamlFile, "pxf")
		Expect(err).NotTo(HaveOccurred(), out)

		// setup old Greenplum cluster
		manifestYaml := GetGreenplumManifestYaml(1, GpYamlOptions{
			PXFService: "my-greenplum-pxf",
		})
		tempDir, gpdbYamlFile = CreateTempFile(manifestYaml)
		SetupGreenplumCluster(gpdbYamlFile, OldOperatorVersion)
		Expect(kubewait.ForClusterReady(true)).To(Succeed())
		VerifyGreenplumForKubernetesVersion("master-0", OldOperatorVersion)
		AddHbaToAllowAccessToGreenplumThroughService("master-0")
		LoadData()

		Expect(err).NotTo(HaveOccurred(), out)

		// wait for the services to be available
		Expect(kubewait.ForReplicasReady("deployment", "my-greenplum-pxf")).To(Succeed())

		oldVWHSvcName := GetVWHSvcName()

		// upgrade greenplum-operator
		SetupOperator(CurrentOperatorOptions().WithUpgrade())
		Expect(kubewait.ForOperatorPod()).To(Succeed())
		Expect(kubewait.ForServiceTo(oldVWHSvcName, kubewait.NotExist)).To(Succeed())

		suiteFailed = false
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			suiteFailed = true
		}
		// print operator logs and cluster information to console
		log.Info("*****************************Start DEBUG LOGS********************************")
		out, _ := exec.Command("kubectl", "logs", "-l", "app=greenplum-operator", "--tail", "50").CombinedOutput()
		fmt.Println(string(out))
		log.Info("*****************************Greenplum Cluster*******************************")
		out, _ = exec.Command("kubectl", "describe", "greenplumcluster", "my-greenplum").CombinedOutput()
		fmt.Println(string(out))
		out, _ = exec.Command("kubectl", "get", "all").CombinedOutput()
		fmt.Println(string(out))
		log.Info("*******************************End DEBUG LOGS********************************")
	})

	AfterSuite(func() {
		if !suiteFailed {
			// After a successful run of this test suite, all the components will already be deleted.
			// Therefore, we do not delete any of the components here
			CleanUpK8s(*OperatorImageTag)
			VerifyPVCsAreDeleted()
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}
	})

	When(fmt.Sprintf("the operator is upgraded from %s to latest, leaving an existing GreenplumCluster instance running", OldOperatorVersion), func() {
		var gpdbYamlFile string

		It("should not affect the existing cluster", func() {
			Expect(kubewait.ForClusterReady(true)).To(Succeed())
			Expect(kubewait.ForGreenplumInitializationWithService()).To(Succeed())
			VerifyDataThroughService()
		})

		It("prevents updates to the existing GreenplumCluster", func() {
			manifestYaml := GetGreenplumManifestYaml(2, GpYamlOptions{})
			_, gpdbYamlFile = CreateTempFile(manifestYaml)
			result, err := ApplyManifest(gpdbYamlFile, "greenplumcluster")
			Expect(err).To(HaveOccurred())
			Expect(result).To(ContainSubstring(admission.UpgradeClusterHelpMsg))
		})
	})

	When(fmt.Sprintf("the operator is upgraded from %s to latest, leaving an existing GreenplumPXFService running", OldOperatorVersion), func() {
		It("should not affect the existing PXF", func() {
			CreatePXFExternalTableNoS3()
			pxfQueryResult, err := QueryWithRetry("master-0", "SELECT count(*) FROM pxf_read_test;")
			if err != nil {
				fmt.Println(string(pxfQueryResult))
				Fail("query PXF table failed")
			}
			Expect(string(pxfQueryResult)).To(Equal("6\n"))
		})
		It("prevents updates to the existing GreenplumPXFService", func() {
			pxfManifestYaml := GetPXFManifestYaml(3, PXFYamlOptions{})
			pxfTempDir, pxfYamlFile = CreateTempFile(pxfManifestYaml)
			pxfUpdateResult, err := ApplyManifest(pxfYamlFile, "pxf")
			Expect(err).To(HaveOccurred())
			Expect(pxfUpdateResult).To(ContainSubstring(admission.UpgradePXFHelpMsg))
		})
	})

	Context("GreenplumCluster is upgraded successfully", func() {
		BeforeEach(func() {
			log.Info("deleting greenplum cluster", "operatorVersion", OldOperatorVersion)
			DeleteGreenplumCluster()
			log.Info("deleting greenplum cluster successful", "operatorVersion", OldOperatorVersion)

			log.Info("creating greenplum cluster", "imageTag", *GreenplumImageTag)
			out, err := ApplyManifest(gpdbYamlFile, "greenplumcluster")
			Expect(err).NotTo(HaveOccurred(), out)
			log.Info("creating greenplum cluster successful", "imageTag", *GreenplumImageTag)

			Expect(kubewait.ForClusterReady(true)).To(Succeed())
			CheckCleanClusterStartup()
			Expect(kubewait.ForGreenplumService("greenplum")).To(Succeed())
			AddHbaToAllowAccessToGreenplumThroughService("master-0")
			Expect(kubewait.ForGreenplumInitializationWithService()).To(Succeed())
		})

		It("should always contain the latest image, verify previous data, sets Status.InstanceImage / Status.OperatorVersion to latest and sets Status.Phase to Running", func() {
			VerifyGreenplumForKubernetesVersion("master-0", *GreenplumImageTag)
			VerifyDataThroughService()
			VerifyStatusInstanceImage(*GreenplumImageTag)
			VerifyStatusOperatorVersion(*OperatorImageTag)
			Expect(kubewait.ForGreenplumClusterStatus(greenplumv1.GreenplumClusterPhaseRunning)).To(Succeed())
		})
	})

	Context("GreenplumPXFService is upgraded successfully", func() {
		BeforeEach(func() {
			CleanupComponentService(pxfTempDir, pxfYamlFile, "my-greenplum-pxf")
			pxfManifestYaml := GetPXFManifestYaml(2, PXFYamlOptions{})
			pxfTempDir, pxfYamlFile = CreateTempFile(pxfManifestYaml)
			pxfCreateResult, err := ApplyManifest(pxfYamlFile, "pxf")
			Expect(err).NotTo(HaveOccurred(), pxfCreateResult)
		})
		It("allows updates to the GreenplumPXFService and continues to work", func() {
			pxfManifestYaml := GetPXFManifestYaml(3, PXFYamlOptions{})
			pxfTempDir, pxfYamlFile = CreateTempFile(pxfManifestYaml)
			pxfUpdateResult, err := ApplyManifest(pxfYamlFile, "pxf")

			Expect(err).NotTo(HaveOccurred(), pxfUpdateResult)
			Expect(kubewait.ForReplicasReady("deployment", "my-greenplum-pxf")).To(Succeed())

			CreatePXFExternalTableNoS3()

			pxfQueryResult, err := QueryWithRetry("master-0", "SELECT count(*) FROM pxf_read_test;")
			if err != nil {
				fmt.Println(string(pxfQueryResult))
				Fail("query PXF table failed")
			}
			Expect(string(pxfQueryResult)).To(Equal("6\n"))
		})
	})

	When("The greenplum cluster is deleted and helm uninstall greenplum-operator is run", func() {
		// note: we are expecting the operator and cluster to be left behind from the test before
		BeforeEach(func() {
			CleanupComponentService(pxfTempDir, pxfYamlFile, "my-greenplum-pxf")
			CleanupGreenplumCluster(gpdbYamlFile)
			log.Info("BEGIN cleanup Greenplum operator")
			DeleteAllCharts()
			Expect(kubewait.ForPodDestroyed("greenplum-operator")).To(Succeed())
			log.Info("END cleanup Greenplum operator")
		})
		It("should delete all resources including the CRD", func() {
			cmd := exec.Command("bash", "-c", "kubectl get crd -o name | grep greenplum")
			output, _ := cmd.CombinedOutput()
			Expect(strings.TrimSpace(string(output))).To(BeEmpty(), "output: %s", string(output))
		})
	})
})
