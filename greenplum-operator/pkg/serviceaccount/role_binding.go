package serviceaccount

import (
	greenplumv1 "github.com/pivotal/greenplum-for-kubernetes/greenplum-operator/api/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func ModifyRoleBinding(roleBinding *rbacv1.RoleBinding) {
	if roleBinding.Labels == nil {
		roleBinding.Labels = make(map[string]string)
	}
	roleBinding.Labels["app"] = greenplumv1.AppName

	roleBinding.Subjects = []rbacv1.Subject{{
		Kind:      "ServiceAccount",
		APIGroup:  "",
		Name:      "greenplum-system-pod",
		Namespace: roleBinding.Namespace,
	}}
	roleBinding.RoleRef.APIGroup = "rbac.authorization.k8s.io"
	roleBinding.RoleRef.Kind = "Role"
	roleBinding.RoleRef.Name = "greenplum-system-pod"
}
