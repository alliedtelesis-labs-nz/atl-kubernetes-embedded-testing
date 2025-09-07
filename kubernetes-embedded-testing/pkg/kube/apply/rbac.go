package apply

import (
	"context"
	"fmt"
	"strings"

	"testrunner/pkg/kube/generate"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RBAC creates the necessary RBAC resources for the test namespace
func RBAC(ctx context.Context, client *kubernetes.Clientset, namespace string) error {
	serviceAccount := generate.ServiceAccount(namespace)
	role := generate.Role(namespace)
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
