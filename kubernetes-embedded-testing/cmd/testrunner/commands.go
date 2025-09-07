package main

import (
	"context"
	"fmt"

	"testrunner/pkg/launcher"

	"github.com/spf13/cobra"
)

// createRootCommand creates the main root command
func createRootCommand(ctx context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ket",
		Short: "Kubernetes Embedded Testing - Deploy and run tests in Kubernetes",
		Long: `Kubernetes Embedded Testing (ket) is a CLI tool that enables running tests in isolated Kubernetes environments

OVERVIEW:
  ket provides a way to run integration tests that require Kubernetes resources 
  without setting up complex test infrastructure. It handles namespace creation, 
  RBAC configuration, and cleanup automatically.

WORKFLOW:
  1. Creates a unique test namespace with proper RBAC permissions
  2. Mounts your source code into the test runner pod
  3. Executes your specified test command in the pod
  4. Streams test output back to your terminal
  5. Automatically cleans up resources when tests complete
`,
	}

	addRootFlags(rootCmd)

	launchCmd := createLaunchCommand(ctx)
	rootCmd.AddCommand(launchCmd)

	manifestCmd := createManifestCommand(ctx)
	rootCmd.AddCommand(manifestCmd)

	envCmd := createEnvCommand()
	rootCmd.AddCommand(envCmd)

	return rootCmd
}

// createLaunchCommand creates the launch command
func createLaunchCommand(ctx context.Context) *cobra.Command {
	launchCmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch tests in Kubernetes",
		Long: `Launch and execute tests in an isolated Kubernetes environment.

This command creates a temporary namespace, mounts your source code, and runs 
your test command in a containerized pod. Test output is streamed back to 
your terminal in real-time.

WORKFLOW:
  1. Creates a unique test namespace with RBAC permissions
  2. Mounts your source code into the test runner pod
  3. Executes your specified test command
  4. Streams test output back to stdout
  5. Automatically cleans up resources when complete

EXAMPLES:
  # Run npm tests in current directory
  ket launch --test-command "npm test"
  
  # Run Go tests in src directory
  ket launch --project-root "src" --test-command "go test -v ./..."
  
  # Run with custom image and timeout
  ket launch --image "node:18-alpine" --active-deadline 3600 --test-command "npm run test:integration"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeLaunch(ctx)
		},
		// Disable help display on errors since test failures are expected
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add launch-specific flags
	addLaunchFlags(launchCmd)

	return launchCmd
}

// executeLaunch handles the launch command execution
func executeLaunch(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	default:
	}

	cfg := buildConfig()
	cfg.Ctx = ctx
	if err := launcher.RunLaunch(*cfg); err != nil {
		if testErr, ok := err.(*launcher.TestExecutionError); ok {
			testExitCode = testErr.ExitCode
			return fmt.Errorf("test execution failed with exit code %d: %s", testErr.ExitCode, testErr.Message)
		}
		return fmt.Errorf("launch failed: %w", err)
	}
	return nil
}

// createManifestCommand creates the manifest command
func createManifestCommand(ctx context.Context) *cobra.Command {
	manifestCmd := &cobra.Command{
		Use:   "manifest",
		Short: "Generate Kubernetes manifests that would be applied with launch command",
		Long: `Generate Kubernetes manifests that would be applied when running tests.

This command creates the same Kubernetes resources that would be deployed 
during a test run, but outputs them as YAML manifests to stdout instead 
of applying them to the cluster.

OUTPUT:
  The generated manifests include:
  - Namespace for test isolation
  - ServiceAccount with appropriate RBAC permissions
  - Role and RoleBinding for test runner access
  - Job specification for test execution

EXAMPLES:
  # Generate manifests and save to file
  ket manifest --test-command "npm test" > test-manifests.yaml
  
  # Generate manifests and apply directly
  ket manifest --test-command "go test ./..." | kubectl apply -f -
  
  # Generate manifests for specific project directory
  ket manifest --project-root "src" --test-command "python -m pytest"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeManifest(ctx)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add manifest-specific flags (same as launch command)
	addLaunchFlags(manifestCmd)

	return manifestCmd
}

