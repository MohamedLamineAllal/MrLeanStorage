//go:build linux

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/utils"
)

const (
	serviceName = "mls"
	unitTemplate = `[Unit]
Description=MrLeanStorage Background Agent
After=network.target

[Service]
ExecStart=%s serve
Restart=always
StandardOutput=append:%s
StandardError=append:%s

[Install]
WantedBy=default.target
`
)

// getUnitPath returns the path to the systemd user unit file for the agent.
// It ensures the necessary directory structure exists.
func getUnitPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(home, ".config/systemd/user")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create systemd user directory: %w", err)
	}
	return filepath.Join(dir, serviceName+".service"), nil
}

// getLogPath returns the standard path for the agent's log file on Linux.
func getLogPath() string {
	return filepath.Join(utils.GetAppCacheDir(), "agent.log")
}

// InstallAgent installs the background agent as a systemd user service on Linux.
func InstallAgent() error {
	unitPath, err := getUnitPath()
	if err != nil {
		return err
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	logPath := getLogPath()
	content := fmt.Sprintf(unitTemplate, executable, logPath, logPath)
	if err := os.WriteFile(unitPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Reload systemd and enable service
	if output, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w (output: %s)", err, string(output))
	}
	if output, err := exec.Command("systemctl", "--user", "enable", serviceName).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable agent: %w (output: %s)", err, string(output))
	}

	fmt.Printf("Agent installed: %s\n", unitPath)
	fmt.Printf("Agent logs redirected to: %s\n", logPath)
	return nil
}

// UninstallAgent removes the background agent's systemd service and configuration.
func UninstallAgent() error {
	unitPath, err := getUnitPath()
	if err != nil {
		return err
	}

	// Stop and disable service
	_ = exec.Command("systemctl", "--user", "stop", serviceName).Run()
	_ = exec.Command("systemctl", "--user", "disable", serviceName).Run()

	if err := os.Remove(unitPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove service file: %w", err)
		}
	}

	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	fmt.Printf("Agent uninstalled.\n")
	return nil
}

// StartAgent triggers the background agent's systemd service to start.
func StartAgent() error {
	if output, err := exec.Command("systemctl", "--user", "start", serviceName).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start agent: %w (output: %s)", err, string(output))
	}
	fmt.Printf("Agent started.\n")
	return nil
}

// StopAgent halts the background agent's systemd service.
func StopAgent() error {
	if output, err := exec.Command("systemctl", "--user", "stop", serviceName).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop agent: %w (output: %s)", err, string(output))
	}
	fmt.Printf("Agent stopped.\n")
	return nil
}

// StatusAgent queries the status of the background agent's systemd service.
func StatusAgent() error {
	output, err := exec.Command("systemctl", "--user", "status", serviceName).CombinedOutput()
	if err != nil {
		fmt.Printf("Agent is not running or not installed.\n")
		return nil
	}
	fmt.Printf("Agent status:\n%s\n", string(output))
	return nil
}

// GetAgentLogPath returns the path to the background agent log file on Linux.
func GetAgentLogPath() (string, error) {
	return getLogPath(), nil
}
