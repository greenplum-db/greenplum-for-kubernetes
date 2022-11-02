package greenplumcluster_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Reconcile GreenplumCluster status", func() {
	var (
		ctx                    context.Context
		logBuf                 *gbytes.Buffer
		newGreenplumReconciler *greenplumcluster.GreenplumClusterReconciler
	)
	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		logBuf = gbytes.NewBuffer()

		// using a newer version of the reconciler
		newGreenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:        reactiveClient,
			Log:           gplog.ForTest(logBuf),
			SSHCreator:    fakeSecretCreator{},
			PodExec:       &fake.PodExec{},
			InstanceImage: "greenplum-for-kubernetes:new",
			OperatorImage: "greenplum-operator:new",
		}

		CreateClusterWithOldImages(*newGreenplumReconciler)
	})

	When("an outdated cluster is reconciled", func() {
		var (
			reconcileErr      error
			reconciledCluster greenplumv1.GreenplumCluster
		)
		JustBeforeEach(func() {
			_, reconcileErr = newGreenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
			Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
		})
		It("succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})

		DescribeTable(
			"does not change statefulset pod container images",
			func(statefulsetName string) {
				var statefulset appsv1.StatefulSet
				statefulsetKey := types.NamespacedName{Namespace: namespaceName, Name: statefulsetName}
				Expect(reactiveClient.Get(ctx, statefulsetKey, &statefulset)).To(Succeed())
				Expect(statefulset.Spec.Template.Spec.Containers[0].Image).To(Equal("greenplum-for-kubernetes:old"))
			},
			Entry("master", "master"),
			Entry("segment-a", "segment-a"),
		)
	})
})

func CreateClusterWithOldImages(prototypeReconciler greenplumcluster.GreenplumClusterReconciler) {
	// initialize cluster resources with an old version reconciler
	oldGreenplumReconciler := prototypeReconciler
	oldGreenplumReconciler.InstanceImage = "greenplum-for-kubernetes:old"
	oldGreenplumReconciler.OperatorImage = "greenplum-operator:old"

	greenplumCluster := exampleGreenplumCluster.DeepCopy()
	Expect(reactiveClient.Create(nil, greenplumCluster)).To(Succeed())

	_, reconcileErr := oldGreenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
	Expect(reconcileErr).NotTo(HaveOccurred())

	Expect(reactiveClient.Get(nil, greenplumClusterRequest.NamespacedName, greenplumCluster)).To(Succeed())
	Expect(greenplumCluster.Status.InstanceImage).To(Equal("greenplum-for-kubernetes:old"), "sanity")
	Expect(greenplumCluster.Status.OperatorVersion).To(Equal("greenplum-operator:old"), "sanity")
}
