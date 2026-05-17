package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MacosLeanStorage configuration",
}

var openConfigCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the configuration file in the default application",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			return fmt.Errorf("no configuration file found")
		}

		fmt.Printf("Opening configuration file in default application: %s\n", configFile)
		return exec.Command("open", configFile).Run()
	},
}

var revealConfigCmd = &cobra.Command{
	Use:   "reveal",
	Short: "Reveal the configuration file in Finder",
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := viper.ConfigFileUsed()
		if configFile == "" {
			return fmt.Errorf("no configuration file found")
		}

		fmt.Printf("Revealing configuration file in Finder: %s\n", configFile)
		return exec.Command("open", "-R", configFile).Run()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(openConfigCmd)
	configCmd.AddCommand(revealConfigCmd)
}
