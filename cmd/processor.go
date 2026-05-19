package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/engine"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
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
		engine:    engine.NewDefault(logger, ignorePatterns, dryRun),
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

	tp.engine.SetCommandHandler(engine.NewCommandHandler(tp.engine, tp.scheduler, tp.logger))

	var scanTargets []config.TargetConfig
	for _, t := range targets {
		if t.Command == "" {
			scanTargets = append(scanTargets, t)
		}
	}

	scanBar := progressbar.NewOptions(len(scanTargets),
		progressbar.OptionSetDescription("Scanning targets..."),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "█", SaucerPadding: " ", BarStart: "[", BarEnd: "]"}),
		progressbar.OptionSetPredictTime(false),
	)

	hooks := engine.Hooks{
		OnTargetScanEnd: func(name string, result *scanner.Result, error error) {
			scanBar.Add(1)
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

	resultMap, err := tp.engine.Scan(targets, hooks)
	scanBar.Finish()
	if err != nil {
		return err
	}

	for _, t := range targets {
		if t.Command != "" {
			continue
		}
		res, ok := resultMap[t.Name]
		if !ok {
			continue
		}

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

	uniqueCount, totalSize := 0, int64(0)
	if isClean {
		fmt.Printf("\n\n")
		desc := "Cleaning and processing targets..."
		if tp.engine.Cleaner().DryRun() {
			desc = "[DRY RUN] Cleaning and processing targets..."
		}
		totalWork := len(targets)
		cleanBar := progressbar.NewOptions(totalWork,
			progressbar.OptionSetDescription(desc),
			progressbar.OptionShowCount(),
			progressbar.OptionSetTheme(progressbar.Theme{Saucer: "█", SaucerPadding: " ", BarStart: "[", BarEnd: "]"}),
			progressbar.OptionSetPredictTime(false),
		)
		hooks.OnTargetCleaned = func(name string) {
			cleanBar.Add(1)
		}
		hooks.OnNoMatchesTargetCleanSkip = func(name string) {
			cleanBar.Add(1)
		}
		hooks.AfterHandleCommand = func(name string, command string, err error) {
			cleanBar.Add(1)
		}
		firstCommand := true
		hooks.BeforeHandleCommand = func(name string, command string, shouldRunCommand bool) {
			fmt.Printf("\n")
			if firstCommand {
				if tp.engine.Cleaner().DryRun() {
					colorInfo.Printf("\nProcessing commands Targets (DRY RUN):\n\n")
				} else {
					colorInfo.Printf("\nProcessing commands Targets:\n\n")
				}
				firstCommand = false
			}
			colorTarget.Printf("Target: %s", name)
			colorCommand.Printf(" (command: %s)\n", command)
			if !shouldRunCommand {
				colorInfo.Printf("Skipping command for target: %s (not scheduled to run yet)", name)
			}
		}
		hooks.BeforeExecutingCommand = func(name string, command string) {
			if tp.engine.Cleaner().DryRun() {
				colorInfo.Printf("Executing Command (DRY RUN) for Target: %s", name)
			} else {
				colorTarget.Printf("Executing Command for Target: %s", name)
			}
			colorCommand.Printf(" (command: %s)\n", command)
		}
		hooks.AfterExecutingCommand = func(name string, command string, err error) {
			if err != nil {
				tp.logger.Error("Command failed", zap.String("target", name), zap.Error(err))
			}
		}

		uniqueCount, totalSize, err = tp.engine.Clean(resultMap, targets, hooks)
		cleanBar.Finish()
		if err != nil {
			return err
		}
	} else {
		aggregator := &engine.ResultAggregator{UniquePaths: make(map[string]int64)}
		for _, res := range resultMap {
			aggregator.Add(res.Files, res.FileSizes)
		}
		uniqueCount, totalSize = aggregator.GetStats()
	}

	tp.printSummary(uniqueCount, totalSize, isClean, logPath)
	return nil
}

func (tp *TargetProcessor) printSummary(count int, size int64, isClean bool, logPath string) {
	fmt.Printf("\n")
	if isClean {
		colorSuccess.Print("Clean Summary: ")
		if tp.engine.Cleaner().DryRun() {
			fmt.Printf("[DRY RUN]: Would delete %d files, freeing %.2f MB\n", count, float64(size)/(1024*1024))
			color.New(color.FgMagenta).Print("\nTo proceed with actual deletion, please run: ")
			color.New(color.FgHiYellow).Print("`mls clean --dry-run=false`")
			color.New(color.FgMagenta).Println(".")
		} else {
			fmt.Printf("Deleted %d files, freeing %.2f MB\n", count, float64(size)/(1024*1024))
		}
	} else {
		colorSuccess.Print("Summary: ")
		fmt.Printf("Found %d unique files, total size estimation (approx): %.2f MB\n", count, float64(size)/(1024*1024))
	}
	if count > 0 {
		fmt.Printf("\nFull details log written to: ")
		color.New(color.FgHiYellow, color.Underline).Println(logPath)
	}
}
