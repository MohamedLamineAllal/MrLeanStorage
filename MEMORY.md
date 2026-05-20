# Project Memory Index

## Current Status
- **Phase:** Stable / Maintenance
- **Core Functionality:** Performance-optimized scanning (single-pass) and parallelized cleaning.
- **Background Automation:** Full lifecycle management (`install`, `start`, `stop`, `restart`, `uninstall`) for macOS `launchd` background agents.
- **Persistence:** All state files and logs migrated to persistent cache (`~/Library/Caches/mls`).
- **Reliability:** 30-minute missed-task catch-up ticker and graceful config reloading via `SIGHUP`.

## Documentation
- [Cleanup Estimation Discrepancy Analysis](./docs/Stats_Counting.md)
- [Background Service Setup](./README.md#background-automation-macos)
- [Standardized Configuration](./docs/configuration/Examples/default.yml)

## Pending / Future Considerations
- Periodic review of cache locations for newly integrated apps.
- Potential further optimization for extremely high-concurrency target sets.
