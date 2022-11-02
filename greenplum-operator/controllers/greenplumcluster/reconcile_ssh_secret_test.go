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

var _ = Describe("Greenplum Controller for ssh-secrets", func() {
	const (
		secretName = "ssh-secrets"
	)
	var (
		ctx                 context.Context
		logBuf              *gbytes.Buffer
		greenplumReconciler *greenplumcluster.GreenplumClusterReconciler
		greenplumCluster    *greenplumv1.GreenplumCluster

		gpSecretKey = types.NamespacedName{
			Name:      secretName,
			Namespace: namespaceName,
		}
	)
	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		logBuf = gbytes.NewBuffer()

		greenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:     reactiveClient,
			Log:        gplog.ForTest(logBuf),
			SSHCreator: &fakeSecretCreator{},
			PodExec:    &fake.PodExec{},
		}

		greenplumCluster = exampleGreenplumCluster.DeepCopy()
	})

	var reconcileErr error
	JustBeforeEach(func() {
		Expect(reactiveClient.Create(ctx, greenplumCluster)).To(Succeed())
		_, reconcileErr = greenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
	})

	When("we expect a secret to exist after Reconcile", func() {
		var gpSecret corev1.Secret
		JustBeforeEach(func() {
			Expect(reactiveClient.Get(ctx, gpSecretKey, &gpSecret)).To(Succeed())
		})
		When("secret doesn't exist before Reconcile", func() {
			It("succeeds", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
			})
			It("fills in data", func() {
				Expect(gpSecret.Data).NotTo(BeEmpty())
				Expect(gpSecret.Type).To(Equal(corev1.SecretTypeOpaque))
			})
			It("takes ownership", func() {
				Expect(gpSecret.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
			})
		})

		When("secret exists before Reconcile", func() {
			BeforeEach(func() {
				originalSecret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespaceName, Name: secretName},
					Data:       nil,
					Type:       "invalid",
				}
				Expect(reactiveClient.Create(ctx, originalSecret)).To(Succeed())
			})
			It("succeeds", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
			})
			It("overwrites data", func() {
				Expect(gpSecret.Data).NotTo(BeEmpty())
				Expect(gpSecret.Type).To(Equal(corev1.SecretTypeOpaque))
			})
			It("takes ownership", func() {
				Expect(gpSecret.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
			})
		})
	})

	When("Key generation fails", func() {
		BeforeEach(func() {
			greenplumReconciler.SSHCreator = fakeSecretCreator{err: errors.New("ssh-keygen failed")}
		})
		It("returns an error", func() {
			Expect(reconcileErr).To(MatchError("ssh-keygen failed"))
		})
	})

	When("secret creation fails", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("create", "secrets", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("secret creation error")
			})
		})
		It("returns error", func() {
			Expect(reconcileErr).To(MatchError("secret creation error"))
		})
	})

})

type fakeSecretCreator struct {
	err error
}

func (f fakeSecretCreator) GenerateKey() (map[string][]byte, error) {
	fakeSecret := map[string][]byte{
		"id_rsa":     []byte("foo"),
		"id_rsa.pub": []byte("bar"),
	}
	return fakeSecret, f.err
}
