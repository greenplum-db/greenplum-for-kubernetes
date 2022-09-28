package service

import (
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ModifyGreenplumService(clusterName string, greenplumService *corev1.Service) {
	labels := map[string]string{
		"app":               greenplumv1.AppName,
		"greenplum-cluster": clusterName,
	}
	if greenplumService.Labels == nil {
		greenplumService.Labels = make(map[string]string)
	}
	for key, value := range labels {
		greenplumService.Labels[key] = value
	}

	var psqlPort *corev1.ServicePort
	for i, port := range greenplumService.Spec.Ports {
		if port.Name == "psql" || port.Port == int32(5432) || port.TargetPort.IntVal == 5432 {
			psqlPort = &greenplumService.Spec.Ports[i]
		}
	}
	if psqlPort == nil {
		greenplumService.Spec.Ports = append(greenplumService.Spec.Ports, corev1.ServicePort{})
		psqlPort = &greenplumService.Spec.Ports[len(greenplumService.Spec.Ports)-1]
	}
	psqlPort.Name = "psql"
	psqlPort.Port = int32(5432)
	psqlPort.Protocol = corev1.ProtocolTCP
	psqlPort.TargetPort = intstr.IntOrString{IntVal: 5432}

	greenplumService.Spec.Selector = map[string]string{
		"statefulset.kubernetes.io/pod-name": "master-0",
	}
	greenplumService.Spec.Type = corev1.ServiceTypeLoadBalancer
	greenplumService.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	greenplumService.Spec.SessionAffinity = corev1.ServiceAffinityNone
}
