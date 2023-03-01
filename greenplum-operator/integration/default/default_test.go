package default_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubeexecpsql"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/integrationutils/kubewait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Happy path integration greenplum operator", func() {
	var gpdbYamlFile string
	var tempDir string
	var suiteFailed bool
	var log logr.Logger

	BeforeSuite(func() {
		suiteFailed = true // Mark as failed in case we don't get to the end. This is how we can detect a failure in BeforeSuite().
		ctrl.SetLogger(gplog.ForIntegration())
		Expect(kubewait.ForAPIService()).To(Succeed())
		CleanUpK8s(*OperatorImageTag)
		VerifyPVCsAreDeleted()

		manifestYaml := GetGreenplumManifestYaml(2, GpYamlOptions{})
		tempDir, gpdbYamlFile = CreateTempFile(manifestYaml)
		suiteFailed = false

		log = ctrl.Log.WithName("default")
	})

	AfterSuite(func() {
		if !suiteFailed {
			CleanUpK8s(*OperatorImageTag)
			VerifyPVCsAreDeleted()
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			suiteFailed = true
		}
	})

	When("the cluster is running", func() {
		BeforeEach(func() {
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
		It("has Greenplum master and segment-a stateful sets and PVCs, but no segment-b", func() {
			cmd := exec.Command("kubectl", "get", "greenplumclusters")
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("my-greenplum"))

			cmd = exec.Command("kubectl", "get", "pvc", "-l", "greenplum-cluster=my-greenplum")
			out, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("my-greenplum-pgdata-master-0"))
			Expect(string(out)).NotTo(ContainSubstring("my-greenplum-pgdata-master-1"))
			Expect(string(out)).To(ContainSubstring("my-greenplum-pgdata-segment-a-0"))
			Expect(string(out)).To(ContainSubstring("my-greenplum-pgdata-segment-a-1"))
			Expect(string(out)).NotTo(ContainSubstring("my-greenplum-pgdata-segment-b-0"))

			cmd = exec.Command("kubectl", "describe", "greenplumCluster/my-greenplum")
			out, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
		})
		It("shows greenplumClusters in kubectl get all", func() {
			cmd := exec.Command("kubectl", "get", "all")
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("my-greenplum"))
		})
		It("should set Status.InstanceImage to latest", func() {
			VerifyStatusInstanceImage(*GreenplumImageTag)
		})
		It("should set Status.OperatorVersion to latest", func() {
			VerifyStatusOperatorVersion(*OperatorImageTag)
		})
		It("should set Status.Phase to 'Running' ", func() {
			Expect(kubewait.ForGreenplumClusterStatus(greenplumv1.GreenplumClusterPhaseRunning)).To(Succeed())
		})
		It("fails to query through service without HBA auth", func() {
			out, err := ExecutePsqlQueryThroughService("select 42")
			Expect(err).To(HaveOccurred())
			Expect(string(out)).To(ContainSubstring("no pg_hba.conf entry"))
		})
		When("HBA entry is added", func() {
			BeforeEach(func() {
				AddHbaToAllowAccessToGreenplumThroughService("master-0")
			})
			It("queries through service with HBA auth", func() {
				VerifyDataThroughService()
			})
		})
		It("cleans defunct processes", func() {
			out, err := KubeExec("master-0", "ps -ef | grep [d]efunct | wc -l")
			Expect(strconv.Atoi(strings.TrimSpace(string(out)))).To(Equal(0))
			Expect(err).NotTo(HaveOccurred(), "should be able to run command in master-0: %s", string(out))
		})
		When("a segment is restarted", func() {
			BeforeEach(func() {
				AddHbaToAllowAccessToGreenplumThroughService("master-0")
			})
			It("starts postgres and joins the cluster", func() {
				dbHost, dbPortStr, err := GreenplumService()
				Expect(err).NotTo(HaveOccurred())
				connStr := fmt.Sprintf("postgres://gpadmin@%s:%s/gpadmin?sslmode=disable", dbHost, dbPortStr)
				db, err := sql.Open("postgres", connStr)
				Expect(err).NotTo(HaveOccurred())

				var pgBackendPidBefore int64
				err = db.QueryRowContext(context.Background(), "select pg_backend_pid()").Scan(&pgBackendPidBefore)
				Expect(err).NotTo(HaveOccurred())

				Expect(KubeDelete("pod/segment-a-0")).To(Succeed())
				Expect(kubewait.ForConsistentDNSResolution("segment-a-0", "master-0")).To(Succeed())

				var fooData int64
				Eventually(func() error {
					log.Info("running test query")
					ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(6*time.Second))
					defer cancel()
					err = db.QueryRowContext(ctx, "select * from foo").Scan(&fooData)
					if err != nil {
						log.Info("query failed", "error", err.Error())
					}
					return err
				}, 60*time.Second, 1*time.Second).Should(Succeed())
				Expect(fooData).To(Equal(int64(1)))

				// verify the conn is same as that before and query
				var pgBackendPidAfter int64
				err = db.QueryRowContext(context.Background(), "select pg_backend_pid()").Scan(&pgBackendPidAfter)
				Expect(err).NotTo(HaveOccurred())
				Expect(pgBackendPidBefore).To(Equal(pgBackendPidAfter))
			})
		})
		//		When("a segment fails COMMIT_PREPARED", func() {
		//			var db *sql.DB
		//			var pgBackendPidBefore int64
		//			BeforeEach(func() {
		//				AddHbaToAllowAccessToGreenplumThroughService("master-0")
		//				dbHost, dbPortStr, err := GreenplumService()
		//				Expect(err).NotTo(HaveOccurred())
		//				connStr := fmt.Sprintf("postgres://gpadmin@%s:%s/gpadmin?sslmode=disable", dbHost, dbPortStr)
		//				db, err = sql.Open("postgres", connStr)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				err = db.QueryRowContext(context.Background(), "select pg_backend_pid()").Scan(&pgBackendPidBefore)
		//				Expect(err).NotTo(HaveOccurred())
		//
		//				result, err := db.ExecContext(context.Background(), "CREATE EXTENSION IF NOT EXISTS gp_inject_fault;")
		//				Expect(err).NotTo(HaveOccurred())
		//				Expect(result.RowsAffected()).To(Equal(int64(0)))
		//
		//				log.Info("injecting a fault in segment-a-0 that prevents COMMIT_PREPARED from completing")
		//				var injectResult string
		//				err = db.QueryRowContext(
		//					context.Background(),
		//					"SELECT gp_inject_fault_infinite('finish_prepared_start_of_function', 'error' , //dbid) from gp_segment_configuration where content=0 and role='p';",
		//				).Scan(&injectResult)
		//				Expect(err).NotTo(HaveOccurred())
		//				Expect(injectResult).To(Equal("Success:"))
		//
		//				result, err = db.ExecContext(context.Background(), //"CREATE TABLE master_panic_recovery as select generate_series(1,100)i")
		//				Expect(err).Should(MatchError("pq: the database system is in recovery mode"))
		//			})
		//			When("the DTM eventually recovers", func() {
		//				BeforeEach(func() {
		//					log.Info("forcing segment-a-0 to restart")
		//					Expect(KubeDelete("pod/segment-a-0")).To(Succeed())
		//					Expect(kubewait.ForConsistentDNSResolution("segment-a-0", "master-0")).To(Succeed())
		//				})
		//				It("does not lose transaction data after crash recovery", func() {
		//					var fooData int64
		//					Eventually(func() error {
		//						log.Info("running test query")
		//						ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(6*time.Second))
		//						defer cancel()
		//						err := db.QueryRowContext(ctx, "select count(*) from master_panic_recovery").Scan(&fooData)
		//						if err != nil {
		//							log.Info("query failed", "error", err.Error())
		//						}
		//						return err
		//					}, 120*time.Second, 1*time.Second).Should(Succeed())
		//					Expect(fooData).To(Equal(int64(100)))
		//
		//					// verify the original connection was closed
		//					var pgBackendPidAfter int64
		//					err := db.QueryRowContext(context.Background(), "select pg_backend_pid()").Scan(&pgBackendPidAfter)
		//					Expect(err).NotTo(HaveOccurred())
		//					Expect(pgBackendPidBefore).NotTo(Equal(pgBackendPidAfter))
		//				})
		//			})
		//		})
		When("master-0 is restarted", func() {
			BeforeEach(func() {
				Expect(KubeDelete("pod/master-0")).To(Succeed())
			})
			It("starts postgres and joins the cluster", func() {
				Expect(kubewait.ForNetworkReady(2, false)).To(Succeed())
				VerifyData()
			})
		})
		When("primarySegmentCount is increased by 1", func() {
			BeforeEach(func() {
				_, gpdbYamlFile := CreateTempFile(GetGreenplumManifestYaml(3, GpYamlOptions{}))
				SetupGreenplumCluster(gpdbYamlFile, *GreenplumImageTag)
			})
			It("auto-expands the cluster", func() {
				Expect(kubewait.ForJobSuccess("my-greenplum-gpexpand-job")).To(Succeed())
				checkSegmentConfigQuery := "SELECT count(*) FROM gp_segment_configuration WHERE status='u'"
				expectedQueryResult := "4"
				Expect(kubewait.ForPsqlQuery(checkSegmentConfigQuery, expectedQueryResult)).To(Succeed())
			})
		})
		When("primarySegmentCount is increased by 1 without cleaning up gpexpand schema", func() {
			var expandGpdbYamlFile string
			BeforeEach(func() {
				_, expandGpdbYamlFile = CreateTempFile(GetGreenplumManifestYaml(4, GpYamlOptions{}))
			})
			It("rejects the expansion request", func() {
				out, err := ApplyManifest(expandGpdbYamlFile, "greenplumcluster")
				Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
				Expect(out).To(ContainSubstring("admission webhook \"greenplum.pivotal.io\" denied the request: previous expansion schema exists. you must redistribute data and clean up expansion schema prior to performing another expansion"))
			})
		})
		It("rejects duplicate CREATE requests", func() {
			cmd := exec.Command("kubectl", "create", "-f", gpdbYamlFile)
			out, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())
			Expect(string(out)).To(ContainSubstring(`admission webhook "greenplum.pivotal.io" denied the request: only one GreenplumCluster is allowed in namespace default`))
		})
	})

	When("existing gpInstance is deleted, but PVCs are not deleted", func() {
		BeforeEach(func() {
			DeleteGreenplumClusterIfExists()
		})

		When("standby value is changed", func() {
			var standbyGpdbYamlFile string
			BeforeEach(func() {
				manifestYaml := GetGreenplumManifestYaml(3, GpYamlOptions{Standby: "yes"})
				_, standbyGpdbYamlFile = CreateTempFile(manifestYaml)
			})

			It("rejects the redeploy request", func() {
				out, err := ApplyManifest(standbyGpdbYamlFile, "greenplumcluster")
				Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
				Expect(out).To(ContainSubstring("admission webhook \"greenplum.pivotal.io\" denied the request: my-greenplum has PVCs for 1 masters. masterAndStandby.standby cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
		})

		When("creating a cluster with mirrors", func() {
			var segCountGpdbYamlFile string
			BeforeEach(func() {
				manifestYaml := GetGreenplumManifestYaml(3, GpYamlOptions{Mirrors: "yes"})
				_, segCountGpdbYamlFile = CreateTempFile(manifestYaml)
			})

			It("rejects the redeploy request", func() {
				out, err := ApplyManifest(segCountGpdbYamlFile, "greenplumcluster")
				Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
				Expect(out).To(ContainSubstring("admission webhook \"greenplum.pivotal.io\" denied the request: my-greenplum has PVCs for 0 mirrors. segments.mirrors cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
		})

		When("storageclass and storage are updated", func() {
			var storageGpdbYamlFile string
			BeforeEach(func() {
				manifestYaml := GetGreenplumManifestYaml(3, GpYamlOptions{StorageClass: "newclass"})
				_, storageGpdbYamlFile = CreateTempFile(manifestYaml)
			})

			It("rejects the redeploy request", func() {
				out, err := ApplyManifest(storageGpdbYamlFile, "greenplumcluster")
				Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
				Expect(out).To(ContainSubstring("admission webhook \"greenplum.pivotal.io\" denied the request: storageClassName cannot be changed without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
		})

		When("primarySegmentCount is decreased", func() {
			var segCountGpdbYamlFile string
			BeforeEach(func() {
				manifestYaml := GetGreenplumManifestYaml(2, GpYamlOptions{})
				_, segCountGpdbYamlFile = CreateTempFile(manifestYaml)
			})

			It("rejects the redeploy request", func() {
				out, err := ApplyManifest(segCountGpdbYamlFile, "greenplumcluster")
				Expect(err).To(BeAssignableToTypeOf(&exec.ExitError{}))
				Expect(out).To(ContainSubstring("admission webhook \"greenplum.pivotal.io\" denied the request: my-greenplum has PVCs for 3 segments. segments.primarySegmentCount cannot be decreased without first deleting PVCs. This will result in a new, empty Greenplum cluster"))
			})
		})
	})

	// NB: this should come last since it deletes the cluster and operator.
	When("deleting the operator", func() {
		BeforeEach(func() {
			DeleteGreenplumClusterIfExists()
			Expect(exec.Command("helm", "uninstall", "greenplum-operator").Run()).To(Succeed())
		})
		It("deletes owned resources", func() {
			EventuallyResourceShouldBeDeleted("customresourcedefinition/greenplumclusters.greenplum.pivotal.io")
		})
	})
})

func EventuallyResourceShouldBeDeleted(resource string) {
	Eventually(func() string {
		cmd := exec.Command("kubectl", "get", resource)
		out, err := cmd.CombinedOutput()
		Expect(err).To(Or(Not(HaveOccurred()), BeAssignableToTypeOf(&exec.ExitError{})))
		return string(out)
	}, 30*time.Second, 1*time.Second).Should(HavePrefix("Error from server (NotFound):"))
}
