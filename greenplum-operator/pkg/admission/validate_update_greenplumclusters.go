package admission

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/executor"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const UpgradeClusterHelpMsg = "Cannot update greenplumCluster instance -- operator only supports updates to clusters " +
	"at the latest version. Please update greenplumCluster to the latest version in order to make updates"

func (h *Handler) validateUpdateGreenplumCluster(ctx context.Context, oldGreenplum, newGreenplum greenplumv1.GreenplumCluster) (allowed bool, result *metav1.Status) {
	if !equality.Semantic.DeepEqual(oldGreenplum.Spec, newGreenplum.Spec) {
		if oldGreenplum.Status.InstanceImage != h.InstanceImage {
			msg := fmt.Sprintf(`%s; GreenplumCluster has image: %s; Operator supports image: %s`,
				UpgradeClusterHelpMsg, oldGreenplum.Status.InstanceImage, h.InstanceImage)
			result = &metav1.Status{Message: msg}
			return
		}
	}

	if newGreenplum.Spec.MasterAndStandby.Standby != oldGreenplum.Spec.MasterAndStandby.Standby {
		result = &metav1.Status{Message: "standby value cannot be changed after the cluster has been created"}
		return
	}

	if newGreenplum.Spec.MasterAndStandby.HostBasedAuthentication != oldGreenplum.Spec.MasterAndStandby.HostBasedAuthentication {
		result = &metav1.Status{Message: "hostBasedAuthentication cannot be changed after the cluster has been created"}
		return
	}

	if newGreenplum.Spec.MasterAndStandby.CPU.String() != oldGreenplum.Spec.MasterAndStandby.CPU.String() ||
		newGreenplum.Spec.Segments.CPU.String() != oldGreenplum.Spec.Segments.CPU.String() {
		result = &metav1.Status{Message: "CPU reservation cannot be changed after the cluster has been created"}
		return
	}

	if newGreenplum.Spec.MasterAndStandby.Memory.String() != oldGreenplum.Spec.MasterAndStandby.Memory.String() ||
		newGreenplum.Spec.Segments.Memory.String() != oldGreenplum.Spec.Segments.Memory.String() {
		result = &metav1.Status{Message: "Memory reservation cannot be changed after the cluster has been created"}
		return
	}

	if !equality.Semantic.DeepEqual(newGreenplum.Spec.MasterAndStandby.WorkerSelector, oldGreenplum.Spec.MasterAndStandby.WorkerSelector) ||
		!equality.Semantic.DeepEqual(newGreenplum.Spec.Segments.WorkerSelector, oldGreenplum.Spec.Segments.WorkerSelector) {
		result = &metav1.Status{Message: "workerSelector cannot be changed after the cluster has been created"}
		return
	}

	if strings.ToLower(newGreenplum.Spec.MasterAndStandby.AntiAffinity) != strings.ToLower(oldGreenplum.Spec.MasterAndStandby.AntiAffinity) ||
		strings.ToLower(newGreenplum.Spec.Segments.AntiAffinity) != strings.ToLower(oldGreenplum.Spec.Segments.AntiAffinity) {
		result = &metav1.Status{Message: "antiAffinity cannot be changed after the cluster has been created"}
		return
	}

	if strings.ToLower(newGreenplum.Spec.Segments.Mirrors) != strings.ToLower(oldGreenplum.Spec.Segments.Mirrors) {
		result = &metav1.Status{Message: "mirrors cannot be changed after the cluster has been created"}
		return
	}

	if newGreenplum.Spec.MasterAndStandby.Storage != oldGreenplum.Spec.MasterAndStandby.Storage ||
		newGreenplum.Spec.Segments.Storage != oldGreenplum.Spec.Segments.Storage {
		result = &metav1.Status{Message: "storage cannot be changed after the cluster has been created"}
		return
	}

	if newGreenplum.Spec.MasterAndStandby.StorageClassName != oldGreenplum.Spec.MasterAndStandby.StorageClassName ||
		newGreenplum.Spec.Segments.StorageClassName != oldGreenplum.Spec.Segments.StorageClassName {
		result = &metav1.Status{Message: "storageClassName cannot be changed after the cluster has been created"}
		return
	}

	if newGreenplum.Spec.Segments.PrimarySegmentCount < oldGreenplum.Spec.Segments.PrimarySegmentCount {
		result = &metav1.Status{Message: "primarySegmentCount cannot be decreased after the cluster has been created"}
		return
	}

	result = h.validateExpand(ctx, oldGreenplum, newGreenplum)
	if result != nil {
		return
	}

	if newGreenplum.Spec.PXF.ServiceName != oldGreenplum.Spec.PXF.ServiceName {
		result = &metav1.Status{Message: "PXF serviceName cannot be changed after the cluster has been created"}
		return
	}

	allowed = true
	return
}

func (h *Handler) validateExpand(ctx context.Context, oldGreenplum, newGreenplum greenplumv1.GreenplumCluster) (result *metav1.Status) {
	if newGreenplum.Spec.Segments.PrimarySegmentCount > oldGreenplum.Spec.Segments.PrimarySegmentCount {
		// TODO: Actually query the gpdb status server (once it's implemented)
		if oldGreenplum.Status.Phase != greenplumv1.GreenplumClusterPhaseRunning {
			result = &metav1.Status{Message: "updates only supported when cluster is Running"}
			return
		}

		activeMaster := executor.GetCurrentActiveMaster(h.PodCmdExecutor, newGreenplum.Namespace)
		if activeMaster == "" {
			result = &metav1.Status{Message: "failed to contact an active gpdb master"}
			return
		}

		checkExpandSchemaQuery := "SELECT count(*) FROM information_schema.schemata WHERE schema_name = 'gpexpand'"
		checkExpandSchemaCmd := []string{
			"/bin/bash",
			"-c",
			"--",
			fmt.Sprintf(`source /usr/local/greenplum-db/greenplum_path.sh && psql -d postgres -tAc "%s"`, checkExpandSchemaQuery),
		}
		var stdout, stderr bytes.Buffer
		err := h.PodCmdExecutor.Execute(checkExpandSchemaCmd, newGreenplum.Namespace, activeMaster, &stdout, &stderr)
		if err != nil {
			result = &metav1.Status{Message: "failed to check for expansion schema: " + err.Error()}
			return
		}
		queryResult := strings.TrimSpace(stdout.String())
		if queryResult != "0" {
			result = &metav1.Status{Message: "previous expansion schema exists. you must redistribute data and clean up expansion schema prior to performing another expansion"}
			return
		}

		var job batchv1.Job
		jobKey := types.NamespacedName{
			Namespace: newGreenplum.Namespace,
			Name:      fmt.Sprintf("%s-gpexpand-job", newGreenplum.Name),
		}
		// TODO: get a real context
		err = h.KubeClient.Get(ctx, jobKey, &job)
		if err != nil && !apierrs.IsNotFound(err) {
			result = &metav1.Status{Message: "failed to check for previous expand job: " + err.Error()}
			return
		}
		if err == nil {
			if job.Status.Succeeded > 0 {
				return
			}
			if job.Status.Failed > 0 {
				result = &metav1.Status{Message: "cannot expand cluster because previous gpexpand job failed"}
				return
			}

			// Active could be >0, or the Status could be uninitialized.
			result = &metav1.Status{Message: "cannot expand cluster because a gpexpand job is currently running"}
			return
		}
	}
	return
}
