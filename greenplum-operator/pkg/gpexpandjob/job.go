package gpexpandjob

import (
	"strconv"

	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func GenerateJob(image, hostname string, newSegCount int32) (job batchv1.Job) {
	job.Spec.BackoffLimit = heapvalue.NewInt32(0)

	gpexpandPod := &job.Spec.Template.Spec
	gpexpandPod.RestartPolicy = corev1.RestartPolicyNever

	gpexpandPod.Volumes = []corev1.Volume{
		{
			Name: "ssh-key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "ssh-secrets",
					DefaultMode: heapvalue.NewInt32(0444),
				},
			},
		},
	}
	gpexpandPod.ImagePullSecrets = []corev1.LocalObjectReference{
		{
			Name: "regsecret",
		},
	}
	gpexpandPod.Containers = []corev1.Container{
		{
			Name:  "gpexpand",
			Image: image,
			Command: []string{
				"/home/gpadmin/tools/gpexpand_job.sh",
			},
			Env: []corev1.EnvVar{
				{
					Name:      "GPEXPAND_HOST",
					Value:     hostname,
					ValueFrom: nil,
				},
				{
					Name:      "NEW_SEG_COUNT",
					Value:     strconv.FormatInt(int64(newSegCount), 10),
					ValueFrom: nil,
				},
			},
			ImagePullPolicy: corev1.PullIfNotPresent,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "ssh-key",
					ReadOnly:  false,
					MountPath: "/etc/ssh-key",
				},
			},
		},
	}

	return
}
