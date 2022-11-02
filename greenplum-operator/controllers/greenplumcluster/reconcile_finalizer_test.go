package greenplumcluster_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/testing"
)

var _ = Describe("Reconcile stopcluster.greenplumclusters.pivotal.io finalizer", func() {
	var (
		ctx                 context.Context
		logBuf              *gbytes.Buffer
		greenplumReconciler *greenplumcluster.GreenplumClusterReconciler
		greenplumCluster    *greenplumv1.GreenplumCluster
		podExec             *fake.PodExec
	)

	const gpstopCommand = "/bin/bash -c -- source /usr/local/greenplum-db/greenplum_path.sh && gpstop -aM immediate"

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		logBuf = gbytes.NewBuffer()

		podExec = &fake.PodExec{}
		greenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:        reactiveClient,
			Log:           gplog.ForTest(logBuf),
			SSHCreator:    fakeSecretCreator{},
			InstanceImage: "greenplum-for-kubernetes:v1.0",
			PodExec:       podExec,
		}

		greenplumCluster = exampleGreenplumCluster.DeepCopy()
		greenplumCluster.Spec.MasterAndStandby.Standby = "yes"
	})

	var reconcileErr error
	JustBeforeEach(func() {
		Expect(reactiveClient.Create(ctx, greenplumCluster)).To(Succeed())
		_, reconcileErr = greenplumReconciler.Reconcile(context.TODO(), greenplumClusterRequest)
	})

	It("Succeeds", func() {
		Expect(reconcileErr).NotTo(HaveOccurred())
	})

	It("adds a finalizer to a new cluster", func() {
		var reconciledCluster greenplumv1.GreenplumCluster
		Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
		Expect(reconciledCluster.Finalizers).Should(ContainElement(greenplumcluster.StopClusterFinalizer))
	})

	When("patch fails", func() {
		BeforeEach(func() {
			reactiveClient.PrependReactor("patch", "greenplumclusters", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				a := action.(testing.PatchAction)
				patchBytes := a.GetPatch()
				if strings.Contains(string(patchBytes), "finalizers") {
					return true, nil, errors.New("failed to patch finalizers")
				}
				return true, nil, nil
			})
		})
		It("returns an error (and re-queues)", func() {
			Expect(reconcileErr).To(MatchError("adding finalizer: failed to patch finalizers"))
		})
	})

	When("greenplumcluster exists with a finalizer", func() {
		BeforeEach(func() {
			greenplumCluster.Finalizers = []string{
				greenplumcluster.StopClusterFinalizer,
				"another.finalizer",
			}
		})
		It("Succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})
		It("doesn't modify the finalizers", func() {
			var reconciledCluster greenplumv1.GreenplumCluster
			Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
			Expect(reconciledCluster.Finalizers).To(Equal(greenplumCluster.Finalizers))
		})
		It("does not run gpstop", func() {
			Expect(podExec.RecordedCommands).NotTo(ContainElement(gpstopCommand))
		})
	})

	When("a cluster has a deletion timestamp and no finalizer", func() {
		BeforeEach(func() {
			dt := metav1.Date(2020, 2, 20, 13, 40, 49, 0, time.UTC)
			greenplumCluster.DeletionTimestamp = &dt
		})
		It("Succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})
		It("doesn't modify the finalizers", func() {
			var reconciledCluster greenplumv1.GreenplumCluster
			Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
			Expect(reconciledCluster.Finalizers).To(BeEmpty())
		})
		It("does not run gpstop", func() {
			Expect(podExec.RecordedCommands).NotTo(ContainElement(gpstopCommand))
		})
	})

	When("cluster is marked for deletion and has the finalizer", func() {
		BeforeEach(func() {
			dt := metav1.Date(2020, 2, 20, 13, 40, 49, 0, time.UTC)
			greenplumCluster.DeletionTimestamp = &dt
			greenplumCluster.Finalizers = []string{
				greenplumcluster.StopClusterFinalizer,
				"another.finalizer",
			}
		})
		It("Succeeds", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})
		It("removes the finalizer", func() {
			var reconciledCluster greenplumv1.GreenplumCluster
			Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
			Expect(reconciledCluster.Finalizers).ShouldNot(ContainElement(greenplumcluster.StopClusterFinalizer))
		})
		When("master-0 is active master", func() {
			It("runs gpstop on master-0", func() {
				Expect(podExec.CalledPodName).To(Equal("master-0"))
				Expect(podExec.RecordedCommands).To(ContainElement(gpstopCommand))
			})
			It("logs to indicate progress", func() {
				logs, err := DecodeLogs(bytes.NewReader(logBuf.Contents()))
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("initiating shutdown of the greenplum cluster")}))
				Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("success shutting down the greenplum cluster")}))
			})
			It("sets greenplumcluster status to `Deleting`", func() {
				var reconciledCluster greenplumv1.GreenplumCluster
				Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
				Expect(reconciledCluster.Status.Phase).To(Equal(greenplumv1.GreenplumClusterPhaseDeleting))
			})
			When("gpstop fails", func() {
				BeforeEach(func() {
					podExec.ErrorMsgOnCommand = "failed to run gpstop"
				})
				It("logs an error message", func() {
					logs, err := DecodeLogs(bytes.NewReader(logBuf.Contents()))
					Expect(err).NotTo(HaveOccurred())
					Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("initiating shutdown of the greenplum cluster")}))
					Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("greenplum cluster did not shutdown cleanly. Please check gpAdminLogs for more info.")}))
				})
			})
		})
		When("master-1 is active master", func() {
			BeforeEach(func() {
				podExec.ErrorMsgOnMaster0 = "not active"
			})
			It("runs gpstop on master-1", func() {
				Expect(podExec.CalledPodName).To(Equal("master-1"))
				Expect(podExec.RecordedCommands).To(ContainElement(gpstopCommand))
			})
		})
		When("there is no active master", func() {
			BeforeEach(func() {
				podExec.ErrorMsgOnMaster0 = "not active"
				podExec.ErrorMsgOnMaster1 = "not active"
			})
			It("does not run gpstop", func() {
				Expect(podExec.RecordedCommands).NotTo(ContainElement(gpstopCommand))
			})
		})
		When("patch fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("patch", "greenplumclusters", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					a := action.(testing.PatchAction)
					patchBytes := a.GetPatch()
					if strings.Contains(string(patchBytes), "finalizers") {
						return true, nil, errors.New("failed to patch finalizers")
					}
					return true, nil, nil
				})
			})
			It("returns an error (and re-queues)", func() {
				Expect(reconcileErr).To(MatchError("removing finalizer: failed to patch finalizers"))
			})
		})
		When("patch fails when removing because cluster does not exist", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("patch", "greenplumclusters", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					a := action.(testing.PatchAction)
					patchBytes := a.GetPatch()
					if strings.Contains(string(patchBytes), "finalizers") {
						return true, nil, apierrs.NewNotFound(a.GetResource().GroupResource(), a.GetName())
					}
					return true, nil, nil
				})
			})
			It("does not return an error and logs", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
				logs, err := DecodeLogs(bytes.NewReader(logBuf.Contents()))
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("attempted to remove finalizer, but GreenplumCluster was not found")}))
			})
		})
		When("the cluster was created by a previous-version operator", func() {
			BeforeEach(func() {
				greenplumCluster.Status.InstanceImage = "old-image"
			})
			It("Succeeds", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
			})
			It("runs gpstop", func() {
				Expect(podExec.RecordedCommands).To(ContainElement(gpstopCommand))
			})
			It("removes the finalizer", func() {
				var reconciledCluster greenplumv1.GreenplumCluster
				Expect(reactiveClient.Get(ctx, greenplumClusterRequest.NamespacedName, &reconciledCluster)).To(Succeed())
				Expect(reconciledCluster.Finalizers).ShouldNot(ContainElement(greenplumcluster.StopClusterFinalizer))
			})
		})
	})
})
