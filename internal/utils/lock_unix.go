//go:build !windows

// Package utils implements common helper functions for filesystem paths, caching,
// and cross-platform process locking.
package utils

import (
	"fmt"
	"os"
	"syscall"
)

// AcquireProcessLock attempts to acquire an exclusive, non-blocking lock on the file
// at the specified path using syscall.Flock. This ensures single-instance execution
// of background cleanup processes on macOS and Linux.
//
// If the lock is successfully acquired, it returns the locked file handle and 0.
// If the lock is already held by another active process, it returns nil and the PID
// of that running process (read from the lockfile).
func AcquireProcessLock(path string) (*os.File, int, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open lockfile: %w", err)
	}

	// Attempt to acquire an exclusive lock in a non-blocking manner.
	// If another process holds the lock, this call immediately returns syscall.EWOULDBLOCK.
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		// Read the PID of the existing running process from the lockfile
		var pid int
		_, _ = fmt.Fscanf(file, "%d", &pid)
		file.Close()
		return nil, pid, fmt.Errorf("process lock is already held by another instance")
	}

	// Truncate the file and write the current process's PID to it
	_ = file.Truncate(0)
	_, _ = file.Seek(0, 0)
	_, _ = fmt.Fprintf(file, "%d\n", os.Getpid())
	_ = file.Sync()

	return file, 0, nil
}
