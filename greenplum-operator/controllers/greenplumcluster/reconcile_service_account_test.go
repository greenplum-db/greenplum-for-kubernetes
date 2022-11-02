package greenplumcluster_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"
)

var _ = Describe("Reconcile pod service account for GreenplumCluster", func() {
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

	When("we expect a service account, role, and rolebinding for greenplum cluster pods to exist after Reconcile", func() {
		var (
			serviceAccount corev1.ServiceAccount
			role           rbacv1.Role
			roleBinding    rbacv1.RoleBinding
		)
		JustBeforeEach(func() {
			serviceAccountKey := types.NamespacedName{Namespace: namespaceName, Name: "greenplum-system-pod"}
			Expect(reactiveClient.Get(ctx, serviceAccountKey, &serviceAccount)).To(Succeed())

			roleKey := types.NamespacedName{Namespace: namespaceName, Name: "greenplum-system-pod"}
			Expect(reactiveClient.Get(ctx, roleKey, &role)).To(Succeed())

			roleBindingKey := types.NamespacedName{Namespace: namespaceName, Name: "greenplum-system-pod"}
			Expect(reactiveClient.Get(ctx, roleBindingKey, &roleBinding)).To(Succeed())
		})

		It("succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})

		It("creates a serviceaccount", func() {
			Expect(serviceAccount.ObjectMeta.Name).To(Equal("greenplum-system-pod"))
			Expect(serviceAccount.ObjectMeta.Namespace).To(Equal(namespaceName))
		})

		It("creates a role with permission to get and patch pvcs, get pods, and watch and list endpoints", func() {
			Expect(role.Rules).To(ConsistOf(
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Resources": ConsistOf("persistentvolumeclaims"),
					"Verbs":     ConsistOf("get", "patch"),
				}),
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Resources": ConsistOf("pods"),
					"Verbs":     ConsistOf("get"),
				}),
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Resources": ConsistOf("endpoints"),
					"Verbs":     ConsistOf("watch", "list"),
				}),
			))
		})

		It("creates a rolebinding between the service account and role", func() {
			Expect(roleBinding.RoleRef.Kind).To(Equal("Role"))
			Expect(roleBinding.RoleRef.Name).To(Equal("greenplum-system-pod"))

			Expect(roleBinding.Subjects).To(HaveLen(1))
			Expect(roleBinding.Subjects[0].Kind).To(Equal("ServiceAccount"))
			Expect(roleBinding.Subjects[0].Name).To(Equal("greenplum-system-pod"))
			Expect(roleBinding.Subjects[0].Namespace).To(Equal(namespaceName))
		})

		It("takes ownership", func() {
			Expect(role.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
			Expect(roleBinding.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
			Expect(serviceAccount.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
		})

		When("the serviceaccount, role, and rolebinding already exist", func() {
			var updateOccurred bool
			BeforeEach(func() {
				_, reconcileErr = greenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
				Expect(reconcileErr).NotTo(HaveOccurred())

				reactiveClient.PrependReactor("update", "serviceaccounts", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					updateOccurred = true
					return false, nil, nil
				})

				reactiveClient.PrependReactor("update", "roles", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					updateOccurred = true
					return false, nil, nil
				})

				reactiveClient.PrependReactor("update", "rolebindings", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					updateOccurred = true
					return false, nil, nil
				})
			})

			It("does not update them", func() {
				Expect(updateOccurred).To(BeFalse())
				Expect(reconcileErr).NotTo(HaveOccurred())
			})
		})
	})

	When("we fail to create the service account", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("create", "serviceaccounts", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("failed to create service account")
			})
		})
		It("handles the error", func() {
			Expect(reconcileErr).To(MatchError("failed to create service account"))
		})
	})

	When("we fail to create the role", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("create", "roles", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("failed to create role")
			})
		})
		It("handles the error", func() {
			Expect(reconcileErr).To(MatchError("failed to create role"))
		})
	})

	When("we fail to create the rolebinding", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("create", "rolebindings", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("failed to create rolebinding")
			})
		})
		It("handles the error", func() {
			Expect(reconcileErr).To(MatchError("failed to create rolebinding"))
		})
	})
})
