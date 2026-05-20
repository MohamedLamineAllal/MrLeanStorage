# Architecture: MrLeanStorage

## 1. Project Goal
Build a high-performance, safe, and efficient storage cleanup tool for macOS, Linux, and Windows, focused on developer and browser data.

## 2. Go Project Structure
```text
/cmd
  /mls          - Main CLI entry point
/internal
  /scanner      - Logic for traversing directories and analyzing file metadata
  /cleaner      - Safe deletion logic with goroutine worker pools
  /engine       - Orchestration layer for scan-and-clean workflows
  /config       - Configuration management (YAML + Viper)
  /scheduler    - Background execution and cron-like logic
  /utils        - Shared utilities (cross-platform path resolution, locking)
```

## 3. Design Choices

### High-Performance Orchestration
* **Unified Engine**: The `Engine.ScanAndClean()` method provides a concurrent orchestration layer that overlaps scanning and cleaning. This minimizes idle I/O time by allowing the cleaner to start processing as soon as the scanner identifies the first batch of targets.
* **Single-Pass Traversal**: Uses high-performance `os.ReadDir` instead of recursive `filepath.Walk`. This reduces system call overhead and allows for efficient result aggregation in a single pass.

### Cross-Platform Background Automation
Instead of writing a custom service manager, `mls` leverages native OS primitives for background execution to ensure maximum reliability and minimum resource overhead:
* **macOS (`launchd`)**: Uses `LaunchAgents` via `.plist` files. This is the **best option** for macOS as it integrates with the system's power management and lifecycle, ensuring the agent is revived if it crashes and respects system sleep cycles.
* **Linux (`systemd`)**: Uses `systemd` user units. This is the **best option** for modern Linux distributions, providing robust logging (via `journalctl`), dependency management, and process isolation without requiring root privileges.
* **Windows (Scheduled Tasks)**: Uses `schtasks`. This is the **best option** for Windows as it provides a native way to run background tasks at logon or specific intervals with built-in retry logic and minimal impact on system performance.

**Potential Improvements**: 
- For Windows, implementing a full **Windows Service** (using `kardianos/service`) could offer better lifecycle management for multi-user environments, though Scheduled Tasks are currently preferred for their simplicity and low overhead for per-user cleanup.
- For Linux, adding support for `OpenRC` or `runit` could expand compatibility for non-systemd distributions.

### Event-Driven Configuration Reload
* **Unix (SIGHUP)**: Uses standard POSIX signals to trigger hot-reloads. This is the **best option** as it is handled natively by the kernel and consumes zero resources while idle.
* **Windows (TCP Loopback)**: Since Windows lacks a direct equivalent to SIGHUP for CLI tools, `mls` implements a local loopback TCP listener that blocks on `Accept()`. 
    - **Why it's best**: Guarantees zero CPU usage while waiting for a reload signal, unlike file polling which consumes I/O and CPU cycles.
    - **Potential Better Solution**: Using **Named Pipes** on Windows could be slightly more efficient and more idiomatic for IPC, though TCP loopback is highly portable and robust.

### Single-Instance Enforcement
* **Cross-Platform Locking**: Uses `syscall.Flock` on Unix and exclusive file creation on Windows.
    - **Why it's best**: Prevents race conditions and database corruption by ensuring only one `mls serve` instance runs at a time.
    - **Potential Better Solution**: A socket-based lock (binding to a specific port) could also provide a way to communicate with the running instance (e.g., to query status), but file-based locking is more resilient to abrupt terminations.

## 4. Library Choices
* **CLI**: `github.com/spf13/cobra` - Industry standard for Go CLIs.
* **Config**: `github.com/spf13/viper` - Excellent for YAML/JSON and env vars.
* **Scheduling**: `github.com/robfig/cron/v3` - Supports 6-field standard (including seconds).
* **Logging**: `go.uber.org/zap` - For high-performance structured logging.
* **Patterns**: `github.com/bmatcuk/doublestar/v4` - High-performance recursive globbing.

## 5. File/Directory Monitoring
While `fsnotify` is great for real-time, storage cleanup is better suited for **Periodic Polling** (Cron) because:
1. Cache files are created/deleted constantly; real-time monitoring would be too noisy.
2. We care about "staleness" over time, not instantaneous change.
3. Polling allows for batching deletions, which is significantly more I/O efficient than individual deletions.
