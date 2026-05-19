package cmd

import (
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

func init() {
	agentCmd.AddCommand(installCmd)
	agentCmd.AddCommand(uninstallCmd)
	agentCmd.AddCommand(startCmd)
	agentCmd.AddCommand(stopCmd)
	agentCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(agentCmd)
}
