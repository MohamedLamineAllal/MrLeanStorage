package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/cleaner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scheduler"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// serveCmd represents the serve command which starts a background scheduler
// to perform cleanup tasks at regular intervals.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the background cleanup scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.Schedule == "" {
			return fmt.Errorf("no schedule configured")
		}

		s := scheduler.New(logger)

		task := func() error {
			logger.Info("Starting scheduled cleanup")

			sc := scanner.New(logger, cfg.IgnorePatterns)
			cl := cleaner.New(logger, cfg.DryRun, cfg.IgnorePatterns)

			var allPaths []string
			for _, t := range cfg.Targets {
				target := scanner.Target{
					Name:        t.Name,
					Path:        t.Path,
					Threshold:   time.Duration(t.Threshold) * 24 * time.Hour,
					SafetyLevel: t.SafetyLevel,
					Type:        t.Type,
				}

				result, err := sc.Scan(target, t.IgnorePatterns)
				if err != nil {
					logger.Error("Scan failed", zap.String("target", t.Name), zap.Error(err))
					continue
				}
				allPaths = append(allPaths, result.Files...)
			}

			if len(allPaths) > 0 {
				_, _, err := cl.Clean(allPaths, nil)
				return err
			}
			logger.Info("No files found to clean")
			return nil
		}

		err = s.AddTask(cfg.Schedule, task)
		if err != nil {
			return err
		}

		s.CheckForMissedTasks(task)
		s.Start()
		defer s.Stop()

		colorSuccess.Printf("Scheduler started with schedule: %s\n", cfg.Schedule)
		fmt.Println("Press Ctrl+C to stop")

		// Wait for interruption
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Periodic check for missed tasks every 30 minutes
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		go func() {
			for {
				select {
				case <-ticker.C:
					s.CheckForMissedTasks(task)
				case <-sigChan:
					return
				}
			}
		}()

		<-sigChan

		colorWarning.Println("\nStopping scheduler...")
		return nil
	},
}

// init adds the serve command to the root command.
func init() {
	rootCmd.AddCommand(serveCmd)
}
