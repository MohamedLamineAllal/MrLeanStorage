package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Targets  []TargetConfig `mapstructure:"targets"`
	DryRun   bool           `mapstructure:"dry_run"`
	Schedule string         `mapstructure:"schedule"` // Cron expression, e.g., "0 0 * * *" (daily at midnight)
}

// TargetConfig defines cleanup rules for a specific path
type TargetConfig struct {
	Name        string `mapstructure:"name"`
	Path        string `mapstructure:"path"`
	Threshold   int    `mapstructure:"threshold_days"`
	SafetyLevel int    `mapstructure:"safety_level"`
}

// Load reads the configuration from viper
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

// GetDefaultConfigPath returns the default path for the config file
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".MacosLeanStorage.yaml"), nil
}

// CreateDefaultConfig creates a skeleton config file if it doesn't exist
func CreateDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // Already exists
	}

	defaultConfig := `targets:
  - name: "VSCode Caches"
    path: "~/Library/Caches/com.microsoft.VSCode"
    threshold_days: 7
    safety_level: 1
  - name: "Chrome Caches"
    path: "~/Library/Caches/Google/Chrome/Default/Cache"
    threshold_days: 14
    safety_level: 1
dry_run: true
schedule: "0 0 * * *" # Daily at midnight
`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}
