package greenplumcluster

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/gpexpandjob"
	batchv1 "k8s.io/api/batch/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *GreenplumClusterReconciler) handleExpand(ctx context.Context, greenplumCluster *greenplumv1.GreenplumCluster, activeMaster string) error {
	segmentCount, err := r.getCurrentSegmentCount(greenplumCluster.Namespace, activeMaster)
	if err != nil {
		return err
	}
	if greenplumCluster.Spec.Segments.PrimarySegmentCount <= segmentCount {
		return nil
	}

	jobKey := types.NamespacedName{
		Namespace: greenplumCluster.Namespace,
		Name:      fmt.Sprintf("%s-gpexpand-job", greenplumCluster.Name),
	}

	var existingJob batchv1.Job
	if err := r.Get(ctx, jobKey, &existingJob); err == nil {
		// Job already exists, and is not complete yet
		if existingJob.Status.Succeeded < 1 {
			return nil
		}

		// A job already exists, and has completed successfully
		err = r.Delete(ctx, &existingJob, client.GracePeriodSeconds(0), client.PropagationPolicy(metav1.DeletePropagationBackground))
		if err != nil {
			return err
		}
	} else {
		if !apierrs.IsNotFound(err) {
			return err
		}
	}

	activeMasterFQDN := fmt.Sprintf("%s.agent.%s.svc.cluster.local", activeMaster, greenplumCluster.Namespace)
	job := gpexpandjob.GenerateJob(r.InstanceImage, activeMasterFQDN, greenplumCluster.Spec.Segments.PrimarySegmentCount)
	job.Namespace = jobKey.Namespace
	job.Name = jobKey.Name

	if err := ctrl.SetControllerReference(greenplumCluster, &job, r.Scheme()); err != nil {
		// not tested: not really possible to fail here
		return err
	}
	return r.Create(ctx, &job)
}

func (r *GreenplumClusterReconciler) getCurrentSegmentCount(namespace, masterPodName string) (int32, error) {
	getSegmentCountCommand := []string{
		"/bin/bash",
		"-c",
		"--",
		`source /usr/local/greenplum-db/greenplum_path.sh && psql -t -U gpadmin -c "SELECT COUNT(*) FROM gp_segment_configuration WHERE hostname LIKE 'segment-a%'"`,
	}
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	err := r.PodExec.Execute(getSegmentCountCommand, namespace, masterPodName, stdoutBuf, stderrBuf)
	if err != nil {
		return 0, err
	}
	segCount, err := strconv.ParseInt(strings.TrimSpace(stdoutBuf.String()), 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(segCount), nil
}
