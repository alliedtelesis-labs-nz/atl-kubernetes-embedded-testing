package generate

import (
	"testing"

	"testrunner/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestNamespace_GeneratesCorrectManifest(t *testing.T) {
	namespace := "test-namespace"
	ns := Namespace(namespace)

	assert.Equal(t, "v1", ns.APIVersion)
	assert.Equal(t, "Namespace", ns.Kind)
	assert.Equal(t, namespace, ns.Name)
}

func TestServiceAccount_GeneratesCorrectManifest(t *testing.T) {
	namespace := "test-namespace"
	sa := ServiceAccount(namespace)

	assert.Equal(t, "v1", sa.APIVersion)
	assert.Equal(t, "ServiceAccount", sa.Kind)
	assert.Equal(t, "default", sa.Name)
	assert.Equal(t, namespace, sa.Namespace)
}

func TestRole_GeneratesCorrectManifest(t *testing.T) {
	namespace := "test-namespace"
	role := Role(namespace)

	assert.Equal(t, "rbac.authorization.k8s.io/v1", role.APIVersion)
	assert.Equal(t, "Role", role.Kind)
	assert.Equal(t, "ket-test-runner", role.Name)
	assert.Equal(t, namespace, role.Namespace)
	assert.NotEmpty(t, role.Rules)
}

func TestRoleBinding_GeneratesCorrectManifest(t *testing.T) {
	namespace := "test-namespace"
	rb := RoleBinding(namespace)

	assert.Equal(t, "rbac.authorization.k8s.io/v1", rb.APIVersion)
	assert.Equal(t, "RoleBinding", rb.Kind)
	assert.Equal(t, "ket-test-runner", rb.Name)
	assert.Equal(t, namespace, rb.Namespace)
	assert.Len(t, rb.Subjects, 1)
	assert.Equal(t, "ServiceAccount", rb.Subjects[0].Kind)
	assert.Equal(t, "default", rb.Subjects[0].Name)
	assert.Equal(t, namespace, rb.Subjects[0].Namespace)
	assert.Equal(t, "Role", rb.RoleRef.Kind)
	assert.Equal(t, "ket-test-runner", rb.RoleRef.Name)
}

func TestJob_GeneratesCorrectManifest(t *testing.T) {
	cfg := config.Config{
		ProjectRoot:     "test-project",
		Image:           "test-image:latest",
		TestCommand:     "npm test",
		BackoffLimit:    2,
		ActiveDeadlineS: 1800,
		WorkspacePath:   "/workspace",
	}
	namespace := "test-namespace"

	job, err := Job(cfg, namespace)
	require.NoError(t, err)

	assert.Equal(t, "batch/v1", job.APIVersion)
	assert.Equal(t, "Job", job.Kind)
	assert.Equal(t, "ket-test-project", job.Name)
	assert.Equal(t, namespace, job.Namespace)
	assert.Equal(t, int32(2), *job.Spec.BackoffLimit)
	assert.Equal(t, int64(1800), *job.Spec.ActiveDeadlineSeconds)

	// Verify container configuration
	container := job.Spec.Template.Spec.Containers[0]
	assert.Equal(t, "test-runner", container.Name)
	assert.Equal(t, "test-image:latest", container.Image)
	assert.Equal(t, []string{"/bin/sh", "-c", "npm test"}, container.Command)
	assert.Equal(t, "/workspace/test-project", container.WorkingDir)

	// Verify environment variables
	envVars := make(map[string]string)
	for _, env := range container.Env {
		envVars[env.Name] = env.Value
	}
	assert.Equal(t, namespace, envVars["KET_TEST_NAMESPACE"])
	assert.Equal(t, "test-project", envVars["KET_PROJECT_ROOT"])
	assert.Equal(t, "/workspace", envVars["KET_WORKSPACE_PATH"])

	// Verify volume mounts
	assert.Len(t, container.VolumeMounts, 2)
	mountPaths := make(map[string]string)
	for _, mount := range container.VolumeMounts {
		mountPaths[mount.Name] = mount.MountPath
	}
	assert.Equal(t, "/workspace", mountPaths["source-code"])
	assert.Equal(t, "/reports", mountPaths["reports"])

	// Verify volumes
	assert.Len(t, job.Spec.Template.Spec.Volumes, 2)
	volumeNames := make(map[string]bool)
	for _, vol := range job.Spec.Template.Spec.Volumes {
		volumeNames[vol.Name] = true
	}
	assert.True(t, volumeNames["source-code"])
	assert.True(t, volumeNames["reports"])
}

func TestJob_WorkingDirectoryCalculation(t *testing.T) {
	tests := []struct {
		name          string
		projectRoot   string
		workspacePath string
		expectedDir   string
	}{
		{
			name:          "current directory",
			projectRoot:   ".",
			workspacePath: "/workspace",
			expectedDir:   "/workspace",
		},
		{
			name:          "subdirectory",
			projectRoot:   "src",
			workspacePath: "/workspace",
			expectedDir:   "/workspace/src",
		},
		{
			name:          "nested directory",
			projectRoot:   "backend/api",
			workspacePath: "/workspace",
			expectedDir:   "/workspace/backend/api",
		},
		{
			name:          "custom workspace path",
			projectRoot:   "app",
			workspacePath: "/app",
			expectedDir:   "/app/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				ProjectRoot:   tt.projectRoot,
				WorkspacePath: tt.workspacePath,
			}

			job, err := Job(cfg, "test-namespace")
			require.NoError(t, err)

			container := job.Spec.Template.Spec.Containers[0]
			assert.Equal(t, tt.expectedDir, container.WorkingDir)
		})
	}
}

func TestJob_ProjectNameGeneration(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		expectedJob string
	}{
		{
			name:        "subdirectory",
			projectRoot: "my-app",
			expectedJob: "ket-my-app",
		},
		{
			name:        "nested directory",
			projectRoot: "backend/api",
			expectedJob: "ket-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				ProjectRoot: tt.projectRoot,
			}

			job, err := Job(cfg, "test-namespace")
			require.NoError(t, err)

			assert.Equal(t, tt.expectedJob, job.Name)
		})
	}
}

func TestGetTestRunnerRBACRules_ContainsExpectedRules(t *testing.T) {
	rules := GetTestRunnerRBACRules()

	// Verify we have rules for different API groups
	apiGroups := make(map[string]bool)
	for _, rule := range rules {
		for _, group := range rule.APIGroups {
			apiGroups[group] = true
		}
	}

	assert.True(t, apiGroups[""]) // core API group
	assert.True(t, apiGroups["apps"])
	assert.True(t, apiGroups["batch"])
	assert.True(t, apiGroups["networking.k8s.io"])

	// Verify core resources
	coreRule := findRuleByAPIGroup(rules, "")
	require.NotNil(t, coreRule)
	assert.Contains(t, coreRule.Resources, "pods")
	assert.Contains(t, coreRule.Resources, "services")

	// Verify batch resources (jobs are in batch API group)
	batchRule := findRuleByAPIGroup(rules, "batch")
	require.NotNil(t, batchRule)
	assert.Contains(t, batchRule.Resources, "jobs")
}

// Helper function to find rule by API group
func findRuleByAPIGroup(rules []rbacv1.PolicyRule, apiGroup string) *rbacv1.PolicyRule {
	for _, rule := range rules {
		if len(rule.APIGroups) == 1 && rule.APIGroups[0] == apiGroup {
			return &rule
		}
	}
	return nil
}
