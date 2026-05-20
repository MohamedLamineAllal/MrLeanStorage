# Project Memory Index

## Current Status
- **Phase:** Stable / Maintenance / Deployment & Distribution
- **Core Functionality:** Performance-optimized directory scanning (single-pass) and concurrent deletion (goroutine worker pools).
- **Background Automation:** Full daemon lifecycle command suite (`install`, `start`, `stop`, `status`, `restart`, `uninstall`) for macOS `launchd` background services.
- **Reliability:** Graceful hot-reloads via `SIGHUP` signal and 30-minute missed-task catch-up ticker to handle sleep/wake schedules.
- **Persistence:** Full migration from volatile `/tmp` to local application caches (`~/Library/Caches/mls`).
- **Distribution:** Automatic release builds for Darwin/Linux/Windows via GoReleaser and deployment to Homebrew Tap `homebrew-mls`.

## Completed Tasks & Milestones
- [x] Engine Scan & Clean Optimization: Single-pass scanning (`os.ReadDir`) and concurrent cleaning (worker pools).
- [x] Comprehensive Testing: Thread-safety validation using `go test -race` for scan/clean engine and commands.
- [x] Background Daemon: Standardized `mls agent` commands to control and install/uninstall macOS background service plist.
- [x] Sleep-Wake Ticker: Missed task recovery ticker running every 30 minutes in background daemon.
- [x] Dynamic Hot-Reload: Instant runtime configuration reloading via `SIGHUP`.
- [x] Storage Persistence: App Cache Directory relocation (`~/Library/Caches/mls`).
- [x] Stats discrepancy audit: Analysis on scanned vs deleted stats (`docs/Stats_Counting.md`).
- [x] Automated Releases: End-to-end multi-platform build and deployment setup using GitHub Actions and GoReleaser.
- [x] Homebrew Integration: Setup and deployment via custom Tap `homebrew-mls` utilizing fine-grained PAT and quarantine-bypassing post-install hook.
- [x] Cask Distribution Alignment: Resolved formula checksum mismatch, synchronized tap cache, and fully updated docs (INSTALL.md, USER_GUIDE.md, ARCHITECTURE.md, README.md) to reflect the GoReleaser v2 Cask standard.
- [x] Active Background Deletion Mode: Configured the background scheduler (`mls serve` / `mls agent`) to always run in active deletion mode (`dry_run: false`) regardless of global configuration settings. Added highly visible warning callouts in the README.md and USER_GUIDE.md.
- [x] Background Engine Alignment: Refactored the background serve daemon (`cmd/serve.go`) to utilize the unified orchestration `Engine.ScanAndClean()`, adding support for concurrent scanning/deletion, results deduplication, and execution of system command targets on schedule.
- [x] Multi-PID Config Reload & Silenced CLI Usage: Enhanced config reload command to support signaling multiple active serve processes, resolved channel race conditions for SIGHUP to ensure stable runtime hot-reloads without process termination, and globally silenced confusing Cobra CLI usage menus on runtime/operational errors.
- [x] Single-Instance Serve Enforcement: Implemented a robust cross-platform locking mechanism using syscall.Flock (macOS/Linux) and exclusive-file fallbacks (Windows) to prevent concurrent instances of `mls serve` from running, printing a friendly already-running message (with the active PID) and exiting cleanly.


## Core Project Documentation
- [Cleanup Estimation Discrepancy Analysis](./docs/Stats_Counting.md)
- [Background Service Setup](./README.md#background-automation-macos)
- [Standardized Configuration](./docs/configuration/Examples/default.yml)
- [Release Process Guidelines](./docs/RELEASE_PROCESS.md)
- [Testing Guide](./docs/dev/TESTING.md)
- [System Architecture](./docs/architecture/ARCHITECTURE.md)

## Pending / Future Considerations
- Regular audit and updates of target app cache paths in `default.yml` & `Extensive.yml`.
- Monitoring Homebrew tap installations and binary signature notarization for newer macOS updates.
- Refinement of cleanup duration stats to provide execution-time metrics.
