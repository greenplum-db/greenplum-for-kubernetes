package controllers

import (
	"bytes"
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1beta1"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/gplog"
	. "github.com/pivotal/greenplum-for-kubernetes/pkg/gplog/testing"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("PXF controller", func() {
	var (
		ctx           context.Context
		logBuf        *gbytes.Buffer
		pxfReconciler *GreenplumPXFServiceReconciler
		pxf           *v1beta1.GreenplumPXFService
		examplePxf    = &v1beta1.GreenplumPXFService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pxf",
				Namespace: "test-ns",
			},
			Spec: v1beta1.GreenplumPXFServiceSpec{
				Replicas: 2,
				CPU:      resource.MustParse("2.0"),
				Memory:   resource.MustParse("1.5Gi"),
			},
		}
		pxfRequest = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: examplePxf.Namespace,
				Name:      examplePxf.Name,
			},
		}
		myPxfKey = pxfRequest.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), struct{ key string }{"test"}, CurrentGinkgoTestDescription().TestText)
		logBuf = gbytes.NewBuffer()

		pxfReconciler = &GreenplumPXFServiceReconciler{
			Client:        reactiveClient,
			Log:           gplog.ForTest(logBuf),
			InstanceImage: "greenplum-for-kubernetes:v1.7.5",
		}

		pxf = examplePxf.DeepCopy()
	})

	When("creating GreenplumPXFService", func() {
		When("reconciliation succeeds", func() {
			It("Creates a Deployment and Service when none exist", func() {
				Expect(reactiveClient.Create(ctx, pxf)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())

				var service corev1.Service
				Expect(reactiveClient.Get(ctx, myPxfKey, &service)).To(Succeed())
				serviceRefs := service.GetOwnerReferences()
				Expect(serviceRefs).To(HaveLen(1))
				Expect(serviceRefs[0].Name).To(Equal("my-pxf"))
				Expect(serviceRefs[0].Kind).To(Equal("GreenplumPXFService"))

				var deployment appsv1.Deployment
				Expect(reactiveClient.Get(ctx, myPxfKey, &deployment)).To(Succeed())
				deploymentRefs := deployment.GetOwnerReferences()
				Expect(deploymentRefs).To(HaveLen(1))
				Expect(deploymentRefs[0].Name).To(Equal("my-pxf"))
				Expect(deploymentRefs[0].Kind).To(Equal("GreenplumPXFService"))
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal("greenplum-for-kubernetes:v1.7.5"))

				logs, err := DecodeLogs(bytes.NewReader(logBuf.Contents()))
				Expect(err).NotTo(HaveOccurred())
				Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("PXF Service created")}))
				Expect(logs).To(ContainLogEntry(gstruct.Keys{"msg": Equal("PXF Deployment created")}))
			})
		})

		When("getting pxf fails because it doesn't exist", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("get", "greenplumpxfservices", func(action testing.Action) (bool, runtime.Object, error) {
					getAction := action.(testing.GetAction)
					return true, nil, apierrs.NewNotFound(getAction.GetResource().GroupResource(), getAction.GetName())
				})
			})
			It("does not return an error (do not requeue)", func() {
				Expect(reactiveClient.Create(ctx, pxf)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("getting the pxf fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("get", "greenplumpxfservices", func(action testing.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("injected error")
				})
			})
			It("returns the error", func() {
				Expect(reactiveClient.Create(ctx, pxf)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).To(MatchError("unable to fetch GreenplumPXFService: injected error"))
			})
		})

		When("creating service fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("create", "services", func(action testing.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("injected error")
				})
			})
			It("returns the error", func() {
				Expect(reactiveClient.Create(ctx, pxf)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).To(MatchError("unable to CreateOrUpdate PXF Service: injected error"))
			})
		})

		When("creating deployment fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("create", "deployments", func(action testing.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("injected error")
				})
			})
			It("returns the error", func() {
				Expect(reactiveClient.Create(ctx, pxf)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).To(MatchError("unable to CreateOrUpdate PXF Deployment: injected error"))
			})
		})
	})

	When("updating GreenplumPXFService", func() {
		BeforeEach(func() {
			Expect(reactiveClient.Create(ctx, pxf)).To(Succeed())
			_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
			Expect(err).NotTo(HaveOccurred())
			Expect(logBuf).To(gbytes.Say("PXF Service created"))
			Expect(reactiveClient.Get(ctx, myPxfKey, &corev1.Service{})).To(Succeed())
			Expect(logBuf).To(gbytes.Say("PXF Deployment created"))
			Expect(reactiveClient.Get(ctx, myPxfKey, &appsv1.Deployment{})).To(Succeed())
		})

		When("reconciliation succeeds", func() {
			FIt("updates attributes", func() {
				pxf.Spec.Replicas = 99
				pxf.Spec.CPU = resource.MustParse("42")
				pxf.Spec.Memory = resource.MustParse("36Gi")
				pxf.Spec.WorkerSelector = map[string]string{
					"label": "value",
				}

				Expect(reactiveClient.Update(ctx, pxf)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())

				var deployment appsv1.Deployment
				Expect(reactiveClient.Get(ctx, myPxfKey, &deployment)).To(Succeed())

				Expect(deployment.Spec.Replicas).To(gstruct.PointTo(Equal(int32(99))))
				Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
				container0 := deployment.Spec.Template.Spec.Containers[0]
				Expect(container0.Resources.Limits[corev1.ResourceCPU]).To(Equal(resource.MustParse("42")))
				Expect(container0.Resources.Limits[corev1.ResourceMemory]).To(Equal(resource.MustParse("36Gi")))
				Expect(deployment.Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("label", "value"))
				deploymentRefs := deployment.GetOwnerReferences()
				Expect(deploymentRefs).To(HaveLen(1))
				Expect(deploymentRefs[0].Name).To(Equal("my-pxf"))
				Expect(deploymentRefs[0].Kind).To(Equal("GreenplumPXFService"))
			})
		})

		When("updating deployment fails", func() {
			BeforeEach(func() {
				reactiveClient.PrependReactor("update", "deployments", func(action testing.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("injected error")
				})
			})
			It("returns the error", func() {
				pxf.Spec.Replicas = 99
				Expect(reactiveClient.Update(ctx, pxf)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).To(MatchError("unable to CreateOrUpdate PXF Deployment: injected error"))
			})
		})

		When("the deployment was created by a previous version operator", func() {
			var deploymentUpdated bool
			BeforeEach(func() {
				pxfReconciler.InstanceImage = "greenplum-for-kubernetes:new-version"

				reactiveClient.PrependReactor("update", "deployments", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					deploymentUpdated = true
					return false, nil, nil
				})
			})
			It("skips reconcile", func() {
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())
				Expect(deploymentUpdated).To(BeFalse(), "should not update deployment")
				var deployment appsv1.Deployment
				Expect(reactiveClient.Get(ctx, myPxfKey, &deployment)).To(Succeed())
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).To(Equal("greenplum-for-kubernetes:v1.7.5"))
			})
		})
	})

	Context("Status", func() {
		BeforeEach(func() {
			Expect(reactiveClient.Create(ctx, pxf)).To(Succeed())
			_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
			Expect(err).NotTo(HaveOccurred())
		})
		When("Deployment readyReplics = 0", func() {
			It("sets status to Pending", func() {
				var resultGreenplumPXF v1beta1.GreenplumPXFService
				Expect(reactiveClient.Get(ctx, myPxfKey, &resultGreenplumPXF)).To(Succeed())
				Expect(resultGreenplumPXF.Status.Phase).To(Equal(v1beta1.GreenplumPXFServicePhasePending))
			})
		})
		When("Deployment readyReplicas > 0 and unavailableReplicas > 0", func() {
			BeforeEach(func() {
				var pxfDeployment appsv1.Deployment
				Expect(reactiveClient.Get(ctx, myPxfKey, &pxfDeployment)).To(Succeed())
				pxfDeployment.Status.ReadyReplicas = int32(pxf.Spec.Replicas) - 1
				pxfDeployment.Status.UnavailableReplicas = 1
				Expect(reactiveClient.Update(ctx, &pxfDeployment)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())
			})
			It("sets status to Degraded", func() {
				var resultGreenplumPXF v1beta1.GreenplumPXFService
				Expect(reactiveClient.Get(ctx, myPxfKey, &resultGreenplumPXF)).To(Succeed())
				Expect(resultGreenplumPXF.Status.Phase).To(Equal(v1beta1.GreenplumPXFServicePhaseDegraded))
			})
		})
		When("Deployment readyReplicas > 0 and updatedReplicas < PXF desired replicas", func() {
			BeforeEach(func() {
				var pxfDeployment appsv1.Deployment
				Expect(reactiveClient.Get(ctx, myPxfKey, &pxfDeployment)).To(Succeed())
				pxfDeployment.Status.ReadyReplicas = int32(pxf.Spec.Replicas) - 1
				pxfDeployment.Status.UpdatedReplicas = int32(pxf.Spec.Replicas) - 1
				Expect(reactiveClient.Update(ctx, &pxfDeployment)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())
			})
			It("sets status to Degraded", func() {
				var resultGreenplumPXF v1beta1.GreenplumPXFService
				Expect(reactiveClient.Get(ctx, myPxfKey, &resultGreenplumPXF)).To(Succeed())
				Expect(resultGreenplumPXF.Status.Phase).To(Equal(v1beta1.GreenplumPXFServicePhaseDegraded))
			})
		})
		When("Deployment readyReplicas > 0, unavailableReplicas = 0, and updatedReplicas = PXF desired replicas", func() {
			BeforeEach(func() {
				var pxfDeployment appsv1.Deployment
				Expect(reactiveClient.Get(ctx, myPxfKey, &pxfDeployment)).To(Succeed())
				pxfDeployment.Status.ReadyReplicas = int32(pxf.Spec.Replicas)
				pxfDeployment.Status.UpdatedReplicas = int32(pxf.Spec.Replicas)
				Expect(reactiveClient.Update(ctx, &pxfDeployment)).To(Succeed())
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())
			})
			It("sets status to Running", func() {
				var resultGreenplumPXF v1beta1.GreenplumPXFService
				Expect(reactiveClient.Get(ctx, myPxfKey, &resultGreenplumPXF)).To(Succeed())
				Expect(resultGreenplumPXF.Status.Phase).To(Equal(v1beta1.GreenplumPXFServicePhaseRunning))
			})
		})
		When("there is no need for a status change", func() {
			var patchCalled bool
			BeforeEach(func() {
				reactiveClient.PrependReactor("patch", "greenplumpxfservices", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					patchCalled = true
					return false, nil, nil
				})
				_, err := pxfReconciler.Reconcile(ctx, pxfRequest)
				Expect(err).NotTo(HaveOccurred())
			})
			It("does not update status", func() {
				Expect(patchCalled).To(BeFalse())
			})
		})
	})
})
