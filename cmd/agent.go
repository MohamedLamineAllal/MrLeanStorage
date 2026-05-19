package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	agentLabel = "com.mls.serve"
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
</dict>
</plist>`
)

// InstallAgent installs the background launch agent.
func InstallAgent() error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library/LaunchAgents", agentLabel+".plist")
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	content := fmt.Sprintf(agentPlist, agentLabel, executable)
	if err := os.WriteFile(plistPath, []byte(content), 0644); err != nil {
		return err
	}

	exec.Command("launchctl", "load", plistPath).Run()
	fmt.Printf("Agent installed: %s\n", plistPath)
	return nil
}

// UninstallAgent removes the background launch agent.
func UninstallAgent() error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library/LaunchAgents", agentLabel+".plist")
	
	exec.Command("launchctl", "bootout", "gui/"+fmt.Sprint(os.Getuid()), plistPath).Run()
	os.Remove(plistPath)
	
	fmt.Printf("Agent uninstalled.\n")
	return nil
}
func StartAgent() error {
	cmd := exec.Command("launchctl", "kickstart", "-k", "gui/"+fmt.Sprint(os.Getuid())+"/"+agentLabel)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}
	fmt.Printf("Agent started.\n")
	return nil
}

// StopAgent stops the background launch agent.
func StopAgent() error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library/LaunchAgents", agentLabel+".plist")
	
	cmd := exec.Command("launchctl", "bootout", "gui/"+fmt.Sprint(os.Getuid()), plistPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop agent: %w", err)
	}
	fmt.Printf("Agent stopped.\n")
	return nil
}

// StatusAgent checks the status of the agent.
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
