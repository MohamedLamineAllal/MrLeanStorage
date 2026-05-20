# MrLeanStorage User Guide

## Introduction

MrLeanStorage (`mls`) is a high-performance cleaning tool with a low memory footprint, written in Go, designed to safely and efficiently reclaim disk space. Cleaning is driven entirely by an easy-to-use configuration file. `mls` comes out of the box with a sensible default configuration file that makes it incredibly easy to get started, and it is highly extensible so you can easily add or update targets to fit your own specific needs. Additionally, `mls` goes beyond simple file deletion by providing the ability to execute custom system commands (such as package manager pruning) directly within the cleanup cycle, making it a comprehensive and powerful cleanup solution. You can also explore the configuration examples provided in our repository for advanced setups.

## Getting Started

### Installation

Refer to [Installation Guide](./INSTALL.md)

### Initial Setup

When you first run `mls`, it creates a default configuration file in your home directory: `~/.MrLeanStorage.yaml`.

To see where it is or to reveal it in your system's file manager (Finder / File Explorer), run:

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

1. Set `dry_run: false` in your `~/.MrLeanStorage.yaml`.
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

> [!WARNING]
> **Active Deletion Mode Enabled by Default in Background Automation**
>
> The `serve` command (and background launchd services started via `mls agent install` / `mls agent start`) **will physically delete matched files** (running with `dry_run: false` regardless of global config settings). This ensures background automation performs actual cleanups.
>
> Always verify your target patterns using `mls scan` first before initiating background automation!

Instead of using `mls serve` you can setup a background agent that the system make sure it will be always running in background. Check Background Automation section bellow.

## ⏰ Background Automation (macOS)

Manage the background daemon seamlessly using standard `launchd` controls built right into the CLI:

```bash
# Install the launchd background agent
mls agent install

# Start the background service
mls agent start

# Check background daemon status
mls agent status

# Restart / Hot reload configuration
mls agent restart

# Stop the background service
mls agent stop

# Uninstall the background agent completely
mls agent uninstall
```

### `config open`

Opens the configuration file in the default application (e.g., your preferred text editor). (Cross-platform)

```bash
mls config open
```

### `config reveal`

Reveals the configuration file in your system file explorer (Finder on macOS, File Explorer on Windows, or parent directory on Linux). (Cross-platform)

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
- `interval_days`: Minimum days to wait before running this command again. `mls` tracks the last execution time in a file inside the local application cache directory (e.g., `~/Library/Caches/mls/mls-cmd-<name>.lastrun`).
- `threshold_days`: Items older than this many days will be targeted for cleanup.
- `type`: Defines how the target should be cleaned. Available values:
  - `"file"` (default): Scans inside the path and deletes individual old files.
  - `"folder"`: Treats the matched path itself as the target. If the folder is old enough, the entire folder is deleted.
  - `"both"`: Deletes both old files inside and the folder itself if staleness criteria are met.
- `safety_level`: (Reserved for future use) Intended to define how aggressive the cleanup should be.

### `dry_run`

Global safety switch for manual CLI cleanups (e.g., `mls clean`). If `true`, manual runs will not delete files.

*Note: This switch is ignored by the automated background scheduler (`mls serve` and `mls agent ...`), which always runs in active deletion mode (`dry_run: false`) to ensure background automation is effective.*

### `schedule`

A cron expression (e.g., `"0 0 * * *"`) defining when the automated cleanup should run.

## Tips for Safe Cleanup

1. **Always scan first**: Before running `clean`, run `scan` to see what files are being targeted.
2. **Start with high thresholds**: Set `threshold_days` to a higher value (e.g., 30) and gradually decrease it as you gain confidence.
3. **Use Descriptive Names**: Help yourself by naming targets clearly (e.g., "Arc Browser Cache").

## Configuration Examples & Sharing

We provide both standard and extensive configuration presets to assist you:

- [Default Configuration](file:///Users/mohamedlamineallal/repos/MacosLeanStorage/docs/configuration/Examples/default.yml) — The basic config file created automatically on first startup.
- [Extensive Configuration](file:///Users/mohamedlamineallal/repos/MacosLeanStorage/docs/configuration/Examples/Extensive.yml) — An exhaustive list of targets that you can reference or copy from.

We highly encourage sharing your custom config profiles! If you have built an optimized cleanup configuration, please share it by submitting a [Pull Request (PR)](https://github.com/MohamedLamineAllal/MrLeanStorage/pulls). We accept and merge configuration templates from our community.

## Troubleshooting

If `mls` fails to delete a file, it might be because the file is currently in use by another application. Close the application and try again.
