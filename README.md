# MacosLeanStorage (mls)

`mls` is a high-performance storage cleanup tool for macOS, designed to safely and efficiently clean up large cache and temporary files. Written in Go, it features a small memory footprint, a seamless user experience, and extensive CLI helper commands. It includes a built-in daemon mode that runs daily cleanup tasks automatically.

## Performance

`mls` utilizes a parallel worker pool pattern to scan multiple cleanup targets concurrently, significantly reducing scan times. The scanning engine is optimized for high-throughput I/O and resource-efficient processing, ensuring maximum performance on modern multi-core systems.

## Features

- **Daemon Mode**: Built-in scheduler to perform cleanup tasks automatically.
- **Configurable**: Easily modify your cleanup targets via a simple YAML configuration file.
- **Customizable**: Extensive CLI helpers allow you to manage targets, agent status, and settings easily.
- **Dry-run**: Defaults to dry-run mode to prevent accidental data loss.

## Installation

See the [Installation Guide](./docs/INSTALL.md) for detailed instructions on installing `mls` on macOS, Linux, and Windows, including pre-built binaries and building from source.

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

### 4. Background Automation (macOS)

`mls` can run automatically in the background using `launchd`.

#### Manage Background Agent

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

## Configuration

You can modify or create your configuration file at `~/.MacosLeanStorage.yaml`:

it looks like

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
ignore_patterns:
  - ".DS_Store"
  - "._*"
  - ".Spotlight-V100"
  - ".Trashes"
  - ".fseventsd"
schedule: "0 0 0 * * *"
```

Check the examples on [Configuration Examples](./docs/configuration/Examples/).

- [Extensive Configuration](./docs/configuration/Examples/Extensive.yml)
  - Extensive and growing, we update it from time to time. (It's what the author is using for his own use case)
- [Default configuration](./docs/configuration/Examples/default.yml)
  - The Default configuration if you don't set yours (Updated with time)

## Testing

Run the full test suite:

```bash
go test ./...
```

To run tests with the **Go Race Detector** (recommended for verifying concurrency safety):

```bash
go test -race ./...
```

## Releases

Refer to [docs/RELEASE_PROCESS.md](./docs/RELEASE_PROCESS.md) for information on versioning, release workflows, and binary distribution best practices.

## License

MIT
