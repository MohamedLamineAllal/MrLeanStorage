package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application's root configuration structure.
// It includes global settings and a list of specific cleanup targets.
type Config struct {
	Targets        []TargetConfig `mapstructure:"targets"`
	IgnorePatterns []string       `mapstructure:"ignore_patterns"`
	DryRun         bool           `mapstructure:"dry_run"`
	Schedule       string         `mapstructure:"schedule"` // Cron expression, e.g., "0 0 * * *" (daily at midnight)
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

// Load unmarshals the configuration from Viper into the Config struct.
// It also applies default values for missing optional fields.
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

// GetDefaultConfigPath returns the standard absolute path for the application configuration file.
// It usually resolves to ~/.MacosLeanStorage.yaml.
func GetDefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".MacosLeanStorage.yaml"), nil
}

// CreateDefaultConfig generates a default configuration file with a predefined set of cleanup targets.
// If the file already exists, it does nothing.
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
  # Arc Browser
  - name: "Arc CacheStorage"
    path: "~/Library/Application Support/Arc/User Data/**/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Arc File System"
    path: "~/Library/Application Support/Arc/User Data/**/File System/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Arc IndexedDB"
    path: "~/Library/Application Support/Arc/User Data/**/IndexedDB/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Google Chrome
  - name: "Chrome Global Cache"
    path: "~/Library/Caches/Google/Chrome/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome CacheStorage"
    path: "~/Library/Application Support/Google/Chrome/**/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome Updater crx Cache"
    path: "~/Library/Application Support/Google/GoogleUpdater/crx_cache/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome File System"
    path: "~/Library/Application Support/Google/Chrome/**/File System/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome IndexedDB"
    path: "~/Library/Application Support/Google/Chrome/**/IndexedDB/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Communication Tools
  - name: "Discord Cache"
    path: "~/Library/Application Support/discord/Cache/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Slack CacheStorage"
    path: "~/Library/Containers/com.tinyspeck.slackmacgap/Data/Library/Application Support/Slack/Service Worker/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Microsoft Teams CacheStorage"
    path: "~/Library/Containers/com.microsoft.teams2/Data/Library/Application Support/Microsoft/MSTeams/**/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Development Tools
  - name: "Cursor Cache"
    path: "~/Library/Application Support/Cursor/Cache/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Cursor Service Worker Cache"
    path: "~/Library/Application Support/Cursor/Service Worker/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Cursor CachedExtensionVSIXs"
    path: "~/Library/Application Support/Cursor/CachedExtensionVSIXs/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "VSCode CachedData"
    path: "~/Library/Application Support/Code/CachedData/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "VSCode Service Worker Cache"
    path: "~/Library/Application Support/Code/Service Worker/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "VSCode WebStorage Cache"
    path: "~/Library/Application Support/Code/WebStorage/**/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "VSCode CachedExtensionVSIXs"
    path: "~/Library/Application Support/Code/CachedExtensionVSIXs/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # AI & Other
  - name: "OpenAI Atlas Cache"
    path: "~/Library/Caches/com.openai.atlas/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "OpenAI Atlas Service Worker Cache"
    path: "~/Library/Application Support/com.openai.atlas/**/Service Worker/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "OpenAI Atlas File System"
    path: "~/Library/Application Support/com.openai.atlas/**/File System/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "OpenAI Atlas Extensions Cache"
    path: "~/Library/Application Support/com.openai.atlas/**/extensions_crx_cache/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Telegram Cache"
    path: "~/Library/Caches/ru.keepcoder.Telegram/*"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Telegram Desktop Media Cache"
    path: "~/Library/Application Support/Telegram Desktop/tdata/user_data/media_cache/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Figma Local Storage"
    path: "~/Library/Application Support/Figma/Local Storage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Spotify Cache"
    path: "~/Library/Caches/com.spotify.client/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # System/Build
  - name: "Go Build Cache"
    path: "~/Library/Caches/go-build/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Homebrew Cache"
    path: "~/Library/Caches/Homebrew/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "npm/node-gyp"
    path: "~/Library/Caches/node-gyp/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "pip/pnpm Cache"
    path: "~/Library/Caches/pip/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Microsoft OneNote Backup"
    path: "~/Library/Containers/com.microsoft.onenote.mac/Data/Library/Application Support/Microsoft User Data/OneNote/*/Backup/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Application Updaters Cache"
    path: "~/Library/Application Support/Caches/**/updater/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Commands
  - name: "PNPM Store Prune"
    command: "pnpm store prune"
    interval_days: 30
    safety_level: 1
  - name: "npm clean cache"
    command: "npm cache clean --force"
    interval_days: 30
    safety_level: 1
dry_run: true
ignore_patterns:
  - ".DS_Store"
  - "._*"
  - ".Spotlight-V100"
  - ".Trashes"
  - ".fseventsd"
schedule: "0 0 * * *"
`
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}
