package cleaner

import (
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// Cleaner handles the deletion of files
type Cleaner struct {
	logger *zap.Logger
	dryRun bool
}

// New creates a new Cleaner
func New(logger *zap.Logger, dryRun bool) *Cleaner {
	return &Cleaner{
		logger: logger,
		dryRun: dryRun,
	}
}

// Clean deletes the provided list of file paths
func (c *Cleaner) Clean(paths []string) (int, int64, error) {
	var deletedCount int
	var freedSpace int64

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			c.logger.Error("Failed to stat path", zap.String("path", path), zap.Error(err))
			continue
		}

		size := info.Size()

		if c.dryRun {
			c.logger.Info("[DRY RUN] Would delete", zap.String("path", path), zap.Int64("size", size))
			deletedCount++
			freedSpace += size
			continue
		}

		c.logger.Info("Deleting", zap.String("path", path), zap.Int64("size", size))
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

// ExecuteCommand runs a shell command
func (c *Cleaner) ExecuteCommand(command string) error {
	if c.dryRun {
		c.logger.Info("[DRY RUN] Would execute command", zap.String("command", command))
		return nil
	}

	c.logger.Info("Executing command", zap.String("command", command))
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
