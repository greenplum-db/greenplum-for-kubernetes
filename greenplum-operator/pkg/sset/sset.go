package sset

import (
	"fmt"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/pkg/heapvalue"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const headlessServiceName = "agent"

type StatefulSetType string

const (
	TypeMaster   StatefulSetType = "master"
	TypeSegmentA StatefulSetType = "segment-a"
	TypeSegmentB StatefulSetType = "segment-b"
)

type GreenplumStatefulSetParams struct {
	Type          StatefulSetType
	ClusterName   string
	Replicas      int32
	InstanceImage string
	GpPodSpec     greenplumv1.GreenplumPodSpec
}

func GenerateStatefulSetParams(ssetType StatefulSetType, cluster *greenplumv1.GreenplumCluster, instanceImage string) *GreenplumStatefulSetParams {
	var replicaCount int32
	var gpPodSpec greenplumv1.GreenplumPodSpec

	if ssetType == TypeMaster {
		if cluster.Spec.MasterAndStandby.Standby == "yes" {
			replicaCount = 2
		} else {
			replicaCount = 1
		}
		gpPodSpec = cluster.Spec.MasterAndStandby.GreenplumPodSpec
	} else {
		replicaCount = cluster.Spec.Segments.PrimarySegmentCount
		gpPodSpec = cluster.Spec.Segments.GreenplumPodSpec
	}

	return &GreenplumStatefulSetParams{
		Type:          ssetType,
		ClusterName:   cluster.Name,
		Replicas:      replicaCount,
		InstanceImage: instanceImage,
		GpPodSpec:     gpPodSpec,
	}
}

func ModifyGreenplumStatefulSet(params *GreenplumStatefulSetParams, sset *appsv1.StatefulSet) {
	labels := generateGPClusterLabels(sset.Name, params.ClusterName)

	if sset.Labels == nil {
		sset.Labels = make(map[string]string)
	}
	for key, value := range labels {
		sset.Labels[key] = value
	}

	sset.Spec.Replicas = &params.Replicas
	sset.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels,
	}
	sset.Spec.ServiceName = headlessServiceName
	sset.Spec.PodManagementPolicy = appsv1.ParallelPodManagement
	sset.Spec.VolumeClaimTemplates = modifyGreenplumPVC(params, sset.Spec.VolumeClaimTemplates)

	if sset.Spec.Template.Labels == nil {
		sset.Spec.Template.Labels = make(map[string]string)
	}
	for key, value := range labels {
		sset.Spec.Template.Labels[key] = value
	}

	templateSpec := &sset.Spec.Template.Spec
	templateSpec.DNSConfig = &corev1.PodDNSConfig{
		Searches: []string{headlessServiceName + "." + sset.Namespace + ".svc.cluster.local"},
	}
	if len(params.GpPodSpec.WorkerSelector) > 0 {
		templateSpec.NodeSelector = params.GpPodSpec.WorkerSelector
	}
	templateSpec.ImagePullSecrets = []corev1.LocalObjectReference{
		{
			Name: "regsecret",
		},
	}
	templateSpec.Containers = modifyGreenplumContainer(params, templateSpec.Containers)
	templateSpec.Volumes = getVolumeDefinition()
	if params.GpPodSpec.AntiAffinity == "yes" {
		templateSpec.Affinity = getAffinityDefinition(params.Type, sset.Namespace)
	}
	templateSpec.ServiceAccountName = "greenplum-system-pod"
}

func modifyGreenplumPVC(params *GreenplumStatefulSetParams, pvcs []corev1.PersistentVolumeClaim) []corev1.PersistentVolumeClaim {
	var pvc *corev1.PersistentVolumeClaim
	if len(pvcs) == 0 {
		pvcs = make([]corev1.PersistentVolumeClaim, 1)
	}
	pvc = &pvcs[0]
	pvc.Name = params.ClusterName + "-pgdata"
	pvc.Spec.StorageClassName = &params.GpPodSpec.StorageClassName
	pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	pvc.Spec.Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceStorage: params.GpPodSpec.Storage,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceStorage: params.GpPodSpec.Storage,
		},
	}
	return pvcs
}

