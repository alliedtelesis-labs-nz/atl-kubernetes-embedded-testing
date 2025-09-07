package config

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
)

type LoggingConfig struct {
	Prefix    bool `mapstructure:"prefix" yaml:"prefix" json:"prefix"`
	Timestamp bool `mapstructure:"timestamp" yaml:"timestamp" json:"timestamp"`
}

type Config struct {
	Mode            string          `mapstructure:"mode" yaml:"mode" json:"mode"`
	ProjectRoot     string          `mapstructure:"projectRoot" yaml:"projectRoot" json:"projectRoot"`
	Image           string          `mapstructure:"image" yaml:"image" json:"image"`
	Debug           bool            `mapstructure:"debug" yaml:"debug" json:"debug"`
	TestCommand     string          `mapstructure:"testCommand" yaml:"testCommand" json:"testCommand"`
	KeepNamespace  bool            `mapstructure:"keepNamespace" yaml:"keepNamespace" json:"keepNamespace"`
	BackoffLimit    int32           `mapstructure:"backoffLimit" yaml:"backoffLimit" json:"backoffLimit"`
	ActiveDeadlineS int64           `mapstructure:"activeDeadlineS" yaml:"activeDeadlineS" json:"activeDeadlineS"`
	WorkspacePath   string          `mapstructure:"clusterWorkspacePath" yaml:"clusterWorkspacePath" json:"clusterWorkspacePath"`
	Logging         LoggingConfig   `mapstructure:"logging" yaml:"logging" json:"logging"`
	Ctx             context.Context `mapstructure:"-" yaml:"-" json:"-"`
}

func LoadFromFile(path string) (*Config, error) {
	v := viper.New()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("ket-config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.ket")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func LoadFromViper(v *viper.Viper) (*Config, error) {
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
