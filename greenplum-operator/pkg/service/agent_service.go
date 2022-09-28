package service

import (
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ModifyGreenplumAgentService(clusterName string, agentService *corev1.Service) {
	labels := map[string]string{
		"app":               greenplumv1.AppName,
		"greenplum-cluster": clusterName,
	}
	if agentService.Labels == nil {
		agentService.Labels = make(map[string]string)
	}
	for key, value := range labels {
		agentService.Labels[key] = value
	}

	if len(agentService.Spec.Ports) != 1 {
		agentService.Spec.Ports = make([]corev1.ServicePort, 1)
	}
	sshPort := &agentService.Spec.Ports[0]
	sshPort.Port = int32(22)
	sshPort.Protocol = corev1.ProtocolTCP
	sshPort.TargetPort = intstr.IntOrString{IntVal: 22}

	agentService.Spec.Selector = labels
	agentService.Spec.Type = corev1.ServiceTypeClusterIP
	agentService.Spec.ClusterIP = corev1.ClusterIPNone
}