func modifyGreenplumContainer(params *GreenplumStatefulSetParams, containers []corev1.Container) []corev1.Container {
	var container *corev1.Container
	if len(containers) == 0 {
		containers = make([]corev1.Container, 1)
	}
	container = &containers[0]
	container.Name = greenplumv1.AppName
	container.Args = []string{"/home/gpadmin/tools/startGreenplumContainer"}
	container.Image = params.InstanceImage
	container.ImagePullPolicy = corev1.PullIfNotPresent
	container.Ports = []corev1.ContainerPort{
		{
			ContainerPort: 22,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if container.ReadinessProbe == nil {
		container.ReadinessProbe = &corev1.Probe{}
	}
	container.ReadinessProbe.ProbeHandler = corev1.ProbeHandler{
		TCPSocket: &corev1.TCPSocketAction{
			Port: intstr.FromInt(22),
		},
	}
	container.ReadinessProbe.InitialDelaySeconds = 5

	if container.Resources.Limits == nil {
		container.Resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
	}
	if params.GpPodSpec.Memory.Cmp(container.Resources.Limits[corev1.ResourceMemory]) != 0 {
		container.Resources.Limits[corev1.ResourceMemory] = params.GpPodSpec.Memory
	}
	if params.GpPodSpec.CPU.Cmp(container.Resources.Limits[corev1.ResourceCPU]) != 0 {
		container.Resources.Limits[corev1.ResourceCPU] = params.GpPodSpec.CPU
	}

	container.Env = []corev1.EnvVar{
		{
			Name:      "MASTER_DATA_DIRECTORY",
			Value:     "/greenplum/data-1",
			ValueFrom: nil,
		},
	}

	container.VolumeMounts = []corev1.VolumeMount{
		{
			Name:      "ssh-key-volume",
			MountPath: "/etc/ssh-key",
		},
		{
			Name:      "config-volume",
			MountPath: "/etc/config",
		},
		{
			Name:      params.ClusterName + "-pgdata",
			MountPath: "/greenplum",
		},
		{
			Name:      "cgroups",
			MountPath: "/sys/fs/cgroup",
		},
		{
			Name:      "podinfo",
			MountPath: "/etc/podinfo",
		},
	}

	return containers
}

func getVolumeDefinition() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "ssh-key-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "ssh-secrets",
					DefaultMode: heapvalue.NewInt32(0444),
				},
			},
		},
		{
			Name: "config-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "greenplum-config",
					},
					DefaultMode: heapvalue.NewInt32(corev1.ConfigMapVolumeSourceDefaultMode),
				},
			},
		},
		{
			Name: "cgroups",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/sys/fs/cgroup",
					Type: heapvalue.NewHostPathType(corev1.HostPathUnset),
				},
			},
		},
		{
			Name: "podinfo",
			VolumeSource: corev1.VolumeSource{
				DownwardAPI: &corev1.DownwardAPIVolumeSource{
					Items: []corev1.DownwardAPIVolumeFile{
						{
							Path: "namespace",
							FieldRef: &corev1.ObjectFieldSelector{
								APIVersion: "v1",
								FieldPath:  "metadata.namespace",
							},
						},
						{
							Path: "greenplumClusterName",
							FieldRef: &corev1.ObjectFieldSelector{
								APIVersion: "v1",
								FieldPath:  "metadata.labels['greenplum-cluster']",
							},
						},
					},
					DefaultMode: heapvalue.NewInt32(corev1.DownwardAPIVolumeSourceDefaultMode),
				},
			},
		},
	}
}

func getAffinityDefinition(typ StatefulSetType, namespace string) *corev1.Affinity {
	var nodeSelectorKey string
	var nodeSelectorValues []string
	var podAntiAffinity *corev1.PodAntiAffinity
	switch typ {
	case TypeMaster:
		nodeSelectorKey = fmt.Sprintf("greenplum-affinity-%s-master", namespace)
		nodeSelectorValues = []string{"true"}
		podAntiAffinity = &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "type",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"master"},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		}
	case TypeSegmentA:
		nodeSelectorKey = fmt.Sprintf("greenplum-affinity-%s-segment", namespace)
		nodeSelectorValues = []string{"a"}
	case TypeSegmentB:
		nodeSelectorKey = fmt.Sprintf("greenplum-affinity-%s-segment", namespace)
		nodeSelectorValues = []string{"b"}
	default:
		panic("unexpected value for StatefulSetType: " + typ)
	}

	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      nodeSelectorKey,
								Operator: corev1.NodeSelectorOpIn,
								Values:   nodeSelectorValues,
							},
						},
					},
				},
			},
		},
		PodAntiAffinity: podAntiAffinity,
	}
}

func generateGPClusterLabels(typ, clusterName string) map[string]string {
	return map[string]string{
		"app":               greenplumv1.AppName,
		"type":              typ,
		"greenplum-cluster": clusterName,
	}
}
