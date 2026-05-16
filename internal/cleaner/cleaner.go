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

func New(logger *zap.Logger, dryRun bool) *Cleaner {
	return &Cleaner{
		logger: logger,
		dryRun: dryRun,
	}
}

// Clean deletes the provided list of file paths
func (c *Cleaner) Clean(paths []string) error {
	for _, path := range paths {
		if c.dryRun {
			c.logger.Info("[DRY RUN] Would delete", zap.String("path", path))
			continue
		}
		
		c.logger.Info("Deleting", zap.String("path", path))
		err := os.RemoveAll(path)
		if err != nil {
			c.logger.Error("Failed to delete", zap.String("path", path), zap.Error(err))
		}
	}
	return nil
}
