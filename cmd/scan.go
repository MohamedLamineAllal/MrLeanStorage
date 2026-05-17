package cmd

import (
	"fmt"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

		s := scanner.New(logger)

		totalFiles := 0
		var totalSize int64

		for _, t := range cfg.Targets {
			target := scanner.Target{
				Name:        t.Name,
				Path:        t.Path,
				Threshold:   time.Duration(t.Threshold) * 24 * time.Hour,
				SafetyLevel: t.SafetyLevel,
				Type:        t.Type,
			}

			result, err := s.Scan(target)
			if err != nil {
				logger.Error("Scan failed for target", zap.String("name", t.Name), zap.Error(err))
				continue
			}

			verbose, _ := cmd.Flags().GetBool("verbose")

			fmt.Printf("\nTarget: %s (%s, type: %s)\n", result.TargetName, t.Path, t.Type)
			if len(result.Files) == 0 {
				fmt.Println("  No files match cleanup criteria.")
				continue
			}

			if verbose || len(result.Files) <= 10 {
				for _, file := range result.Files {
					fmt.Printf("  [MATCH] %s\n", file)
				}
			} else {
				fmt.Printf("  Found %d matches (use --verbose to list all)\n", len(result.Files))
			}
			fmt.Printf("  Total size: %.2f MB\n", float64(result.TotalSize)/(1024*1024))

			totalFiles += len(result.Files)
			totalSize += result.TotalSize
		}

		fmt.Printf("\nSummary: Found %d items across all targets, total size: %.2f MB\n", totalFiles, float64(totalSize)/(1024*1024))
		if cfg.DryRun {
			fmt.Println("Running in DRY RUN mode. No files were deleted.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
