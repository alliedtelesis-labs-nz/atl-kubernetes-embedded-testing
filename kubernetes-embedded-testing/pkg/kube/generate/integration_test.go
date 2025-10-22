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



func TestRole_GeneratesCorrectManifest(t *testing.T) {
	role := ClusterRole()

	assert.Equal(t, "rbac.authorization.k8s.io/v1", role.APIVersion)
	assert.Equal(t, "ClusterRole", role.Kind)
	assert.Equal(t, "ket-test-runner", role.Name)
	assert.NotEmpty(t, role.Rules)
}

func TestRoleBinding_GeneratesCorrectManifest(t *testing.T) {
	namespace := "test-namespace"
	rb := ClusterRoleBinding(namespace)

	assert.Equal(t, "rbac.authorization.k8s.io/v1", rb.APIVersion)
	assert.Equal(t, "ClusterRoleBinding", rb.Kind)
	assert.Equal(t, "ket-test-runner", rb.Name)
	assert.Len(t, rb.Subjects, 1)
	assert.Equal(t, "ServiceAccount", rb.Subjects[0].Kind)
	assert.Equal(t, "default", rb.Subjects[0].Name)
	assert.Equal(t, namespace, rb.Subjects[0].Namespace)
	assert.Equal(t, "ClusterRole", rb.RoleRef.Kind)
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
	coreRules := findAllRulesByAPIGroup(rules, "")
	require.NotEmpty(t, coreRules)
	
	// Check for pods, services in main core rule
	foundPods := false
	foundEvents := false
	for _, rule := range coreRules {
		for _, resource := range rule.Resources {
			if resource == "pods" {
				foundPods = true
			}
			if resource == "events" {
				foundEvents = true
			}
		}
	}
	assert.True(t, foundPods, "pods resource should be present")
	assert.True(t, foundEvents, "events resource should be present")

	// Verify batch resources (jobs are in batch API group)
	batchRule := findRuleByAPIGroup(rules, "batch")
	require.NotNil(t, batchRule)
	assert.Contains(t, batchRule.Resources, "jobs")
}

// Helper function to find all rules by API group
func findAllRulesByAPIGroup(rules []rbacv1.PolicyRule, apiGroup string) []rbacv1.PolicyRule {
	var matched []rbacv1.PolicyRule
	for _, rule := range rules {
		for _, group := range rule.APIGroups {
			if group == apiGroup {
				matched = append(matched, rule)
				break
			}
		}
	}
	return matched
}

func TestMergeRBACRules(t *testing.T) {
	defaultRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get", "list"},
		},
	}

	additionalRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments"},
			Verbs:     []string{"create", "update"},
		},
	}

	merged := MergeRBACRules(defaultRules, additionalRules)

	assert.Len(t, merged, 2)
	assert.Equal(t, defaultRules[0], merged[0])
	assert.Equal(t, additionalRules[0], merged[1])
}

func TestMergeRBACRules_WithEmptyAdditional(t *testing.T) {
	defaultRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get"},
		},
	}

	merged := MergeRBACRules(defaultRules, nil)

	assert.Equal(t, defaultRules, merged)
}

func TestRole_WithAdditionalRules(t *testing.T) {
	additionalRules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{"custom.io"},
			Resources: []string{"customresources"},
			Verbs:     []string{"get", "list"},
		},
	}

	role := ClusterRole(additionalRules...)

	assert.Equal(t, "rbac.authorization.k8s.io/v1", role.APIVersion)
	assert.Equal(t, "ClusterRole", role.Kind)
	assert.Equal(t, "ket-test-runner", role.Name)
	
	// Should have default rules plus additional rules
	defaultRuleCount := len(GetTestRunnerRBACRules())
	assert.Len(t, role.Rules, defaultRuleCount+len(additionalRules))
	
	// Verify the additional rule is present
	foundCustomRule := false
	for _, rule := range role.Rules {
		if len(rule.APIGroups) > 0 && rule.APIGroups[0] == "custom.io" {
			foundCustomRule = true
			break
		}
	}
	assert.True(t, foundCustomRule, "Additional custom rule should be present")
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