// executeManifest handles the manifest command execution
func executeManifest(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled")
	default:
	}

	cfg := buildConfig()
	cfg.Ctx = ctx

	if err := launcher.RunManifest(*cfg); err != nil {
		return fmt.Errorf("manifest generation failed: %w", err)
	}
	return nil
}

// createEnvCommand creates the environment variables documentation command
func createEnvCommand() *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Show environment variables available in test runner pod",
		Long: `Display detailed information about environment variables and volume mounts 
available in the test runner pod during test execution.

This command provides comprehensive documentation for developers writing test 
scripts that will run inside the Kubernetes test runner pod. It shows all 
available environment variables, volume mounts, and usage examples.

INFORMATION INCLUDED:
  - Environment variables with descriptions and examples
  - Volume mount points and their purposes
  - Practical usage examples for test scripts
  - Best practices for accessing Kubernetes resources`,
		Run: func(cmd *cobra.Command, args []string) {
			showEnvironmentDocumentation()
		},
	}

	return envCmd
}

// showEnvironmentDocumentation displays detailed environment variable documentation
func showEnvironmentDocumentation() {
	fmt.Println("The following environment variables are automatically set in the test runner pod:")
	fmt.Println()
	fmt.Println("ENVIRONMENT VARIABLES IN TEST RUNNER POD")
	fmt.Println()
	fmt.Println("  KET_TEST_NAMESPACE")
	fmt.Println("    Description: The Kubernetes namespace where tests are running")
	fmt.Println("    Example:     kubernetes-embedded-test-my-project-a1b2c3d4")
	fmt.Println("    Usage:       Use this to target Kubernetes resources in your test namespace")
	fmt.Println()
	fmt.Println("  KET_PROJECT_ROOT")
	fmt.Println("    Description: The project root directory path (relative to workspace)")
	fmt.Println("    Example:     src/my-app or . (for current directory)")
	fmt.Println("    Usage:       Use this to determine the correct working directory")
	fmt.Println()
	fmt.Println("  KET_WORKSPACE_PATH")
	fmt.Println("    Description: The absolute workspace path mounted in the pod")
	fmt.Println("    Example:     /workspace")
	fmt.Println("    Usage:       Use this to reference the mounted source code location")
	fmt.Println()
	fmt.Println("VOLUME MOUNTS")
	fmt.Println()
	fmt.Println("  /workspace")
	fmt.Println("    Description: Your source code mounted from the host")
	fmt.Println("    Type:        HostPath volume")
	fmt.Println("    Usage:       Contains your project files for testing")
	fmt.Println()
	fmt.Println("  /reports")
	fmt.Println("    Description: Empty directory for test reports and artifacts")
	fmt.Println("    Type:        EmptyDir volume")
	fmt.Println("    Usage:       Write test results, coverage reports, etc. here")
	fmt.Println()
	fmt.Println("EXAMPLE USAGE IN TEST SCRIPTS")
	fmt.Println()
	fmt.Println("  # Get the test namespace")
	fmt.Println("  NAMESPACE=${KET_TEST_NAMESPACE}")
	fmt.Println("  kubectl get pods -n $NAMESPACE")
	fmt.Println()
	fmt.Println("  # Set working directory")
	fmt.Println("  cd ${KET_WORKSPACE_PATH}/${KET_PROJECT_ROOT}")
	fmt.Println()
	fmt.Println("  # Write test reports")
	fmt.Println("  echo 'Test completed' > /reports/test-results.txt")
	fmt.Println("  cp coverage.xml /reports/")
	fmt.Println()
	fmt.Println("  # Access project files")
	fmt.Println("  ls -la ${KET_WORKSPACE_PATH}")
	fmt.Println("  cat ${KET_WORKSPACE_PATH}/package.json")
}
