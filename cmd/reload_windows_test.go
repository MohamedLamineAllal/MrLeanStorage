//go:build windows

package cmd

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestStartReloadListenerAndProcesses verifies the platform-specific reload mechanisms on Windows.
// It spins up the TCP listener, sends a reload trigger connection, and asserts the callback is fired.
func TestStartReloadListenerAndProcesses(t *testing.T) {
	logger := zap.NewNop()

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
}
