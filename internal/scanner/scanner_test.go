package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestScan(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mls-scan-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create some files
	oldFile := filepath.Join(tempDir, "old.txt")
	newFile := filepath.Join(tempDir, "new.txt")

	err = os.WriteFile(oldFile, []byte("old content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(newFile, []byte("new content"), 0644)
	assert.NoError(t, err)

	// Set old file modification time to 10 days ago
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	assert.NoError(t, err)

	s := New(zap.NewNop())
	target := Target{
		Name:      "Test Target",
		Path:      tempDir,
		Threshold: 7 * 24 * time.Hour,
	}

	result, err := s.Scan(target)
	assert.NoError(t, err)
	assert.Equal(t, "Test Target", result.TargetName)
	assert.Len(t, result.Files, 1)
	assert.Equal(t, oldFile, result.Files[0])
}

func TestExpandPath(t *testing.T) {
	// Test ~/ expansion
	home, _ := os.UserHomeDir()
	path, err := expandPath("~/Library")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "Library"), path)

	// Test absolute path
	absPath, _ := filepath.Abs("/tmp")
	path, err = expandPath("/tmp")
	assert.NoError(t, err)
	assert.Equal(t, absPath, path)
}

func TestScan_NonExistentPath(t *testing.T) {
	s := New(zap.NewNop())
	target := Target{
		Name:      "Non Existent",
		Path:      "/tmp/this-directory-definitely-does-not-exist-123456789",
		Threshold: 7 * 24 * time.Hour,
	}

	result, err := s.Scan(target)
	assert.NoError(t, err)
	assert.Equal(t, "Non Existent", result.TargetName)
	assert.Len(t, result.Files, 0)
}
