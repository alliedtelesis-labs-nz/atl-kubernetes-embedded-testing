package main

import (
	"fmt"
	"testrunner/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func addRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to config file (YAML/JSON)")
	cmd.PersistentFlags().StringVarP(&projectRoot, "project-root", "r", ".", "Project root path")
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Enable debug logging")
	cmd.PersistentFlags().StringVarP(&workspacePath, "cluster-workspace-path", "w", "/workspace",
		"Absolute path where the local project directory is mounted inside the test runner pod.\n"+
			"Defaults to '/workspace', matching Kind/K3D volume mounts (e.g., '$(pwd):/workspace').\n"+
			"Used for syncing source code between local and cluster environments.")
	cmd.PersistentFlags().BoolVarP(&logPrefix, "log-prefix", "", true, "Show log prefixes ([INFO] [LAUNCHER], etc.)")
	cmd.PersistentFlags().BoolVarP(&logTimestamp, "log-timestamp", "", true, "Show timestamps in logs")
}

func addLaunchFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&image, "image", "i", "atidyshirt/kubernetes-embedded-test-runner-base:latest",
		"Container image used for the test runner pod."+
			"Must include dependencies for the test command (e.g., mocha or other test runners).")
	cmd.Flags().StringVarP(&testCommand, "test-command", "t", "",
		"Command to execute inside the test runner pod (e.g., 'mocha **/*.spec.ts').")
	cmd.Flags().BoolVarP(&keepNamespace, "keep-namespace", "k", false,
		"If set, the test namespace will not be deleted after the run for debugging purposes.")
	cmd.Flags().Int32VarP(&backoffLimit, "backoff-limit", "b", 1,
		"Maximum number of retry attempts for a failed Kubernetes job.")
	cmd.Flags().Int64VarP(&activeDeadline, "active-deadline-seconds", "d", 1800,
		"Maximum duration in seconds the job is allowed to run before termination.")
}

func setupViper() *viper.Viper {
	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		v.SetConfigName("ket-config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.ket")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		}
	}

	return v
}

func buildConfig() *config.Config {
	v := setupViper()

	v.SetDefault("mode", "launch")
	v.SetDefault("image", "atidyshirt/kubernetes-embedded-test-runner-base:latest")
	v.SetDefault("projectRoot", ".")
	v.SetDefault("clusterWorkspacePath", "/workspace")
	v.SetDefault("backoffLimit", int32(1))
	v.SetDefault("activeDeadlineS", int64(1800))
	v.SetDefault("debug", false)

	if projectRoot != "." {
		v.Set("projectRoot", projectRoot)
	}
	if workspacePath != "/workspace" {
		v.Set("clusterWorkspacePath", workspacePath)
	}
	if debug {
		v.Set("debug", debug)
	}
	if image != "atidyshirt/kubernetes-embedded-test-runner-base:latest" {
		v.Set("image", image)
	}
	if testCommand != "" {
		v.Set("testCommand", testCommand)
	}
	if keepNamespace {
		v.Set("keepNamespace", keepNamespace)
	}
	if backoffLimit != 1 {
		v.Set("backoffLimit", backoffLimit)
	}
	if activeDeadline != 1800 {
		v.Set("activeDeadlineS", activeDeadline)
	}

	if !logPrefix {
		v.Set("logging.prefix", false)
	}
	if !logTimestamp {
		v.Set("logging.timestamp", false)
	}

	if debug {
		fmt.Printf("Setting config values:\n")
		fmt.Printf("  projectRoot: %s\n", projectRoot)
		fmt.Printf("  clusterWorkspacePath: %s\n", workspacePath)
		fmt.Printf("  image: %s\n", image)
		fmt.Printf("  testCommand: %s\n", testCommand)
	}

	cfg, err := config.LoadFromViper(v)
	if err != nil {
		cfg = &config.Config{
			Mode:            "launch",
			Image:           image,
			ProjectRoot:     projectRoot,
			WorkspacePath:   workspacePath,
			BackoffLimit:    backoffLimit,
			ActiveDeadlineS: activeDeadline,
			Debug:           debug,
			TestCommand:     testCommand,
			KeepNamespace:  keepNamespace,
		}
	}

	return cfg
}
