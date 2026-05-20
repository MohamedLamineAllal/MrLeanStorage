//go:build !darwin && !windows && !linux

package cmd

import (
	"fmt"
	"runtime"
)

// InstallAgent returns an error indicating that the background agent is not supported on the current platform.
func InstallAgent() error {
	return fmt.Errorf("background agent is not supported on %s", runtime.GOOS)
}

// UninstallAgent returns an error indicating that the background agent is not supported on the current platform.
func UninstallAgent() error {
	return fmt.Errorf("background agent is not supported on %s", runtime.GOOS)
}

// StartAgent returns an error indicating that the background agent is not supported on the current platform.
func StartAgent() error {
	return fmt.Errorf("background agent is not supported on %s", runtime.GOOS)
}

// StopAgent returns an error indicating that the background agent is not supported on the current platform.
func StopAgent() error {
	return fmt.Errorf("background agent is not supported on %s", runtime.GOOS)
}

// StatusAgent returns an error indicating that the background agent is not supported on the current platform.
func StatusAgent() error {
	return fmt.Errorf("background agent is not supported on %s", runtime.GOOS)
}
