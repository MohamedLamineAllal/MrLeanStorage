package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Target represents a directory to be scanned
type Target struct {
	Name        string
	Path        string
	Threshold   time.Duration
	SafetyLevel int
}

// Result contains information about a scanned target
type Result struct {
	TargetName string
	Files      []string
	TotalSize  int64
}

// Scanner handles the directory traversal and analysis
type Scanner struct {
	logger *zap.Logger
}

// New creates a new Scanner
func New(logger *zap.Logger) *Scanner {
	return &Scanner{logger: logger}
}

// Scan analyzes a target and returns a list of files that match cleanup criteria
func (s *Scanner) Scan(target Target) (*Result, error) {
	expandedPath, err := expandPath(target.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path %s: %w", target.Path, err)
	}

	paths, err := filepath.Glob(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to glob path %s: %w", expandedPath, err)
	}

	result := &Result{
		TargetName: target.Name,
		Files:      []string{},
	}

	now := time.Now()

	for _, p := range paths {
		err = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				if os.IsPermission(err) {
					return nil
				}
				return err
			}

			// Skip the root path itself
			if path == p {
				return nil
			}

			// If it's a directory, we might want to clean it if it's old enough,
			// but usually we clean files. For simplicity, we'll list files.
			if !info.IsDir() {
				if now.Sub(info.ModTime()) > target.Threshold {
					result.Files = append(result.Files, path)
					result.TotalSize += info.Size()
				}
			}

			return nil
		})

		if err != nil {
			s.logger.Warn("Partial walk failed", zap.String("path", p), zap.Error(err))
		}
	}

	return result, nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return filepath.Abs(path)
}
