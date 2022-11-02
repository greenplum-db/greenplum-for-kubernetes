/*
.
*/

package greenplumcluster

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/configmap"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/service"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/serviceaccount"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/sset"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/sshkeygen"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	StopClusterFinalizer           = "stopcluster.greenplumcluster.pivotal.io"
	SupportedGreenplumMajorVersion = "6"
)

// GreenplumClusterReconciler reconciles a GreenplumCluster object
type GreenplumClusterReconciler struct {
	client.Client
	Log           logr.Logger
	SSHCreator    sshkeygen.SSHSecretCreator
	InstanceImage string
	OperatorImage string
	PodExec       executor.PodExecInterface
}

var _ client.Client = &GreenplumClusterReconciler{}

// +kubebuilder:rbac:groups=greenplum.pivotal.io,resources=greenplumclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=greenplum.pivotal.io,resources=greenplumclusters/status,verbs=get;update;patch

func (r *GreenplumClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&greenplumv1.GreenplumCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}

func (r *GreenplumClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("greenplumcluster", req.NamespacedName)

	// GreenplumCluster
	var greenplumCluster greenplumv1.GreenplumCluster
	if err := r.Get(ctx, req.NamespacedName, &greenplumCluster); err != nil {
		if apierrs.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to fetch GreenplumCluster: %w", err)
	}
	SetDefaultGreenplumClusterValues(&greenplumCluster)

	activeMaster := executor.GetCurrentActiveMaster(r.PodExec, greenplumCluster.Namespace)
	log.V(1).Info("current active master", "activeMaster", activeMaster)

	if err := r.handleFinalizer(ctx, &greenplumCluster, &activeMaster); err != nil {
		return ctrl.Result{}, err
	}

	if greenplumCluster.Status.InstanceImage != "" &&
		greenplumCluster.Status.InstanceImage != r.InstanceImage {
		return ctrl.Result{}, nil
	}

	clusterExists, err := r.clusterExists(ctx, greenplumCluster)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to check if GreenplumCluster resources exist: %w", err)
	}
	if !clusterExists {
		if err := handleAntiAffinity(ctx, r, greenplumCluster); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.createOrUpdateClusterResources(ctx, greenplumCluster); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileStatus(ctx, &greenplumCluster); err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Decide when to set status to greenplumv1.GreenplumClusterPhaseFailed

	if greenplumCluster.Status.Phase == greenplumv1.GreenplumClusterPhasePending && activeMaster != "" {
		r.setStatus(ctx, &greenplumCluster, greenplumv1.GreenplumClusterPhaseRunning)
	}

	if activeMaster == "" {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if err := r.handleExpand(ctx, &greenplumCluster, activeMaster); err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to run gpexpand: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *GreenplumClusterReconciler) createOrUpdateClusterResources(ctx context.Context, greenplumCluster greenplumv1.GreenplumCluster) error {
	ns := greenplumCluster.Namespace
	gpName := greenplumCluster.Name

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "greenplum-config",
			Namespace: ns,
		},
	}

	operationResult, err := ctrl.CreateOrUpdate(ctx, r, configMap, func() error {
		configmap.ModifyConfigMap(&greenplumCluster, configMap)
		return controllerutil.SetControllerReference(&greenplumCluster, configMap, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, configMap)

	sshSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-secrets",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, sshSecret, func() error {
		var keyData map[string][]byte
		if sshSecret.Data == nil {
			keyData, err = r.SSHCreator.GenerateKey()
			if err != nil {
				return err
			}
		}
		sshkeygen.ModifySecret(gpName, sshSecret, keyData)
		return ctrl.SetControllerReference(&greenplumCluster, sshSecret, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, sshSecret)

	agentService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, agentService, func() error {
		service.ModifyGreenplumAgentService(gpName, agentService)
		return ctrl.SetControllerReference(&greenplumCluster, agentService, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, agentService)

	greenplumService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "greenplum",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, greenplumService, func() error {
		service.ModifyGreenplumService(gpName, greenplumService)
		return ctrl.SetControllerReference(&greenplumCluster, greenplumService, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, greenplumService)

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "greenplum-system-pod",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, serviceAccount, func() error {
		return ctrl.SetControllerReference(&greenplumCluster, serviceAccount, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, serviceAccount)

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "greenplum-system-pod",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, role, func() error {
		if err := serviceaccount.ModifyRole(role); err != nil {
			return err
		}
		return ctrl.SetControllerReference(&greenplumCluster, role, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, role)

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "greenplum-system-pod",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, roleBinding, func() error {
		serviceaccount.ModifyRoleBinding(roleBinding)
		return ctrl.SetControllerReference(&greenplumCluster, roleBinding, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, roleBinding)

	masterStatefulSetParams := sset.GenerateStatefulSetParams(sset.TypeMaster, &greenplumCluster, r.InstanceImage)
	masterStatefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "master",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, masterStatefulSet, func() error {
		sset.ModifyGreenplumStatefulSet(masterStatefulSetParams, masterStatefulSet)
		return ctrl.SetControllerReference(&greenplumCluster, masterStatefulSet, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, masterStatefulSet)

	primaryStatefulSetParams := sset.GenerateStatefulSetParams(sset.TypeSegmentA, &greenplumCluster, r.InstanceImage)
	primaryStatefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "segment-a",
			Namespace: ns,
		},
	}
	operationResult, err = ctrl.CreateOrUpdate(ctx, r, primaryStatefulSet, func() error {
		sset.ModifyGreenplumStatefulSet(primaryStatefulSetParams, primaryStatefulSet)
		return controllerutil.SetControllerReference(&greenplumCluster, primaryStatefulSet, r.Scheme())
	})
	if err != nil {
		return err
	}
	r.logReconcileResult(operationResult, primaryStatefulSet)

	if greenplumCluster.Spec.Segments.Mirrors == "yes" {
		mirrorStatefulSetParams := sset.GenerateStatefulSetParams(sset.TypeSegmentB, &greenplumCluster, r.InstanceImage)
		mirrorStatefulSet := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "segment-b",
				Namespace: ns,
			},
		}
		operationResult, err = ctrl.CreateOrUpdate(ctx, r, mirrorStatefulSet, func() error {
			sset.ModifyGreenplumStatefulSet(mirrorStatefulSetParams, mirrorStatefulSet)
			return ctrl.SetControllerReference(&greenplumCluster, mirrorStatefulSet, r.Scheme())
		})
		if err != nil {
			return err
		}
		r.logReconcileResult(operationResult, mirrorStatefulSet)
	}

	return nil
}

func (r *GreenplumClusterReconciler) logReconcileResult(operationResult controllerutil.OperationResult, obj runtime.Object) {
	if operationResult == controllerutil.OperationResultNone {
		return
	}
	gvk, err := apiutil.GVKForObject(obj, r.Scheme())
	if err != nil {
		panic(err)
	}
	mo, err := meta.Accessor(obj)
	if err != nil {
		panic(err)
	}
	r.Log.Info("reconciled",
		"op", operationResult,
		"kind", gvk.Kind,
		"name", mo.GetName(),
		"namespace", mo.GetNamespace())
}

// deprecated: This function is only used as a crutch for antiaffinity v1. It will go away soon (hopefully)
func (r *GreenplumClusterReconciler) clusterExists(ctx context.Context, greenplumCluster greenplumv1.GreenplumCluster) (bool, error) {
	var ssetList appsv1.StatefulSetList
	labelMatcher := client.MatchingLabels{"greenplum-cluster": greenplumCluster.Name}
	err := r.List(ctx, &ssetList, labelMatcher, client.InNamespace(greenplumCluster.Namespace))
	if err != nil {
		return false, err
	}
	return len(ssetList.Items) > 0, nil
}
