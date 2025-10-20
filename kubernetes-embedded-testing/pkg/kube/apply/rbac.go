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
	serviceAccount := generate.ServiceAccount(namespace)
	
	// Load additional RBAC rules from file if specified
	var additionalRules []rbacv1.PolicyRule
	if cfg != nil && cfg.RbacFile != "" {
		rules, err := generate.LoadRBACRulesFromFile(cfg.RbacFile)
		if err != nil {
			return fmt.Errorf("failed to load RBAC rules from file: %w", err)
		}
		additionalRules = rules
	}
	
	role := generate.Role(namespace, additionalRules...)
	roleBinding := generate.RoleBinding(namespace)

	_, err := client.CoreV1().ServiceAccounts(namespace).Create(ctx, serviceAccount, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create service account: %w", err)
		}
	}

	_, err = client.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create role: %w", err)
		}
	}

	_, err = client.RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create role binding: %w", err)
		}
	}

	return nil
}
