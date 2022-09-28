package greenplumcluster

import (
	"context"
	"fmt"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *GreenplumClusterReconciler) reconcileStatus(ctx context.Context, greenplumCluster *greenplumv1.GreenplumCluster) error {
	originalGreenplumCluster := greenplumCluster.DeepCopy()
	greenplumCluster.Status.OperatorVersion = r.OperatorImage
	greenplumCluster.Status.InstanceImage = r.InstanceImage
	if greenplumCluster.Status.Phase == "" {
		greenplumCluster.Status.Phase = greenplumv1.GreenplumClusterPhasePending
	}

	if equality.Semantic.DeepEqual(greenplumCluster, originalGreenplumCluster) {
		return nil
	}

	if err := r.Patch(ctx, greenplumCluster, client.MergeFrom(originalGreenplumCluster)); err != nil {
		return fmt.Errorf("updating status: %w", err)
	}

	return nil
}

func (r *GreenplumClusterReconciler) setStatus(ctx context.Context, greenplumCluster *greenplumv1.GreenplumCluster, status greenplumv1.GreenplumClusterPhase) {
	if greenplumCluster.Status.Phase != status {
		originalGreenplumCluster := greenplumCluster.DeepCopy()
		greenplumCluster.Status.Phase = status
		if err := r.Patch(ctx, greenplumCluster, client.MergeFrom(originalGreenplumCluster)); err != nil {
			r.Log.Error(err, "failed to set GreenplumCluster status", "status", status)
		} else {
			r.Log.Info("set GreenplumCluster status", "status", status)
		}
	}
}
