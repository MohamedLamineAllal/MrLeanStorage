package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MrLeanStorage/internal/utils"
	"github.com/spf13/cobra"
)

// configCmd represents the base command for managing MrLeanStorage configurations.
// It acts as a grouping command for open, reveal, and reload subcommands.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

// openCmd represents the config open subcommand.
// It opens the default configuration YAML file in the system's default text editor.
// This allows users to inspect or manually edit their configurations quickly.
var openCmd = &cobra.Command{
	Use:          "open",
	Short:        "Open the configuration file",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetDefaultConfigPath()
		if err != nil {
			return err
		}
		return utils.OpenPath(path)
	},
}

// revealCmd represents the config reveal subcommand.
// It highlights/reveals the default configuration file in your system's file explorer
// (macOS Finder, Windows File Explorer, or opening the parent directory on Linux).
var revealCmd = &cobra.Command{
	Use:          "reveal",
	Short:        "Reveal the configuration file in your system's file explorer",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetDefaultConfigPath()
		if err != nil {
			return err
		}
		return utils.RevealPath(path)
	},
}

// reloadCmd represents the config reload subcommand.
// It signals all running background 'mls serve' processes to dynamically reload
// their configurations from the disk. This allows instant updates to targets/schedules
// cross-platform without having to restart the background services manually.
var reloadCmd = &cobra.Command{
	Use:          "reload",
	Short:        "Reload the configuration for the running background agent",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Write the cross-platform reload signal file
		sigPath := filepath.Join(utils.GetAppCacheDir(), "reload.signal")
		timestamp := time.Now().Format(time.RFC3339Nano)
		if err := os.WriteFile(sigPath, []byte(timestamp), 0644); err != nil {
			return fmt.Errorf("failed to write reload signal file: %w", err)
		}
		colorSuccess.Println("Reload signal file updated successfully.")

		// 2. Trigger platform-specific signaling (SIGHUP on Unix, no-op on Windows)
		if err := reloadProcesses(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(reloadCmd)
	configCmd.AddCommand(openCmd)
	configCmd.AddCommand(revealCmd)
	rootCmd.AddCommand(configCmd)
}
