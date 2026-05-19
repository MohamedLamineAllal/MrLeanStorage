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

// LogEvent represents a structured event for logging callbacks.
type LogEvent struct {
	Type    string
	Message string
	Path    string
	Size    int64
}

// Engine orchestrates scanning and cleaning targets.
type Engine struct {
	scanner *scanner.Scanner
	cleaner *cleaner.Cleaner
	logger  *zap.Logger
}

// NewEngine creates a new Engine.
func NewEngine(logger *zap.Logger, ignorePatterns []string, dryRun bool) *Engine {
	return &Engine{
		scanner: scanner.New(logger, ignorePatterns),
		cleaner: cleaner.New(logger, dryRun, ignorePatterns),
		logger:  logger,
	}
}

// Hooks defines the callback interface for engine events.
type Hooks struct {
	OnTargetScanStart func(name string, path string)
	OnMatchFound      func(targetName string, files []string)
	OnFileCleaned     func(path string, freed int64, err error)
}

// RunOptions configures the execution of the Engine.
type RunOptions struct {
	IsClean bool
	DryRun  bool
	Hooks   Hooks
}

// Scan orchestrates the parallel scanning of targets.
func (e *Engine) Scan(targets []config.TargetConfig, hooks Hooks) (map[string]*scanner.Result, error) {
	resultMap := e.ScanTargets(targets)

	// Optional: fire hooks for each target scanned if needed
	return resultMap, nil
}

// Clean executes the cleanup of identified scan results.
func (e *Engine) Clean(resultMap map[string]*scanner.Result, targets []config.TargetConfig, hooks Hooks) (int, int64, error) {
	aggregator := &ResultAggregator{uniquePaths: make(map[string]int64)}

	for _, t := range targets {
		res, ok := resultMap[t.Name]
		if !ok || len(res.Files) == 0 {
			continue
		}

		if hooks.OnTargetScanStart != nil {
			hooks.OnTargetScanStart(t.Name, t.Path)
		}
		if hooks.OnMatchFound != nil {
			hooks.OnMatchFound(t.Name, res.Files)
		}

		aggregator.Add(res.Files, res.FileSizes)

		_, _, err := e.cleaner.CleanWithHook(res.Files, hooks.OnFileCleaned)
		if err != nil {
			e.logger.Error("Clean failed", zap.String("target", t.Name), zap.Error(err))
		}
	}

	uniqueCount := len(aggregator.uniquePaths)
	return uniqueCount, aggregator.totalSize, nil
}

// ScanAndClean performs a full scan followed by a clean operation.
func (e *Engine) ScanAndClean(targets []config.TargetConfig, hooks Hooks) (int, int64, error) {
	resultMap, err := e.Scan(targets, hooks)
	if err != nil {
		return 0, 0, err
	}
	return e.Clean(resultMap, targets, hooks)
}



// ScanTargets processes multiple targets in parallel and returns scan results.
func (e *Engine) ScanTargets(targets []config.TargetConfig) map[string]*scanner.Result {
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
				res, err := e.scanner.Scan(target, t.IgnorePatterns)
				results <- struct {
					Name string
					Res  *scanner.Result
					Err  error
				}{t.Name, res, err}
			}
		}()
	}

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
	return resultMap
}

// Cleaner returns the underlying Cleaner instance.
func (e *Engine) Cleaner() *cleaner.Cleaner {
	return e.cleaner
}

// ResultAggregator tracks unique file stats.
type ResultAggregator struct {
	mu           sync.RWMutex
	uniquePaths  map[string]int64
	totalSize    int64
}

func (ra *ResultAggregator) Add(files []string, sizes []int64) {
	ra.mu.Lock()
	defer ra.mu.Unlock()
	for i, file := range files {
		if _, exists := ra.uniquePaths[file]; !exists {
			ra.uniquePaths[file] = sizes[i]
			ra.totalSize += sizes[i]
		}
	}
}

// GetStats returns the unique file count and aggregated size.
func (ra *ResultAggregator) GetStats() (int, int64) {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return len(ra.uniquePaths), ra.totalSize
}

// GetUniquePaths returns the list of all unique file paths.
func (ra *ResultAggregator) GetUniquePaths() []string {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	paths := make([]string, 0, len(ra.uniquePaths))
	for path := range ra.uniquePaths {
		paths = append(paths, path)
	}
	return paths
}
