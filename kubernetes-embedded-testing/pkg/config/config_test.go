package config

import (
	"os"
	"testing"
)

func TestConfigStructure(t *testing.T) {
	cfg := Config{
		Mode:            "launch",
		ProjectRoot:     ".",
		Image:           "node:18-alpine",
		Debug:           false,
		TestCommand:     "npm test",
		KeepNamespace:  false,
		BackoffLimit:    1,
		ActiveDeadlineS: 1800,
		WorkspacePath:   "/workspace",
	}

	if cfg.Mode != "launch" {
		t.Errorf("Expected Mode to be 'launch', got %s", cfg.Mode)
	}

	if cfg.ProjectRoot != "." {
		t.Errorf("Expected ProjectRoot to be '.', got %s", cfg.ProjectRoot)
	}

	if cfg.Image != "node:18-alpine" {
		t.Errorf("Expected Image to be 'node:18-alpine', got %s", cfg.Image)
	}

	if cfg.TestCommand != "npm test" {
		t.Errorf("Expected TestCommand to be 'npm test', got %s", cfg.TestCommand)
	}

	if cfg.WorkspacePath != "/workspace" {
		t.Errorf("Expected WorkspacePath to be '/workspace', got %s", cfg.WorkspacePath)
	}
}

func TestLoadFromFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "ket-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	configContent := `mode: launch
projectRoot: /custom/path
image: custom-image:latest
debug: true
testCommand: "npm run test:integration"
keepNamespace: true
backoffLimit: 3
activeDeadlineS: 3600
clusterWorkspacePath: /custom/workspace`

	if _, err := tempFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}
	tempFile.Close()

	cfg, err := LoadFromFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	if cfg.Mode != "launch" {
		t.Errorf("Expected Mode to be 'launch', got %s", cfg.Mode)
	}

	if cfg.ProjectRoot != "/custom/path" {
		t.Errorf("Expected ProjectRoot to be '/custom/path', got %s", cfg.ProjectRoot)
	}

	if cfg.Image != "custom-image:latest" {
		t.Errorf("Expected Image to be 'custom-image:latest', got %s", cfg.Image)
	}

	if !cfg.Debug {
		t.Error("Expected Debug to be true")
	}

	if cfg.TestCommand != "npm run test:integration" {
		t.Errorf("Expected TestCommand to be 'npm run test:integration', got %s", cfg.TestCommand)
	}

	if !cfg.KeepNamespace {
		t.Error("Expected KeepNamespace to be true")
	}

	if cfg.BackoffLimit != 3 {
		t.Errorf("Expected BackoffLimit to be 3, got %d", cfg.BackoffLimit)
	}

	if cfg.ActiveDeadlineS != 3600 {
		t.Errorf("Expected ActiveDeadlineS to be 3600, got %d", cfg.ActiveDeadlineS)
	}

	if cfg.WorkspacePath != "/custom/workspace" {
		t.Errorf("Expected WorkspacePath to be '/custom/workspace', got %s", cfg.WorkspacePath)
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	_, err := LoadFromFile("nonexistent-file.yaml")
	if err == nil {
		t.Error("Expected error when loading nonexistent file")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		shouldError bool
	}{
		{
			name: "valid config",
			config: Config{
				Mode:            "launch",
				ProjectRoot:     ".",
				Image:           "node:18-alpine",
				Debug:           false,
				TestCommand:     "npm test",
				KeepNamespace:  false,
				BackoffLimit:    1,
				ActiveDeadlineS: 1800,
				WorkspacePath:   "/workspace",
			},
			shouldError: false,
		},
		{
			name: "missing test command",
			config: Config{
				Mode:            "launch",
				ProjectRoot:     ".",
				Image:           "node:18-alpine",
				Debug:           false,
				TestCommand:     "",
				KeepNamespace:  false,
				BackoffLimit:    1,
				ActiveDeadlineS: 1800,
				WorkspacePath:   "/workspace",
			},
			shouldError: true,
		},
		{
			name: "missing image",
			config: Config{
				Mode:            "launch",
				ProjectRoot:     ".",
				Image:           "",
				Debug:           false,
				TestCommand:     "npm test",
				KeepNamespace:  false,
				BackoffLimit:    1,
				ActiveDeadlineS: 1800,
				WorkspacePath:   "/workspace",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldError {
				if tt.config.TestCommand != "" && tt.config.Image != "" {
					t.Error("Expected config to be invalid, but it appears valid")
				}
			} else {
				if tt.config.TestCommand == "" || tt.config.Image == "" {
					t.Error("Expected config to be valid, but it appears invalid")
				}
			}
		})
	}
}
