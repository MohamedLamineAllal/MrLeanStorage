package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ExpandPath expands tilde (~) to the user home directory and environment variables
// (e.g., %LOCALAPPDATA% on Windows).
func ExpandPath(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	// Expand environment variables (e.g., %LOCALAPPDATA%)
	path = os.ExpandEnv(path)

	// Expand tilde (~)
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}

	return filepath.Clean(path), nil
}

// GetAppCacheDir returns the persistent cache directory for the application.
// It falls back to the system temp directory if the user cache dir is unavailable.
func GetAppCacheDir() string {
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = os.TempDir()
	}
	path := filepath.Join(cache, "mls")
	_ = os.MkdirAll(path, 0755)
	return path
}

// OpenPath opens a file or directory using the default system handler in a
// cross-platform way. It supports macOS (open), Windows (cmd start), and Linux (xdg-open).
func OpenPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default: // linux, freebsd, openbsd, etc.
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Run()
}

// RevealPath opens the system file manager and highlights/reveals the file.
// It supports macOS Finder selection (open -R) and Windows Explorer selection (explorer /select).
// On Linux/Unix, it falls back to opening the parent directory of the file using xdg-open.
func RevealPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	case "windows":
		cmd = exec.Command("explorer", "/select,", path)
	default: // linux, freebsd, etc.
		parentDir := filepath.Dir(path)
		cmd = exec.Command("xdg-open", parentDir)
	}
	return cmd.Run()
}
