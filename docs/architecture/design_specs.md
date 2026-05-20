# MrLeanStorage (mls) Design Specifications

This document outlines the low-level design decisions for the `mls` tool, focusing on performance, efficiency, and safety.

## 1. Concurrent Orchestration Architecture

### Overlapping Scan-and-Clean
To maximize I/O throughput and minimize execution time:
- **Streaming Pipeline**: The `Engine` uses channels to stream discovered targets from the scanner directly to the cleaner.
- **Worker Pools**: 
    - **Scanner**: Uses a pool of goroutines to perform high-speed `os.ReadDir` traversals.
    - **Cleaner**: Uses a separate pool of goroutines to perform deletions in parallel, preventing I/O wait in one thread from blocking others.
- **Aggregator**: Collects real-time statistics (size, count) as the pipeline executes, providing immediate feedback.

### Performance Optimizations
- **Batched I/O**: Uses `os.ReadDir` instead of `filepath.Walk` to reduce system calls by reading multiple directory entries in a single operation.
- **Memory Efficiency**: Avoids loading the entire file tree into memory. Only identified cleanup targets are buffered in channels.
- **Lock-Free Concurrency**: Uses atomic counters and thread-safe maps for stats collection to avoid mutex contention in high-volume scans.

## 2. Background Automation (Agent)

### Native Lifecycle Management
`mls` delegates process management to the OS to ensure high quality and system integration:
- **macOS (`launchd`)**:
    - **Mechanism**: `LaunchAgents` via `.plist`.
    - **Why**: Native support for "KeepAlive" and "RunAtLoad". It's the standard way to run user-scoped background tasks on macOS.
- **Linux (`systemd`)**:
    - **Mechanism**: User units (`systemctl --user`).
    - **Why**: Provides superior logging integration (`journald`) and dependency management (e.g., start after network or local-fs).
- **Windows (Scheduled Tasks)**:
    - **Mechanism**: `schtasks` with `/sc onlogon`.
    - **Why**: Most reliable way to run a background task in a user session without the complexity of a full Windows Service.

## 3. Communication & Control

### Zero-Idle Signal Handling
- **Unix (SIGHUP)**: Instant reload without process restart.
- **Windows (TCP Loopback)**: 
    - **Design**: The `mls serve` process opens a random local port and writes it to `mls.port` in the cache directory. `mls config reload` reads this port and sends a single byte to trigger the reload.
    - **Performance**: Zero CPU cycles spent polling; the listener blocks in the kernel until a connection arrives.

## 4. Safety & Integrity

### Process Isolation
- **Advisory Locking**: Uses `syscall.Flock` (Unix) and exclusive file handles (Windows) to prevent data corruption from concurrent instances.
- **Dry-Run Enforcement**: The background agent is hardcoded to run with `dry_run: false` to ensure automation is effective, while manual CLI commands default to `true` for user safety.

## 5. Library Selections (Finalized)

- **Engine**: Standard Library (`os`, `sync/atomic`, `runtime`).
- **Patterns**: `github.com/bmatcuk/doublestar/v4` (Performance-optimized globbing).
- **CLI**: `github.com/spf13/cobra`.
- **Config**: `github.com/spf13/viper`.
- **Logging**: `go.uber.org/zap`.
- **Progress**: `github.com/schollz/progressbar/v3`.
