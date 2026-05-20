//go:build windows

package cmd

import (
	"os"
	"syscall"
)

// reloadSignals lists the signals that trigger a configuration reload on Windows systems (none).
var reloadSignals = []os.Signal{}

// listenSignals lists the signals we subscribe to for shutdown on Windows.
var listenSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

// reloadProcesses is a no-op on Windows because SIGHUP is unsupported.
// Configuration reload on Windows relies purely on the reload.signal file-based mechanism.
func reloadProcesses() error {
	return nil
}
