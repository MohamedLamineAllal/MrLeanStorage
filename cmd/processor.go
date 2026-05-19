package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/engine"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scheduler"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
)

type TargetProcessor struct {
	engine    *engine.Engine
	scheduler *scheduler.Scheduler
	logger    *zap.Logger
}

func NewTargetProcessor(logger *zap.Logger, ignorePatterns []string, dryRun bool) *TargetProcessor {
	return &TargetProcessor{
		engine:    engine.New(logger, ignorePatterns, dryRun),
		scheduler: scheduler.New(logger),
		logger:    logger,
	}
}

func (tp *TargetProcessor) Run(targets []config.TargetConfig, isClean bool, verbose bool) error {
	logPath := filepath.Join(os.TempDir(), "mls-last-run.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		tp.logger.Error("Failed to create log file", zap.Error(err))
	} else {
		defer logFile.Close()
	}

	hooks := engine.Hooks{
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
			} else if logFile != nil {
				prefix := ""
				if tp.engine.Cleaner().DryRun() {
					prefix = "[DRY RUN] "
				}
				fmt.Fprintf(logFile, "%sDeleted: %s\n", prefix, path)
			}
		},
	}

	scanBar := progressbar.Default(-1, "Scanning targets...")
	resultMap, err := tp.engine.Scan(targets, hooks)
	scanBar.Finish()
	if err != nil {
		return err
	}

	aggregator := &engine.ResultAggregator{UniquePaths: make(map[string]int64)}
	for _, t := range targets {
		if t.Command != "" {
			continue
		}
		res, ok := resultMap[t.Name]
		if !ok {
			continue
		}
		aggregator.Add(res.Files, res.FileSizes)
		
		colorTarget.Printf("\nTarget: %s ", t.Name)
		colorPath.Printf("(%s)\n", t.Path)
		if len(res.Files) == 0 {
			fmt.Println("  No files match cleanup criteria.")
		} else {
			if isClean {
				fmt.Printf("  %d files to delete, freeing %.2f MB\n", len(res.Files), float64(res.TotalSize)/(1024*1024))
			} else {
				fmt.Printf("  Found %d files, total size: %.2f MB\n", len(res.Files), float64(res.TotalSize)/(1024*1024))
			}
		}
	}

	uniqueCount, totalSize := aggregator.GetStats()
	if isClean {
		cleanBar := progressbar.Default(-1, "Cleaning...")
		uniqueCount, totalSize, err = tp.engine.Clean(resultMap, targets, hooks)
		cleanBar.Finish()
		if err != nil {
			return err
		}
	}

	tp.printSummary(uniqueCount, totalSize, isClean, logPath)

	for _, t := range targets {
		if t.Command != "" {
			tp.handleCommand(t, nil, nil)
		}
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
		color.New(color.FgHiYellow, color.Underline).Println(logPath)
	}
}

func (tp *TargetProcessor) handleCommand(t config.TargetConfig, allCommands *[]string, commandNames *[]string) {
	fmt.Printf("\n")
	colorTarget.Printf("Target: %s", t.Name)
	colorCommand.Printf(" (command: %s)\n", t.Command)
	if t.IntervalDays > 0 {
		fmt.Printf("  Interval: %d days\n", t.IntervalDays)
	} else {
		fmt.Println("  Interval: Not scheduled")
	}
	if tp.scheduler.ShouldRunCommand(t.Name, t.IntervalDays) {
		err := tp.engine.Cleaner().ExecuteCommand(t.Command)
		if err == nil {
			tp.scheduler.UpdateCommandRunTime(t.Name)
		}
	}
}
