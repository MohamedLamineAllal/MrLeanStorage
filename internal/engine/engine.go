// Package engine coordinates the scanning and cleanup operations.
// It manages lifecycle hooks, parallel task execution, and result aggregation for cleanup targets.
package engine

import (
	"runtime"
	"sync"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/cleaner"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
	"go.uber.org/zap"
)

// ScannerInterface abstracts the scanner functionality for dependency injection.
type ScannerInterface interface {
	Scan(target scanner.Target, ignorePatterns []string) (*scanner.Result, error)
}

// CleanerInterface abstracts the cleaner functionality for dependency injection.
type CleanerInterface interface {
	Clean(paths []string, hook func(path string, freed int64, err error)) (int, int64, error)
	DryRun() bool
}

// Hooks provides lifecycle callbacks for engine operations, enabling progress tracking and logging.
type Hooks struct {
	OnTargetScanStart          func(name string, path string)
	OnTargetScanEnd            func(name string, result *scanner.Result, err error)
	OnFileCleaned              func(path string, freed int64, err error)
	OnNoMatchesTargetCleanSkip func(name string)
	OnTargetCleaned            func(name string)
	BeforeHandleCommand        func(name string, command string, shouldExecuteCommand bool)
	AfterHandleCommand         func(name string, command string, err error)
	BeforeExecutingCommand     func(name string, command string)
	AfterExecutingCommand      func(name string, command string, err error)
}

// Engine encapsulates the orchestration logic for scanning and cleaning.
type Engine struct {
	scanner        ScannerInterface
	cleaner        CleanerInterface
	commandHandler *CommandHandler
	logger         *zap.Logger
}

// New creates a new Engine instance with the provided dependencies.
func New(logger *zap.Logger, s ScannerInterface, c CleanerInterface, commandHandler *CommandHandler) *Engine {
	return &Engine{
		scanner:        s,
		cleaner:        c,
		commandHandler: commandHandler,
		logger:         logger,
	}
}

// SetCommandHandler allows injecting the command handler into the engine.
func (e *Engine) SetCommandHandler(ch *CommandHandler) {
	e.commandHandler = ch
}

// CommandHandler returns the configured command handler for the engine.
func (e *Engine) CommandHandler() *CommandHandler {
	return e.commandHandler
}

// Scan performs parallel scanning of the provided cleanup targets using a worker pool.
func (e *Engine) Scan(targets []config.TargetConfig, hooks Hooks) (map[string]*scanner.Result, error) {
	numWorkers := runtime.NumCPU()
	jobs := make(chan config.TargetConfig, len(targets))
	results := make(chan struct {
		Name string
		Res  *scanner.Result
		Err  error
	}, len(targets))

	var wg sync.WaitGroup
	// Initialize workers for parallel scanning
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
				// Trigger lifecycle hook before scanning
				if hooks.OnTargetScanStart != nil {
					hooks.OnTargetScanStart(t.Name, t.Path)
				}

				res, err := e.scanner.Scan(target, t.IgnorePatterns)

				// Trigger lifecycle hook after scanning
				if hooks.OnTargetScanEnd != nil {
					hooks.OnTargetScanEnd(t.Name, res, err)
				}

				results <- struct {
					Name string
					Res  *scanner.Result
					Err  error
				}{t.Name, res, err}
			}
		}()
	}

	// Dispatch targets to jobs queue
	for _, t := range targets {
		if t.Command == "" {
			jobs <- t
		}
	}
	close(jobs)
	wg.Wait()
	close(results)

	// Collect and map results
	resultMap := make(map[string]*scanner.Result)
	for res := range results {
		if res.Err == nil {
			resultMap[res.Name] = res.Res
		}
	}
	return resultMap, nil
}

// Clean executes the cleanup process for the identified scan results in parallel.
func (e *Engine) Clean(resultMap map[string]*scanner.Result, targets []config.TargetConfig, hooks Hooks) (int, int64, error) {
	aggregator := &ResultAggregator{UniquePaths: make(map[string]int64)}
	numWorkers := runtime.NumCPU()
	
	type job struct {
		target config.TargetConfig
		res    *scanner.Result
	}
	jobs := make(chan job, len(targets))
	
	var wg sync.WaitGroup
	// Initialize workers for parallel cleanup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				// Use thread-safe aggregator for unique path tracking
				aggregator.Add(j.res.Files, j.res.FileSizes)
				
				// Execute parallel cleanup
				_, _, err := e.cleaner.Clean(j.res.Files, hooks.OnFileCleaned)
				if err != nil {
					e.logger.Error("Clean failed", zap.String("target", j.target.Name), zap.Error(err))
				}
				// Trigger lifecycle hook
				if hooks.OnTargetCleaned != nil {
					hooks.OnTargetCleaned(j.target.Name)
				}
			}
		}()
	}

	// Dispatch cleanup jobs
	for _, t := range targets {
		res, ok := resultMap[t.Name]
		if !ok || len(res.Files) == 0 {
			if hooks.OnNoMatchesTargetCleanSkip != nil {
				hooks.OnNoMatchesTargetCleanSkip(t.Name)
			}
			continue
		}
		jobs <- job{target: t, res: res}
	}
	close(jobs)
	wg.Wait()

	// Execute command-based tasks after file cleanup
	e.ProcessCommands(targets, hooks)

	uniqueCount, totalSize := aggregator.GetStats()
	return uniqueCount, totalSize, nil
}

// ProcessCommands executes commands associated with targets sequentially.
func (e *Engine) ProcessCommands(targets []config.TargetConfig, hooks Hooks) {
	if e.commandHandler == nil {
		return
	}

	commandHooks := CommandHooks{
		BeforeHandleCommand:    hooks.BeforeHandleCommand,
		AfterHandleCommand:     hooks.AfterHandleCommand,
		BeforeExecutingCommand: hooks.BeforeExecutingCommand,
		AfterExecutingCommand:  hooks.AfterExecutingCommand,
	}

	for _, t := range targets {
		if t.Command != "" {
			e.commandHandler.Handle(t, commandHooks)
		}
	}
}

// ScanAndClean runs both Scan and Clean sequentially.
func (e *Engine) ScanAndClean(targets []config.TargetConfig, hooks Hooks) (int, int64, error) {
	resMap, err := e.Scan(targets, hooks)
	if err != nil {
		return 0, 0, err
	}
	return e.Clean(resMap, targets, hooks)
}

// ResultAggregator safely aggregates unique scan results to avoid double-counting.
type ResultAggregator struct {
	mu          sync.RWMutex
	UniquePaths map[string]int64
	totalSize   int64
}

// Add safely adds results to the aggregator, ensuring paths are only counted once.
func (ra *ResultAggregator) Add(files []string, sizes []int64) {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	for i, file := range files {
		if _, exists := ra.UniquePaths[file]; !exists {
			ra.UniquePaths[file] = sizes[i]
			ra.totalSize += sizes[i]
		}
	}
}

// GetStats returns the current deduplicated statistics of aggregated results.
func (ra *ResultAggregator) GetStats() (int, int64) {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return len(ra.UniquePaths), ra.totalSize
}

// Cleaner returns the underlying Cleaner instance.
func (e *Engine) Cleaner() CleanerInterface {
	return e.cleaner
}

// NewDefault creates a new Engine instance with the default scanner and cleaner.
// This is used for backward compatibility.
func NewDefault(logger *zap.Logger, ignorePatterns []string, dryRun bool) *Engine {
	return &Engine{
		scanner: scanner.New(logger, ignorePatterns),
		cleaner: cleaner.New(logger, dryRun, ignorePatterns),
		logger:  logger,
	}
}
