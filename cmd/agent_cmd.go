package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the background agent",
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the background agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return InstallAgent()
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the background agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return UninstallAgent()
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return StartAgent()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		return StopAgent()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check background agent status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return StatusAgent()
	},
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Manage and view background agent logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		pathOnly, _ := cmd.Flags().GetBool("path")
		live, _ := cmd.Flags().GetBool("live")

		logPath, err := GetAgentLogPath()
		if err != nil {
			return err
		}

		if pathOnly {
			fmt.Println(logPath)
			return nil
		}

		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			return fmt.Errorf("agent log file does not exist yet: %s", logPath)
		}

		if live {
			return streamLog(logPath)
		}

		// Default behavior: show the last 20 lines
		content, err := os.ReadFile(logPath)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}
		lines := splitLines(string(content))
		start := len(lines) - 20
		if start < 0 {
			start = 0
		}
		for i := start; i < len(lines); i++ {
			fmt.Println(lines[i])
		}
		return nil
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the background agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := StopAgent(); err != nil {
			return err
		}
		return StartAgent()
	},
}

func init() {
	logCmd.Flags().Bool("path", false, "Show the path to the agent log file")
	logCmd.Flags().Bool("live", false, "Stream the agent log in real-time")

	agentCmd.AddCommand(installCmd)
	agentCmd.AddCommand(uninstallCmd)
	agentCmd.AddCommand(startCmd)
	agentCmd.AddCommand(stopCmd)
	agentCmd.AddCommand(statusCmd)
	agentCmd.AddCommand(restartCmd)
	agentCmd.AddCommand(logCmd)
	rootCmd.AddCommand(agentCmd)
}

// streamLog provides a simple cross-platform way to stream a file's content in real-time.
// It uses polling to detect new data, ensuring compatibility without external dependencies.
func streamLog(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Start at the end of the file
	_, err = file.Seek(0, 2)
	if err != nil {
		return err
	}

	fmt.Printf("Streaming logs from %s (Press Ctrl+C to stop)...\n", path)
	buffer := make([]byte, 4096)
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			fmt.Print(string(buffer[:n]))
		}
		if err != nil && err.Error() != "EOF" {
			return err
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func splitLines(s string) []string {
	var lines []string
	var line string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, line)
			line = ""
		} else if r != '\r' {
			line += string(r)
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
