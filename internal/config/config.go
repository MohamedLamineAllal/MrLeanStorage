package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config/defaults"
)

// Config represents the application's root configuration structure.
// It encapsulates global application settings and a list of cleanup targets to process.
type Config struct {
	Targets        []TargetConfig `mapstructure:"targets"`
	IgnorePatterns []string       `mapstructure:"ignore_patterns"`
	DryRun         bool           `mapstructure:"dry_run"`
	// Schedule is the cron expression (e.g., "0 0 0 * * *" for daily at midnight).
	Schedule string `mapstructure:"schedule"`
}

// TargetConfig defines the cleanup rules and metadata for a specific filesystem path or command.
type TargetConfig struct {
	Name           string   `mapstructure:"name"`
	Path           string   `mapstructure:"path"`
	Threshold      int      `mapstructure:"threshold_days"`
	IntervalDays   int      `mapstructure:"interval_days"`
	SafetyLevel    int      `mapstructure:"safety_level"`
	Type           string   `mapstructure:"type"` // "file", "folder", or "both"
	Command        string   `mapstructure:"command"`
	IgnorePatterns []string `mapstructure:"ignore_patterns"`
}

// Load unmarshals the configuration from the globally configured viper instance into the Config struct.
// It applies default values, such as setting the cleanup Type to "file" if left unspecified.
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure all targets have a valid Type
	for i := range cfg.Targets {
		if cfg.Targets[i].Type == "" {
			cfg.Targets[i].Type = "file"
		}
	}

	return &cfg, nil
}

// GetDefaultConfigPath returns the standard absolute path for the application configuration file,
// which defaults to ~/.MrLeanStorage.yaml.
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".MrLeanStorage.yaml"), nil
}

// CreateDefaultConfig generates a default configuration file with a predefined set of cleanup targets.
// If the configuration file already exists, this function returns nil without overwriting it.
func CreateDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // File already exists
	}

	// Ensure parent directory exists before creation
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(path, []byte(defaults.GetDefaultConfig()), 0644)
}
