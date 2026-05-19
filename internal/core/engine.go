package core

import (
	"runtime"
	"sync"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/mohamedlamineallal/MacosLeanStorage/internal/scanner"
	"go.uber.org/zap"
)

// Engine orchestrates the parallel scanning of configuration targets.
type Engine struct {
	scanner *scanner.Scanner
	logger  *zap.Logger
}

// NewEngine creates a new Engine.
func NewEngine(logger *zap.Logger, ignorePatterns []string) *Engine {
	return &Engine{
		scanner: scanner.New(logger, ignorePatterns),
		logger:  logger,
	}
}

// ScanTargets processes multiple targets in parallel and returns results in the original order.
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
		} else {
			e.logger.Error("Scan failed", zap.String("target", res.Name), zap.Error(res.Err))
		}
	}
	return resultMap
}
