package cmd

import (
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan targets for old files",
	Long:  `Scans the configured targets and lists files that exceed the age threshold.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		processor := NewTargetProcessor(logger, cfg.IgnorePatterns, cfg.DryRun)
		return processor.Run(cfg.Targets, false)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
