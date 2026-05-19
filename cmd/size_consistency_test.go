package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mohamedlamineallal/MacosLeanStorage/internal/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestScanAndCleanSizeConsistency(t *testing.T) {
	logger := zap.NewNop()
	
	// Setup a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "mls-consistency-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a structure:
	// tempDir/
	//   old_file.txt (10 bytes, stale)
	//   new_file.txt (20 bytes, not stale)
	//   subfolder/ (stale folder)
	//     file1.txt (30 bytes)
	//     .ds_store (ignored, 100 bytes)
	
	oldFile := filepath.Join(tempDir, "old_file.txt")
	newFile := filepath.Join(tempDir, "new_file.txt")
	subFolder := filepath.Join(tempDir, "subfolder")
	subFile := filepath.Join(subFolder, "file1.txt")
	ignoredFile := filepath.Join(subFolder, ".DS_Store")

	err = os.MkdirAll(subFolder, 0755)
	assert.NoError(t, err)

	err = os.WriteFile(oldFile, make([]byte, 10), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(newFile, make([]byte, 20), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(subFile, make([]byte, 30), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(ignoredFile, make([]byte, 100), 0644)
	assert.NoError(t, err)

	// Set mtimes
	staleTime := time.Now().Add(-40 * 24 * time.Hour)
	err = os.Chtimes(oldFile, staleTime, staleTime)
	assert.NoError(t, err)
	err = os.Chtimes(subFolder, staleTime, staleTime)
	assert.NoError(t, err)
	err = os.Chtimes(subFile, staleTime, staleTime)
	assert.NoError(t, err)
	err = os.Chtimes(ignoredFile, staleTime, staleTime)
	assert.NoError(t, err)

	ignorePatterns := []string{".DS_Store"}
	tp := NewTargetProcessor(logger, ignorePatterns, true) // Dry run
// ...
targets := []config.TargetConfig{
	{
		Name:           "Test Target",
		Path:           tempDir + "/*", // Glob to catch files and subfolder
		Threshold:      30,
		Type:           "both",
		IgnorePatterns: ignorePatterns,
	},
}

resultMap := tp.engine.ScanTargets(targets)
result := resultMap["Test Target"]
assert.NotNil(t, result)

// Expected: old_file.txt (10) + subfolder (30, because .DS_Store is ignored) = 40 bytes
assert.Equal(t, int64(40), result.TotalSize)

// 2. Run the cleaner part
count, size, err := tp.cleaner.Clean(result.Files)
assert.NoError(t, err)

// Both should report exactly 40 bytes
assert.Equal(t, int64(2), int64(count))
assert.Equal(t, result.TotalSize, size, "Scan and Clean sizes should match exactly")
fmt.Printf("Scan size: %d, Clean size: %d\n", result.TotalSize, size)
}

