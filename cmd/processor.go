package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/cleaner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scheduler"
	"go.uber.org/zap"
)

type TargetProcessor struct {
	scanner   *scanner.Scanner
	cleaner   *cleaner.Cleaner
	scheduler *scheduler.Scheduler
	logger    *zap.Logger
}

func NewTargetProcessor(logger *zap.Logger, ignorePatterns []string, dryRun bool) *TargetProcessor {
	return &TargetProcessor{
		scanner:   scanner.New(logger, ignorePatterns),
		cleaner:   cleaner.New(logger, dryRun, ignorePatterns),
		scheduler: scheduler.New(logger),
		logger:    logger,
	}
}

// Result holds the findings of a scan
type Result struct {
	Paths    []string
	Commands []string
	Names    []string
	Size     int64
}

func (tp *TargetProcessor) Run(targets []config.TargetConfig, isClean bool, verbose bool) error {
	var allPaths []string
	var allCommands []string
	var commandNames []string
	var totalSize int64

	// Initialize log file
	logPath := filepath.Join(os.TempDir(), "mls-last-run.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err == nil {
		defer logFile.Close()
		tp.cleaner.SetLogFile(logFile)
	}

	for _, t := range targets {
		if t.Command != "" {
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
				allCommands = append(allCommands, t.Command)
				commandNames = append(commandNames, t.Name)
			}
			continue
		}

		target := scanner.Target{
			Name:        t.Name,
			Path:        t.Path,
			Threshold:   time.Duration(t.Threshold) * 24 * time.Hour,
			SafetyLevel: t.SafetyLevel,
			Type:        t.Type,
		}

		result, err := tp.scanner.Scan(target, t.IgnorePatterns)
		if err != nil {
			tp.logger.Error("Scan failed for target", zap.String("name", t.Name), zap.Error(err))
			continue
		}

		fmt.Printf("\n")
		colorTarget.Printf("Target: %s", t.Name)
		fmt.Print(" (")
		colorPath.Print(t.Path)
		fmt.Printf(", type: %s)\n", t.Type)

		if len(result.Files) == 0 {
			fmt.Println("  No files match cleanup criteria.")
		} else {
			// Group matches by parent directory for smarter organization in the log
			dirGroups := make(map[string][]string)
			for _, p := range result.Files {
				parent := filepath.Dir(p)
				dirGroups[parent] = append(dirGroups[parent], p)
			}

			// Maintain order for the log
			parents := make([]string, 0, len(dirGroups))
			for k := range dirGroups {
				parents = append(parents, k)
			}
			for i := 0; i < len(parents); i++ {
				for j := i + 1; j < len(parents); j++ {
					if parents[i] > parents[j] {
						parents[i], parents[j] = parents[j], parents[i]
					}
				}
			}

			displayCount := 0
			maxDisplay := 5
			
			for _, parent := range parents {
				group := dirGroups[parent]
				for _, file := range group {
					// Always log to file if it was initialized
					if logFile != nil {
						fmt.Fprintf(logFile, "  [MATCH] %s\n", file)
					}

					// Truncated display to console (ONLY in scan mode)
					if !isClean {
						if verbose || displayCount < maxDisplay {
							colorMatch.Print("  [MATCH] ")
							fmt.Println(file)
							displayCount++
						} else if displayCount == maxDisplay {
							fmt.Printf("    ... and %d more matches (see log for full list)\n", len(result.Files)-maxDisplay)
							displayCount++ // only print the summary once
						}
					}
				}
			}
			if !isClean {
				fmt.Printf("  Total size: %.2f MB\n", float64(result.TotalSize)/(1024*1024))
			}
		}

		allPaths = append(allPaths, result.Files...)
		totalSize += result.TotalSize

		// Perform cleaning for this target if in clean mode
		if isClean && len(result.Files) > 0 {
			count, size, err := tp.cleaner.Clean(result.Files)
			if err != nil {
				tp.logger.Error("Clean failed for target", zap.String("name", t.Name), zap.Error(err))
			}
			_ = count // aggregate summary at the end using allPaths and totalSize
			_ = size
		}
	}

	if !isClean {
		fmt.Printf("\n")
		colorSuccess.Print("Summary: ")
		fmt.Printf("Found %d files, total size: %.2f MB, %d commands scheduled\n", len(allPaths), float64(totalSize)/(1024*1024), len(allCommands))
		return nil
	}

	// Final summary for clean mode
	fmt.Printf("\n")
	colorSuccess.Print("Clean Summary: ")
	if tp.cleaner.DryRun() {
		fmt.Printf("Would delete %d files, freeing %.2f MB\n", len(allPaths), float64(totalSize)/(1024*1024))
	} else {
		fmt.Printf("Deleted %d files, freed %.2f MB\n", len(allPaths), float64(totalSize)/(1024*1024))
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
		fmt.Printf("If you want to perform the cleaning, run: `mls clean --confirm` (or use --dry-run=false)\n")
	} else {
		colorSuccess.Print("LIVE\n")
	}

	logPath = filepath.Join(os.TempDir(), "mls-last-run.log")
	fmt.Printf("\nFull log written to: ")
	colorPath.Println(logPath)

	return nil
}
