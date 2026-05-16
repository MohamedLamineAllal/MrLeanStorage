# MacosLeanStorage (mls)

A high-performance storage cleanup tool for macOS, designed to safely and efficiently clean up large cache and temporary files.

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

### 5. Manage Configuration
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

## Testing
Run the test suite:
```bash
go test ./...
```
See [testing_report.md](./testing_report.md) for the latest test results.

## License
MIT
