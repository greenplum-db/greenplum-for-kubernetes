package component

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("component")

var _ = Describe("Components", func() {
	It("installs and checks madlib successfully", func() {
		log.Info("Installing madlib")
		out, err := KubeExec("master-0", "source /usr/local/greenplum-db/greenplum_path.sh && /usr/local/madlib/bin/madpack -p greenplum install")
		if err != nil {
			log.Info("madlib install failed")
			fmt.Println(string(out))
		}
		Expect(err).NotTo(HaveOccurred())
		log.Info("madlib installed")
		log.Info("Running madlib install-check")
		out, err = KubeExec("master-0", "source /usr/local/greenplum-db/greenplum_path.sh && /usr/local/madlib/bin/madpack -p greenplum install-check")
		if err != nil {
			log.Info("madlib install check failed")
			fmt.Println(string(out))
		}
		Expect(err).NotTo(HaveOccurred())
		log.Info("madlib install-check passed")
	})

	When("using PXF", func() {
		BeforeEach(func() {
			CreatePXFExternalTable()
			Expect(kubewait.ForReplicasReady("deployment", "my-greenplum-pxf")).To(Succeed())
		})

		It("can download data from S3 bucket", func() {
			log.Info("Querying PXF...")
			out, err := Query("master-0", "SELECT * FROM lineitem_s3_1 limit 10")
			Expect(err).NotTo(HaveOccurred(), string(out))
		})

		AfterEach(func() {
			DropPXFExternalTable("lineitem_s3_1")
		})
	})

	Context("Idempotency", func() {
		When("reconciling gpdb", func() {
			var previousOperatorLogs string
			BeforeEach(func() {
				previousOperatorLogs = getOperatorLogs()
				AnnotateObject("greenplumcluster/my-greenplum", "test=test")
			})
			It("should not unnecessarily UPDATE subresources", func() {
				Consistently(func() string {
					afterOperatorLogs := getOperatorLogs()
					return strings.TrimPrefix(afterOperatorLogs, previousOperatorLogs)
				}, 5*time.Second, 1*time.Second).ShouldNot(ContainSubstring("updated"))
				afterOperatorLogs := strings.TrimPrefix(getOperatorLogs(), previousOperatorLogs)
				Expect(DecodeLogs(strings.NewReader(afterOperatorLogs))).To(ContainLogEntry(gstruct.Keys{
					"msg":        Equal("Successfully Reconciled"),
					"controller": Equal("greenplumcluster"),
					"name":       Equal("my-greenplum"),
				}))
			})
		})
		When("reconciling pxf", func() {
			var previousOperatorLogs string
			BeforeEach(func() {
				previousOperatorLogs = getOperatorLogs()
				AnnotateObject("greenplumpxfservice/my-greenplum-pxf", "test=test")
			})
			It("should not unnecessarily UPDATE subresources", func() {
				Consistently(func() string {
					afterOperatorLogs := getOperatorLogs()
					return strings.TrimPrefix(afterOperatorLogs, previousOperatorLogs)
				}, 5*time.Second, 1*time.Second).ShouldNot(ContainSubstring("updated"))
				afterOperatorLogs := strings.TrimPrefix(getOperatorLogs(), previousOperatorLogs)
				Expect(DecodeLogs(strings.NewReader(afterOperatorLogs))).To(ContainLogEntry(gstruct.Keys{
					"msg":        Equal("Successfully Reconciled"),
					"controller": Equal("greenplumpxfservice"),
					"name":       Equal("my-greenplum-pxf"),
				}))
			})
		})
	})
})

func getOperatorLogs() string {
	operatorLogs, _ := exec.Command("kubectl", "logs", "-l", "app=greenplum-operator", "--tail", "-1").CombinedOutput()
	return string(operatorLogs)
}
