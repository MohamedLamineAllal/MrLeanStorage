// Package cleaner provides robust and parallelized file system cleanup capabilities.
// It features dry-run support, recursive size calculation, and hook-based event reporting.
package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"go.uber.org/zap"
)

// Cleaner handles the deletion of files and execution of cleanup processes.
// It maintains configurations for logging, safety (dry-run), and ignore patterns.
type Cleaner struct {
	logger         *zap.Logger
	dryRun         bool
	ignorePatterns []string
	logFile        *os.File
}

// New creates a new Cleaner instance with the provided logger, dry-run setting, and ignore patterns.
func New(logger *zap.Logger, dryRun bool, ignorePatterns []string) *Cleaner {
	return &Cleaner{
		logger:         logger,
		dryRun:         dryRun,
		ignorePatterns: ignorePatterns,
	}
}

// SetLogFile sets the file where detailed cleanup logs will be written.
func (c *Cleaner) SetLogFile(f *os.File) {
	c.logFile = f
}

// logToFile writes a formatted message to the configured log file, if any.
func (c *Cleaner) logToFile(format string, a ...interface{}) {
	if c.logFile != nil {
		fmt.Fprintf(c.logFile, format+"\n", a...)
	}
}

// isIgnored checks if a file or directory name matches any of the configured ignore patterns
// using standard filepath glob matching.
func (c *Cleaner) isIgnored(name string) bool {
	for _, pattern := range c.ignorePatterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Clean deletes the provided list of file paths in parallel, invoking an optional hook for each path.
// It uses a worker pool based on the available CPU cores to maximize performance.
func (c *Cleaner) Clean(paths []string, hook func(path string, freed int64, err error)) (int, int64, error) {
	numWorkers := runtime.NumCPU()
	pathChan := make(chan string, len(paths))
	type result struct {
		path  string
		freed int64
		err   error
	}
	resChan := make(chan result, len(paths))

	var wg sync.WaitGroup
	// Spawn workers to process paths concurrently
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathChan {
				info, err := os.Stat(path)
				if err != nil {
					// Skip paths that don't exist, log errors for others
					if !os.IsNotExist(err) {
						c.logger.Debug("Failed to stat path", zap.String("path", path), zap.Error(err))
					}
					continue
				}

				// Calculate space that will be freed
				var size int64
				if info.IsDir() {
					size, _ = c.getDirSize(path)
				} else {
					size = info.Size()
				}

				prefix := ""
				if c.dryRun {
					prefix = "[DRY RUN] "
				}
				c.logToFile("%sWould delete: %s", prefix, path)

				// Perform removal if not in dry-run mode
				if !c.dryRun {
					err = os.RemoveAll(path)
					if err != nil {
						if !os.IsNotExist(err) {
							c.logger.Error("Failed to delete", zap.String("path", path), zap.Error(err))
						}
						continue
					}
				}
				// Callback to notify progress
				if hook != nil {
					hook(path, size, nil)
				}
				resChan <- result{path, size, nil}
			}
		}()
	}

	// Distribute work to channels
	for _, path := range paths {
		pathChan <- path
	}
	close(pathChan)

	// Wait for workers and aggregate results
	go func() {
		wg.Wait()
		close(resChan)
	}()

	var deletedCount int
	var freedSpace int64
	for res := range resChan {
		deletedCount++
		freedSpace += res.freed
	}

	return deletedCount, freedSpace, nil
}

// getDirSize calculates the total size of all non-ignored files within a directory recursively.
func (c *Cleaner) getDirSize(path string) (int64, error) {
	var size int64
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		// Respect ignore patterns to avoid counting sensitive or system files
		if c.isIgnored(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			subSize, err := c.getDirSize(filepath.Join(path, entry.Name()))
			if err == nil {
				size += subSize
			}
		} else {
			size += info.Size()
		}
	}
	return size, nil
}

// SetDryRun toggles the dry run mode of the Cleaner.
func (c *Cleaner) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// DryRun returns true if the Cleaner is currently in dry-run mode.
func (c *Cleaner) DryRun() bool {
	return c.dryRun
}
