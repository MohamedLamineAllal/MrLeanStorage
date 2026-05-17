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

	s := New(zap.NewNop(), nil)
	target := Target{
		Name:      "Test Target",
		Path:      tempDir,
		Threshold: 7 * 24 * time.Hour,
		Type:      "file",
	}

	result, err := s.Scan(target, nil)
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

func TestScan_Glob(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mls-glob-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create structure: tempDir/profile1/Cache/old.txt and tempDir/profile2/Cache/old.txt
	profile1Cache := filepath.Join(tempDir, "profile1", "Cache")
	profile2Cache := filepath.Join(tempDir, "profile2", "Cache")
	err = os.MkdirAll(profile1Cache, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(profile2Cache, 0755)
	assert.NoError(t, err)

	oldFile1 := filepath.Join(profile1Cache, "old.txt")
	oldFile2 := filepath.Join(profile2Cache, "old.txt")

	err = os.WriteFile(oldFile1, []byte("old content"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(oldFile2, []byte("old content"), 0644)
	assert.NoError(t, err)

	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	os.Chtimes(oldFile1, oldTime, oldTime)
	os.Chtimes(oldFile2, oldTime, oldTime)

	s := New(zap.NewNop(), nil)
	target := Target{
		Name:      "Glob Test",
		Path:      filepath.Join(tempDir, "*", "Cache"),
		Threshold: 7 * 24 * time.Hour,
		Type:      "file",
	}

	result, err := s.Scan(target, nil)
	assert.NoError(t, err)
	assert.Len(t, result.Files, 2)
	assert.Contains(t, result.Files, oldFile1)
	assert.Contains(t, result.Files, oldFile2)
}

func TestScan_IgnorePatterns(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mls-ignore-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create files
	ignoredFile := filepath.Join(tempDir, ".DS_Store")
	keptFile := filepath.Join(tempDir, "data.txt")
	os.WriteFile(ignoredFile, []byte("ignore me"), 0644)
	os.WriteFile(keptFile, []byte("keep me"), 0644)

	// Make both old
	oldTime := time.Now().Add(-20 * 24 * time.Hour)
	os.Chtimes(ignoredFile, oldTime, oldTime)
	os.Chtimes(keptFile, oldTime, oldTime)

	s := New(zap.NewNop(), []string{".DS_Store"})
	target := Target{
		Name:      "Ignore Test",
		Path:      tempDir,
		Threshold: 7 * 24 * time.Hour,
		Type:      "file",
	}

	result, err := s.Scan(target, nil)
	assert.NoError(t, err)
	assert.Len(t, result.Files, 1)
	assert.Equal(t, keptFile, result.Files[0])
}

