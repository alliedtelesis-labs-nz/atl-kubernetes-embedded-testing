package apply

import (
	"context"
	"fmt"

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

func DeleteRBAC(ctx context.Context, client *kubernetes.Clientset, namespace string) error {
	// Explicitly delete the ClusterRoleBinding.
	// This is a cluster-scoped object and is not cleaned up by namespace deletion.
	fmt.Println("Deleting ClusterRoleBinding ket-test-runner...")
	if err := client.RbacV1().ClusterRoleBindings().Delete(ctx, "ket-test-runner", metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete ClusterRoleBinding ket-test-runner: %w", err)
	}

	// Explicitly delete the ClusterRole.
	// This is a cluster-scoped object and is not cleaned up by namespace deletion.
	fmt.Println("Deleting ClusterRole ket-test-runner...")
	if err := client.RbacV1().ClusterRoles().Delete(ctx, "ket-test-runner", metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete ClusterRole ket-test-runner: %w", err)
	}

	// Deleting the ServiceAccount is not strictly necessary if you delete the namespace,
	// but including it ensures a complete cleanup if the namespace deletion fails for some reason.
	// A ServiceAccount is a namespaced resource.
	fmt.Printf("Deleting ServiceAccount default in namespace %s...\n", namespace)
	if err := client.CoreV1().ServiceAccounts(namespace).Delete(ctx, "default", metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete ServiceAccount default: %w", err)
	}
	return nil
}

