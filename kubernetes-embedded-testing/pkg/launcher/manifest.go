package launcher

import (
	"fmt"
	"os"

	"testrunner/pkg/config"
	"testrunner/pkg/kube/manifest"
	"testrunner/pkg/logger"
)

// RunManifest generates and outputs Kubernetes manifests
func RunManifest(cfg config.Config) error {
	// Set log level to SILENT in manifest mode for clean output
	logger.SetGlobalLevel(logger.SILENT)

	namespace := generateTestNamespace(cfg.ProjectRoot)

	manifests, err := manifest.All(cfg, namespace)
	if err != nil {
		return fmt.Errorf("failed to generate test manifests: %w", err)
	}

	for _, manifest := range manifests {
		fmt.Fprintf(os.Stdout, "---\n%s\n", manifest)
	}

	return nil
}
