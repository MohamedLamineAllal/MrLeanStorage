//go:build windows

// Package utils implements common helper functions for filesystem paths, caching,
// and cross-platform process locking.
package utils

import (
	"fmt"
	"os"
)

// AcquireProcessLock attempts to open the file at the specified path exclusively.
// On Windows, since syscall.Flock is unavailable, we rely on exclusive file open/creation
// flags (os.O_EXCL) to prevent concurrent instances of background processes from executing.
//
// If the lock is successfully acquired, it returns the locked file handle and 0.
// If the lock is already held by another active process, it returns nil and the PID
// of that running process.
func AcquireProcessLock(path string) (*os.File, int, error) {
	// Attempt exclusive file creation/opening to detect if the lockfile already exists
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		// Read the PID of the existing running process from the lockfile
		existingFile, readErr := os.Open(path)
		var pid int
		if readErr == nil {
			_, _ = fmt.Fscanf(existingFile, "%d", &pid)
			existingFile.Close()
		}
		return nil, pid, fmt.Errorf("process lock is already held by another instance")
	}

	// Write the current process's PID to the lockfile
	_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
	_ = file.Sync()

	return file, 0, nil
}
