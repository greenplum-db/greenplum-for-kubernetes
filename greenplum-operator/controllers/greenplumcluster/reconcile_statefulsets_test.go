package greenplumcluster_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"
)

var _ = Describe("Reconcile statefulsets for GreenplumCluster", func() {
	var (
		ctx                 context.Context
		logBuf              *gbytes.Buffer
		greenplumReconciler *greenplumcluster.GreenplumClusterReconciler
		greenplumCluster    *greenplumv1.GreenplumCluster

		masterCPULimit  = resource.MustParse("1.0")
		segmentCPULimit = resource.MustParse("2.0")
	)
	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		logBuf = gbytes.NewBuffer()

		greenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:        reactiveClient,
			Log:           gplog.ForTest(logBuf),
			SSHCreator:    fakeSecretCreator{},
			PodExec:       &fake.PodExec{},
			InstanceImage: "greenplum-for-kubernetes:v1.0",
		}

		greenplumCluster = exampleGreenplumCluster.DeepCopy()
		greenplumCluster.Spec.MasterAndStandby.CPU = masterCPULimit
		greenplumCluster.Spec.Segments.CPU = segmentCPULimit

	})

	var reconcileErr error
	JustBeforeEach(func() {
		Expect(reactiveClient.Create(ctx, greenplumCluster)).To(Succeed())
		_, reconcileErr = greenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
	})

	for _, ss := range []struct {
		mirrors         string
		statefulsetName string
		cpuLimit        resource.Quantity
	}{
		{mirrors: "", statefulsetName: "master", cpuLimit: masterCPULimit},
		{mirrors: "", statefulsetName: "segment-a", cpuLimit: segmentCPULimit},
		{mirrors: "yes", statefulsetName: "master", cpuLimit: masterCPULimit},
		{mirrors: "yes", statefulsetName: "segment-a", cpuLimit: segmentCPULimit},
		{mirrors: "yes", statefulsetName: "segment-b", cpuLimit: segmentCPULimit},
	} {
		mirrors := ss.mirrors
		statefulsetName := ss.statefulsetName
		cpuLimit := ss.cpuLimit

		When("mirrors: \""+mirrors+"\" and we expect the "+statefulsetName+" statefulset to exist after Reconcile", func() {
			var statefulset appsv1.StatefulSet
			BeforeEach(func() {
				greenplumCluster.Spec.Segments.Mirrors = mirrors
			})
			JustBeforeEach(func() {
				statefulsetKey := types.NamespacedName{Namespace: namespaceName, Name: statefulsetName}
				Expect(reactiveClient.Get(ctx, statefulsetKey, &statefulset)).To(Succeed())
			})

			When(statefulsetName+" statefulset doesn't exist before Reconcile", func() {
				It("succeeds", func() {
					Expect(reconcileErr).NotTo(HaveOccurred())
				})
				It("fills in pod template with the right image", func() {
					Expect(statefulset.Spec.Template.Spec.Containers[0].Image).To(Equal("greenplum-for-kubernetes:v1.0"))
				})
				It("fills in pod template with the correct resource request", func() {
					ssetCPULimit := statefulset.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU]
					Expect(ssetCPULimit.Equal(cpuLimit)).To(BeTrue())
				})
				It("fills in pod template with the right service account name", func() {
					Expect(statefulset.Spec.Template.Spec.ServiceAccountName).To(Equal("greenplum-system-pod"))
				})
				It("takes ownership", func() {
					Expect(statefulset.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
				})
			})

			When(statefulsetName+" statefulset exists before Reconcile", func() {
				BeforeEach(func() {
					originalStatefulSet := &appsv1.StatefulSet{
						ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: statefulsetName},
						Spec: appsv1.StatefulSetSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{{
										Image: "oops-wrong-image",
										Resources: corev1.ResourceRequirements{
											Limits: map[corev1.ResourceName]resource.Quantity{
												corev1.ResourceCPU: cpuLimit,
											},
										},
									}},
									ServiceAccountName: "wrong-service-account-name",
								},
							}},
					}
					Expect(reactiveClient.Create(ctx, originalStatefulSet)).To(Succeed())
				})
				It("succeeds", func() {
					Expect(reconcileErr).NotTo(HaveOccurred())
				})
				It("overwrites pod template with the right image", func() {
					Expect(statefulset.Spec.Template.Spec.Containers[0].Image).To(Equal("greenplum-for-kubernetes:v1.0"))
				})
				It("overwrites pod template with the correct resource request", func() {
					ssetCPULimit := statefulset.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU]
					Expect(ssetCPULimit.Equal(cpuLimit)).To(BeTrue())
				})
				It("overwrites pod template with the correct service account", func() {
					Expect(statefulset.Spec.Template.Spec.ServiceAccountName).To(Equal("greenplum-system-pod"))
				})
				It("takes ownership", func() {
					Expect(statefulset.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
				})
			})
		})

		When("Creating "+statefulsetName+" statefulset fails with mirrors: \""+mirrors+"\"", func() {
			BeforeEach(func() {
				greenplumCluster.Spec.Segments.Mirrors = mirrors
				reactiveClient.PrependReactor("create", "statefulsets", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					om, err := meta.Accessor(action.(testing.CreateAction).GetObject())
					Expect(err).NotTo(HaveOccurred())
					if om.GetNamespace() == namespaceName && om.GetName() == statefulsetName {
						return true, nil, errors.New("error creating " + statefulsetName + " statefulset")
					}
					return false, nil, nil
				})
			})
			It("returns the error", func() {
				Expect(reconcileErr).To(MatchError("error creating " + statefulsetName + " statefulset"))
			})
		})
	}

	When(`mirrors: "no"`, func() {
		BeforeEach(func() {
			greenplumCluster.Spec.Segments.Mirrors = "no"
		})
		It("doesn't create statefulsets/segment-b", func() {
			var statefulset appsv1.StatefulSet
			statefulsetKey := types.NamespacedName{Namespace: namespaceName, Name: "segment-b"}
			Expect(reactiveClient.Get(ctx, statefulsetKey, &statefulset)).To(MatchError(`statefulsets.apps "segment-b" not found`))
		})
	})
})
