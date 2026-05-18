package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"go.uber.org/zap"
)

var (
	colorDryRun = color.New(color.FgHiBlack, color.Italic)
	colorDelete = color.New(color.FgRed)
	colorPath   = color.New(color.FgBlue)
	colorCmd    = color.New(color.FgYellow)
)

// Cleaner handles the deletion of files
type Cleaner struct {
	logger         *zap.Logger
	dryRun         bool
	ignorePatterns []string
}

// New creates a new Cleaner
func New(logger *zap.Logger, dryRun bool, ignorePatterns []string) *Cleaner {
	return &Cleaner{
		logger:         logger,
		dryRun:         dryRun,
		ignorePatterns: ignorePatterns,
	}
}

// isIgnored checks if a file name matches any of the ignore patterns
func (c *Cleaner) isIgnored(name string) bool {
	for _, pattern := range c.ignorePatterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Clean deletes the provided list of file paths
func (c *Cleaner) Clean(paths []string) (int, int64, error) {
	var deletedCount int
	var freedSpace int64

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			c.logger.Debug("Failed to stat path", zap.String("path", path), zap.Error(err))
			continue
		}

		var size int64
		if info.IsDir() {
			size, _ = c.getDirSize(path)
		} else {
			size = info.Size()
		}

		if c.dryRun {
			colorDryRun.Print("  [DRY RUN] ")
			fmt.Print("Would delete: ")
			colorPath.Println(path)
			deletedCount++
			freedSpace += size
			continue
		}

		colorDelete.Print("  Deleting: ")
		colorPath.Println(path)
		err = os.RemoveAll(path)
		if err != nil {
			c.logger.Error("Failed to delete", zap.String("path", path), zap.Error(err))
			continue
		}
		deletedCount++
		freedSpace += size
	}

	return deletedCount, freedSpace, nil
}

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

// ExecuteCommand runs a shell command
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


// SetDryRun toggles the dry run mode
func (c *Cleaner) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// DryRun returns the current dry run mode
func (c *Cleaner) DryRun() bool {
	return c.dryRun
}
