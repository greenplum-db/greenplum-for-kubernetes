package ha_test

import (
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("ha")

var _ = Describe("Happy path integration greenplum operator", func() {
	It("has Greenplum master, segment-a and segment-b stateful sets and PVCs", func() {
		cmd := exec.Command("kubectl", "get", "greenplumclusters")
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(out)).To(ContainSubstring("my-greenplum"))

		cmd = exec.Command(
			"kubectl", "get", "pvc",
			"-l", "greenplum-cluster=my-greenplum",
			"-o", `jsonpath={range .items[*]}{.metadata.name}{"\tgreenplum-major-version="}{.metadata.labels.greenplum-major-version}{"\n"}{end}`)
		out, err = cmd.Output()
		Expect(err).NotTo(HaveOccurred())

		Expect(string(out)).To(ContainSubstring("my-greenplum-pgdata-master-0\tgreenplum-major-version=6"))
		Expect(string(out)).To(ContainSubstring("my-greenplum-pgdata-master-1\tgreenplum-major-version=6"))
		Expect(string(out)).To(ContainSubstring("my-greenplum-pgdata-segment-a-0\tgreenplum-major-version=6"))
		Expect(string(out)).To(ContainSubstring("my-greenplum-pgdata-segment-b-0\tgreenplum-major-version=6"))

		cmd = exec.Command("kubectl", "describe", "greenplumCluster/my-greenplum")
		out, err = cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())
	})

	It("properly assigns pods to nodes based on anti-affinity rules", func() {
		if IsSingleNode() {
			return
		}

		podsOnNodes := GetPodsOnNodes()
		for nodeName, podsOnNode := range podsOnNodes {
			var segmentA, segmentB, master0, master1 int
			for _, pod := range podsOnNode {
				if strings.Contains(pod, "master-0") {
					master0++
				} else if strings.Contains(pod, "master-1") {
					master1++
				} else if strings.Contains(pod, "segment-a") {
					segmentA++
				} else if strings.Contains(pod, "segment-b") {
					segmentB++
				}
			}

			log.Info("pods on node", "node", nodeName, "master-0", master0, "master-1", master1, "segment-a", segmentA, "segment-b", segmentB)

			// print additional info for debugging purposes
			if ((segmentA > 0) && (segmentB > 0)) || ((master0 > 0) && (master1 > 0)) {
				cmd := exec.Command("kubectl", "get", "nodes", "--show-labels")
				out, err := cmd.CombinedOutput()
				fmt.Println(string(out))
				Expect(err).NotTo(HaveOccurred())

				cmd = exec.Command("kubectl", "get", "po", "-o", "wide", "--show-labels")
				out, err = cmd.CombinedOutput()
				fmt.Println(string(out))
				Expect(err).NotTo(HaveOccurred())
			}

			// can't have segment-a and segment-b on the same node
			Expect((segmentA > 0) && (segmentB > 0)).To(BeFalse())
			// can't have primary and standby on the same node
			Expect((master0 > 0) && (master1 > 0)).To(BeFalse())
		}
	})

	When("primarySegmentCount is increased by 1", func() {
		var expandedGpdbYamlFile string

		BeforeEach(func() {
			_, expandedGpdbYamlFile = CreateTempFile(GetGreenplumManifestYaml(3, GpYamlOptions{Standby: "yes", Mirrors: "yes"}))
			SetupGreenplumCluster(expandedGpdbYamlFile, *GreenplumImageTag)
		})

		It("auto-expands the cluster", func() {
			Expect(kubewait.ForJobSuccess("my-greenplum-gpexpand-job")).To(Succeed())
			checkSegmentConfigQuery := "SELECT count(*) FROM gp_segment_configuration WHERE status='u'"
			expectedQueryResult := "8"
			Expect(kubewait.ForPsqlQuery(checkSegmentConfigQuery, expectedQueryResult)).To(Succeed())
		})

		AfterEach(func() {
			DeleteGreenplumCluster()
			Expect(KubeDelete("--all", "pvc")).To(Succeed())
			VerifyPVCsAreDeleted()
		})
	})
})
