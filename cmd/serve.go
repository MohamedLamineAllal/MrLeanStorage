// Package cmd implements the CLI commands for MacosLeanStorage.
// It provides the entry point for the "serve" command which runs the cleanup agent.
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

		// Define the core cleanup task
		task := func() error {
			logger.Info("Starting scheduled cleanup")

			sc := scanner.New(logger, cfg.IgnorePatterns)
			cl := cleaner.New(logger, cfg.DryRun, cfg.IgnorePatterns)

			var allPaths []string
			// Collect paths from all configured targets
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

			// Execute cleanup if files were found
			if len(allPaths) > 0 {
				_, _, err := cl.Clean(allPaths, nil)
				return err
			}
			logger.Info("No files found to clean")
			return nil
		}

		// Schedule the task and handle missed executions
		err = s.AddTask(cfg.Schedule, task)
		if err != nil {
			return err
		}

		s.CheckForMissedTasks(task)
		s.Start()
		defer s.Stop()

		colorSuccess.Printf("Scheduler started with schedule: %s\n", cfg.Schedule)
		fmt.Println("Press Ctrl+C to stop")

		// Setup signal handling for graceful shutdown and configuration reloading
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		// Periodic check for missed tasks (e.g., wake from sleep) every 30 minutes
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		// Background loop for signal handling and catch-up ticker
		go func() {
			for {
				select {
				case <-ticker.C:
					s.CheckForMissedTasks(task)
				case sig := <-sigChan:
					if sig == syscall.SIGHUP {
						logger.Info("Reloading configuration...")
						newCfg, err := config.Load()
						if err != nil {
							logger.Error("Failed to reload config", zap.Error(err))
							continue
						}
						cfg = newCfg
						logger.Info("Configuration reloaded successfully")
					} else {
						return
					}
				}
			}
		}()

		// Block until shutdown signal
		<-sigChan

		colorWarning.Println("\nStopping scheduler...")
		return nil
	},
}

// init adds the serve command to the root command.
func init() {
	rootCmd.AddCommand(serveCmd)
}
