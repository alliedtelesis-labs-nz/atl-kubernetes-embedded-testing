package apply

import (
	"context"
	"fmt"
	"strings"

	"testrunner/pkg/kube/generate"
	"testrunner/pkg/logger"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Namespace creates a namespace in the cluster
func Namespace(ctx context.Context, client *kubernetes.Clientset, namespace string) (string, error) {
	ns := generate.Namespace(namespace)

	created, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return "", fmt.Errorf("failed to create namespace: %w", err)
		}
		return namespace, nil
	}

	return created.Name, nil
}

// DeleteNamespace deletes a namespace
func DeleteNamespace(ctx context.Context, client *kubernetes.Clientset, namespace string) error {
	logger.KubeLogger.Info("Deleting namespace: %s", namespace)

	err := client.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.KubeLogger.Info("Namespace %s not found, already deleted", namespace)
			return nil
		}
		return fmt.Errorf("failed to delete namespace %s: %w", namespace, err)
	}

	logger.KubeLogger.Info("Namespace %s deletion requested", namespace)
	return nil
}
