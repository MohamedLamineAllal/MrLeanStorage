// Package cmd implements the CLI commands for MrLeanStorage.
// It provides the entry point for the "serve" command which runs the cleanup agent.
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MrLeanStorage/internal/engine"
	"github.com/mohamedlamineallal/MrLeanStorage/internal/scheduler"
	"github.com/mohamedlamineallal/MrLeanStorage/internal/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// serveCmd represents the serve command which starts a background scheduler
// to perform cleanup tasks at regular intervals.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the background cleanup scheduler",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Enforce single-instance execution of serve using a cross-platform lockfile
		pidPath := filepath.Join(utils.GetAppCacheDir(), "mls.pid")
		lockFile, existingPid, err := utils.AcquireProcessLock(pidPath)
		if err != nil {
			if existingPid > 0 {
				fmt.Printf("Another mls serve process is already running (PID %d).\n", existingPid)
			} else {
				fmt.Println("Another mls serve process is already running.")
			}
			return nil
		}
		defer func() {
			if lockFile != nil {
				_ = lockFile.Close()
				_ = os.Remove(pidPath)
			}
		}()

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.Schedule == "" {
			return fmt.Errorf("no schedule configured")
		}

		s := scheduler.New(logger)

		// Define the core cleanup task using the unified Engine orchestration
		task := func() error {
			logger.Info("Starting scheduled cleanup")

			// Initialize the default Engine and CommandHandler with dryRun = false
			eng := engine.NewDefault(logger, cfg.IgnorePatterns, false)
			ch := engine.NewCommandHandler(eng, s, logger)
			eng.SetCommandHandler(ch)

			// Setup event-driven Hooks to log background execution progress
			hooks := engine.Hooks{
				OnTargetScanStart: func(name string, path string) {
					logger.Info("Scanning target", zap.String("name", name), zap.String("path", path))
				},
				OnFileCleaned: func(path string, freed int64, err error) {
					if err != nil {
						logger.Error("Failed to delete file", zap.String("path", path), zap.Error(err))
					} else {
						logger.Info("Deleted file", zap.String("path", path), zap.Int64("freed_bytes", freed))
					}
				},
				OnTargetCleaned: func(name string) {
					logger.Info("Target cleaned successfully", zap.String("target", name))
				},
				OnNoMatchesTargetCleanSkip: func(name string) {
					logger.Info("No files found to clean for target", zap.String("target", name))
				},
				BeforeExecutingCommand: func(name string, command string) {
					logger.Info("Executing Command for Target", zap.String("target", name), zap.String("command", command))
				},
				AfterExecutingCommand: func(name string, command string, err error) {
					if err != nil {
						logger.Error("Command failed", zap.String("target", name), zap.String("command", command), zap.Error(err))
					} else {
						logger.Info("Command completed successfully", zap.String("target", name), zap.String("command", command))
					}
				},
			}

			count, size, err := eng.ScanAndClean(cfg.Targets, hooks)
			if err != nil {
				return err
			}

			logger.Info("Scheduled cleanup finished", zap.Int("deleted_files_count", count), zap.Float64("freed_space_mb", float64(size)/(1024*1024)))
			return nil
		}

		// Schedule the task and handle missed executions
		err = s.AddTask(cfg.Schedule, task)
		if err != nil {
			return err
		}

		// Initialize the reload signal file state
		sigPath := filepath.Join(utils.GetAppCacheDir(), "reload.signal")
		var lastReloadTime time.Time
		if info, err := os.Stat(sigPath); err == nil {
			lastReloadTime = info.ModTime()
		} else {
			// If the file doesn't exist, establish a baseline modification time
			_ = os.WriteFile(sigPath, []byte(time.Now().Format(time.RFC3339Nano)), 0644)
			if info, err := os.Stat(sigPath); err == nil {
				lastReloadTime = info.ModTime()
			}
		}

		s.CheckForMissedTasks(task)
		s.Start()
		defer s.Stop()

		colorSuccess.Printf("Scheduler started with schedule: %s\n", cfg.Schedule)
		fmt.Println("Press Ctrl+C to stop")

		// Setup signal handling for graceful shutdown and configuration reloading
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, listenSignals...)

		// doneChan is used to coordinate graceful shutdown without channel races
		doneChan := make(chan struct{})

		// Periodic check for missed tasks (e.g., wake from sleep) every 30 minutes
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		// Periodic check for config reload signal file every 2 seconds
		reloadTicker := time.NewTicker(2 * time.Second)
		defer reloadTicker.Stop()

		// Background loop for signal handling, catch-up ticker, and reload file watcher
		go func() {
			for {
				select {
				case <-ticker.C:
					s.CheckForMissedTasks(task)
				case <-reloadTicker.C:
					if info, err := os.Stat(sigPath); err == nil {
						if info.ModTime().After(lastReloadTime) {
							lastReloadTime = info.ModTime()
							logger.Info("Reloading configuration (file signal detected)...")
							newCfg, err := config.Load()
							if err != nil {
								logger.Error("Failed to reload config", zap.Error(err))
								continue
							}
							cfg = newCfg
							logger.Info("Configuration reloaded successfully")
						}
					}
				case sig := <-sigChan:
					// Check if signal matches SIGHUP/reloadSignals
					isReload := false
					for _, rs := range reloadSignals {
						if sig == rs {
							isReload = true
							break
						}
					}
					if isReload {
						logger.Info("Reloading configuration (process signal SIGHUP received)...")
						newCfg, err := config.Load()
						if err != nil {
							logger.Error("Failed to reload config", zap.Error(err))
							continue
						}
						cfg = newCfg
						logger.Info("Configuration reloaded successfully")
					} else {
						// Signal the main thread to shut down on SIGINT or SIGTERM
						close(doneChan)
						return
					}
				}
			}
		}()

		// Block until shutdown signal
		<-doneChan

		colorWarning.Println("\nStopping scheduler...")
		return nil
	},
}

// init adds the serve command to the root command.
func init() {
	rootCmd.AddCommand(serveCmd)
}
