package apply

import (
	"context"
	"fmt"

	"testrunner/pkg/config"
	"testrunner/pkg/kube/generate"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Job creates a job in the cluster
func Job(ctx context.Context, client *kubernetes.Clientset, cfg config.Config, namespace string) (*batchv1.Job, error) {
	job, err := generate.Job(cfg, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate job manifest: %w", err)
	}

	created, err := client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return created, nil
}
