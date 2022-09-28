package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/service"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("GreenplumCluster service spec", func() {
	var greenplumService *corev1.Service
	BeforeEach(func() {
		greenplumService = &corev1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      "greenplum",
				Namespace: NamespaceName,
			},
		}
	})
	It("adds the psql port to a new greenplum service", func() {
		service.ModifyGreenplumService(ClusterName, greenplumService)
		Expect(greenplumService.Name).To(Equal("greenplum"))
		Expect(greenplumService.Namespace).To(Equal(NamespaceName))
		Expect(greenplumService.Spec.Type).To(Equal(corev1.ServiceTypeLoadBalancer))
		Expect(greenplumService.Spec.ExternalTrafficPolicy).To(Equal(corev1.ServiceExternalTrafficPolicyTypeLocal))
		Expect(greenplumService.Spec.Selector["statefulset.kubernetes.io/pod-name"]).To(Equal("master-0"))
		Expect(greenplumService.Spec.Ports).To(HaveLen(1))
		Expect(greenplumService.Spec.Ports[0].Name).To(Equal("psql"))
		Expect(greenplumService.Spec.Ports[0].Port).To(Equal(int32(5432)))
		Expect(greenplumService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
		Expect(greenplumService.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(5432)))
		Expect(greenplumService.ObjectMeta.Labels["app"]).To(Equal("greenplum"))
		Expect(greenplumService.ObjectMeta.Labels["greenplum-cluster"]).To(Equal("my-greenplum"))
	})
	When("the greenplum service already has another port, but the psql port does not exist", func() {
		BeforeEach(func() {
			greenplumService.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "somethingelse",
					Protocol:   "TCP",
					Port:       9999,
					TargetPort: intstr.IntOrString{IntVal: 9999},
				},
			}
		})
		It("adds the psql port", func() {
			service.ModifyGreenplumService(ClusterName, greenplumService)
			Expect(greenplumService.Spec.Ports).To(HaveLen(2))
			Expect(greenplumService.Spec.Ports[0].Name).To(Equal("somethingelse"))
			Expect(greenplumService.Spec.Ports[0].Port).To(Equal(int32(9999)))
			Expect(greenplumService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(greenplumService.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(9999)))
			Expect(greenplumService.Spec.Ports[1].Name).To(Equal("psql"))
			Expect(greenplumService.Spec.Ports[1].Port).To(Equal(int32(5432)))
			Expect(greenplumService.Spec.Ports[1].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(greenplumService.Spec.Ports[1].TargetPort.IntVal).To(Equal(int32(5432)))
		})

	})
	DescribeTable("psql port is modified",
		func(name string, port, targetPort int32) {
			greenplumService.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "somethingelse",
					Protocol:   "TCP",
					Port:       9999,
					TargetPort: intstr.IntOrString{IntVal: 9999},
				},
				{
					Name:       name,
					Protocol:   "TCP",
					Port:       port,
					TargetPort: intstr.IntOrString{IntVal: targetPort},
				},
			}
			service.ModifyGreenplumService(ClusterName, greenplumService)
			Expect(greenplumService.Spec.Ports).To(HaveLen(2))
			Expect(greenplumService.Spec.Ports[0].Name).To(Equal("somethingelse"))
			Expect(greenplumService.Spec.Ports[0].Port).To(Equal(int32(9999)))
			Expect(greenplumService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(greenplumService.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(9999)))
			Expect(greenplumService.Spec.Ports[1].Name).To(Equal("psql"))
			Expect(greenplumService.Spec.Ports[1].Port).To(Equal(int32(5432)))
			Expect(greenplumService.Spec.Ports[1].Protocol).To(Equal(corev1.ProtocolTCP))
			Expect(greenplumService.Spec.Ports[1].TargetPort.IntVal).To(Equal(int32(5432)))
		},
		Entry("name is changed", "foo", int32(5432), int32(5432)),
		Entry("port is changed", "psql", int32(1111), int32(5432)),
		Entry("targetPort is changed", "psql", int32(5432), int32(1111)),
	)
})
