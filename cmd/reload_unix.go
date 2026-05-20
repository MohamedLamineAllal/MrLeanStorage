//go:build !windows

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

// reloadSignals lists the signals that trigger a configuration reload on Unix systems.
var reloadSignals = []os.Signal{syscall.SIGHUP}

// listenSignals lists the signals we subscribe to for shutdown or configuration reloading on Unix.
var listenSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}

// startReloadListener is a no-op on Unix since SIGHUP is natively event-driven via the OS.
func startReloadListener(logger *zap.Logger, reloadFn func()) {
	// No-op on Unix since SIGHUP signals are used natively.
}

// reloadProcesses sends a SIGHUP signal to all active "mls serve" processes on Unix systems.
func reloadProcesses() error {
	pids, err := findMLSServePIDs()
	if err != nil {
		return err
	}

	var signalErrors []string
	for _, pid := range pids {
		if pid == os.Getpid() {
			continue
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			signalErrors = append(signalErrors, fmt.Sprintf("PID %d: %v", pid, err))
			continue
		}
		if err := process.Signal(syscall.SIGHUP); err != nil {
			signalErrors = append(signalErrors, fmt.Sprintf("PID %d: %v", pid, err))
		} else {
			colorSuccess.Printf("Reload signal sent to process %d successfully\n", pid)
		}
	}

	if len(signalErrors) > 0 {
		return fmt.Errorf("failed to signal some processes: %s", strings.Join(signalErrors, "; "))
	}
	return nil
}

// findMLSServePIDs queries the system using pgrep to find all process IDs (PIDs)
// matching "mls serve". It filters out empty lines and invalid PIDs to ensure
// signal execution targets only valid running server processes.
func findMLSServePIDs() ([]int, error) {
	out, err := exec.Command("pgrep", "-f", "mls serve").Output()
	if err != nil {
		// pgrep exits with status 1 if no process matches
		return nil, fmt.Errorf("no running mls serve process found")
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var pids []int
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err == nil {
			pids = append(pids, pid)
		}
	}
	if len(pids) == 0 {
		return nil, fmt.Errorf("no running mls serve process found")
	}
	return pids, nil
}
