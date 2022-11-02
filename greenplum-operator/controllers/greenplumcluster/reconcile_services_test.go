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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"
)

var _ = Describe("Reconcile services for GreenplumCluster", func() {
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

	for _, sn := range []string{"agent", "greenplum"} {
		serviceName := sn

		When("we expect the "+serviceName+" service to exist after Reconcile", func() {
			var service corev1.Service
			JustBeforeEach(func() {
				serviceKey := types.NamespacedName{Namespace: namespaceName, Name: serviceName}
				Expect(reactiveClient.Get(ctx, serviceKey, &service)).To(Succeed())
			})

			When(serviceName+" service doesn't exist before Reconcile", func() {
				It("succeeds", func() {
					Expect(reconcileErr).NotTo(HaveOccurred())
				})
				It("fills in data", func() {
					Expect(service.Spec.Ports).NotTo(BeEmpty())
				})
				It("takes ownership", func() {
					Expect(service.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
				})
			})

			When(serviceName+" service exists before Reconcile", func() {
				BeforeEach(func() {
					originalService := &corev1.Service{
						ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: serviceName},
						Spec:       corev1.ServiceSpec{Ports: nil},
					}
					Expect(reactiveClient.Create(ctx, originalService)).To(Succeed())
				})
				It("succeeds", func() {
					Expect(reconcileErr).NotTo(HaveOccurred())
				})
				It("overwrites data", func() {
					Expect(service.Spec.Ports).NotTo(BeEmpty())
				})
				It("takes ownership", func() {
					Expect(service.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
				})
			})
		})

		When("Creating "+serviceName+" service fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("create", "services", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					om, err := meta.Accessor(action.(testing.CreateAction).GetObject())
					Expect(err).NotTo(HaveOccurred())
					if om.GetNamespace() == namespaceName && om.GetName() == serviceName {
						return true, nil, errors.New("error creating " + serviceName + " service")
					}
					return false, nil, nil
				})
			})
			It("returns the error", func() {
				Expect(reconcileErr).To(MatchError("error creating " + serviceName + " service"))
			})
		})
	}
})
