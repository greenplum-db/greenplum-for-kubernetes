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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"
)

var _ = Describe("Reconcile configmap for GreenplumCluster", func() {
	var (
		ctx                 context.Context
		logBuf              *gbytes.Buffer
		greenplumReconciler *greenplumcluster.GreenplumClusterReconciler
		greenplumCluster    *greenplumv1.GreenplumCluster
	)
	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		logBuf = gbytes.NewBuffer()

		greenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:     reactiveClient,
			Log:        gplog.ForTest(logBuf),
			SSHCreator: fakeSecretCreator{},
			PodExec:    &fake.PodExec{},
		}

		greenplumCluster = exampleGreenplumCluster.DeepCopy()
	})

	var reconcileErr error
	JustBeforeEach(func() {
		Expect(reactiveClient.Create(ctx, greenplumCluster)).To(Succeed())
		_, reconcileErr = greenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
	})

	When("we expect a configmap to exist after Reconcile", func() {
		var configMap corev1.ConfigMap
		JustBeforeEach(func() {
			gpconfigKey := types.NamespacedName{Namespace: namespaceName, Name: "greenplum-config"}
			Expect(reactiveClient.Get(ctx, gpconfigKey, &configMap)).To(Succeed())
		})

		When("configmap doesn't exist before Reconcile", func() {
			It("succeeds", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
			})
			It("fills in data", func() {
				Expect(configMap.Data).NotTo(BeEmpty())
			})
			It("takes ownership", func() {
				Expect(configMap.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
			})
		})

		When("configmap exists before Reconcile", func() {
			BeforeEach(func() {
				originalConfigMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: "greenplum-config"},
					Data:       nil,
				}
				Expect(reactiveClient.Create(ctx, originalConfigMap)).To(Succeed())
			})
			It("succeeds", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
			})
			It("overwrites data", func() {
				Expect(configMap.Data).NotTo(BeEmpty())
			})
			It("takes ownership", func() {
				Expect(configMap.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
			})
		})
	})

	When("Creating configmap fails", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("create", "configmaps", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("error creating configmap")
			})
		})
		It("returns the error", func() {
			Expect(reconcileErr).To(MatchError("error creating configmap"))
		})
	})

})
