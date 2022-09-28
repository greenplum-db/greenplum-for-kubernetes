package greenplumcluster

import (
	"context"
	"fmt"
	"io/ioutil"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *GreenplumClusterReconciler) handleFinalizer(ctx context.Context, greenplumCluster *greenplumv1.GreenplumCluster, activeMaster *string) error {
	originalGreenplumCluster := greenplumCluster.DeepCopy()
	var needsPatch bool
	var verb string
	if greenplumCluster.DeletionTimestamp.IsZero() {
		if !sliceContainsString(greenplumCluster.Finalizers, StopClusterFinalizer) {
			greenplumCluster.Finalizers = append(greenplumCluster.Finalizers, StopClusterFinalizer)
			verb = "adding"
			needsPatch = true
		}
	} else {
		if sliceContainsString(greenplumCluster.Finalizers, StopClusterFinalizer) {
			r.setStatus(ctx, greenplumCluster, greenplumv1.GreenplumClusterPhaseDeleting)
			r.ensureGreenplumClusterStopped(greenplumCluster, *activeMaster)
			*activeMaster = ""
			greenplumCluster.Finalizers = removeStringFromSlice(greenplumCluster.Finalizers, StopClusterFinalizer)
			verb = "removing"
			needsPatch = true
		}
	}
	if needsPatch {
		if err := r.Patch(ctx, greenplumCluster, client.MergeFrom(originalGreenplumCluster)); err != nil {
			if verb == "removing" && apierrs.IsNotFound(err) {
				r.Log.Info("attempted to remove finalizer, but GreenplumCluster was not found")
				return nil
			}
			return fmt.Errorf("%s finalizer: %w", verb, err)
		}
	}
	return nil
}

func sliceContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeStringFromSlice(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *GreenplumClusterReconciler) ensureGreenplumClusterStopped(greenplumCluster *greenplumv1.GreenplumCluster, activeMaster string) {
	if activeMaster != "" {
		r.Log.Info("initiating shutdown of the greenplum cluster")
		gpStopCommand := []string{
			"/bin/bash",
			"-c",
			"--",
			"source /usr/local/greenplum-db/greenplum_path.sh && gpstop -aM immediate",
		}
		stdout, stderr := ioutil.Discard, ioutil.Discard
		err := r.PodExec.Execute(gpStopCommand, greenplumCluster.Namespace, activeMaster, stdout, stderr)
		if err != nil {
			r.Log.Info("greenplum cluster did not shutdown cleanly. Please check gpAdminLogs for more info.")
		} else {
			r.Log.Info("success shutting down the greenplum cluster")
		}
	}
}
