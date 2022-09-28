package serviceaccount

import (
	"sort"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/kubectl/pkg/util/rbac"
)

func ModifyRole(role *rbacv1.Role) error {
	return modifyRules(&role.Rules)
}

func modifyRules(rules *[]rbacv1.PolicyRule) error {
	var breakdownRules []rbacv1.PolicyRule
	for _, rule := range *rules {
		breakdownRules = append(breakdownRules, rbac.BreakdownRule(rule)...)
	}

	compactRules, err := rbac.CompactRules(breakdownRules)
	if err != nil {
		return err
	}

	mergePolicyRules(&compactRules, []rbacv1.PolicyRule{
		{
			Verbs:     []string{"get", "patch"},
			APIGroups: []string{""},
			Resources: []string{"persistentvolumeclaims"},
		},
		{
			Verbs:     []string{"get"},
			APIGroups: []string{""},
			Resources: []string{"pods"},
		},
		{
			Verbs:     []string{"list", "watch"},
			APIGroups: []string{""},
			Resources: []string{"endpoints"},
		},
	})

	sort.Stable(rbac.SortableRuleSlice(compactRules))
	*rules = compactRules
	return nil
}

// NB: Only supports adding rules that match APIGroup & Resource (not ResourceNames or NonResourceURLs).
func mergePolicyRules(mergeTo *[]rbacv1.PolicyRule, add []rbacv1.PolicyRule) {
	existing := *mergeTo

	found := make([]bool, len(add))
	for i, existingRule := range existing {
		for j, addRule := range add {
			if existingRule.Resources[0] == addRule.Resources[0] &&
				existingRule.APIGroups[0] == addRule.APIGroups[0] &&
				len(existingRule.ResourceNames) == 0 {
				found[j] = true
				for _, verb := range addRule.Verbs {
					if !sliceContainsString(existingRule.Verbs, verb) {
						existingRule.Verbs = append(existingRule.Verbs, verb)
					}
				}
				existing[i] = existingRule
			}
		}
	}
	for j, addRule := range add {
		if !found[j] {
			existing = append(existing, addRule)
		}
	}

	*mergeTo = existing
}

func sliceContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
