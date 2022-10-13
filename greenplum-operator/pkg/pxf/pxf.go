package pxf

import (
	"fmt"

	greenplumv1beta1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1beta1"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ModifyDeployment(greenplumPXF greenplumv1beta1.GreenplumPXFService, deployment *appsv1.Deployment, image string) {
	labels := generateLabels(greenplumPXF.Name)

	deployment.Labels = labels
	deployment.Spec.Replicas = heapvalue.NewInt32(greenplumPXF.Spec.Replicas)
	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	deployment.Spec.Template.Labels = labels
	templateSpec := &deployment.Spec.Template.Spec
	templateSpec.NodeSelector = greenplumPXF.Spec.WorkerSelector
	templateSpec.ImagePullSecrets = []corev1.LocalObjectReference{
		{
			Name: "regsecret",
		},
	}
	var container *corev1.Container
	if len(templateSpec.Containers) == 0 {
		templateSpec.Containers = make([]corev1.Container, 1)
	}
	container = &templateSpec.Containers[0]

	envVars := []corev1.EnvVar{{
		Name:  "PXF_JVM_OPTS",
		Value: "-XX:MaxRAMPercentage=75.0",
	}}
	if greenplumPXF.Spec.PXFConf != nil && greenplumPXF.Spec.PXFConf.S3Source.Secret != "" {
		envVars = append(envVars, generateS3Env(greenplumPXF)...)
	}
	container.Env = envVars

	container.Name = "pxf"
	container.Args = []string{"/home/gpadmin/tools/startPXF"}
	container.Image = image
	container.ImagePullPolicy = corev1.PullIfNotPresent
	container.Ports = []corev1.ContainerPort{
		{
			ContainerPort: 5888,
			Protocol:      corev1.ProtocolTCP,
		},
	}
	if container.ReadinessProbe == nil {
		container.ReadinessProbe = &corev1.Probe{}
	}
	container.ReadinessProbe.ProbeHandler = corev1.ProbeHandler{
		Exec: &corev1.ExecAction{Command: []string{"/usr/local/pxf-gp6/bin/pxf", "status"}},
	}
	container.ReadinessProbe.InitialDelaySeconds = 30
	container.ReadinessProbe.TimeoutSeconds = 5

	if container.Resources.Limits.Cpu().Cmp(greenplumPXF.Spec.CPU) != 0 ||
		container.Resources.Limits.Memory().Cmp(greenplumPXF.Spec.Memory) != 0 {
		container.Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    greenplumPXF.Spec.CPU,
				corev1.ResourceMemory: greenplumPXF.Spec.Memory,
			},
		}
	}
}

func generateS3Env(greenplumPXF greenplumv1beta1.GreenplumPXFService) []corev1.EnvVar {
	s3Source := greenplumPXF.Spec.PXFConf.S3Source

	endpointIsSecure := true
	if s3Source.Protocol == "http" {
		endpointIsSecure = false
	}

	return []corev1.EnvVar{
		{
			Name: "S3_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: s3Source.Secret,
					},
					Key: "secret_access_key",
				},
			},
		},
		{
			Name: "S3_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: s3Source.Secret,
					},
					Key: "access_key_id",
				},
			},
		},
		{
			Name:  "S3_BUCKET",
			Value: s3Source.Bucket,
		},
		{
			Name:  "S3_ENDPOINT",
			Value: s3Source.EndPoint,
		},
		{
			Name:  "S3_ENDPOINT_IS_SECURE",
			Value: fmt.Sprintf("%v", endpointIsSecure),
		},
		{
			Name:  "S3_FOLDER",
			Value: s3Source.Folder,
		},
	}
}

func ModifyService(greenplumPXF greenplumv1beta1.GreenplumPXFService, service *corev1.Service) {
	labels := generateLabels(greenplumPXF.Name)

	service.Labels = labels
	service.Spec.Selector = labels
	service.Spec.Ports = []corev1.ServicePort{{
		Port:       int32(5888),
		Protocol:   "TCP",
		TargetPort: intstr.IntOrString{IntVal: 5888},
	}}
}

func generateLabels(name string) map[string]string {
	return map[string]string{
		"app":           greenplumv1beta1.PXFAppName,
		"greenplum-pxf": name,
	}
}
