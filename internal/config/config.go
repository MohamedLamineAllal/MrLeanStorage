package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Targets        []TargetConfig `mapstructure:"targets"`
	IgnorePatterns []string       `mapstructure:"ignore_patterns"`
	DryRun         bool           `mapstructure:"dry_run"`
	Schedule       string         `mapstructure:"schedule"` // Cron expression, e.g., "0 0 * * *" (daily at midnight)
}

// TargetConfig defines cleanup rules for a specific path
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

// Load reads the configuration from viper
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Default Type to "file" if not specified
	for i := range cfg.Targets {
		if cfg.Targets[i].Type == "" {
			cfg.Targets[i].Type = "file"
		}
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

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfig := `targets:
  - name: "Arc Cache"
    path: "~/Library/Application Support/Arc/User Data/**/*Cache*/**"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Arc Cache"
    path: "~/Library/Application Support/Arc/User Data/**/CacheStorage/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Chrome Global Cache"
    path: "~/Library/Caches/Google/Chrome/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Chrome CacheStorage Cache"
    path: "~/Library/Application Support/Google/Chrome/**/CacheStorage/**"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Chrome crx Cache"
    path: "~/Library/Application Support/Google/GoogleUpdater/crx_cache/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Discord Cache"
    path: "~/Library/Application Support/discord/Cache/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Cursor CachedData"
    path: "~/Library/Application Support/Cursor/CachedData/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "VSCode CachedData"
    path: "~/Library/Application Support/Code/CachedData/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "VSCode Workspace Storage"
    path: "~/Library/Application Support/Code/User/workspaceStorage/*"
    threshold_days: 7
    safety_level: 2
    type: "both"
  - name: "OpenAI Atlas Cache"
    path: "~/Library/Caches/com.openai.atlas/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Telegram Cache"
    path: "~/Library/Caches/ru.keepcoder.Telegram/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Figma Local Storage"
    path: "~/Library/Application Support/Figma/Local Storage/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Spotify Cache"
    path: "~/Library/Caches/com.spotify.client/*"
    threshold_days: 3
    safety_level: 1
    type: "both"
  - name: "Go Build Cache"
    path: "~/Library/Caches/go-build/*"
    threshold_days: 7
    safety_level: 1
    type: "both"
  - name: "Homebrew Cache"
    path: "~/Library/Caches/Homebrew/*"
    threshold_days: 14
    safety_level: 1
    type: "both"
  - name: "npm/node-gyp"
    path: "~/Library/Caches/node-gyp/*"
    threshold_days: 14
    safety_level: 1
    type: "both"
  - name: "pip/pnpm Cache"
    path: "~/Library/Caches/pip/*"
    threshold_days: 14
    safety_level: 1
    type: "both"
dry_run: true
schedule: "0 0 * * *" # Daily at midnight
`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}
