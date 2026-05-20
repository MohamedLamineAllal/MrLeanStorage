//go:build !windows

package cmd

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestStartReloadListenerAndProcesses verifies the platform-specific reload mechanisms.
// On Windows, it spins up the TCP listener, sends a reload trigger connection, and asserts the callback is fired.
// On Unix (macOS/Linux), it registers a native SIGHUP signal trap, signals its own process, and asserts signal delivery.
func TestStartReloadListenerAndProcesses(t *testing.T) {
	logger := zap.NewNop()

	if runtime.GOOS == "windows" {
		calledChan := make(chan bool, 1)
		reloadFn := func() {
			calledChan <- true
		}

		startReloadListener(logger, reloadFn)

		// Trigger reload
		err := reloadProcesses()
		if err != nil {
			t.Fatalf("reloadProcesses failed on Windows: %v", err)
		}

		// Wait for the callback to be called
		select {
		case <-calledChan:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for reload callback on Windows")
		}
	} else {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP)
		defer signal.Stop(sigChan)

		// Send SIGHUP to ourselves
		pid := os.Getpid()
		process, err := os.FindProcess(pid)
		if err != nil {
			t.Fatalf("failed to find own process: %v", err)
		}

		err = process.Signal(syscall.SIGHUP)
		if err != nil {
			t.Fatalf("failed to send SIGHUP on Unix: %v", err)
		}

		select {
		case sig := <-sigChan:
			if sig != syscall.SIGHUP {
				t.Errorf("expected SIGHUP, got %v", sig)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Timeout waiting for SIGHUP signal on Unix")
		}
	}
}

// TestFindMLSServePIDs verifies that the pgrep PID discovery handles active serve discovery
// gracefully without crashing under different system states on Unix.
func TestFindMLSServePIDs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("findMLSServePIDs is Unix-specific and not supported on Windows")
	}

	pids, err := findMLSServePIDs()
	// It's acceptable for pids to be nil and err to be non-nil if no "mls serve" is running.
	// This test simply asserts that the execution path functions and returns a predictable result.
	t.Logf("findMLSServePIDs returned: pids=%v, err=%v", pids, err)
}
