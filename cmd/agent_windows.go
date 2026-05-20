//go:build windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/utils"
)

const agentTaskName = "mls-agent"

// getLogPath returns the standard path for the agent's log file on Windows.
func getLogPath() string {
	return filepath.Join(utils.GetAppCacheDir(), "agent.log")
}

// InstallAgent installs the background agent as a Windows Scheduled Task.
// It configures the task to run at user logon and redirect output to a log file.
func InstallAgent() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	logPath := getLogPath()
	// Create a scheduled task that runs at logon.
	// On Windows, schtasks doesn't natively support output redirection,
	// so we wrap the command in a shell to handle redirection.
	taskCmd := fmt.Sprintf("cmd /c \"\"%s\" serve >> \"%s\" 2>&1\"", executable, logPath)

	// Create a scheduled task that runs at logon
	// /create: Create a new task
	// /tn: Task Name
	// /tr: Task Run (command)
	// /sc: Schedule (onlogon)
	// /f: Force (overwrite if exists)
	cmd := exec.Command("schtasks", "/create", "/tn", agentTaskName, "/tr", taskCmd, "/sc", "onlogon", "/f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install agent: %w (output: %s)", err, string(output))
	}

	fmt.Printf("Agent installed as scheduled task: %s\n", agentTaskName)
	fmt.Printf("Agent logs redirected to: %s\n", logPath)
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

// GetAgentLogPath returns the path to the background agent log file on Windows.
func GetAgentLogPath() (string, error) {
	return getLogPath(), nil
}
