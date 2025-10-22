package generate

import (
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetTestRunnerRBACRules returns the RBAC rules for the test runner ClusterRole
func GetTestRunnerRBACRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "services", "configmaps", "secrets", "persistentvolumeclaims", "endpoints"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods/log", "pods/portforward", "pods/exec"},
			Verbs:     []string{"get", "list", "create"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"events"},
			Verbs:     []string{"get", "list", "watch"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "statefulsets", "daemonsets"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		},
		{
			APIGroups: []string{"batch"},
			Resources: []string{"jobs", "cronjobs"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		},
		{
			APIGroups: []string{"networking.k8s.io"},
			Resources: []string{"ingresses", "networkpolicies"},
			Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
		},
		{
			APIGroups: []string{""}, // The core API group for Namespaces
			Resources: []string{"namespaces"},
			Verbs:     []string{"get", "delete"},
		},
	}
}

// ClusterRole generates a ClusterRole manifest
func ClusterRole(additionalRules ...rbacv1.PolicyRule) *rbacv1.ClusterRole {
	rules := MergeRBACRules(GetTestRunnerRBACRules(), additionalRules)

	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ket-test-runner",
		},
		Rules: rules,
	}
}

// ClusterRoleBinding generates a ClusterRoleBinding manifest
func ClusterRoleBinding(namespace string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ket-test-runner",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "default",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "ket-test-runner",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}
