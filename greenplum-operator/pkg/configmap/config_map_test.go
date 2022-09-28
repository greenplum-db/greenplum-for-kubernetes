package configmap_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/configmap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GreenplumCluster configmap spec", func() {
	var (
		configMap *corev1.ConfigMap
		cluster   *greenplumv1.GreenplumCluster

		clusterNamespace = "test-ns"
	)
	BeforeEach(func() {
		configMap = &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      "greenplum-config",
				Namespace: clusterNamespace,
			},
		}
		cluster = &greenplumv1.GreenplumCluster{
			ObjectMeta: v1.ObjectMeta{
				Name:      "my-test-cluster-name",
				Namespace: clusterNamespace,
				Labels: map[string]string{
					"app":               "greenplum",
					"greenplum-cluster": "my-test-cluster-name",
				},
			},
			Spec: greenplumv1.GreenplumClusterSpec{
				MasterAndStandby: greenplumv1.GreenplumMasterAndStandbySpec{
					GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
						Memory:         resource.MustParse("500Mi"),
						WorkerSelector: nil,
					},
					HostBasedAuthentication: "host based authentication",
					Standby:                 "no",
				},
				Segments: greenplumv1.GreenplumSegmentsSpec{
					GreenplumPodSpec: greenplumv1.GreenplumPodSpec{
						Memory:         resource.MustParse("600Mi"),
						WorkerSelector: nil,
					},
					PrimarySegmentCount: 6,
					Mirrors:             "yes",
				},
				PXF: greenplumv1.GreenplumPXFSpec{
					ServiceName: "my-pxf-service",
				},
			},
		}
	})
	JustBeforeEach(func() {
		configmap.ModifyConfigMap(cluster, configMap)
	})
	It("creates a greenplum configmap", func() {
		Expect(configMap.Name).To(Equal("greenplum-config"))
		Expect(configMap.Namespace).To(Equal(clusterNamespace))
		Expect(configMap.Data[configmap.SegmentCount]).To(Equal("6"))
		Expect(configMap.Data[configmap.Mirrors]).To(Equal("true"))
		Expect(configMap.Data[configmap.Standby]).To(Equal("false"))
		Expect(configMap.Data[configmap.HostBasedAuthentication]).To(Equal("host based authentication"))
		Expect(configMap.Data[configmap.GUCs]).To(Equal("gp_resource_manager = group\ngp_resource_group_memory_limit = 1.0"))
		Expect(configMap.Data[configmap.PXFServiceName]).To(Equal("my-pxf-service"))
		Expect(configMap.ObjectMeta.Labels["app"]).To(Equal("greenplum"))
		Expect(configMap.ObjectMeta.Labels["greenplum-cluster"]).To(Equal("my-test-cluster-name"))

	})
})
