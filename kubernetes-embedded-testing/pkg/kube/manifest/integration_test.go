package manifest

import (
	"strings"
	"testing"

	"testrunner/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAll_GeneratesValidYAML(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:     "test-project",
		Image:           "test-image:latest",
		TestCommand:     "npm test",
		BackoffLimit:    2,
		ActiveDeadlineS: 1800,
		WorkspacePath:   "/workspace",
	}
	namespace := "test-namespace"

	manifests, err := All(cfg, namespace)
	require.NoError(t, err)
	require.Len(t, manifests, 5) // Namespace, ServiceAccount, Role, RoleBinding, Job

	// Verify all manifests are valid YAML
	for i, manifest := range manifests {
		assert.NotEmpty(t, manifest, "Manifest %d should not be empty", i)
		assert.Contains(t, manifest, "apiVersion:", "Manifest %d should contain apiVersion", i)
		assert.Contains(t, manifest, "kind:", "Manifest %d should contain kind", i)
		assert.Contains(t, manifest, "metadata:", "Manifest %d should contain metadata", i)
	}

	// Verify specific resource types
	allManifests := strings.Join(manifests, "\n")
	assert.Contains(t, allManifests, "kind: Namespace")
	assert.Contains(t, allManifests, "kind: ServiceAccount")
	assert.Contains(t, allManifests, "kind: ClusterRole")
	assert.Contains(t, allManifests, "kind: ClusterRoleBinding")
	assert.Contains(t, allManifests, "kind: Job")
}

func TestAll_ContainsExpectedConfiguration(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:     "my-app",
		Image:           "node:18-alpine",
		TestCommand:     "npm run test:integration",
		BackoffLimit:    3,
		ActiveDeadlineS: 3600,
		WorkspacePath:   "/workspace",
	}
	namespace := "test-namespace"

	manifests, err := All(cfg, namespace)
	require.NoError(t, err)

	allManifests := strings.Join(manifests, "\n")

	// Verify namespace
	assert.Contains(t, allManifests, "name: "+namespace)

	// Verify service account
	assert.Contains(t, allManifests, "name: default")

	// Verify role and role binding
	assert.Contains(t, allManifests, "name: ket-test-runner")

	// Verify job configuration
	assert.Contains(t, allManifests, "name: ket-my-app")
	assert.Contains(t, allManifests, "image: node:18-alpine")
	assert.Contains(t, allManifests, "npm run test:integration")
	assert.Contains(t, allManifests, "workingDir: /workspace/my-app")
	assert.Contains(t, allManifests, "backoffLimit: 3")
	assert.Contains(t, allManifests, "activeDeadlineSeconds: 3600")

	// Verify environment variables
	assert.Contains(t, allManifests, "name: KET_TEST_NAMESPACE")
	assert.Contains(t, allManifests, "name: KET_PROJECT_ROOT")
	assert.Contains(t, allManifests, "name: KET_WORKSPACE_PATH")

	// Verify volume mounts
	assert.Contains(t, allManifests, "mountPath: /workspace")
	assert.Contains(t, allManifests, "mountPath: /reports")
}

func TestAll_ProjectRootVariations(t *testing.T) {
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

			manifests, err := All(cfg, "test-namespace")
			require.NoError(t, err)

			allManifests := strings.Join(manifests, "\n")
			assert.Contains(t, allManifests, "workingDir: "+tt.expectedDir)
		})
	}
}

func TestAll_YAMLStructure(t *testing.T) {
	cfg := config.Config{
		ProjectRoot: "test-project",
	}

	manifests, err := All(cfg, "test-namespace")
	require.NoError(t, err)

	for i, manifest := range manifests {
		// Each manifest should be a valid YAML document
		lines := strings.Split(strings.TrimSpace(manifest), "\n")
		assert.NotEmpty(t, lines, "Manifest %d should not be empty", i)

		// First line should be "---" or start with "apiVersion:"
		firstLine := strings.TrimSpace(lines[0])
		assert.True(t,
			firstLine == "---" || strings.HasPrefix(firstLine, "apiVersion:"),
			"Manifest %d first line should be '---' or start with 'apiVersion:', got: %s", i, firstLine)
	}
}
