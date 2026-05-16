package cleaner

import (
	"os"

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

// SetDryRun toggles the dry run mode
func (c *Cleaner) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}
