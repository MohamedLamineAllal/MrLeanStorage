# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.7] - 2026-05-20

### Added
- **Made default config Cross-Platform**: Added default configurations for Linux and Windows. And updated docs.

## [0.1.6] - 2026-05-20

### Added
- **Cross-Platform Agent Log**: New `mls agent log` command with `--live`, `--path`, and `--open` support across macOS, Linux, and Windows.

## [0.1.5] - 2026-05-20

### Added
- **Cross-Platform Agent Management**: Native background service management for Linux (`systemd`) and Windows (`Scheduled Tasks`).
- **High-Performance Orchestration**: Overlapping scan-and-clean engine to minimize I/O wait times.
- **Event-Driven Reloads**: Zero-idle CPU config reloading using `SIGHUP` (Unix) and TCP loopback (Windows).
- **Single-Instance Enforcement**: Cross-platform process locking to prevent concurrent instances of `mls serve`.

### Changed
- **Architecture Docs**: Comprehensive update with technical justifications for cross-platform design choices.
- **Agent Guidelines**: Mandated append-only logging for better traceability in `AGENTS.md`.

### Fixed
- **Windows Tests**: Resolved symbol resolution errors in `reload_test.go` by splitting platform-specific test logic.

## [0.1.4] - 2026-05-20

### Added
- **Cross-Platform Helpers**: Robust `OpenPath` and `RevealPath` utilities for macOS, Windows, and Linux.
- **Homebrew Cask**: Transitioned to Cask distribution to bypass macOS quarantine via post-install hooks.

### Fixed
- **Config Reload**: Resolved SIGHUP channel race conditions ensuring stable hot-reloads without process termination.
- **PID Discovery**: Enhanced PID lookup to support signaling multiple active `mls serve` instances.

## [0.1.0] - 2026-05-19

### Added
- **Background Agent**: Full lifecycle management (`install`, `start`, `stop`, `restart`, `status`, `uninstall`) for macOS `launchd` background agents.
- **Auto-Catchup**: Periodic 30-minute ticker to check and run missed scheduled tasks after wake or suspension.
- **Config Reload**: Support for `mls config reload` (via `SIGHUP`) to refresh configuration without restarting the daemon.
- **CI/CD**: GitHub Actions workflow for cross-platform binary distribution (macOS, Linux, Windows).
- **Documentation**: Rigorous GoDoc pass and release process documentation.

### Changed
- **Scanner**: Refactored to single-pass traversal for significantly better performance.
- **Cleaner**: Parallelized deletion using a worker pool.
- **Storage**: Migrated all temporary state files (`/tmp` -> `~/Library/Caches/mls`) for better persistence.
- **Configuration**: Standardized target configurations and added default template examples.

### Fixed
- **Scheduling**: Fixed cron expression parsing (updated to 6-field format).
- **Agent Lifecycle**: Resolved startup hangs by correctly loading and kicking off `launchd` services.
