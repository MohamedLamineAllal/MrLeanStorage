package cleaner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClean(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mls-clean-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	err = os.WriteFile(file1, []byte("content1"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0644)
	assert.NoError(t, err)

	c := New(zap.NewNop(), false, nil) // Not dry run
	count, size, err := c.Clean([]string{file1, file2}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, int64(16), size)

	_, err = os.Stat(file1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(file2)
	assert.True(t, os.IsNotExist(err))
}

func TestCleanDryRun(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mls-clean-dry-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	file1 := filepath.Join(tempDir, "file1.txt")
	err = os.WriteFile(file1, []byte("content1"), 0644)
	assert.NoError(t, err)

	c := New(zap.NewNop(), true, nil) // Dry run
	count, size, err := c.Clean([]string{file1}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, int64(8), size)

	_, err = os.Stat(file1)
	assert.NoError(t, err) // Should still exist
}
