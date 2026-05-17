package cmd

import (
	"fmt"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/cleaner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

		s := scanner.New(logger, cfg.IgnorePatterns)
		c := cleaner.New(logger, cfg.DryRun)

		var allPaths []string

		for _, t := range cfg.Targets {
			target := scanner.Target{
				Name:        t.Name,
				Path:        t.Path,
				Threshold:   time.Duration(t.Threshold) * 24 * time.Hour,
				SafetyLevel: t.SafetyLevel,
				Type:        t.Type,
			}

			result, err := s.Scan(target, t.IgnorePatterns)
			if err != nil {
				logger.Error("Scan failed for target", zap.String("name", t.Name), zap.Error(err))
				continue
			}

			allPaths = append(allPaths, result.Files...)
		}

		if len(allPaths) == 0 {
			fmt.Println("No files found to clean.")
			return nil
		}

		fmt.Printf("Cleaning %d files...\n", len(allPaths))
		count, size, err := c.Clean(allPaths)
		if err != nil {
			return err
		}

		if cfg.DryRun {
			fmt.Printf("\nDRY RUN Summary: Would have deleted %d files, freeing %.2f MB\n", count, float64(size)/(1024*1024))
		} else {
			fmt.Printf("\nClean Summary: Deleted %d files, freed %.2f MB\n", count, float64(size)/(1024*1024))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
