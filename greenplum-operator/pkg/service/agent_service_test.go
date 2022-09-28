package service_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/service"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GreenplumCluster agent service spec", func() {
	var agentService *corev1.Service
	BeforeEach(func() {
		agentService = &corev1.Service{
			ObjectMeta: v1.ObjectMeta{
				Name:      "agent",
				Namespace: NamespaceName,
			},
		}
	})
	It("creates a greenplum agent service", func() {
		service.ModifyGreenplumAgentService(ClusterName, agentService)
		Expect(agentService.Name).To(Equal("agent"))
		Expect(agentService.Namespace).To(Equal(NamespaceName))
		Expect(agentService.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
		Expect(agentService.Spec.ClusterIP).To(Equal(corev1.ClusterIPNone))
		Expect(agentService.Spec.Selector["app"]).To(Equal(AppName))
		Expect(agentService.Spec.Selector["greenplum-cluster"]).To(Equal(ClusterName))
		Expect(agentService.Spec.Ports[0].Port).To(Equal(int32(22)))
		Expect(agentService.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
		Expect(agentService.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(22)))
		Expect(agentService.ObjectMeta.Labels["app"]).To(Equal(AppName))
		Expect(agentService.ObjectMeta.Labels["greenplum-cluster"]).To(Equal(ClusterName))
	})
})
