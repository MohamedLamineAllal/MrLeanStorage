# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
