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
  - name: "Arc Cache"
    path: "~/Library/Application Support/Arc/User Data/*/Cache"
    threshold_days: 3
    safety_level: 1
  - name: "Chrome Global Cache"
    path: "~/Library/Caches/Google/Chrome"
    threshold_days: 3
    safety_level: 1
  - name: "Discord Cache"
    path: "~/Library/Application Support/discord/Cache"
    threshold_days: 3
    safety_level: 1
  - name: "Cursor CachedData"
    path: "~/Library/Application Support/Cursor/CachedData"
    threshold_days: 3
    safety_level: 1
  - name: "VSCode CachedData"
    path: "~/Library/Application Support/Code/CachedData"
    threshold_days: 3
    safety_level: 1
  - name: "VSCode Workspace Storage"
    path: "~/Library/Application Support/Code/User/workspaceStorage"
    threshold_days: 7
    safety_level: 2
  - name: "OpenAI Atlas Cache"
    path: "~/Library/Caches/com.openai.atlas"
    threshold_days: 3
    safety_level: 1
  - name: "Telegram Cache"
    path: "~/Library/Caches/ru.keepcoder.Telegram"
    threshold_days: 3
    safety_level: 1
  - name: "Figma Local Storage"
    path: "~/Library/Application Support/Figma/Local Storage"
    threshold_days: 3
    safety_level: 1
  - name: "Spotify Cache"
    path: "~/Library/Caches/com.spotify.client"
    threshold_days: 3
    safety_level: 1
  - name: "Go Build Cache"
    path: "~/Library/Caches/go-build"
    threshold_days: 7
    safety_level: 1
  - name: "Homebrew Cache"
    path: "~/Library/Caches/Homebrew"
    threshold_days: 14
    safety_level: 1
  - name: "npm/node-gyp"
    path: "~/Library/Caches/node-gyp"
    threshold_days: 14
    safety_level: 1
  - name: "pip/pnpm Cache"
    path: "~/Library/Caches/pip"
    threshold_days: 14
    safety_level: 1
dry_run: true
schedule: "0 0 * * *" # Daily at midnight
`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}
