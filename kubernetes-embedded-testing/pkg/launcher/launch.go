package launcher

import (
	"context"
	"fmt"
	"time"

	"testrunner/pkg/config"
	"testrunner/pkg/kube/apply"
	"testrunner/pkg/logger"
)

// TestExecutionError represents a test execution failure with an exit code
type TestExecutionError struct {
	ExitCode int
	Message  string
}

func (e *TestExecutionError) Error() string {
	return fmt.Sprintf("%s (exit code: %d)", e.Message, e.ExitCode)
}

// RunLaunch executes tests in Kubernetes
func RunLaunch(cfg config.Config) error {
	ctx := context.Background()
	if cfg.Ctx != nil {
		ctx = cfg.Ctx
	}

	logger.ConfigureFromConfig(cfg.Logging.Prefix, cfg.Logging.Timestamp)

	if cfg.Debug {
		logger.SetGlobalLevel(logger.DEBUG)
	}

	client, err := apply.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	namespace := generateTestNamespace(cfg)
	logger.LauncherLogger.Info("Using test namespace: %s", namespace)

	// Track what resources were created for cleanup
	var (
		namespaceCreated = false
		rbacCreated      = false
	)

	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cleanupCancel()

		// Only cleanup resources that were actually created
		if rbacCreated {
			if err := apply.DeleteRBAC(cleanupCtx, client, namespace); err != nil {
				logger.LauncherLogger.Warn("Failed to cleanup RBAC resources: %v", err)
			}
		}

		if namespaceCreated && !cfg.KeepNamespace {
			logger.LauncherLogger.Info("Cleaning up test namespace %s", namespace)
			if err := apply.DeleteNamespace(cleanupCtx, client, namespace); err != nil {
				logger.LauncherLogger.Warn("Failed to cleanup namespace %s: %v", namespace, err)
			}
		}
	}()

	createdNamespace, err := apply.Namespace(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}
	namespaceCreated = true

	err = apply.RBAC(ctx, client, createdNamespace, &cfg)
	if err != nil {
		return fmt.Errorf("failed to create RBAC resources: %w", err)
	}
	rbacCreated = true

	job, err := apply.Job(ctx, client, cfg, createdNamespace)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	if err := apply.StreamTestOutputToHost(ctx, client, job); err != nil {
		return fmt.Errorf("failed to stream test output: %w", err)
	}

	result, err := apply.WaitForTestCompletion(ctx, client, job)
	if err != nil {
		return fmt.Errorf("failed to wait for test completion: %w", err)
	}

	if !result.Success {
		return &TestExecutionError{
			ExitCode: result.ExitCode,
			Message:  result.Error.Error(),
		}
	}

	logger.LauncherLogger.Info("Test execution completed successfully")
	return nil
}
