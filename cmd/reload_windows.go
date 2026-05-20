//go:build windows

package cmd

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/mohamedlamineallal/MrLeanStorage/internal/utils"
	"go.uber.org/zap"
)

// reloadSignals lists the signals that trigger a configuration reload on Windows systems (none).
var reloadSignals = []os.Signal{}

// listenSignals lists the signals we subscribe to for shutdown on Windows.
var listenSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

// startReloadListener starts a local loopback TCP listener on Windows to receive reload signals natively.
func startReloadListener(logger *zap.Logger, reloadFn func()) {
	portPath := filepath.Join(utils.GetAppCacheDir(), "mls.port")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logger.Error("Failed to start reload port listener", zap.Error(err))
		return
	}

	// Write the port to mls.port
	addr := listener.Addr().(*net.TCPAddr)
	portStr := strconv.Itoa(addr.Port)
	if err := os.WriteFile(portPath, []byte(portStr), 0644); err != nil {
		logger.Error("Failed to write reload port file", zap.Error(err))
		listener.Close()
		return
	}

	go func() {
		defer func() {
			listener.Close()
			_ = os.Remove(portPath)
		}()
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Listener closed on exit
			}
			conn.Close() // Close connection immediately
			logger.Info("Reload signal received via local port")
			reloadFn()
		}
	}()
}

// reloadProcesses connects to the running "mls serve" process's local port to trigger configuration reloading on Windows.
func reloadProcesses() error {
	portPath := filepath.Join(utils.GetAppCacheDir(), "mls.port")
	data, err := os.ReadFile(portPath)
	if err != nil {
		return fmt.Errorf("no running mls serve process found (failed to read port: %w)", err)
	}
	portStr := strings.TrimSpace(string(data))
	conn, err := net.Dial("tcp", "127.0.0.1:"+portStr)
	if err != nil {
		return fmt.Errorf("could not connect to running mls serve process: %w", err)
	}
	conn.Close()
	colorSuccess.Println("Reload signal sent to running serve process successfully.")
	return nil
}
