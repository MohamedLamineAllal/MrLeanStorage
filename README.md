# MacosLeanStorage (mls)

A high-performance storage cleanup tool for macOS, designed to safely and efficiently clean up large cache and temporary files.

## Performance
`mls` utilizes a parallel worker pool pattern to scan multiple cleanup targets concurrently, significantly reducing scan times, especially for deep directory trees and numerous cache locations. The scanning engine is optimized for high-throughput I/O and resource-efficient processing, ensuring maximum performance on modern multi-core systems.

## Features
- **Safety First**: Defaults to dry-run mode. No files are deleted unless explicitly requested.
- **Configurable**: Define targets and age thresholds in a simple YAML configuration.
- **macOS Optimized**: Handles `~/` path expansion and targets common macOS cache locations.
- **Detailed Reporting**: Shows exactly what will be deleted and how much space will be freed.

## Installation

```bash
go build -o mls main.go
sudo mv mls /usr/local/bin/
```

## Usage

### 1. Initialize Configuration
The tool automatically creates a default configuration file at `~/.MacosLeanStorage.yaml` on the first run.

### 2. Scan for Old Files
```bash
mls scan
```

### 3. Clean Files (Dry Run)
```bash
mls clean
```

### 4. Clean Files (Actual Deletion)
Edit your config file to set `dry_run: false` or use the flag (if implemented/planned):
```bash
mls clean --dry-run=false
```

### 5. Automated Cleanup
Start the background scheduler to perform cleanup automatically:
```bash
mls serve
```

### 6. Manage Configuration
Open the configuration file in Finder:
```bash
mls config open
```

## Configuration
Example `~/.MacosLeanStorage.yaml`:
```yaml
targets:
  - name: "VSCode Caches"
    path: "~/Library/Caches/com.microsoft.VSCode"
    threshold_days: 7
    safety_level: 1
  - name: "Chrome Caches"
    path: "~/Library/Caches/Google/Chrome/Default/Cache"
    threshold_days: 14
    safety_level: 1
dry_run: true
```

### Configuration Patterns
- The tool supports standard file globbing.
- **Recursive Globbing**: Use the `**` pattern to match directories recursively (e.g., `~/Library/Application Support/MyApp/**/Cache/*`). This is powered by the `doublestar` library.
- **Command-based Cleanup**: Define a `command` field in your target (e.g., `command: "pnpm store prune"`) to run system-level cleanup tasks. `mls` persists the last run time in the system temporary directory and respects the `interval_days` setting to avoid frequent execution.


## Background Automation (macOS)
`mls` can run automatically in the background using `launchd`.

### Manage Background Agent
```bash
# Install the agent
mls agent install

# Start the agent
mls agent start

# Restart the agent
mls agent restart

# Check agent status
mls agent status

# Stop the agent
mls agent stop

# Uninstall the agent
mls agent uninstall
```

## License
MIT
