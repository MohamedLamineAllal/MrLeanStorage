# Project Memory Index

## Current Status
- **Phase:** Stable / Maintenance / Deployment & Distribution
- **Version:** v0.1.6 (Planned/Development)
- **Core Functionality:** Performance-optimized directory scanning (single-pass) and concurrent deletion (goroutine worker pools).
- **Background Automation:** Native background service management for macOS (launchd), Linux (systemd), and Windows (Scheduled Tasks).
- **Reliability:** Graceful hot-reloads via `SIGHUP` signal/TCP loopback and 30-minute missed-task catch-up ticker to handle sleep/wake schedules.
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
- [x] Cross-Platform Config Open & Reveal: Created robust OpenPath and RevealPath system helpers in internal/utils supporting macOS, Windows, and Linux natively, and refactored config commands to use them.
- [x] Cross-Platform Documentation & CLI Help Alignment: Fully updated README.md, docs/USER_GUIDE.md, and cmd/config.go help text to remove macOS-only notes for config open/reveal and describe the cross-platform file explorers.
- [x] Strict Prompt Logging Rule: Updated AGENTS.md to mandate that user prompts are logged exactly as they are given, without paraphrasing or truncating.
- [x] Purely Event-Driven Cross-Platform Config Reload: Refactored the config reload mechanism to be purely event-driven, removing all periodic file polling/tickers. Uses OS-native SIGHUP signaling on Unix (macOS/Linux) and an OS-allocated local loopback TCP port listener on Windows that blocks on Accept(), guaranteeing zero idle CPU and memory usage.
- [x] Cross-Platform Agent Management (v0.1.5): Implemented native background service management for macOS (launchd), Linux (systemd), and Windows (Scheduled Tasks). Split `agent.go` into platform-specific files (`agent_darwin.go`, `agent_windows.go`, `agent_linux.go`, and `agent_other.go`) and updated documentation to reflect cross-platform support.
- [x] Agent Guidelines Update (v0.1.5): Mandated append-only logging for `Prompts.log` and `ACTIONS.log` in `AGENTS.md` to ensure complete interaction history.
- [x] Cross-Platform Agent Logs (v0.1.6): Implemented `mls agent log` command with `--live`, `--path`, and `--open` support across macOS, Linux, and Windows. Updated background service definitions to redirect output to `agent.log`.


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
- Implemented multi-platform configuration support via compile-time build tags.
- [x] Fixed build import error: Corrected incorrect module path in internal/config/config.go to match go.mod.
- Added ExpandPath utility for environment variable and tilde expansion.
- Organized documentation examples for macOS, Windows, and Linux.
