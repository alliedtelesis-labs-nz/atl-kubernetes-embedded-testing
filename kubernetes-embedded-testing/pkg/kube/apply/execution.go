package apply

import (
	"context"
	"fmt"
	"time"

	"testrunner/pkg/logger"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TestResult contains the result of a test execution
type TestResult struct {
	ExitCode int
	Success  bool
	Error    error
}

// WaitForTestCompletion waits for the injected test runner job to complete and returns the test results
func WaitForTestCompletion(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job) (*TestResult, error) {
	logger.KubeLogger.Info("Waiting for Test Runner Job %s to complete", job.Name)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.KubeLogger.Debug("Context cancelled, stopping wait for test completion")
			return nil, ctx.Err()
		case <-ticker.C:
			currentJob, err := client.BatchV1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
			if err != nil {
				logger.KubeLogger.Warn("Failed to get job status: %v", err)
				continue
			}

			if currentJob.Status.Succeeded > 0 {
				logger.KubeLogger.Info("Test Runner Job %s completed successfully", job.Name)
				exitCode, err := getPodExitCode(ctx, client, job.Name, job.Namespace)
				if err != nil {
					logger.KubeLogger.Warn("Could not determine pod exit code: %v", err)
					exitCode = 0
				}
				return &TestResult{ExitCode: exitCode, Success: true}, nil
			}

			if currentJob.Status.Failed > 0 {
				logger.KubeLogger.Error("Test Runner Job %s failed", job.Name)
				exitCode, err := getPodExitCode(ctx, client, job.Name, job.Namespace)
				if err != nil {
					logger.KubeLogger.Warn("Could not determine pod exit code: %v", err)
					exitCode = 1
				}
				return &TestResult{ExitCode: exitCode, Success: false, Error: fmt.Errorf("Test Runner Job %s failed", job.Name)}, nil
			}

			logger.KubeLogger.Debug("Test Runner Job %s is still running...", job.Name)
		}
	}
}

// getPodExitCode attempts to get the exit code from the pod's container
func getPodExitCode(ctx context.Context, client *kubernetes.Clientset, jobName, namespace string) (int, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "job-name=" + jobName,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list pods for job: %w", err)
	}

	if len(pods.Items) == 0 {
		return 0, fmt.Errorf("no pods found for job %s", jobName)
	}

	pod := pods.Items[0]
	if len(pod.Status.ContainerStatuses) == 0 {
		return 0, fmt.Errorf("no container statuses found for pod %s", pod.Name)
	}

	containerStatus := pod.Status.ContainerStatuses[0]
	if containerStatus.State.Terminated != nil {
		return int(containerStatus.State.Terminated.ExitCode), nil
	}

	return 0, fmt.Errorf("container not terminated yet")
}

// StreamTestOutputToHost streams the test output from the injected test runner pod back to the host machine
func StreamTestOutputToHost(ctx context.Context, client *kubernetes.Clientset, job *batchv1.Job) error {
	logger.KubeLogger.Info("Waiting for test runner pod to be ready...")

	timeout := time.After(120 * time.Second)  // Increased timeout for image pulling
	ticker := time.NewTicker(2 * time.Second) // Reduced frequency to avoid spam
	defer ticker.Stop()

	lastStatus := ""
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for Test Runner Job %s to become active", job.Name)
		case <-ticker.C:
			attempts++
			pods, err := client.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{
				LabelSelector: "job-name=" + job.Name,
			})
			if err != nil {
				logger.KubeLogger.Warn("Failed to list pods for Test Runner Job: %v", err)
				continue
			}

			if len(pods.Items) > 0 {
				pod := pods.Items[0]
				currentStatus := getPodStatus(pod)

				// Only log status changes to avoid spam
				if currentStatus != lastStatus {
					logger.KubeLogger.Info("Pod %s status: %s", pod.Name, currentStatus)
					lastStatus = currentStatus
				}

				// Check if pod is ready to stream logs
				if isPodReadyForLogs(pod) {
					logger.KubeLogger.Info("Pod %s is ready, starting log stream", pod.Name)
					if err := streamPodLogs(ctx, client, pod, job.Namespace); err != nil {
						logger.KubeLogger.Warn("Failed to stream logs from pod %s: %v", pod.Name, err)
						continue
					}
					return nil
				}
			} else if attempts%5 == 0 { // Log every 10 seconds (5 * 2 second ticker)
				logger.KubeLogger.Info("Waiting for pod to be created...")
			}
		}
	}
}

// getPodStatus returns a human-readable status of the pod
func getPodStatus(pod corev1.Pod) string {
	if len(pod.Status.ContainerStatuses) == 0 {
		return "ContainerCreating"
	}

	containerStatus := pod.Status.ContainerStatuses[0]
	if containerStatus.State.Waiting != nil {
		if containerStatus.State.Waiting.Reason == "ContainerCreating" {
			return "ContainerCreating"
		}
		return fmt.Sprintf("Waiting: %s", containerStatus.State.Waiting.Reason)
	}

	if containerStatus.State.Running != nil {
		return "Running"
	}

	if containerStatus.State.Terminated != nil {
		return fmt.Sprintf("Terminated: %s", containerStatus.State.Terminated.Reason)
	}

	return "Unknown"
}

// isPodReadyForLogs checks if the pod is ready to stream logs from
func isPodReadyForLogs(pod corev1.Pod) bool {

	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}

	containerStatus := pod.Status.ContainerStatuses[0]

	if containerStatus.State.Terminated != nil {
		return true
	}

	return containerStatus.State.Running != nil && containerStatus.Ready
}

// streamPodLogs attempts to stream logs from a specific pod
func streamPodLogs(ctx context.Context, client *kubernetes.Clientset, pod corev1.Pod, namespace string) error {
	logger.KubeLogger.Info("Streaming test output from pod %s", pod.Name)

	req := client.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Follow: true,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to get test output stream: %w", err)
	}
	defer stream.Close()

	logger.TestRunnerLogger.StreamLogs(stream)
	return nil
}
