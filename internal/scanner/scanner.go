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
	Type        string // "file", "folder", or "both"
}

// Result contains information about a scanned target
type Result struct {
	TargetName string
	Files      []string
	TotalSize  int64
}

// Scanner handles the directory traversal and analysis
type Scanner struct {
	logger         *zap.Logger
	ignorePatterns []string
}

// New creates a new Scanner
func New(logger *zap.Logger, ignorePatterns []string) *Scanner {
	return &Scanner{logger: logger, ignorePatterns: ignorePatterns}
}

// isIgnored checks if a file name matches any of the ignore patterns
func (s *Scanner) isIgnored(name string) bool {
	for _, pattern := range s.ignorePatterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Scan analyzes a target and returns a list of files that match cleanup criteria
func (s *Scanner) Scan(target Target, targetIgnorePatterns []string) (*Result, error) {
	// Merge global ignore patterns with target-specific ones
	allIgnorePatterns := append(s.ignorePatterns, targetIgnorePatterns...)
	// Temporarily override the scanner's ignore patterns for this scan
	originalPatterns := s.ignorePatterns
	s.ignorePatterns = allIgnorePatterns
	defer func() { s.ignorePatterns = originalPatterns }()

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
		info, err := os.Stat(p)
		if err != nil {
			s.logger.Warn("Failed to stat glob match", zap.String("path", p), zap.Error(err))
			continue
		}

		if (target.Type == "folder" || target.Type == "both") && info.IsDir() {
			isStale, err := s.checkStaleness(p, target.Threshold, now)
			if err != nil {
				s.logger.Warn("Failed to check folder staleness", zap.String("path", p), zap.Error(err))
				continue
			}

			if isStale {
				size, err := s.getDirSize(p)
				if err != nil {
					s.logger.Warn("Failed to calculate directory size", zap.String("path", p), zap.Error(err))
				}
				result.Files = append(result.Files, p)
				result.TotalSize += size
			}
			continue
		}

		if target.Type == "file" || target.Type == "both" || target.Type == "" {
			err = s.walkFiles(p, target.Threshold, &result.Files, &result.TotalSize, now)
			if err != nil {
				s.logger.Warn("Failed to walk files", zap.String("path", p), zap.Error(err))
			}
		}
	}

	return result, nil
}

func (s *Scanner) walkFiles(path string, threshold time.Duration, matches *[]string, totalSize *int64, now time.Time) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsPermission(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if s.isIgnored(entry.Name()) {
			continue
		}
		fullPath := filepath.Join(path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			if err := s.walkFiles(fullPath, threshold, matches, totalSize, now); err != nil {
				s.logger.Debug("Subdirectory walk failed", zap.String("path", fullPath), zap.Error(err))
			}
		} else {
			if now.Sub(info.ModTime()) > threshold {
				*matches = append(*matches, fullPath)
				*totalSize += info.Size()
			}
		}
	}
	return nil
}

func (s *Scanner) checkStaleness(path string, threshold time.Duration, now time.Time) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// 1. Fast Path: If the folder itself is stale, it's definitely stale.
	if now.Sub(info.ModTime()) > threshold {
		return true, nil
	}

	// 2. Slow Path: Folder mtime is recent, perform a deep check.
	// A folder is stale only if NONE of its files are newer than the threshold.
	stale := true
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if s.isIgnored(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			if now.Sub(info.ModTime()) <= threshold {
				stale = false
				return fmt.Errorf("not stale")
			}
		}
		return nil
	})

	if err != nil && err.Error() == "not stale" {
		return false, nil
	}
	return stale, err
}

func (s *Scanner) getDirSize(path string) (int64, error) {
	var size int64
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			subSize, err := s.getDirSize(filepath.Join(path, entry.Name()))
			if err == nil {
				size += subSize
			}
		} else {
			size += info.Size()
		}
	}
	return size, nil
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
