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
		cleaner:   cleaner.New(logger, dryRun),
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

	for _, t := range targets {
		if t.Command != "" {
			fmt.Printf("\nTarget: %s (command: %s)\n", t.Name, t.Command)
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

		fmt.Printf("\nTarget: %s (%s, type: %s)\n", t.Name, t.Path, t.Type)
		if len(result.Files) == 0 {
			fmt.Println("  No files match cleanup criteria.")
		} else {
			if verbose || len(result.Files) <= 10 {
				for _, file := range result.Files {
					fmt.Printf("  [MATCH] %s\n", file)
				}
			} else {
				fmt.Printf("  Found %d matches (use --verbose to list all)\n", len(result.Files))
			}
			fmt.Printf("  Total size: %.2f MB\n", float64(result.TotalSize)/(1024*1024))
		}

		allPaths = append(allPaths, result.Files...)
		totalSize += result.TotalSize
	}

	if !isClean {
		fmt.Printf("\nSummary: Found %d files, total size: %.2f MB, %d commands scheduled\n", len(allPaths), float64(totalSize)/(1024*1024), len(allCommands))
		return nil
	}
	// ... (rest of the cleaner logic)

	// Perform cleaning
	if len(allPaths) > 0 {
		fmt.Printf("Cleaning %d files...\n", len(allPaths))
		count, size, err := tp.cleaner.Clean(allPaths)
		if err != nil {
			return err
		}
		fmt.Printf("Clean Summary: Deleted %d files, freed %.2f MB\n", count, float64(size)/(1024*1024))
	}

	for i, cmd := range allCommands {
		err := tp.cleaner.ExecuteCommand(cmd)
		if err == nil {
			tp.scheduler.UpdateCommandRunTime(commandNames[i])
		}
	}

	fmt.Printf("Mode: %s\n", map[bool]string{true: "DRY RUN", false: "LIVE"}[tp.cleaner.DryRun()])
	fmt.Printf("If you want to perform the cleaning, run: `mls clean --dry-run=false`")
	return nil
}
