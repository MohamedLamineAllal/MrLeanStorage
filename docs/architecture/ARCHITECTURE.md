# Architecture: MacosLeanStorage

## 1. Project Goal
Build a high-performance, safe, and efficient storage cleanup tool for macOS, focused on developer and browser data.

## 2. Go Project Structure
```text
/cmd
  /mls          - Main CLI entry point
/internal
  /scanner      - Logic for traversing directories and analyzing file metadata
  /cleaner      - Safe deletion logic with dry-run support
  /config       - Configuration management (YAML)
  /scheduler    - Background execution and cron-like logic
  /models       - Shared data structures (Rule, Target, etc.)
/pkg
  /utils        - Shared utilities (size formatting, path expansion)
```

## 3. Design Choices

### Staleness Check Optimization
* **Hybrid Adaptive Staleness Check**: For folder-level cleanup, the scanner first checks the folder `mtime` (Fast Path). If the folder `mtime` is recent, a recursive deep scan (Slow Path) is triggered only if necessary to confirm content staleness. This ensures high performance for most directories while maintaining accuracy even when system metadata is updated by external factors (e.g., Finder access).

### Safety
*   **Dry Run**: Every operation defaults to a dry run. The user must explicitly pass a flag to delete.
*   **Immutable Rules**: Rules for "100% safe" directories are hardcoded or verified by a strict schema.
*   **Exclusion List**: Built-in protection for critical system folders (`/System`, `/Library/CoreServices`).

### Exclusion/Ignore Mechanism
* **Ignore Patterns**: A flexible system to ignore files or folders based on user-provided glob patterns (e.g., `.DS_Store`). This is configurable both globally and per-target, and is integrated into both file walking and deep staleness checks for high performance.
* **Recursive Globbing**: Supports the `**` recursive wildcard pattern (e.g., `**/Cache/**`) via `doublestar` to allow for deep path resolution within complex directory structures.
* **Command-based Cleanup**: Allows for execution of pre-defined system commands (e.g., `pnpm store prune`) to supplement file-based cleanup tasks. Commands are executed during the `clean` phase, respecting the global dry-run mode and an optional `interval_days` property to avoid frequent, unnecessary execution.
    * **Persistence**: The last execution time of a command target is persisted in the system temporary directory as a file named `mls-cmd-<target_name>.lastrun`. The scheduler checks this timestamp against `interval_days` before deciding whether to execute the command.



## 4. Library Choices
*   **CLI**: `github.com/spf13/cobra` - Industry standard for Go CLIs.
*   **Config**: `github.com/spf13/viper` - Excellent for YAML/JSON and env vars.
*   **Scheduling**: `github.com/robfig/cron/v3` - For background daemon tasks.
*   **Formatting**: `github.com/dustin/go-humanize` - For readable bytes (e.g., "1.2 GB").
*   **Logging**: `go.uber.org/zap` - For high-performance structured logging.

## 5. File/Directory Monitoring
While `fsnotify` is great for real-time, storage cleanup is better suited for **Periodic Polling** (Cron) because:
1.  Cache files are created/deleted constantly; real-time monitoring would be too noisy.
2.  We care about "staleness" over time, not instantaneous change.
3.  Polling every 24 hours is much lighter on CPU/Battery than keeping a file watcher active on `~/Library`.
