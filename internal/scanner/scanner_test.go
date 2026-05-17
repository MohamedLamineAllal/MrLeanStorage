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
		Type:      "file",
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

	s := New(zap.NewNop())
	target := Target{
		Name:      "Glob Test",
		Path:      filepath.Join(tempDir, "*", "Cache"),
		Threshold: 7 * 24 * time.Hour,
		Type:      "file",
	}

	result, err := s.Scan(target)
	assert.NoError(t, err)
	assert.Len(t, result.Files, 2)
	assert.Contains(t, result.Files, oldFile1)
	assert.Contains(t, result.Files, oldFile2)
}

func TestScan_Folder(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mls-folder-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a subfolder with files
	cacheDir := filepath.Join(tempDir, "Cache")
	err = os.MkdirAll(cacheDir, 0755)
	assert.NoError(t, err)

	oldFile := filepath.Join(cacheDir, "old.txt")
	err = os.WriteFile(oldFile, []byte("old content"), 0644)
	assert.NoError(t, err)

	// Set both folder and file modification time to 10 days ago
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	assert.NoError(t, err)
	err = os.Chtimes(cacheDir, oldTime, oldTime)
	assert.NoError(t, err)

	s := New(zap.NewNop())
	target := Target{
		Name:      "Folder Test",
		Path:      cacheDir,
		Threshold: 7 * 24 * time.Hour,
		Type:      "folder",
	}

	result, err := s.Scan(target)
	assert.NoError(t, err)
	assert.Len(t, result.Files, 1)
	assert.Equal(t, cacheDir, result.Files[0])
	assert.Equal(t, int64(len("old content")), result.TotalSize)
}
