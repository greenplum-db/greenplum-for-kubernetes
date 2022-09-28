package admission

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MaxLabelLen = 63

func validateWorkerSelector(workerSelector map[string]string, typ string) (result *metav1.Status) {
	for k, v := range workerSelector {
		if len(k) > MaxLabelLen || len(v) > MaxLabelLen {
			result = &metav1.Status{Message: fmt.Sprintf("%s workerSelector key/value is longer than %d characters", typ, MaxLabelLen)}
			return
		}
	}
	return
}

func validateResourceQuantity(quantity resource.Quantity, typ, field string) (result *metav1.Status) {
	if quantity.Sign() == -1 {
		result = &metav1.Status{Message: fmt.Sprintf(`invalid %s %s value: "%s": must be greater than or equal to 0`, typ, field, quantity.String())}
	}
	return
}

func (h *Handler) validateStorageHelper(pvcList *corev1.PersistentVolumeClaimList, newStorage resource.Quantity, newStorageClassName, parentObjectType string) (result *metav1.Status) {
	if len(pvcList.Items) > 0 {
		pvc := &pvcList.Items[0]
		pvcStorage := pvc.Spec.Resources.Limits[corev1.ResourceStorage]
		if pvcStorage.Cmp(newStorage) != 0 {
			result = &metav1.Status{Message: generateShortPVCErrStr("storage", "changed", parentObjectType)}
			return
		}

		if *pvc.Spec.StorageClassName != newStorageClassName {
			result = &metav1.Status{Message: generateShortPVCErrStr("storageClassName", "changed", parentObjectType)}
			return
		}
	}
	return
}

const (
	pvcInfoFmt       = "%s has PVCs for %d %s."
	pvcErrFmt        = "%s cannot be %s without first deleting PVCs. This will result in a new, empty %s cluster"
	pvcVersionErrFmt = "the existing PVCs for my-gp-instance are not compatible with this controller. Expected PVCs to have greenplum-major-version=%s; found %s"
)

func generateLongPVCErrStr(parentObjectType, parentObject string, previousCount int, childObjectName, childObjectPath, verb string) string {
	info := fmt.Sprintf(pvcInfoFmt, parentObject, previousCount, childObjectName)
	err := generateShortPVCErrStr(childObjectPath, verb, parentObjectType)
	return info + " " + err
}

func generateShortPVCErrStr(childObjectPath, verb, parentObjectType string) string {
	return fmt.Sprintf(pvcErrFmt, childObjectPath, verb, parentObjectType)
}

func (h *Handler) getPVCs(ctx context.Context, namespace string, labelMatcher client.MatchingLabels) (*corev1.PersistentVolumeClaimList, error) {
	var pvcList corev1.PersistentVolumeClaimList
	err := h.KubeClient.List(ctx, &pvcList, labelMatcher, client.InNamespace(namespace))
	if err != nil {
		return nil, err
	}
	return &pvcList, nil
}
