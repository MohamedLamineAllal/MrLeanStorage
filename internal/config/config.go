package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Targets []TargetConfig `mapstructure:"targets"`
	DryRun  bool           `mapstructure:"dry_run"`
}

type TargetConfig struct {
	Path        string `mapstructure:"path"`
	Threshold   int    `mapstructure:"threshold_days"`
	SafetyLevel int    `mapstructure:"safety_level"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
