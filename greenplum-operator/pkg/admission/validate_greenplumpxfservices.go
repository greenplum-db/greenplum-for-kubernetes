package admission

import (
	"context"
	"fmt"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const UpgradePXFHelpMsg = "Cannot update GreenplumPXFService instance -- operator only supports updates to greenplumpxfservices " +
	"at the latest version. Please update GreenplumPXFService to the latest version in order to make updates"

func (h *Handler) validateGreenplumPXFService(ctx context.Context, oldPXF, newPXF *greenplumv1.GreenplumPXFService) (allowed bool, result *metav1.Status) {
	isUpdate := false
	if oldPXF != nil {
		isUpdate = true
	}

	if isUpdate {
		if !equality.Semantic.DeepEqual(oldPXF.Spec, newPXF.Spec) {
			var oldDeployment appsv1.Deployment
			deploymentKey := types.NamespacedName{Namespace: oldPXF.Namespace, Name: oldPXF.Name}
			err := h.KubeClient.Get(ctx, deploymentKey, &oldDeployment)
			if err != nil {
				result = &metav1.Status{Message: "failed to get PXF Deployment. Try again later: " + err.Error()}
				return
			}
			oldImage := oldDeployment.Spec.Template.Spec.Containers[0].Image
			if oldImage != h.InstanceImage {
				msg := fmt.Sprintf(`%s; GreenplumPXFService has image: %s; Operator supports image: %s`,
					UpgradePXFHelpMsg, oldImage, h.InstanceImage)
				result = &metav1.Status{Message: msg}
				return
			}
		}
	}

	if result = validateWorkerSelector(newPXF.Spec.WorkerSelector, "pxf"); result != nil {
		return
	}

	if result = validateResourceQuantity(newPXF.Spec.CPU, "pxf", "cpu"); result != nil {
		return
	}
	if result = validateResourceQuantity(newPXF.Spec.Memory, "pxf", "memory"); result != nil {
		return
	}

	allowed = true
	return
}
