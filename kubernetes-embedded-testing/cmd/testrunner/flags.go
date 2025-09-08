package main

import (
	"fmt"
	"testrunner/pkg/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getShortFlag returns the short flag character for a given flag name
func getShortFlag(flagName string) string {
	shortFlags := map[string]string{
		"project-root":            "r",
		"cluster-workspace-path":  "w",
		"debug":                   "v",
		"image":                   "i",
		"test-command":            "t",
		"keep-namespace":          "k",
		"backoff-limit":           "b",
		"active-deadline-seconds": "d",
	}
	if short, exists := shortFlags[flagName]; exists {
		return short
	}
	return ""
}

func addRootFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("config", "", "Path to config file (YAML/JSON)")

	for flagName, config := range FlagMapping.RootFlags {
		switch v := config.Default.(type) {
		case string:
			cmd.PersistentFlags().StringP(flagName, getShortFlag(flagName), v, config.Description)
		case bool:
			cmd.PersistentFlags().BoolP(flagName, getShortFlag(flagName), v, config.Description)
		case int32:
			cmd.PersistentFlags().Int32P(flagName, getShortFlag(flagName), v, config.Description)
		case int64:
			cmd.PersistentFlags().Int64P(flagName, getShortFlag(flagName), v, config.Description)
		}
	}
}

func addLaunchFlags(cmd *cobra.Command) {
	for flagName, config := range FlagMapping.LaunchFlags {
		switch v := config.Default.(type) {
		case string:
			cmd.Flags().StringP(flagName, getShortFlag(flagName), v, config.Description)
		case bool:
			cmd.Flags().BoolP(flagName, getShortFlag(flagName), v, config.Description)
		case int32:
			cmd.Flags().Int32P(flagName, getShortFlag(flagName), v, config.Description)
		case int64:
			cmd.Flags().Int64P(flagName, getShortFlag(flagName), v, config.Description)
		}
	}
}

func setupViper(cmd *cobra.Command) *viper.Viper {
	v := viper.New()

	// Get config file from command flag
	configFile, _ := cmd.PersistentFlags().GetString("config")
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

func buildConfig(cmd *cobra.Command) *config.Config {
	v := setupViper(cmd)
	v.SetDefault("mode", "launch")

	for _, config := range FlagMapping.RootFlags {
		v.SetDefault(config.ViperKey, config.Default)
	}

	for _, config := range FlagMapping.LaunchFlags {
		v.SetDefault(config.ViperKey, config.Default)
	}

	bindFlagsToViper(v, cmd)

	cfg, err := config.LoadFromViper(v)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	return cfg
}
