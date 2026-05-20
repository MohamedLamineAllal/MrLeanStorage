//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

const agentTaskName = "mls-agent"

// InstallAgent installs the background agent as a Windows Scheduled Task.
// It configures the task to run at user logon.
func InstallAgent() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// Create a scheduled task that runs at logon
	// /create: Create a new task
	// /tn: Task Name
	// /tr: Task Run (command)
	// /sc: Schedule (onlogon)
	// /f: Force (overwrite if exists)
	cmd := exec.Command("schtasks", "/create", "/tn", agentTaskName, "/tr", fmt.Sprintf("\"%s\" serve", executable), "/sc", "onlogon", "/f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install agent: %w (output: %s)", err, string(output))
	}

	fmt.Printf("Agent installed as scheduled task: %s\n", agentTaskName)
	return nil
}

// UninstallAgent removes the background agent's scheduled task from Windows.
func UninstallAgent() error {
	cmd := exec.Command("schtasks", "/delete", "/tn", agentTaskName, "/f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to uninstall agent: %w (output: %s)", err, string(output))
	}
	fmt.Printf("Agent uninstalled.\n")
	return nil
}

// StartAgent triggers the background agent's scheduled task to run immediately.
func StartAgent() error {
	cmd := exec.Command("schtasks", "/run", "/tn", agentTaskName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start agent: %w (output: %s)", err, string(output))
	}
	fmt.Printf("Agent started.\n")
	return nil
}

// StopAgent halts the background agent's scheduled task.
func StopAgent() error {
	cmd := exec.Command("schtasks", "/end", "/tn", agentTaskName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop agent: %w (output: %s)", err, string(output))
	}
	fmt.Printf("Agent stopped.\n")
	return nil
}

// StatusAgent queries the status of the background agent's scheduled task.
func StatusAgent() error {
	// /query: Query tasks
	// /tn: Task Name
	// /fo: Format (LIST)
	cmd := exec.Command("schtasks", "/query", "/tn", agentTaskName, "/fo", "LIST")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Agent is not running or not installed.\n")
		return nil
	}
	fmt.Printf("Agent status:\n%s\n", string(output))
	return nil
}
