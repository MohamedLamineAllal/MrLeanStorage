package cmd

import (
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Scan and clean old files",
	Long:  `Scans the configured targets and deletes files that exceed the age threshold.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		processor := NewTargetProcessor(logger, cfg.IgnorePatterns, cfg.DryRun)
		return processor.Run(cfg.Targets, true)
	},

}

func init() {
rootCmd.AddCommand(cleanCmd)
}

