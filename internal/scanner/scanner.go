package scanner

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// Target represents a directory to be scanned
type Target struct {
	Path       string
	Name       string
	SafetyLevel int // 0: Always safe, 1: Safe after 3 days, 2: Safe after 7 days
}

// Scanner handles the directory traversal and analysis
type Scanner struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) *Scanner {
	return &Scanner{logger: logger}
}

// Scan analyzes a target and returns a list of files that match cleanup criteria
func (s *Scanner) Scan(target Target, threshold time.Duration) ([]string, error) {
	// Implementation will use concurrent workers to scan paths
	s.logger.Info("Scanning target", zap.String("path", target.Path))
	return nil, nil
}
