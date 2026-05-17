# Memory Index

## Current Phase: Implementation & Refinement
The project is currently in the **Development & Feature Refinement** phase. We have successfully scaffolded the core components, implemented the scanner, cleaner, and scheduler, and recently integrated the full cleanup inventory.

## Completed Tasks
- **System Architecture**: Established Go project structure (`cmd`, `internal/scanner`, `internal/cleaner`, `internal/config`, `internal/scheduler`).
- **Configuration Management**: Implemented YAML-based config with `Viper`.
- **Scanner**: Implemented robust, concurrent directory traversal with support for glob patterns for multi-profile directory support.
- **Cleaner**: Implemented safe file deletion with dry-run support by default.
- **Scheduler**: Implemented a periodic task scheduler with persistent last-run state tracking to handle missed task catch-up on wake/startup.
- **Cleanup Inventory**: Populated configuration with the full list of safe cleanup targets (Browsers, IDEs, Build Tools).
- **Compliance**: Restored architectural compliance to `AGENTS.md` mandates, including `Prompts.log` enforcement.

## Pending Tasks
- Add more robust error handling and telemetry (Zap).
- Perform stress testing on directory traversal (large caches).
- Future: Add automated release/build process (Cobra/Go).

## Decisions & Notes
- **Glob Support**: Scanner now uses `filepath.Glob` to handle paths like `.../User Data/*/Cache`.
- **Catch-up Logic**: Scheduler tracks last successful run via `~/.MacosLeanStorage.lastrun` and checks for execution on startup if the last run was > 23 hours ago.
- **Safety First**: Dry-run mode remains the default for all cleanup operations.
