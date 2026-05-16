# MacosLeanStorage User Guide

## Introduction
MacosLeanStorage (mls) is a command-line tool designed to help macOS users reclaim storage space by cleaning up old cache and temporary files. It specifically targets directories that tend to grow large over time, such as browser caches and developer tool temporary files.

## Getting Started

### Prerequisites
- macOS
- Go (if building from source)

### Initial Setup
When you first run `mls`, it creates a default configuration file in your home directory: `~/.MacosLeanStorage.yaml`. 

To see where it is or to edit it, run:
```bash
mls config open
```
This will reveal the file in Finder.

## Commands

### `scan`
The `scan` command analyzes all configured targets and lists files that exceed the `threshold_days`. It does not delete anything. Use this to see what `mls` has found.

```bash
mls scan
```

### `clean`
The `clean` command performs a scan and then deletes the matched files. 
**Note:** By default, `clean` runs in **Dry Run** mode. It will show you what it *would* delete but won't actually perform the deletion.

```bash
mls clean
```

To actually delete the files, you have two options:
1.  Set `dry_run: false` in your `~/.MacosLeanStorage.yaml`.
2.  Use the flag: `mls clean --dry-run=false`

### `serve`
The `serve` command starts a background scheduler that runs the cleanup process according to the `schedule` defined in your configuration file.

```bash
mls serve
```
This is ideal for keeping your Mac lean without manual intervention.

### `config open`
Opens the configuration file in Finder for easy editing.

## Configuration Guide

The configuration file is written in YAML.

### `targets`
A list of directories to monitor.
- `name`: A descriptive name for the target.
- `path`: The absolute path or a home-relative path (using `~/`).
- `threshold_days`: Files older than this many days will be targeted for cleanup.
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
