# MacosLeanStorage User Guide

## Introduction
MacosLeanStorage (mls) is a command-line tool designed to help macOS users reclaim storage space by cleaning up old cache and temporary files. It specifically targets directories that tend to grow large over time, such as browser caches and developer tool temporary files.

## Getting Started

### Prerequisites
- macOS
- Go (if building from source)

### Initial Setup
When you first run `mls`, it creates a default configuration file in your home directory: `~/.MacosLeanStorage.yaml`. 

To see where it is or to reveal it in Finder, run:
```bash
mls config reveal
```
To open it in your default editor, run:
```bash
mls config open
```

## Commands

### `scan`
The `scan` command analyzes configured targets and lists files that match cleanup criteria.

**Output Information**:
- For path-based targets, it displays matched files and the total size that would be cleaned.
- For command-based targets, it displays the command to be executed, the configured interval, and the scheduled "Next Run" time.

```bash
mls scan
```

**New in v1.1:**
- **Concise Output**: `mls scan` now provides a per-target summary by default. 
- **Detailed Match Listing**: Individual files/folders are only listed if there are 10 or fewer matches.
- **Verbose Flag**: To see a full list of all matching items regardless of the count, use the `--verbose` or `-v` flag:
  ```bash
  mls scan --verbose
  ```

### `clean`
The `clean` command performs a scan and then deletes the matched files. 
**Note:** By default, `clean` runs in **Dry Run** mode. It will show you what it *would* delete but won't actually perform the deletion.

```bash
mls clean
```

To actually delete the files, you have two options:
1. Set `dry_run: false` in your `~/.MacosLeanStorage.yaml`.
2. Use the command line flag to override the default:
   ```bash
   mls clean --dry-run=false
   ```

### `serve`
The `serve` command starts a background scheduler that runs the cleanup process according to the `schedule` defined in your configuration file.

```bash
mls serve
```
This is ideal for keeping your Mac lean without manual intervention.

### `config open`
Opens the configuration file in the default application (e.g., your preferred text editor).

```bash
mls config open
```

### `config reveal`
Reveals the configuration file in Finder.

```bash
mls config reveal
```

### Configuration Guide

The configuration file is written in YAML.

### `path`
The directory path to monitor. You can use standard file globbing.
- **Recursive Globbing**: You can use the `**` pattern to match directories recursively (e.g., `~/Library/Application Support/MyApp/**/Cache/*`). This allows for deep path resolution.


### `targets`
A list of directories to monitor or commands to execute.
- `name`: A descriptive name for the target.
- `path`: The absolute path or a home-relative path (using `~/`).
- `command`: A system command to run (e.g., `pnpm store prune`).
- `interval_days`: Minimum days to wait before running this command again.
- `threshold_days`: Items older than this many days will be targeted for cleanup.
- `type`: Defines how the target should be cleaned. Available values:
    - `"file"` (default): Scans inside the path and deletes individual old files.
    - `"folder"`: Treats the matched path itself as the target. If the folder is old enough, the entire folder is deleted.
    - `"both"`: Deletes both old files inside and the folder itself if staleness criteria are met.
- `safety_level`: (Reserved for future use) Intended to define how aggressive the cleanup should be.

### `dry_run`
Global safety switch. If `true`, no files are ever deleted.

### `schedule`
A cron expression (e.g., `"0 0 * * *"`) defining when the automated cleanup should run.

## Tips for Safe Cleanup
1.  **Always scan first**: Before running `clean`, run `scan` to see what files are being targeted.
2.  **Start with high thresholds**: Set `threshold_days` to a higher value (e.g., 30) and gradually decrease it as you gain confidence.
3.  **Use Descriptive Names**: Help yourself by naming targets clearly (e.g., "Arc Browser Cache").

## Troubleshooting
If `mls` fails to delete a file, it might be because the file is currently in use by another application. Close the application and try again.
