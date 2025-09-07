package launcher

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTestNamespace_Uniqueness(t *testing.T) {
	projectRoot := "test-project"

	// Generate multiple namespaces
	ns1 := generateTestNamespace(projectRoot)
	ns2 := generateTestNamespace(projectRoot)
	ns3 := generateTestNamespace(projectRoot)

	// All should be different
	assert.NotEqual(t, ns1, ns2)
	assert.NotEqual(t, ns2, ns3)
	assert.NotEqual(t, ns1, ns3)

	// All should follow the expected pattern
	pattern := regexp.MustCompile(`^kubernetes-embedded-test-test-project-[a-f0-9]{8}$`)
	assert.True(t, pattern.MatchString(ns1))
	assert.True(t, pattern.MatchString(ns2))
	assert.True(t, pattern.MatchString(ns3))
}

func TestGenerateTestNamespace_ProjectRootVariations(t *testing.T) {
	tests := []struct {
		name            string
		projectRoot     string
		expectedPattern string
	}{
		{
			name:            "simple project name",
			projectRoot:     "my-app",
			expectedPattern: `^kubernetes-embedded-test-my-app-[a-f0-9]{8}$`,
		},
		{
			name:            "project with hyphens",
			projectRoot:     "my-awesome-project",
			expectedPattern: `^kubernetes-embedded-test-my-awesome-project-[a-f0-9]{8}$`,
		},
		{
			name:            "project with underscores",
			projectRoot:     "my_awesome_project",
			expectedPattern: `^kubernetes-embedded-test-my-awesome-project-[a-f0-9]{8}$`,
		},
		{
			name:            "project with special characters",
			projectRoot:     "my@awesome#project",
			expectedPattern: `^kubernetes-embedded-test-my-awesome-project-[a-f0-9]{8}$`,
		},
		{
			name:            "project with spaces",
			projectRoot:     "my awesome project",
			expectedPattern: `^kubernetes-embedded-test-my-awesome-project-[a-f0-9]{8}$`,
		},
		{
			name:            "current directory",
			projectRoot:     ".",
			expectedPattern: `^kubernetes-embedded-test-.*-[a-f0-9]{8}$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespace := generateTestNamespace(tt.projectRoot)
			pattern := regexp.MustCompile(tt.expectedPattern)
			assert.True(t, pattern.MatchString(namespace),
				"Namespace %s should match pattern %s", namespace, tt.expectedPattern)
		})
	}
}

func TestGenerateTestNamespace_LengthAndFormat(t *testing.T) {
	namespace := generateTestNamespace("test-project")

	// Should not be empty
	assert.NotEmpty(t, namespace)

	// Should start with expected prefix
	assert.True(t, strings.HasPrefix(namespace, "kubernetes-embedded-test-"))

	// Should end with 8-character UUID
	// Format: kubernetes-embedded-test-{cleanName}-{uuid}
	// Find the last hyphen and extract the UUID part
	lastHyphenIndex := strings.LastIndex(namespace, "-")
	uuidPart := namespace[lastHyphenIndex+1:]
	assert.Len(t, uuidPart, 8)

	// UUID part should be alphanumeric
	uuidPattern := regexp.MustCompile(`^[a-f0-9]{8}$`)
	assert.True(t, uuidPattern.MatchString(uuidPart))
}

func TestGenerateTestNamespace_EmptyProjectRoot(t *testing.T) {
	// Test with empty string (should default to "default")
	namespace := generateTestNamespace("")

	// Should still generate a valid namespace
	assert.NotEmpty(t, namespace)
	assert.True(t, strings.HasPrefix(namespace, "kubernetes-embedded-test-"))
}
