package configmap

import (
	"fmt"
	"strings"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	Standby                 = "standby"
	SegmentCount            = "segmentCount"
	Mirrors                 = "mirrors"
	HostBasedAuthentication = "hostBasedAuthentication"
	GUCs                    = "GUCs"
	PXFServiceName          = "pxfServiceName"
)

func ModifyConfigMap(cluster *greenplumv1.GreenplumCluster, config *corev1.ConfigMap) {
	segmentCount := cluster.Spec.Segments.PrimarySegmentCount
	mirrors := cluster.Spec.Segments.Mirrors == "yes"
	standby := cluster.Spec.MasterAndStandby.Standby == "yes"

	gucsList := []string{
		"gp_resource_manager = group",
		"gp_resource_group_memory_limit = 1.0",
	}
	gucs := strings.Join(gucsList, "\n")

	labels := map[string]string{
		"app":               greenplumv1.AppName,
		"greenplum-cluster": cluster.Name,
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	for key, value := range labels {
		config.Labels[key] = value
	}
	config.Data = map[string]string{
		SegmentCount:            fmt.Sprint(segmentCount),
		Standby:                 fmt.Sprint(standby),
		Mirrors:                 fmt.Sprint(mirrors),
		HostBasedAuthentication: cluster.Spec.MasterAndStandby.HostBasedAuthentication,
		GUCs:                    gucs,
		PXFServiceName:          cluster.Spec.PXF.ServiceName,
	}
}
