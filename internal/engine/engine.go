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
	OnTargetScanStart func(name string, path string)
	OnMatchFound      func(targetName string, files []string)
	OnFileCleaned     func(path string, freed int64, err error)
}

// Engine encapsulates the scanning and cleaning logic.
type Engine struct {
	scanner *scanner.Scanner
	cleaner *cleaner.Cleaner
	logger  *zap.Logger
}

// New creates a new Engine instance.
func New(logger *zap.Logger, ignorePatterns []string, dryRun bool) *Engine {
	return &Engine{
		scanner: scanner.New(logger, ignorePatterns),
		cleaner: cleaner.New(logger, dryRun, ignorePatterns),
		logger:  logger,
	}
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
	return resultMap, nil
}

// Clean executes the cleanup process for the identified scan results.
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

		_, _, err := e.cleaner.Clean(res.Files, hooks.OnFileCleaned)
		if err != nil {
			e.logger.Error("Clean failed", zap.String("target", t.Name), zap.Error(err))
		}
	}

	uniqueCount := len(aggregator.uniquePaths)
	return uniqueCount, aggregator.totalSize, nil
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
	uniquePaths map[string]int64
	totalSize   int64
}

// Add safely adds results to the aggregator.
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

// GetStats returns the current statistics of aggregated results.
func (ra *ResultAggregator) GetStats() (int, int64) {
	ra.mu.RLock()
	defer ra.mu.RUnlock()
	return len(ra.uniquePaths), ra.totalSize
}

// Cleaner returns the underlying Cleaner instance.
func (e *Engine) Cleaner() *cleaner.Cleaner {
	return e.cleaner
}
