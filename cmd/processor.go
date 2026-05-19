package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/cleaner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/engine"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scheduler"
	"go.uber.org/zap"
)

// TargetProcessor coordinates the scanning and cleaning of multiple targets.
type TargetProcessor struct {
	engine    *engine.Engine
	cleaner   *cleaner.Cleaner
	scheduler *scheduler.Scheduler
	logger    *zap.Logger
}

// NewTargetProcessor creates a new TargetProcessor.
func NewTargetProcessor(logger *zap.Logger, ignorePatterns []string, dryRun bool) *TargetProcessor {
	eng := engine.NewEngine(logger, ignorePatterns, dryRun)
	return &TargetProcessor{
		engine:    eng,
		cleaner:   eng.Cleaner(), // I need to add a Cleaner method to Engine. Oh, wait, I can just expose it or provide a getter. Let's add a Cleaner() method to Engine.
		scheduler: scheduler.New(logger),
		logger:    logger,
	}
}


// Run executes scanning and cleaning in order.
func (tp *TargetProcessor) Run(targets []config.TargetConfig, isClean bool, verbose bool) error {
	var allCommands []string
	var commandNames []string

	logPath := filepath.Join(os.TempDir(), "mls-last-run.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err == nil {
		defer logFile.Close()
		tp.cleaner.SetLogFile(logFile)
	}

	resultMap := tp.engine.ScanTargets(targets)
	aggregator := NewResultAggregator()

	for _, t := range targets {
		if t.Command != "" {
			tp.handleCommand(t, &allCommands, &commandNames)
			continue
		}

		res, ok := resultMap[t.Name]
		if !ok {
			continue
		}

		fmt.Printf("\n")
		colorTarget.Printf("Target: %s", t.Name)
		fmt.Print(" (")
		colorPath.Print(t.Path)
		fmt.Printf(", type: %s)\n", t.Type)

		if len(res.Files) > 0 && logFile != nil {
			for _, file := range res.Files {
				fmt.Fprintf(logFile, "  [MATCH] %s (Target: %s)\n", file, t.Name)
			}
		}

		aggregator.Add(res.Files, res.FileSizes)

		if len(res.Files) == 0 {
			fmt.Println("  No files match cleanup criteria.")
		} else {
			if isClean {
				fmt.Printf("  %d files will be deleted, freeing %.2f MB\n", len(res.Files), float64(res.TotalSize)/(1024*1024))
			} else {
				fmt.Printf("  Found %d files, total size: %.2f MB\n", len(res.Files), float64(res.TotalSize)/(1024*1024))
			}
		}

		if isClean && len(res.Files) > 0 {
			_, _, err := tp.cleaner.Clean(res.Files)
			if err != nil {
				tp.logger.Error("Clean failed for target", zap.String("name", t.Name), zap.Error(err))
			}
		}
	}

	uniqueCount, totalUniqueSize := aggregator.GetStats()
	allPaths := aggregator.GetUniquePaths()

	if !isClean {
		fmt.Printf("\n")
		colorSuccess.Print("Summary: ")
		fmt.Printf("Found %d unique files, total size estimation (approx): %.2f MB, %d commands scheduled\n", uniqueCount, float64(totalUniqueSize)/(1024*1024), len(allCommands))
		if uniqueCount > 0 {
			fmt.Printf("Full list of matched files available at: ")
			colorPath.Println(logPath)
		}
		return nil
	}

	if tp.cleaner.DryRun() && uniqueCount > 0 {
		fmt.Printf("\nDetails:")
		if uniqueCount <= 20 {
			fmt.Printf("\n")
			for _, path := range allPaths {
				colorDryRun.Print("  [DRY RUN] ")
				fmt.Print("Would delete: ")
				colorPath.Println(path)
			}
		} else {
			fmt.Printf(" List of %d files to be deleted is too large for console. Check log for details.\n", uniqueCount)
		}
	}

	fmt.Printf("\n")
	colorSuccess.Print("Clean Summary: ")
	if tp.cleaner.DryRun() {
		fmt.Printf("Would delete %d files, freeing %.2f MB\n", uniqueCount, float64(totalUniqueSize)/(1024*1024))
	} else {
		fmt.Printf("Deleted %d files, freed %.2f MB\n", uniqueCount, float64(totalUniqueSize)/(1024*1024))
	}

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

	fmt.Printf("\nFull log written to: ")
	colorPath.Println(logPath)

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
