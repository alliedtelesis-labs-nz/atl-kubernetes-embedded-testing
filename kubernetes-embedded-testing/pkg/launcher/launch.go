package launcher

import (
	"context"
	"fmt"

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

	createdNamespace, err := apply.Namespace(ctx, client, namespace)
	if err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	err = apply.RBAC(ctx, client, createdNamespace, &cfg)
	if err != nil {
		return fmt.Errorf("failed to create RBAC resources: %w", err)
	}

	job, err := apply.Job(ctx, client, cfg, createdNamespace)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Ensure cleanup on function exit
	defer func() {
		if !cfg.KeepNamespace {
			logger.LauncherLogger.Info("Cleaning up test namespace %s", createdNamespace)
			if err := apply.DeleteNamespace(ctx, client, createdNamespace); err != nil {
				logger.LauncherLogger.Warn("Failed to cleanup namespace %s: %v", createdNamespace, err)
			}
		}
		if err := apply.DeleteRBAC(ctx, client, createdNamespace); err != nil {
			logger.LauncherLogger.Warn("Failed to cleanup namespace %s: %v", createdNamespace, err)
		}
	}()

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
