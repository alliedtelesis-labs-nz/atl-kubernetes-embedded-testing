package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// FlagConfig defines the mapping between command flags and Viper configuration keys
type FlagConfig struct {
	ViperKey    string
	Description string
	Default     interface{}
}

// FlagMapping contains all flag configurations organized by command type
var FlagMapping = struct {
	RootFlags   map[string]FlagConfig
	LaunchFlags map[string]FlagConfig
}{
	RootFlags: map[string]FlagConfig{
		"project-root": {
			ViperKey:    "projectRoot",
			Description: "Project root path",
			Default:     ".",
		},
		"cluster-workspace-path": {
			ViperKey: "clusterWorkspacePath",
			Description: "Absolute path where the local project directory is mounted inside the test runner pod.\n" +
				"Defaults to '/workspace', matching Kind/K3D volume mounts (e.g., '$(pwd):/workspace').\n" +
				"Used for syncing source code between local and cluster environments.",
			Default: "/workspace",
		},
		"ns-prefix": {
			ViperKey:    "namespacePrefix",
			Description: "Prefix for the namespace name e.g. kubernetes-embedded-test-nodejs-example",
			Default:     "kubernetes-embedded-test",
		},
		"debug": {
			ViperKey:    "debug",
			Description: "Enable debug logging",
			Default:     false,
		},
		"log-prefix": {
			ViperKey:    "logging.prefix",
			Description: "Show log prefixes ([INFO] [LAUNCHER], etc.)",
			Default:     false,
		},
		"log-timestamp": {
			ViperKey:    "logging.timestamp",
			Description: "Show timestamps in logs",
			Default:     false,
		},
	},
	LaunchFlags: map[string]FlagConfig{
		"image": {
			ViperKey: "image",
			Description: "Container image used for the test runner pod." +
				"Must include dependencies for the test command (e.g., mocha or other test runners).",
			Default: "atidyshirt/kubernetes-embedded-test-runner-base:latest",
		},
		"rbac": {
			ViperKey:    "rbac",
			Description: "Extra Yaml configuration to load into the rbac options",
			Default:     "",
		},
		"test-command": {
			ViperKey:    "testCommand",
			Description: "Command to execute inside the test runner pod (e.g., 'mocha **/*.spec.ts').",
			Default:     "",
		},
		"keep-namespace": {
			ViperKey:    "keepNamespace",
			Description: "If set, the test namespace will not be deleted after the run for debugging purposes.",
			Default:     false,
		},
		"backoff-limit": {
			ViperKey:    "backoffLimit",
			Description: "Maximum number of retry attempts for a failed Kubernetes job.",
			Default:     int32(1),
		},
		"active-deadline-seconds": {
			ViperKey:    "activeDeadlineS",
			Description: "Maximum duration in seconds the job is allowed to run before termination.",
			Default:     int64(1800),
		},
	},
}

// bindFlagsToViper binds flags to Viper using the configuration mapping
func bindFlagsToViper(v *viper.Viper, cmd *cobra.Command) {
	if cmd == nil {
		return
	}

	// Get the root command to access persistent flags
	rootCmd := cmd
	for rootCmd.HasParent() {
		rootCmd = rootCmd.Parent()
	}

	// Bind root flags (persistent)
	for flagName, config := range FlagMapping.RootFlags {
		if flag := rootCmd.PersistentFlags().Lookup(flagName); flag != nil {
			v.BindPFlag(config.ViperKey, flag)
		}
	}

	// Bind command-specific flags
	for flagName, config := range FlagMapping.LaunchFlags {
		if flag := cmd.Flags().Lookup(flagName); flag != nil {
			v.BindPFlag(config.ViperKey, flag)
		}
	}
}
