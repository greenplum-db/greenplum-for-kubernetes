package greenplumcluster_test

import (
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gstruct"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/configmap"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor/fake"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/gpexpandjob"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Reconcile expansion", func() {
	var (
		logBuf              *gbytes.Buffer
		greenplumReconciler *greenplumcluster.GreenplumClusterReconciler
		podExec             *fake.PodExec
		ctx                 context.Context
	)
	BeforeEach(func() {
		logBuf = gbytes.NewBuffer()
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)

		podExec = &fake.PodExec{}
		greenplumReconciler = &greenplumcluster.GreenplumClusterReconciler{
			Client:        reactiveClient,
			Log:           gplog.ForTest(logBuf),
			SSHCreator:    fakeSecretCreator{},
			InstanceImage: "greenplum-for-kubernetes:latest",
			OperatorImage: "greenplum-operator:latest",
			PodExec:       podExec,
		}
	})

	var (
		firstGreenplumClusterSpec *greenplumv1.GreenplumCluster
		newGreenplumClusterSpec   *greenplumv1.GreenplumCluster
	)
	BeforeEach(func() {
		By("starting with a 5 segment cluster")
		firstGreenplumClusterSpec = exampleGreenplumCluster.DeepCopy()
		firstGreenplumClusterSpec.Spec.Segments.PrimarySegmentCount = 5

		podExec.ErrorMsgOnMaster0 = "not active"
		podExec.ErrorMsgOnMaster1 = "not active"
		Expect(reactiveClient.Create(nil, firstGreenplumClusterSpec)).To(Succeed())
		_, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
		Expect(err).NotTo(HaveOccurred())

		podExec.ErrorMsgOnMaster0 = ""
		podExec.ErrorMsgOnMaster1 = ""
		podExec.SegmentCount = "5\n"
	})

	When("gpexpand-job does not exist", func() {
		BeforeEach(func() {
			// Sanity check
			var gpexpandJob batchv1.Job
			jobKey := types.NamespacedName{
				Namespace: namespaceName,
				Name:      clusterName + "-gpexpand-job",
			}
			Expect(reactiveClient.Get(nil, jobKey, &gpexpandJob)).To(MatchError(`jobs.batch "my-greenplum-gpexpand-job" not found`))
		})
		When("gpdb cluster size is increased to 6", func() {
			var (
				reconcileErr error
			)
			JustBeforeEach(func() {
				newGreenplumClusterSpec = firstGreenplumClusterSpec.DeepCopy()
				newGreenplumClusterSpec.Spec.Segments.PrimarySegmentCount = 6

				Expect(reactiveClient.Update(nil, newGreenplumClusterSpec)).To(Succeed())
				_, reconcileErr = greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
			})
			It("Creates a job to run gpexpand", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
				var gpexpandJob batchv1.Job
				jobKey := types.NamespacedName{
					Namespace: namespaceName,
					Name:      clusterName + "-gpexpand-job",
				}
				Expect(reactiveClient.Get(nil, jobKey, &gpexpandJob)).To(Succeed())
				gpexpandContainer := gpexpandJob.Spec.Template.Spec.Containers[0]
				By("setting GPEXPAND_HOST to master-0")
				Expect(gpexpandContainer.Image).To(Equal("greenplum-for-kubernetes:latest"))
				Expect(gpexpandContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:      "GPEXPAND_HOST",
					Value:     "master-0.agent.test-ns.svc.cluster.local",
					ValueFrom: nil,
				}))
				By("setting NEW_SEG_COUNT to 6")
				Expect(gpexpandContainer.Env).To(ContainElement(corev1.EnvVar{
					Name:      "NEW_SEG_COUNT",
					Value:     "6",
					ValueFrom: nil,
				}))
				By("setting a controller reference")
				Expect(gpexpandJob.GetOwnerReferences()).To(ConsistOf(beOwnedByGreenplum))
			})
			It("increases the number of replicas in the segment statefulsets", func() {
				var segmentA appsv1.StatefulSet
				segmentAKey := types.NamespacedName{Namespace: namespaceName, Name: "segment-a"}
				Expect(reactiveClient.Get(nil, segmentAKey, &segmentA)).To(Succeed())
				Expect(segmentA.Spec.Replicas).To(PointTo(BeNumerically("==", 6)))
			})
			It("increases the number of segments in the configmap", func() {
				var cm corev1.ConfigMap
				cmKey := types.NamespacedName{Namespace: namespaceName, Name: "greenplum-config"}
				Expect(reactiveClient.Get(nil, cmKey, &cm)).To(Succeed())
				Expect(cm.Data[configmap.SegmentCount]).To(Equal("6"))
			})
			When("there is an error creating the job", func() {
				BeforeEach(func() {
					reactiveClient.PrependReactor("create", "jobs", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New("failed to create job")
					})
				})
				It("returns an error", func() {
					Expect(reconcileErr).To(MatchError("unable to run gpexpand: failed to create job"))
				})
			})
			When("there is an error getting the segment count from the master", func() {
				BeforeEach(func() {
					podExec.SegmentCountErr = errors.New("psql is borken")
				})
				It("returns an error", func() {
					Expect(reconcileErr).To(MatchError("unable to run gpexpand: psql is borken"))
				})
			})
			When("the segment count is not a number", func() {
				BeforeEach(func() {
					podExec.SegmentCount = "I'm a free man\n"
				})
				It("returns an error", func() {
					Expect(reconcileErr).To(MatchError(`unable to run gpexpand: strconv.ParseInt: parsing "I'm a free man": invalid syntax`))
				})
			})
		})

		When("gpdb cluster size is unchanged", func() {
			var reconcileErr error
			BeforeEach(func() {
				podExec.ErrorMsgOnMaster0 = ""
				podExec.ErrorMsgOnMaster1 = ""
				podExec.SegmentCount = "5\n"
				_, reconcileErr = greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
			})
			It("does not create a new job", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
				var gpexpandJob batchv1.Job
				jobKey := types.NamespacedName{
					Namespace: namespaceName,
					Name:      clusterName + "-gpexpand-job",
				}
				Expect(reactiveClient.Get(nil, jobKey, &gpexpandJob)).To(MatchError(`jobs.batch "my-greenplum-gpexpand-job" not found`))
			})
		})

		When("there's no active master", func() {
			var reconcileErr error
			BeforeEach(func() {
				podExec.ErrorMsgOnMaster0 = "not active"
				podExec.ErrorMsgOnMaster1 = "not active"
				podExec.SegmentCount = ""
				_, reconcileErr = greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)
			})
			It("does not create a new job", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
				var gpexpandJob batchv1.Job
				jobKey := types.NamespacedName{
					Namespace: namespaceName,
					Name:      clusterName + "-gpexpand-job",
				}
				Expect(reactiveClient.Get(nil, jobKey, &gpexpandJob)).To(MatchError(`jobs.batch "my-greenplum-gpexpand-job" not found`))
			})
		})
	})

	When("gpexpand-job exists", func() {
		var existingJob batchv1.Job
		var sawCreate bool
		var sawDelete bool
		BeforeEach(func() {
			existingJob = gpexpandjob.GenerateJob(greenplumReconciler.InstanceImage, "master-0", 5)
			existingJob.Namespace = firstGreenplumClusterSpec.Namespace
			existingJob.Name = fmt.Sprintf("%s-gpexpand-job", firstGreenplumClusterSpec.Name)
		})
		JustBeforeEach(func() {
			Expect(reactiveClient.Create(nil, &existingJob)).To(Succeed())
		})
		When("the job has succeeded", func() {
			BeforeEach(func() {
				By("the job succeeding")
				existingJob.Status.Succeeded = 1
			})
			JustBeforeEach(func() {
				reactiveClient.PrependReactor("delete", "jobs", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					deleteAction := action.(testing.DeleteAction)
					if deleteAction.GetNamespace() == existingJob.Namespace &&
						deleteAction.GetName() == existingJob.Name {
						sawDelete = true
					}
					return false, nil, nil
				})

				reactiveClient.PrependReactor("create", "jobs", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					createAction := action.(testing.CreateAction)
					job := createAction.GetObject().(*batchv1.Job)
					if createAction.GetNamespace() == existingJob.Namespace &&
						job.Name == existingJob.Name {
						sawCreate = true
					}
					return false, nil, nil
				})
				newGreenplumClusterSpec = firstGreenplumClusterSpec.DeepCopy()
				newGreenplumClusterSpec.Spec.Segments.PrimarySegmentCount = 6

			})
			It("deletes the old job", func() {
				Expect(reactiveClient.Update(nil, newGreenplumClusterSpec)).To(Succeed())
				Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))

				Expect(sawDelete).To(BeTrue())
			})
			It("creates a new job", func() {
				Expect(reactiveClient.Update(nil, newGreenplumClusterSpec)).To(Succeed())
				Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))

				Expect(sawCreate).To(BeTrue())
			})
			When("getting the job fails", func() {
				JustBeforeEach(func() {
					reactiveClient.PrependReactor("get", "jobs", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						getAction := action.(testing.GetAction)
						if getAction.GetNamespace() == existingJob.Namespace &&
							getAction.GetName() == existingJob.Name {
							return true, nil, errors.New("I failed to get a job")
						}
						return false, nil, nil
					})
				})
				It("returns an error", func() {
					Expect(reactiveClient.Update(nil, newGreenplumClusterSpec)).To(Succeed())
					_, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)

					Expect(err).To(MatchError("unable to run gpexpand: I failed to get a job"))
				})
			})
			When("deleting the job fails", func() {
				JustBeforeEach(func() {
					reactiveClient.PrependReactor("delete", "jobs", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						deleteAction := action.(testing.DeleteAction)
						if deleteAction.GetNamespace() == existingJob.Namespace &&
							deleteAction.GetName() == existingJob.Name {
							return true, nil, errors.New("delete failure")
						}
						return false, nil, nil
					})
				})
				It("returns an error", func() {
					Expect(reactiveClient.Update(nil, newGreenplumClusterSpec)).To(Succeed())
					_, err := greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)

					Expect(err).To(MatchError("unable to run gpexpand: delete failure"))
				})
			})
		})
		When("the job has not yet succeeded", func() {
			var sawCreate bool
			JustBeforeEach(func() {
				reactiveClient.PrependReactor("create", "jobs", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					createAction := action.(testing.CreateAction)
					job := createAction.GetObject().(*batchv1.Job)
					if createAction.GetNamespace() == existingJob.Namespace &&
						job.Name == existingJob.Name {
						sawCreate = true
					}
					return false, nil, nil
				})
				newGreenplumClusterSpec = firstGreenplumClusterSpec.DeepCopy()
				newGreenplumClusterSpec.Spec.Segments.PrimarySegmentCount = 6
			})
			It("does nothing", func() {
				Expect(reactiveClient.Update(nil, newGreenplumClusterSpec)).To(Succeed())
				Expect(greenplumReconciler.Reconcile(ctx, greenplumClusterRequest)).To(Equal(ctrl.Result{}))

				Expect(sawCreate).To(BeFalse(), "should not create a job")
			})
		})
	})
})
