package greenplumcluster_test

import (
	"context"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Reconcile GreenplumCluster status", func() {
	var (
		ctx                 context.Context
		logBuf              *gbytes.Buffer
		greenplumReconciler *greenplumcluster.GreenplumClusterReconciler
		greenplumCluster    *greenplumv1.GreenplumCluster
		podExec             *fake.PodExec
	)
	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		logBuf = gbytes.NewBuffer()

		podExec = &fake.PodExec{}
		greenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:        reactiveClient,
			Log:           gplog.ForTest(logBuf),
			SSHCreator:    fakeSecretCreator{},
			PodExec:       podExec,
			InstanceImage: "greenplum-for-kubernetes:greenplumv1.0",
			OperatorImage: "greenplum-operator:greenplumv1.0",
		}

		greenplumCluster = exampleGreenplumCluster.DeepCopy()
		// Adding this to prevent updates for adding finalizers
		greenplumCluster.Finalizers = []string{greenplumcluster.StopClusterFinalizer}
	})

	var reconcileErr error
	var reconcileResult ctrl.Result
	JustBeforeEach(func() {
		Expect(reactiveClient.Create(ctx, greenplumCluster)).To(Succeed())
		reconcileResult, reconcileErr = greenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
	})

	Context("initial cluster creation", func() {
		var reconciledCluster greenplumv1.GreenplumCluster
		BeforeEach(func() {
			greenplumCluster.Status = greenplumv1.GreenplumClusterStatus{}
			podExec.ErrorMsgOnMaster0 = "not active"
			podExec.ErrorMsgOnMaster1 = "not active"
		})
		JustBeforeEach(func() {
			Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
		})
		It("succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})
		It("requeues after 5 seconds", func() {
			Expect(reconcileResult).To(Equal(ctrl.Result{RequeueAfter: 5 * time.Second}))
		})
		It("sets OperatorVersion in the status", func() {
			Expect(reconciledCluster.Status.OperatorVersion).To(Equal("greenplum-operator:greenplumv1.0"))
		})
		It("sets InstanceImage in the status", func() {
			Expect(reconciledCluster.Status.InstanceImage).To(Equal("greenplum-for-kubernetes:greenplumv1.0"))
		})
		It("sets Phase to Pending", func() {
			Expect(reconciledCluster.Status.Phase).To(Equal(greenplumv1.GreenplumClusterPhasePending))
		})
		When("patching status fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("patch", "greenplumclusters", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					a := action.(testing.PatchAction)
					patchBytes := a.GetPatch()
					if strings.Contains(string(patchBytes), `"status"`) {
						return true, nil, errors.New("patch status error")
					}
					return true, nil, nil
				})
			})
			It("returns the error", func() {
				Expect(reconcileErr).To(MatchError("updating status: patch status error"))
			})
		})
	})

	When("status is already set and cluster is not yet running", func() {
		var statusUpdated = false
		BeforeEach(func() {
			greenplumCluster.Status = greenplumv1.GreenplumClusterStatus{
				InstanceImage:   greenplumReconciler.InstanceImage,
				OperatorVersion: greenplumReconciler.OperatorImage,
				Phase:           "a-phase",
			}
			reactiveClient.PrependReactor("patch", "greenplumclusters", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				a := action.(testing.PatchAction)
				patchBytes := a.GetPatch()
				if strings.Contains(string(patchBytes), "status") {
					statusUpdated = true
				}
				return true, nil, nil
			})
			podExec.ErrorMsgOnMaster0 = "not active"
			podExec.ErrorMsgOnMaster1 = "not active"
		})
		It("succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})
		It("requeues after 5 seconds", func() {
			Expect(reconcileResult).To(Equal(ctrl.Result{RequeueAfter: 5 * time.Second}))
		})
		It("does not update status", func() {
			Expect(statusUpdated).To(BeFalse(), "should not update")
		})
	})

	When("status is pending and cluster is now running", func() {
		BeforeEach(func() {
			greenplumCluster.Status = greenplumv1.GreenplumClusterStatus{
				InstanceImage:   greenplumReconciler.InstanceImage,
				OperatorVersion: greenplumReconciler.OperatorImage,
				Phase:           greenplumv1.GreenplumClusterPhasePending,
			}
		})
		It("succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})
		It("updates the status to `Running`", func() {
			var reconciledCluster greenplumv1.GreenplumCluster
			Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
			Expect(reconciledCluster.Status.Phase).To(Equal(greenplumv1.GreenplumClusterPhaseRunning))
		})
	})
})
