package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/config"
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
		return exec.Command("open", path).Run()
	},
}

// revealCmd represents the config reveal subcommand.
// It highlights/reveals the default configuration file in macOS Finder,
// which is useful for dragging the configuration file, copying it, or opening it via custom programs.
var revealCmd = &cobra.Command{
	Use:          "reveal",
	Short:        "Reveal the configuration file in Finder",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.GetDefaultConfigPath()
		if err != nil {
			return err
		}
		return exec.Command("open", "-R", path).Run()
	},
}

// reloadCmd represents the config reload subcommand.
// It signals all running background 'mls serve' processes to dynamically reload
// their configurations from the disk. This allows instant updates to targets/schedules
// without having to restart the macOS daemon manually.
var reloadCmd = &cobra.Command{
	Use:          "reload",
	Short:        "Reload the configuration for the running background agent",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find the PIDs of all running mls serve processes
		pids, err := findMLSServePIDs()
		if err != nil {
			return err
		}

		var signalErrors []string
		for _, pid := range pids {
			if pid == os.Getpid() {
				continue
			}
			process, err := os.FindProcess(pid)
			if err != nil {
				signalErrors = append(signalErrors, fmt.Sprintf("PID %d: %v", pid, err))
				continue
			}
			if err := process.Signal(syscall.SIGHUP); err != nil {
				signalErrors = append(signalErrors, fmt.Sprintf("PID %d: %v", pid, err))
			} else {
				colorSuccess.Printf("Reload signal sent to process %d successfully\n", pid)
			}
		}

		if len(signalErrors) > 0 {
			return fmt.Errorf("failed to signal some processes: %s", strings.Join(signalErrors, "; "))
		}
		return nil
	},
}

// findMLSServePIDs queries the system using pgrep to find all process IDs (PIDs)
// matching "mls serve". It filters out empty lines and invalid PIDs to ensure
// signal execution targets only valid running server processes.
func findMLSServePIDs() ([]int, error) {
	out, err := exec.Command("pgrep", "-f", "mls serve").Output()
	if err != nil {
		// pgrep exits with status 1 if no process matches
		return nil, fmt.Errorf("no running mls serve process found")
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var pids []int
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err == nil {
			pids = append(pids, pid)
		}
	}
	if len(pids) == 0 {
		return nil, fmt.Errorf("no running mls serve process found")
	}
	return pids, nil
}

func init() {
	configCmd.AddCommand(reloadCmd)
	configCmd.AddCommand(openCmd)
	configCmd.AddCommand(revealCmd)
	rootCmd.AddCommand(configCmd)
}
