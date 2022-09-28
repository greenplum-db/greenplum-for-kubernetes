package admission

import (
	"context"
	"fmt"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/controllers/greenplumcluster"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) validateCreateGreenplumCluster(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster) (allowed bool, result *metav1.Status) {
	greenplumcluster.SetDefaultGreenplumClusterValues(&newGreenplum)

	result = h.validateAntiAffinity(newGreenplum)
	if result != nil {
		return
	}
	result = h.validateUniqueCluster(ctx, newGreenplum)
	if result != nil {
		return
	}
	result = h.validateGreenplumStorageFromPVCs(ctx, newGreenplum)
	if result != nil {
		return
	}
	result = h.validatePvcGreenplumVersion(ctx, newGreenplum, "master")
	if result != nil {
		return
	}
	result = h.validatePvcGreenplumVersion(ctx, newGreenplum, "segment-a")
	if result != nil {
		return
	}
	result = h.validatePvcGreenplumVersion(ctx, newGreenplum, "segment-b")
	if result != nil {
		return
	}
	result = h.validateStandby(ctx, newGreenplum)
	if result != nil {
		return
	}
	result = h.validatePrimarySegmentCount(ctx, newGreenplum)
	if result != nil {
		return
	}
	result = h.validateMirrors(ctx, newGreenplum)
	if result != nil {
		return
	}

	result = validateWorkerSelector(newGreenplum.Spec.MasterAndStandby.WorkerSelector, "masterAndStandby")
	if result != nil {
		return
	}
	result = validateWorkerSelector(newGreenplum.Spec.Segments.WorkerSelector, "segments")
	if result != nil {
		return
	}

	result = validateResourceQuantity(newGreenplum.Spec.MasterAndStandby.CPU, "masterAndStandby", "cpu")
	if result != nil {
		return
	}
	result = validateResourceQuantity(newGreenplum.Spec.Segments.CPU, "segments", "cpu")
	if result != nil {
		return
	}

	result = validateResourceQuantity(newGreenplum.Spec.MasterAndStandby.Memory, "masterAndStandby", "memory")
	if result != nil {
		return
	}
	result = validateResourceQuantity(newGreenplum.Spec.Segments.Memory, "segments", "memory")
	if result != nil {
		return
	}

	result = validateResourceQuantity(newGreenplum.Spec.MasterAndStandby.Storage, "masterAndStandby", "storage")
	if result != nil {
		return
	}
	result = validateResourceQuantity(newGreenplum.Spec.Segments.Storage, "segments", "storage")
	if result != nil {
		return
	}

	allowed = true
	return
}

func (h *Handler) validateUniqueCluster(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster) (result *metav1.Status) {
	exists, err := h.clusterExistsInNamespace(ctx, newGreenplum)
	if exists || err != nil {
		if err != nil {
			result = &metav1.Status{Message: "could not check if a cluster exists in namespace " + newGreenplum.Namespace + ". " + err.Error()}
			return
		}
		result = &metav1.Status{Message: "only one GreenplumCluster is allowed in namespace " + newGreenplum.Namespace}
		return
	}
	return
}

func (h *Handler) validatePvcGreenplumVersion(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster, typ string) (result *metav1.Status) {
	pvcList, err := h.getGreenplumPVCs(ctx, newGreenplum, typ)
	if err != nil {
		result = &metav1.Status{Message: err.Error()}
		return
	}

	for _, pvc := range pvcList.Items {
		ver, ok := pvc.Labels["greenplum-major-version"]
		if !ok {
			errMsg := fmt.Sprintf(pvcVersionErrFmt, greenplumcluster.SupportedGreenplumMajorVersion, "no label")
			result = &metav1.Status{Message: errMsg}
		} else if ver != greenplumcluster.SupportedGreenplumMajorVersion {
			errMsg := fmt.Sprintf(pvcVersionErrFmt, greenplumcluster.SupportedGreenplumMajorVersion, "greenplum-major-version="+ver)
			result = &metav1.Status{Message: errMsg}
		}
	}

	return
}

func (h *Handler) validatePrimarySegmentCount(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster) (result *metav1.Status) {
	pvcList, err := h.getGreenplumPVCs(ctx, newGreenplum, "segment-a")
	if err != nil {
		result = &metav1.Status{Message: err.Error()}
		return
	}
	previousPrimarySegmentCount := int32(len(pvcList.Items))
	if newGreenplum.Spec.Segments.PrimarySegmentCount < previousPrimarySegmentCount {
		result = &metav1.Status{
			Message: generateLongPVCErrStr("Greenplum", newGreenplum.Name, int(previousPrimarySegmentCount), "segments", "segments.primarySegmentCount", "decreased"),
		}
		return
	}
	return
}

