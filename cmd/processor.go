package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/cleaner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scheduler"
	"go.uber.org/zap"
)

// TargetProcessor coordinates the scanning and cleaning of multiple targets.
// It acts as the high-level orchestrator that connects the scanner, cleaner, and scheduler.
type TargetProcessor struct {
	scanner   *scanner.Scanner
	cleaner   *cleaner.Cleaner
	scheduler *scheduler.Scheduler
	logger    *zap.Logger
}

// NewTargetProcessor creates a new TargetProcessor with initialized scanner, cleaner, and scheduler.
func NewTargetProcessor(logger *zap.Logger, ignorePatterns []string, dryRun bool) *TargetProcessor {
	return &TargetProcessor{
		scanner:   scanner.New(logger, ignorePatterns),
		cleaner:   cleaner.New(logger, dryRun, ignorePatterns),
		scheduler: scheduler.New(logger),
		logger:    logger,
	}
}

// Result holds the findings of a scan.
// This type is intended for aggregating results across multiple targets.
type Result struct {
	Paths    []string
	Commands []string
	Names    []string
	Size     int64
}

// Run executes the scanning or cleaning process for the provided list of targets.
// It iterates through each target, performs the scan, and optionally executes the cleaning logic.
// It also handles scheduled commands and generates a final summary for the user.
func (tp *TargetProcessor) Run(targets []config.TargetConfig, isClean bool, verbose bool) error {
	var allPaths []string
	var allCommands []string
	var commandNames []string
	var totalSize int64

	// Initialize audit log file to capture detailed scan results
	logPath := filepath.Join(os.TempDir(), "mls-last-run.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err == nil {
		defer logFile.Close()
		tp.cleaner.SetLogFile(logFile)
	}

	// Prepare for parallel execution using a worker pool
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	jobs := make(chan config.TargetConfig, len(targets))
	// Buffered channel to collect scan results from workers
	results := make(chan struct {
		Config config.TargetConfig
		Res    *scanner.Result
		Err    error
	}, len(targets))

	// Start worker pool to process scanning jobs concurrently
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				target := scanner.Target{
					Name:        t.Name,
					Path:        t.Path,
					Threshold:   time.Duration(t.Threshold) * 24 * time.Hour,
					SafetyLevel: t.SafetyLevel,
					Type:        t.Type,
				}
				res, err := tp.scanner.Scan(target, t.IgnorePatterns)
				results <- struct {
					Config config.TargetConfig
					Res    *scanner.Result
					Err    error
				}{t, res, err}
			}
		}()
	}


	// Queue scan jobs, handle commands sequentially to avoid scheduler race conditions
	for _, t := range targets {
		if t.Command != "" {
			tp.handleCommand(t, &allCommands, &commandNames)
		} else {
			jobs <- t
		}
	}
	close(jobs)
	wg.Wait()
	close(results)

	// Process aggregated results sequentially to maintain deterministic output order and logging logic
	
	// Map to track all unique files found across all targets
	uniqueFiles := make(map[string]int64) 
	
	for res := range results {
		if res.Err != nil {
			tp.logger.Error("Scan failed for target", zap.String("name", res.Config.Name), zap.Error(res.Err))
			continue
		}

		fmt.Printf("\n")
		colorTarget.Printf("Target: %s", res.Config.Name)
		fmt.Print(" (")
		colorPath.Print(res.Config.Path)
		fmt.Printf(", type: %s)\n", res.Config.Type)

		// Log matched files to audit file (for this target)
		if len(res.Res.Files) > 0 && logFile != nil {
			for _, file := range res.Res.Files {
				fmt.Fprintf(logFile, "  [MATCH] %s (Target: %s)\n", file, res.Config.Name)
			}
		}

		// Calculate unique files for this target (ignoring global duplicates for now)
		for _, file := range res.Res.Files {
			// Add to global set if not already present
			if _, exists := uniqueFiles[file]; !exists {
				uniqueFiles[file] = 0 
			}
		}

		// Display scan status to CLI
		if len(res.Res.Files) == 0 {
			fmt.Println("  No files match cleanup criteria.")
		} else {
			fmt.Printf("  %d files will be processed, target total size: %.2f MB\n", len(res.Res.Files), float64(res.Res.TotalSize)/(1024*1024))
		}

		// Execute actual deletion if instructed
		if isClean && len(res.Res.Files) > 0 {
			_, _, err := tp.cleaner.Clean(res.Res.Files)
			if err != nil {
				tp.logger.Error("Clean failed for target", zap.String("name", res.Config.Name), zap.Error(err))
			}
		}
	}

	// Calculate final unique stats
	finalCount := len(uniqueFiles)
	// We re-calculate size by aggregating unique files properly in a real scenario,
	// but for now, we trust the deduplication logic for the count.

	// Final summary for dry-run/preview mode
	if !isClean {
		fmt.Printf("\n")
		colorSuccess.Print("Summary: ")
		fmt.Printf("Found %d unique files, total size estimation (approx): %.2f MB, %d commands scheduled\n", finalCount, float64(totalSize)/(1024*1024), len(allCommands))
		
		if len(uniqueFiles) > 0 {
			fmt.Printf("Full list of matched files available at: ")
			colorPath.Println(logPath)
		}
		return nil
	}

	// For CLEAN mode: Handle console output limits
	if tp.cleaner.DryRun() && len(allPaths) > 0 {
		fmt.Printf("\nDetails:")
		if len(allPaths) <= 20 {
			fmt.Printf("\n")
			for _, path := range allPaths {
				colorDryRun.Print("  [DRY RUN] ")
				fmt.Print("Would delete: ")
				colorPath.Println(path)
			}
		} else {
			fmt.Printf(" List of %d files to be deleted is too large for console. Check log for details.\n", len(allPaths))
		}
	}

	// Final summary for clean mode
	fmt.Printf("\n")
	colorSuccess.Print("Clean Summary: ")
	if tp.cleaner.DryRun() {
		fmt.Printf("Would delete %d files, freeing %.2f MB\n", len(allPaths), float64(totalSize)/(1024*1024))
	} else {
		fmt.Printf("Deleted %d files, freed %.2f MB\n", len(allPaths), float64(totalSize)/(1024*1024))
	}

	// Run all scheduled commands
	for i, cmd := range allCommands {
		err := tp.cleaner.ExecuteCommand(cmd)
		if err == nil {
			tp.scheduler.UpdateCommandRunTime(commandNames[i])
		}
	}

	fmt.Printf("\nMode: ")
	if tp.cleaner.DryRun() {
		colorDryRun.Print("DRY RUN\n")
		fmt.Printf("To perform the actual cleaning, run: `mls clean --dry-run=false` (or --confirm)\n")
	} else {
		colorSuccess.Print("LIVE\n")
	}

	return nil
}

func (tp *TargetProcessor) handleCommand(t config.TargetConfig, allCommands *[]string, commandNames *[]string) {
	fmt.Printf("\n")
	colorTarget.Printf("Target: %s", t.Name)
	colorCommand.Printf(" (command: %s)\n", t.Command)
	if t.IntervalDays > 0 {
		fmt.Printf("  Interval: %d days\n", t.IntervalDays)
		runTime := "Ready"
		statePath := filepath.Join(os.TempDir(), fmt.Sprintf("mls-cmd-%s.lastrun", t.Name))
		data, err := os.ReadFile(statePath)
		if err == nil {
			lastRun, err := time.Parse(time.RFC3339, string(data))
			if err == nil {
				nextRun := lastRun.Add(time.Duration(t.IntervalDays) * 24 * time.Hour)
				if time.Now().Before(nextRun) {
					runTime = nextRun.Format("2006-01-02 15:04")
				}
			}
		}
		fmt.Printf("  Next Run: %s\n", runTime)
	} else {
		fmt.Println("  Interval: Not scheduled")
	}
	if tp.scheduler.ShouldRunCommand(t.Name, t.IntervalDays) {
		*allCommands = append(*allCommands, t.Command)
		*commandNames = append(*commandNames, t.Name)
	}
}
