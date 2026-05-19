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

// Hooks provides lifecycle callbacks for engine operations.
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

// Engine encapsulates the scanning and cleaning logic.
type Engine struct {
	scanner        *scanner.Scanner
	cleaner        *cleaner.Cleaner
	commandHandler *CommandHandler
	logger         *zap.Logger
}

// New creates a new Engine instance.
func New(logger *zap.Logger, ignorePatterns []string, dryRun bool) *Engine {
	e := &Engine{
		scanner: scanner.New(logger, ignorePatterns),
		cleaner: cleaner.New(logger, dryRun, ignorePatterns),
		logger:  logger,
	}
	// Note: We'll initialize commandHandler separately or pass a scheduler if needed.
	// For now, keeping it simple as per original design.
	return e
}

// SetCommandHandler allows injecting the command handler.
func (e *Engine) SetCommandHandler(ch *CommandHandler) {
	e.commandHandler = ch
}

func (e *Engine) CommandHandler() *CommandHandler {
	return e.commandHandler
}

// Scan performs parallel scanning of the provided targets.
func (e *Engine) Scan(targets []config.TargetConfig, hooks Hooks) (map[string]*scanner.Result, error) {
	numWorkers := runtime.NumCPU()
	jobs := make(chan config.TargetConfig, len(targets))
	results := make(chan struct {
		Name string
		Res  *scanner.Result
		Err  error
	}, len(targets))

	var wg sync.WaitGroup
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
				// Fire the OnTargetScanStart hook before scanning
				if hooks.OnTargetScanStart != nil {
					hooks.OnTargetScanStart(t.Name, t.Path)
				}

				res, err := e.scanner.Scan(target, t.IgnorePatterns)

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
	// ...

	for _, t := range targets {
		if t.Command == "" {
			jobs <- t
		}
	}
	close(jobs)
	wg.Wait()
	close(results)

	resultMap := make(map[string]*scanner.Result)
	for res := range results {
		if res.Err == nil {
			resultMap[res.Name] = res.Res
		}
	}
	return resultMap, nil
}

// Clean executes the cleanup process for the identified scan results.
func (e *Engine) Clean(resultMap map[string]*scanner.Result, targets []config.TargetConfig, hooks Hooks) (int, int64, error) {
	aggregator := &ResultAggregator{UniquePaths: make(map[string]int64)}

	for _, t := range targets {
		res, ok := resultMap[t.Name]
		if !ok || len(res.Files) == 0 {
			if hooks.OnNoMatchesTargetCleanSkip != nil {
				hooks.OnNoMatchesTargetCleanSkip(t.Name)
			}
			continue
		}

		aggregator.Add(res.Files, res.FileSizes)

		_, _, err := e.cleaner.Clean(res.Files, hooks.OnFileCleaned)
		if err != nil {
			e.logger.Error("Clean failed", zap.String("target", t.Name), zap.Error(err))
		}

		if hooks.OnTargetCleaned != nil {
			hooks.OnTargetCleaned(t.Name)
		}
	}

	// Process commands
	e.ProcessCommands(targets, hooks)

	uniqueCount := len(aggregator.UniquePaths)
	return uniqueCount, aggregator.totalSize, nil
}

// ProcessCommands executes commands associated with targets.
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

// ResultAggregator safely aggregates unique scan results.
type ResultAggregator struct {
	mu          sync.RWMutex
	UniquePaths map[string]int64
	totalSize   int64
}

// Add safely adds results to the aggregator.
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

// GetStats returns the current statistics of aggregated results.
func (ra *ResultAggregator) GetStats() (int, int64) {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return len(ra.UniquePaths), ra.totalSize
}

// Cleaner returns the underlying Cleaner instance.
func (e *Engine) Cleaner() *cleaner.Cleaner {
	return e.cleaner
}
