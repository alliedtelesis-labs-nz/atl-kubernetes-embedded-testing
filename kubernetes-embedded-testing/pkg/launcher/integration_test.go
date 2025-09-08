package launcher

import (
	"io"
	"os"
	"strings"
	"testing"

	"testrunner/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunManifest_GeneratesValidYAML(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:     "test-project",
		Image:           "test-image:latest",
		TestCommand:     "npm test",
		BackoffLimit:    2,
		ActiveDeadlineS: 1800,
		WorkspacePath:   "/workspace",
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	err := RunManifest(cfg)
	require.NoError(t, err)

	// Close pipe and read output
	w.Close()
	output, _ := io.ReadAll(r)

	// Verify output contains all expected resources
	outputStr := string(output)
	assert.Contains(t, outputStr, "kind: Namespace")
	assert.Contains(t, outputStr, "kind: ServiceAccount")
	assert.Contains(t, outputStr, "kind: Role")
	assert.Contains(t, outputStr, "kind: RoleBinding")
	assert.Contains(t, outputStr, "kind: Job")

	// Verify YAML structure
	assert.Contains(t, outputStr, "apiVersion: v1")
	assert.Contains(t, outputStr, "apiVersion: batch/v1")
	assert.Contains(t, outputStr, "apiVersion: rbac.authorization.k8s.io/v1")

	// Verify job configuration
	assert.Contains(t, outputStr, "name: ket-test-project")
	assert.Contains(t, outputStr, "image: test-image:latest")
	assert.Contains(t, outputStr, "npm test")
	assert.Contains(t, outputStr, "workingDir: /workspace/test-project")
}

func TestRunManifest_ProjectRootVariations(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		expectedDir string
	}{
		{
			name:        "current directory",
			projectRoot: ".",
			expectedDir: "/workspace",
		},
		{
			name:        "subdirectory",
			projectRoot: "src",
			expectedDir: "/workspace/src",
		},
		{
			name:        "nested directory",
			projectRoot: "backend/api",
			expectedDir: "/workspace/backend/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				ProjectRoot:   tt.projectRoot,
				WorkspacePath: "/workspace",
			}

			// Capture stdout
			originalStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = originalStdout }()

			err := RunManifest(cfg)
			require.NoError(t, err)

			// Close pipe and read output
			w.Close()
			output, _ := io.ReadAll(r)

			// Verify working directory is correct
			outputStr := string(output)
			assert.Contains(t, outputStr, "workingDir: "+tt.expectedDir)
		})
	}
}

func TestRunManifest_NamespaceGeneration(t *testing.T) {
	cfg := config.Config{
		NamespacePrefix: "kubernetes-embedded-test",
		ProjectRoot: "test-project",
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	err := RunManifest(cfg)
	require.NoError(t, err)

	// Close pipe and read output
	w.Close()
	output, _ := io.ReadAll(r)

	// Verify namespace follows expected pattern
	outputStr := string(output)
	assert.Contains(t, outputStr, "name: kubernetes-embedded-test-")

	// Extract namespace name and verify it's unique
	lines := strings.Split(outputStr, "\n")
	var namespaceName string
	for _, line := range lines {
		if strings.Contains(line, "name: kubernetes-embedded-test-") {
			namespaceName = strings.TrimSpace(strings.Split(line, "name: ")[1])
			break
		}
	}
	require.NotEmpty(t, namespaceName)

	// Verify it contains UUID-like suffix (8 characters)
	assert.Regexp(t, `^kubernetes-embedded-test-[a-f0-9]{8}$`, namespaceName)
}

func TestRunManifest_RBACConfiguration(t *testing.T) {
	cfg := config.Config{
		ProjectRoot: "test-project",
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	err := RunManifest(cfg)
	require.NoError(t, err)

	// Close pipe and read output
	w.Close()
	output, _ := io.ReadAll(r)

	outputStr := string(output)

	// Verify ServiceAccount
	assert.Contains(t, outputStr, "name: default")

	// Verify Role
	assert.Contains(t, outputStr, "name: ket-test-runner")
	assert.Contains(t, outputStr, "resources:")
	assert.Contains(t, outputStr, "- pods")
	assert.Contains(t, outputStr, "- services")
	assert.Contains(t, outputStr, "- jobs")

	// Verify RoleBinding
	assert.Contains(t, outputStr, "kind: ServiceAccount")
	assert.Contains(t, outputStr, "kind: Role")
}

func TestRunManifest_JobConfiguration(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:     "my-project",
		Image:           "node:18-alpine",
		TestCommand:     "npm run test:integration",
		BackoffLimit:    3,
		ActiveDeadlineS: 3600,
		WorkspacePath:   "/workspace",
	}

	// Capture stdout
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = originalStdout }()

	err := RunManifest(cfg)
	require.NoError(t, err)

	// Close pipe and read output
	w.Close()
	output, _ := io.ReadAll(r)

	outputStr := string(output)

	// Verify job name
	assert.Contains(t, outputStr, "name: ket-my-project")

	// Verify container configuration
	assert.Contains(t, outputStr, "image: node:18-alpine")
	assert.Contains(t, outputStr, "npm run test:integration")
	assert.Contains(t, outputStr, "workingDir: /workspace/my-project")

	// Verify job spec
	assert.Contains(t, outputStr, "backoffLimit: 3")
	assert.Contains(t, outputStr, "activeDeadlineSeconds: 3600")

	// Verify environment variables
	assert.Contains(t, outputStr, "name: KET_TEST_NAMESPACE")
	assert.Contains(t, outputStr, "name: KET_PROJECT_ROOT")
	assert.Contains(t, outputStr, "name: KET_WORKSPACE_PATH")

	// Verify volume mounts
	assert.Contains(t, outputStr, "mountPath: /workspace")
	assert.Contains(t, outputStr, "mountPath: /reports")
}
