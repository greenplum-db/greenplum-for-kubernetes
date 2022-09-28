package serviceaccount_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/pkg/serviceaccount"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GreenplumCluster role", func() {
	var role *rbacv1.Role

	JustBeforeEach(func() {
		Expect(serviceaccount.ModifyRole(role)).To(Succeed())
	})

	When("no rule exists", func() {
		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
				},
			}
		})
		It("creates the desired rules", func() {
			Expect(role.Rules).To(ConsistOf(
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Verbs":     ConsistOf("get", "patch"),
					"APIGroups": ConsistOf(""),
					"Resources": ConsistOf("persistentvolumeclaims"),
				}),
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Verbs":     ConsistOf("get"),
					"APIGroups": ConsistOf(""),
					"Resources": ConsistOf("pods"),
				}),
				gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Verbs":     ConsistOf("list", "watch"),
					"APIGroups": ConsistOf(""),
					"Resources": ConsistOf("endpoints"),
				}),
			))
		})
	})

	When("a rule already exists for another object", func() {
		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"watch"},
						APIGroups: []string{"apps/v1"},
						Resources: []string{"deployments"},
					},
				},
			}
		})

		It("has the expected number of rules", func() {
			Expect(role.Rules).To(HaveLen(4))
		})
		It("creates the desired rules", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("get", "patch"),
				"APIGroups": ConsistOf(""),
				"Resources": ConsistOf("persistentvolumeclaims"),
			})))
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("get"),
				"APIGroups": ConsistOf(""),
				"Resources": ConsistOf("pods"),
			})))
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("list", "watch"),
				"APIGroups": ConsistOf(""),
				"Resources": ConsistOf("endpoints"),
			})))

		})
		It("does not remove the existing rule", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("watch"),
				"APIGroups": ConsistOf("apps/v1"),
				"Resources": ConsistOf("deployments"),
			})))
		})
	})

	When("multiple rules exist for the same resource", func() {
		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"watch"},
						APIGroups: []string{""},
						Resources: []string{"persistentvolumeclaims"},
					},
					{
						Verbs:     []string{"get", "patch"},
						APIGroups: []string{""},
						Resources: []string{"persistentvolumeclaims"},
					},
				},
			}
		})
		It("combines them into a single rule", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("get", "patch", "watch"),
				"APIGroups": ConsistOf(""),
				"Resources": ConsistOf("persistentvolumeclaims"),
			})))
		})
	})

	When("a rule exists with a resourceName", func() {
		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:         []string{"get"},
						APIGroups:     []string{""},
						Resources:     []string{"persistentvolumeclaims"},
						ResourceNames: []string{"my-pvc"},
					},
				},
			}
		})
		It("adds a rule for get/patch pvcs", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":         ConsistOf("get", "patch"),
				"APIGroups":     ConsistOf(""),
				"Resources":     ConsistOf("persistentvolumeclaims"),
				"ResourceNames": BeEmpty(),
			})))
		})
		It("does not modify the resourceName rule", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":         ConsistOf("get"),
				"APIGroups":     ConsistOf(""),
				"Resources":     ConsistOf("persistentvolumeclaims"),
				"ResourceNames": ConsistOf("my-pvc"),
			})))
		})
	})

	When("a single rule has multiple resources", func() {
		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"get"},
						APIGroups: []string{""},
						Resources: []string{"persistentvolumeclaims", "serviceaccounts", "pods"},
					},
				},
			}
		})
		It("splits the rule into separate rules", func() {
			Expect(role.Rules).To(HaveLen(4))
		})
		It("preserves the verbs for the extra resource", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("get"),
				"APIGroups": ConsistOf(""),
				"Resources": ConsistOf("serviceaccounts"),
			})))
		})
		It("adds the missing verbs for the pvc resource", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("get", "patch"),
				"APIGroups": ConsistOf(""),
				"Resources": ConsistOf("persistentvolumeclaims"),
			})))
		})
		It("splits out the verb for the pod resource", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     ConsistOf("get"),
				"APIGroups": ConsistOf(""),
				"Resources": ConsistOf("pods"),
			})))
		})
	})

	When("multiple rules exist and are unsorted", func() {
		BeforeEach(func() {
			role = &rbacv1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "greenplum-system-pod",
					Namespace: NamespaceName,
				},
				Rules: []rbacv1.PolicyRule{
					{
						Verbs:     []string{"watch", "patch", "get", "create"},
						APIGroups: []string{""},
						Resources: []string{"serviceaccounts"},
					},
					{
						Verbs:     []string{"watch"},
						APIGroups: []string{"apps/v1"},
						Resources: []string{"deployments"},
					},
					{
						Verbs:     []string{"patch"},
						APIGroups: []string{""},
						Resources: []string{"persistentvolumeclaims"},
					},
					{
						Verbs:     []string{"get"},
						APIGroups: []string{""},
						Resources: []string{"pods"},
					},
				},
			}
		})
		It("sorts the rules", func() {
			Expect(role.Rules[0].Resources).To(ConsistOf("pods"))
			Expect(role.Rules[1].Resources).To(ConsistOf("endpoints"))
			Expect(role.Rules[2].Resources).To(ConsistOf("persistentvolumeclaims"))
			Expect(role.Rules[3].Resources).To(ConsistOf("serviceaccounts"))
			Expect(role.Rules[4].Resources).To(ConsistOf("deployments"))
		})

		It("does not change the order of the verbs", func() {
			Expect(role.Rules).To(ContainElement(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Verbs":     Equal([]string{"watch", "patch", "get", "create"}),
				"Resources": ConsistOf("serviceaccounts"),
			})))
		})
	})
})
