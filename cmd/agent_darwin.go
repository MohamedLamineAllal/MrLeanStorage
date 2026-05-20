//go:build darwin

// Package cmd implements the CLI commands for MrLeanStorage.
// It includes logic for background agent management using launchd and CLI configuration.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/utils"
)

const (
	// agentLabel is the identifier used for the launchd service.
	agentLabel = "com.mls.serve"
	// agentPlist is the template for the launchd property list file.
	// It includes StandardOutPath and StandardErrorPath for logging.
	agentPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>serve</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s</string>
    <key>StandardErrorPath</key>
    <string>%s</string>
</dict>
</plist>`
)

// getPlistPath returns the standard path for the agent's launchd plist file in the user's Library.
func getPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library/LaunchAgents", agentLabel+".plist")
}

// getLogPath returns the standard path for the agent's log file.
func getLogPath() string {
	return filepath.Join(utils.GetAppCacheDir(), "agent.log")
}

// getAgentUninstallCommand returns the command to unload/bootout the agent from launchd.
func getAgentUninstallCommand(plistPath string) *exec.Cmd {
	return exec.Command("launchctl", "bootout", "gui/"+fmt.Sprint(os.Getuid()), plistPath)
}

// getAgentLoadCommand returns the command to load the agent into launchd.
func getAgentLoadCommand(plistPath string) *exec.Cmd {
	return exec.Command("launchctl", "load", plistPath)
}

// InstallAgent installs the background launch agent, generates the plist, and loads it into launchd.
func InstallAgent() error {
	plistPath := getPlistPath()
	logPath := getLogPath()
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(agentPlist, agentLabel, executable, logPath, logPath)
	if err := os.WriteFile(plistPath, []byte(content), 0644); err != nil {
		return err
	}

	// Load the newly created plist file
	if err := getAgentLoadCommand(plistPath).Run(); err != nil {
		return fmt.Errorf("failed to load agent: %w", err)
	}
	fmt.Printf("Agent installed: %s\n", plistPath)
	fmt.Printf("Agent logs redirected to: %s\n", logPath)
	return nil
}

// UninstallAgent removes the background launch agent, unloads it, and deletes the plist file.
func UninstallAgent() error {
	plistPath := getPlistPath()

	// Unload the service before removing the file
	if err := getAgentUninstallCommand(plistPath).Run(); err != nil {
		return fmt.Errorf("failed to uninstall agent: %w", err)
	}
	os.Remove(plistPath)

	fmt.Printf("Agent uninstalled.\n")
	return nil
}

// StartAgent triggers the background launch agent to run using the launchctl kickstart command.
func StartAgent() error {
	plistPath := getPlistPath()

	// Load the service if it's not currently loaded
	exec.Command("launchctl", "load", plistPath).Run()

	// Kickstart the service
	cmd := exec.Command("launchctl", "kickstart", "-k", "gui/"+fmt.Sprint(os.Getuid())+"/"+agentLabel)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}
	fmt.Printf("Agent started.\n")
	return nil
}

// StopAgent halts the background launch agent using launchctl.
func StopAgent() error {
	plistPath := getPlistPath()

	// Use bootout to stop the service
	if err := getAgentUninstallCommand(plistPath).Run(); err != nil {
		return fmt.Errorf("failed to stop agent: %w", err)
	}
	fmt.Printf("Agent stopped.\n")
	return nil
}

// StatusAgent queries the launchd service status for the agent.
func StatusAgent() error {
	cmd := exec.Command("launchctl", "print", "gui/"+fmt.Sprint(os.Getuid())+"/"+agentLabel)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Agent is not running or not installed.\n")
		return nil
	}
	fmt.Printf("Agent status:\n%s\n", string(output))
	return nil
}

// GetAgentLogPath returns the path to the background agent log file on Darwin.
func GetAgentLogPath() (string, error) {
	return getLogPath(), nil
}
