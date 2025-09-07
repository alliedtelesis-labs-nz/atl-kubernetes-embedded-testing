package launcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// generateTestNamespace generates a unique test namespace name using UUID
func generateTestNamespace(projectRoot string) string {
	var basename string

	if projectRoot == "." {
		if cwd, err := os.Getwd(); err == nil {
			basename = filepath.Base(cwd)
		} else {
			basename = "default"
		}
	} else {
		basename = filepath.Base(projectRoot)
	}

	// Clean the basename to make it safe for namespace names
	// Replace any non-alphanumeric characters with hyphens
	cleanName := ""
	for _, r := range basename {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			cleanName += string(r)
		} else {
			cleanName += "-"
		}
	}

	// Remove multiple consecutive hyphens
	for i := 0; i < len(cleanName)-1; i++ {
		if cleanName[i] == '-' && cleanName[i+1] == '-' {
			cleanName = cleanName[:i] + cleanName[i+1:]
			i--
		}
	}

	// Remove leading/trailing hyphens
	if len(cleanName) > 0 && cleanName[0] == '-' {
		cleanName = cleanName[1:]
	}
	if len(cleanName) == 0 {
		cleanName = "default"
	}

	namespaceUUID := uuid.New().String()[:8]
	return fmt.Sprintf("kubernetes-embedded-test-%s-%s", cleanName, namespaceUUID)
}
