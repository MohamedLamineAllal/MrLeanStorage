package cmd

import (
	"fmt"
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

func (tp *TargetProcessor) Run(targets []config.TargetConfig, isClean bool) error {
	var allPaths []string
	var allCommands []string
	var commandNames []string
	var totalSize int64

	for _, t := range targets {
		if t.Command != "" {
			if tp.scheduler.ShouldRunCommand(t.Name, t.IntervalDays) {
				allCommands = append(allCommands, t.Command)
				commandNames = append(commandNames, t.Name)
			} else {
				tp.logger.Info("Skipping command target (interval not met)", zap.String("name", t.Name))
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

		// Print scan details for both commands
		fmt.Printf("\nTarget: %s (%s, type: %s)\n", result.TargetName, t.Path, t.Type)
		if len(result.Files) == 0 {
			fmt.Println("  No files match cleanup criteria.")
		} else {
			for _, file := range result.Files {
				fmt.Printf("  [MATCH] %s\n", file)
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
	return nil
}