func (h *Handler) validateAntiAffinity(newGreenplum greenplumv1.GreenplumCluster) (result *metav1.Status) {
	if newGreenplum.Spec.MasterAndStandby.Standby == "no" {
		if newGreenplum.Spec.MasterAndStandby.AntiAffinity != "no" || newGreenplum.Spec.Segments.AntiAffinity != "no" {
			result = &metav1.Status{Message: `when standby is set to "no", antiAffinity must also be set to "no"`}
			return
		}
	}
	if newGreenplum.Spec.Segments.Mirrors == "no" {
		if newGreenplum.Spec.MasterAndStandby.AntiAffinity != "no" || newGreenplum.Spec.Segments.AntiAffinity != "no" {
			result = &metav1.Status{Message: `when mirrors is set to "no", antiAffinity must also be set to "no"`}
			return
		}
	}
	return
}

func (h *Handler) validateGreenplumStorageFromPVCs(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster) (result *metav1.Status) {
	masterPVCs, err := h.getGreenplumPVCs(ctx, newGreenplum, "master")
	if err != nil {
		result = &metav1.Status{Message: err.Error()}
		return
	}
	result = h.validateStorageHelper(masterPVCs, newGreenplum.Spec.MasterAndStandby.Storage, newGreenplum.Spec.MasterAndStandby.StorageClassName, "Greenplum")
	if result != nil {
		return
	}

	segmentPVCs, err := h.getGreenplumPVCs(ctx, newGreenplum, "segment-a")
	if err != nil {
		result = &metav1.Status{Message: err.Error()}
		return
	}
	return h.validateStorageHelper(segmentPVCs, newGreenplum.Spec.Segments.Storage, newGreenplum.Spec.Segments.StorageClassName, "Greenplum")
}

func (h *Handler) validateStandby(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster) (result *metav1.Status) {
	pvcList, err := h.getGreenplumPVCs(ctx, newGreenplum, "master")
	if err != nil {
		result = &metav1.Status{Message: err.Error()}
		return
	}
	previousNumMasters := int32(len(pvcList.Items))
	if previousNumMasters == 0 {
		// if previousNumMasters=0, there was no previous cluster, so we will allow it
		return
	}

	previousUsedStandby := previousNumMasters == 2
	var previousStandby string
	if previousUsedStandby {
		previousStandby = "yes"
	} else {
		previousStandby = "no"
	}
	if previousStandby != newGreenplum.Spec.MasterAndStandby.Standby {
		result = &metav1.Status{
			Message: generateLongPVCErrStr("Greenplum", newGreenplum.Name, int(previousNumMasters), "masters", "masterAndStandby.standby", "changed"),
		}
		return
	}
	return
}

func (h *Handler) validateMirrors(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster) (result *metav1.Status) {
	primaryPvcList, err := h.getGreenplumPVCs(ctx, newGreenplum, "segment-a")
	if err != nil {
		result = &metav1.Status{Message: err.Error()}
		return
	}
	previousPrimaryCount := len(primaryPvcList.Items)
	if previousPrimaryCount == 0 { // new cluster
		return
	}

	mirrorPvcList, err := h.getGreenplumPVCs(ctx, newGreenplum, "segment-b")
	if err != nil {
		result = &metav1.Status{Message: err.Error()}
		return
	}
	previousMirrorCount := len(mirrorPvcList.Items)
	if (previousMirrorCount == 0 && newGreenplum.Spec.Segments.Mirrors == "yes") ||
		(previousMirrorCount > 0 && newGreenplum.Spec.Segments.Mirrors == "no") {
		result = &metav1.Status{
			Message: generateLongPVCErrStr("Greenplum", newGreenplum.Name, previousMirrorCount, "mirrors", "segments.mirrors", "changed"),
		}
		return
	}

	return
}

func (h *Handler) getGreenplumPVCs(ctx context.Context, newGreenplum greenplumv1.GreenplumCluster, typ string) (*corev1.PersistentVolumeClaimList, error) {
	labelMatcher := client.MatchingLabels{
		"app":               "greenplum",
		"greenplum-cluster": newGreenplum.Name,
		"type":              typ,
	}
	return h.getPVCs(ctx, newGreenplum.Namespace, labelMatcher)
}

func (h *Handler) clusterExistsInNamespace(ctx context.Context, greenplumCluster greenplumv1.GreenplumCluster) (bool, error) {
	var clusterList greenplumv1.GreenplumClusterList
	err := h.KubeClient.List(ctx, &clusterList, client.InNamespace(greenplumCluster.Namespace))
	if err != nil {
		return false, err
	}
	return len(clusterList.Items) > 0, nil
}
