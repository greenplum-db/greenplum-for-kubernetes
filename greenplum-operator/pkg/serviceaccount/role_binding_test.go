package serviceaccount_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/serviceaccount"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GreenplumCluster role binding", func() {
	var roleBinding *rbacv1.RoleBinding
	When("the role binding is empty", func() {
		BeforeEach(func() {
			roleBinding = &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
				},
			}
		})
		It("modifies a role binding", func() {
			serviceaccount.ModifyRoleBinding(roleBinding)

			Expect(roleBinding.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(roleBinding.RoleRef.Kind).To(Equal("Role"))
			Expect(roleBinding.RoleRef.Name).To(Equal("greenplum-system-pod"))

			Expect(roleBinding.Subjects).To(HaveLen(1))
			Expect(roleBinding.Subjects[0].Kind).To(Equal("ServiceAccount"))
			Expect(roleBinding.Subjects[0].Name).To(Equal("greenplum-system-pod"))
			Expect(roleBinding.Subjects[0].Namespace).To(Equal(NamespaceName))

			Expect(roleBinding.ObjectMeta.Labels["app"]).To(Equal(AppName))
		})
	})

	When("there are existing labels", func() {
		BeforeEach(func() {
			roleBinding = &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
					Labels: map[string]string{
						"key":    "value",
						"boring": "label",
					},
				},
			}
		})
		It("does not overwrite them", func() {
			serviceaccount.ModifyRoleBinding(roleBinding)

			Expect(roleBinding.ObjectMeta.Labels["app"]).To(Equal(AppName))
			Expect(roleBinding.ObjectMeta.Labels["key"]).To(Equal("value"))
			Expect(roleBinding.ObjectMeta.Labels["boring"]).To(Equal("label"))
		})
	})
})

// check that we do not overwrite labels
