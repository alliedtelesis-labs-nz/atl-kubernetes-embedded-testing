package launcher

import (
	"regexp"
	"strings"
	"testing"
	"testrunner/pkg/config"

	"github.com/stretchr/testify/assert"
)


func TestGenerateTestNamespace_Uniqueness(t *testing.T) {
	cfg := config.Config{NamespacePrefix: "kubernetes-embedded-test"}
	ns1 := generateTestNamespace(cfg)
	ns2 := generateTestNamespace(cfg)
	ns3 := generateTestNamespace(cfg)
	assert.NotEqual(t, ns1, ns2)
	assert.NotEqual(t, ns2, ns3)
	assert.NotEqual(t, ns1, ns3)

	pattern := regexp.MustCompile(`^kubernetes-embedded-test-[a-f0-9]{8}$`)
	assert.True(t, pattern.MatchString(ns1))
	assert.True(t, pattern.MatchString(ns2))
	assert.True(t, pattern.MatchString(ns3))
}

func TestGenerateTestNamespace_PrefixVariations(t *testing.T) {
	tests := []struct {
		name            string
		prefix          string
		expectedPattern string
	}{
		{
			name:            "simple name",
			prefix:          "my-app",
			expectedPattern: `^my-app-[a-f0-9]{8}$`,
		},
		{
			name:            "with underscores",
			prefix:          "my_app",
			expectedPattern: `^my-app-[a-f0-9]{8}$`,
		},
		{
			name:            "with special chars",
			prefix:          "my@awesome#project",
			expectedPattern: `^my-awesome-project-[a-f0-9]{8}$`,
		},
		{
			name:            "with spaces",
			prefix:          "my awesome project",
			expectedPattern: `^my-awesome-project-[a-f0-9]{8}$`,
		},
		{
			name:            "empty (should default)",
			prefix:          "",
			expectedPattern: `^kubernetes-embedded-test-[a-f0-9]{8}$`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{NamespacePrefix: tt.prefix}
			namespace := generateTestNamespace(cfg)
			pattern := regexp.MustCompile(tt.expectedPattern)
			assert.True(t, pattern.MatchString(namespace),
				"Namespace %s should match pattern %s", namespace, tt.expectedPattern)
		})
	}
}

func TestGenerateTestNamespace_LengthAndFormat(t *testing.T) {
	cfg := config.Config{NamespacePrefix: "kubernetes-embedded-test"}
	namespace := generateTestNamespace(cfg)
	assert.NotEmpty(t, namespace)
	assert.True(t, strings.HasPrefix(namespace, "kubernetes-embedded-test-"))

	lastHyphenIndex := strings.LastIndex(namespace, "-")
	uuidPart := namespace[lastHyphenIndex+1:]
	assert.Len(t, uuidPart, 8)
	uuidPattern := regexp.MustCompile(`^[a-f0-9]{8}$`)
	assert.True(t, uuidPattern.MatchString(uuidPart))
}
