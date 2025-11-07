package launcher

import (
	"fmt"
	"regexp"
	"strings"
	"testrunner/pkg/config"
	"github.com/google/uuid"
)

// generateTestNamespace generates a unique test namespace name using cleaned cfg.NamespacePrefix + UUID
func generateTestNamespace(cfg config.Config) string {
        namespace := cfg.Namespace
	if namespace == "" {
		prefix := cfg.NamespacePrefix
		if prefix == "" {
    		       prefix = "kubernetes-embedded-test"
    	        }
    	        cleanPrefix := toKubeSafe(prefix)
    	        namespaceUUID := uuid.New().String()[:8]
    	        namespace = fmt.Sprintf("%s-%s", cleanPrefix, namespaceUUID)
	}
	return namespace
}

// toKubeSafe converts a string to a Kubernetes-safe form: non-alphanumerics replaced with hyphens
func toKubeSafe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	// Collapse multiple hyphens and trim
	re := regexp.MustCompile(`-+`)
	collapsed := re.ReplaceAllString(b.String(), "-")
	return strings.Trim(collapsed, "-")
}
