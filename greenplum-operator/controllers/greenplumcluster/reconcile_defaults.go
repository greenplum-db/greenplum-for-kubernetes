package greenplumcluster

import (
	"strings"

	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
)

func SetDefaultGreenplumClusterValues(greenplumCluster *greenplumv1.GreenplumCluster) {
	defaultLowercaseFields := []*string{
		&greenplumCluster.Spec.MasterAndStandby.AntiAffinity,
		&greenplumCluster.Spec.Segments.AntiAffinity,
		&greenplumCluster.Spec.MasterAndStandby.Standby,
		&greenplumCluster.Spec.Segments.Mirrors,
	}
	for _, p := range defaultLowercaseFields {
		// It will be easier to deal with these properties later if they are guaranteed to be lowercase
		*p = strings.ToLower(*p)
	}
}
