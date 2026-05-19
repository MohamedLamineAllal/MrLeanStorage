package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	"go.uber.org/zap"
)

var (
	// colorDryRun is used for displaying dry-run specific output.
	colorDryRun = color.New(color.FgHiBlack, color.Italic)
	// colorDelete is used for highlighting deletion actions.
	colorDelete = color.New(color.FgRed)
	// colorPath is used for displaying filesystem paths.
	colorPath = color.New(color.FgBlue)
	// colorCmd is used for displaying shell commands.
	colorCmd = color.New(color.FgYellow)
)

// Cleaner handles the deletion of files and execution of cleanup commands.
// It supports a dry-run mode to prevent accidental data loss.
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

// isIgnored checks if a file or directory name matches any of the configured ignore patterns.
func (c *Cleaner) isIgnored(name string) bool {
	for _, pattern := range c.ignorePatterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Clean deletes the provided list of file paths in parallel, invoking a hook for each file.
// Clean deletes the provided list of file paths in parallel, invoking an optional hook for each path.
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
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathChan {
				info, err := os.Stat(path)
				if err != nil {
					if !os.IsNotExist(err) {
						c.logger.Debug("Failed to stat path", zap.String("path", path), zap.Error(err))
					}
					continue
				}

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

				if !c.dryRun {
					err = os.RemoveAll(path)
					if err != nil {
						if !os.IsNotExist(err) {
							c.logger.Error("Failed to delete", zap.String("path", path), zap.Error(err))
						}
						continue
					}
				}
				if hook != nil {
					hook(path, size, nil)
				}
				resChan <- result{path, size, nil}
			}
		}()
	}

	for _, path := range paths {
		pathChan <- path
	}
	close(pathChan)

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

// ExecuteCommand runs a shell command unless dry-run mode is enabled.
func (c *Cleaner) ExecuteCommand(command string) error {
	if c.dryRun {
		colorDryRun.Print("  [DRY RUN] ")
		fmt.Print("Would execute command: ")
		colorCmd.Println(command)
		return nil
	}

	fmt.Print("  Executing command: ")
	colorCmd.Println(command)
	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], parts[1:]...)
	err := cmd.Run()
	if err != nil {
		c.logger.Error("Failed to execute command", zap.String("command", command), zap.Error(err))
		return err
	}
	return nil
}

// SetDryRun toggles the dry run mode of the Cleaner.
func (c *Cleaner) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// DryRun returns true if the Cleaner is currently in dry-run mode.
func (c *Cleaner) DryRun() bool {
	return c.dryRun
}
