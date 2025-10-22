package apply

import (
	"context"
	"fmt"
	"strings"

	"testrunner/pkg/config"
	"testrunner/pkg/kube/generate"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RBAC creates the necessary RBAC resources for the test namespace
func RBAC(ctx context.Context, client *kubernetes.Clientset, namespace string, cfg *config.Config) error {
	// Load additional RBAC rules from file if specified
	var additionalRules []rbacv1.PolicyRule
	if cfg != nil && cfg.RbacFile != "" {
		rules, err := generate.LoadRBACRulesFromFile(cfg.RbacFile)
		if err != nil {
			return fmt.Errorf("failed to load RBAC rules from file: %w", err)
		}
		additionalRules = rules
	}
	
	role := generate.ClusterRole(additionalRules...)
	roleBinding := generate.ClusterRoleBinding(namespace)

	_, err := client.RbacV1().ClusterRoles().Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	_, err = client.RbacV1().ClusterRoleBindings().Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create role binding: %w", err)
	}

	return nil
}
	}

	return nil
}
