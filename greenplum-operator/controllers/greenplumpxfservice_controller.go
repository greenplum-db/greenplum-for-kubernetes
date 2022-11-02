/*
.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	greenplumv1beta1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1beta1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/pxf"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// GreenplumPXFServiceReconciler reconciles a GreenplumPXFService object
type GreenplumPXFServiceReconciler struct {
	client.Client
	Log           logr.Logger
	InstanceImage string
}

var _ client.Client = &GreenplumPXFServiceReconciler{}

// +kubebuilder:rbac:groups=greenplum.pivotal.io,resources=greenplumpxfservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=greenplum.pivotal.io,resources=greenplumpxfservices/status,verbs=get;update;patch

func (r *GreenplumPXFServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("greenplumpxfservice", req.NamespacedName)

	// check image version
	var pxfDeployment appsv1.Deployment
	err := r.Get(ctx, req.NamespacedName, &pxfDeployment)
	if err == nil {
		currentImage := pxfDeployment.Spec.Template.Spec.Containers[0].Image
		if currentImage != r.InstanceImage {
			return ctrl.Result{}, nil
		}
	} else if !apierrs.IsNotFound(err) {
		return ctrl.Result{}, errors.Wrap(err, "unable to fetch PXF Deployment")
	}

	// GreenplumPXFService
	var greenplumPXF greenplumv1beta1.GreenplumPXFService
	if err := r.Get(ctx, req.NamespacedName, &greenplumPXF); err != nil {
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "unable to fetch GreenplumPXFService")
	}

	// PXF Service
	var pxfService corev1.Service
	pxfService.Name = greenplumPXF.Name
	pxfService.Namespace = greenplumPXF.Namespace
	result, err := ctrl.CreateOrUpdate(ctx, r, &pxfService, func() error {
		pxf.ModifyService(greenplumPXF, &pxfService)
		return controllerutil.SetControllerReference(&greenplumPXF, &pxfService, r.Scheme())
	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to CreateOrUpdate PXF Service")
	}
	if result != controllerutil.OperationResultNone {
		log.Info("PXF Service " + string(result))
	}

	// PXF Deployment
	pxfDeployment = appsv1.Deployment{}
	pxfDeployment.Name = greenplumPXF.Name
	pxfDeployment.Namespace = greenplumPXF.Namespace
	result, err = ctrl.CreateOrUpdate(ctx, r, &pxfDeployment, func() error {
		pxf.ModifyDeployment(greenplumPXF, &pxfDeployment, r.InstanceImage)
		return controllerutil.SetControllerReference(&greenplumPXF, &pxfDeployment, r.Scheme())
	})
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "unable to CreateOrUpdate PXF Deployment")
	}
	if result != controllerutil.OperationResultNone {
		log.Info("PXF Deployment " + string(result))
	}

	// update status
	newPXF := greenplumPXF.DeepCopy()
	desiredReplicas := int32(greenplumPXF.Spec.Replicas)
	readyReplicas := pxfDeployment.Status.ReadyReplicas
	unavailableReplicas := pxfDeployment.Status.UnavailableReplicas
	updatedReplicas := pxfDeployment.Status.UpdatedReplicas
	if readyReplicas == 0 {
		newPXF.Status.Phase = greenplumv1beta1.GreenplumPXFServicePhasePending
	} else if unavailableReplicas != 0 || updatedReplicas < desiredReplicas {
		newPXF.Status.Phase = greenplumv1beta1.GreenplumPXFServicePhaseDegraded
	} else {
		newPXF.Status.Phase = greenplumv1beta1.GreenplumPXFServicePhaseRunning
	}
	if newPXF.Status.Phase != greenplumPXF.Status.Phase {
		err = r.Patch(ctx, newPXF, client.MergeFrom(&greenplumPXF))
		if err != nil {
			log.Error(err, "update failed")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *GreenplumPXFServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&greenplumv1beta1.GreenplumPXFService{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
