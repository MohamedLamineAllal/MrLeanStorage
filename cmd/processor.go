package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/engine"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scheduler"
	"go.uber.org/zap"
)

// TargetProcessor acts as the CLI controller, utilizing the core engine library.
type TargetProcessor struct {
	engine    *engine.Engine
	scheduler *scheduler.Scheduler
	logger    *zap.Logger
}

// NewTargetProcessor initializes a new processor.
func NewTargetProcessor(logger *zap.Logger, ignorePatterns []string, dryRun bool) *TargetProcessor {
	return &TargetProcessor{
		engine:    engine.New(logger, ignorePatterns, dryRun),
		scheduler: scheduler.New(logger),
		logger:    logger,
	}
}

// getHooks returns the standard logging hooks for the CLI.
func (tp *TargetProcessor) getHooks(logFile *os.File) engine.Hooks {
	return engine.Hooks{
		OnTargetScanStart: func(name string, path string) {
			fmt.Printf("\n")
			colorTarget.Printf("Target: %s", name)
			fmt.Print(" (")
			colorPath.Print(path)
			fmt.Printf(")\n")
		},
		OnMatchFound: func(name string, files []string) {
			if logFile != nil {
				for _, file := range files {
					fmt.Fprintf(logFile, "  [MATCH] %s (Target: %s)\n", file, name)
				}
			}
		},
		OnFileCleaned: func(path string, freed int64, err error) {
			if err != nil {
				tp.logger.Error("Failed to delete", zap.String("path", path), zap.Error(err))
				return
			}
			if logFile != nil {
				prefix := ""
				if tp.engine.Cleaner().DryRun() {
					prefix = "[DRY RUN] "
				}
				fmt.Fprintf(logFile, "%sWould delete: %s\n", prefix, path)
			}
		},
	}
}

// Run executes the scanning and optional cleaning process.
func (tp *TargetProcessor) Run(targets []config.TargetConfig, isClean bool, verbose bool) error {
	var allCommands []string
	var commandNames []string

	logPath := filepath.Join(os.TempDir(), "mls-last-run.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		tp.logger.Error("Failed to create log file", zap.Error(err))
	} else {
		defer logFile.Close()
	}

	hooks := tp.getHooks(logFile)

	// Scan phase
	resultMap, err := tp.engine.Scan(targets, hooks)
	if err != nil {
		return err
	}

	// Iterate through targets to print scan summary per target
	for _, t := range targets {
		if t.Command != "" {
			tp.handleCommand(t, &allCommands, &commandNames)
			continue
		}

		res, ok := resultMap[t.Name]
		if !ok {
			continue
		}

		if len(res.Files) == 0 {
			fmt.Printf("\nTarget: %s - No files match cleanup criteria.", t.Name)
			continue
		}

		if isClean {
			fmt.Printf("\nTarget: %s - %d files will be processed, freeing %.2f MB", t.Name, len(res.Files), float64(res.TotalSize)/(1024*1024))
		} else {
			fmt.Printf("\nTarget: %s - Found %d files, total size: %.2f MB", t.Name, len(res.Files), float64(res.TotalSize)/(1024*1024))
		}
	}

	// Clean phase
	if isClean {
		uniqueCount, totalSize, err := tp.engine.Clean(resultMap, targets, hooks)
		if err != nil {
			return err
		}
		tp.printSummary(uniqueCount, totalSize, isClean, logPath)
	} else {
		// Calculate final unique stats for scan only
		aggregator := &engine.ResultAggregator{UniquePaths: make(map[string]int64)}
		for _, res := range resultMap {
			aggregator.Add(res.Files, res.FileSizes)
		}
		uniqueCount, totalSize := aggregator.GetStats()
		tp.printSummary(uniqueCount, totalSize, isClean, logPath)
	}

	return nil
}

func (tp *TargetProcessor) printSummary(count int, size int64, isClean bool, logPath string) {
	fmt.Printf("\n")
	if isClean {
		colorSuccess.Print("Clean Summary: ")
		if tp.engine.Cleaner().DryRun() {
			fmt.Printf("Would delete %d files, freeing %.2f MB\n", count, float64(size)/(1024*1024))
		} else {
			fmt.Printf("Deleted %d files, freed %.2f MB\n", count, float64(size)/(1024*1024))
		}
	} else {
		colorSuccess.Print("Summary: ")
		fmt.Printf("Found %d unique files, total size estimation (approx): %.2f MB\n", count, float64(size)/(1024*1024))
	}
	if count > 0 {
		fmt.Printf("Full log written to: ")
		colorPath.Println(logPath)
	}
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
