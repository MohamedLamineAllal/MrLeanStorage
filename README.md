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

### From Source (Build & Install)

Inside the repo run:

```sh
go build -o mls main.go && \
sudo mv mls /usr/local/bin/mls
```

To build and install `mls` from the source repository:

```bash
# Clone the repository and install
git clone git@github.com:MohamedLamineAllal/MacOSLeanStorage.git /tmp/mls-build && \
cd /tmp/mls-build && \
go build -o mls main.go && \
sudo mv mls /usr/local/bin/mls && \
cd /tmp && rm -rf mls-build
```

### From GitHub Releases (Pre-built)
You can download pre-built binaries for your platform directly from the [GitHub Releases page](https://github.com/MohamedLamineAllal/MacOSLeanStorage/releases).

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
See [Configuration Examples](./docs/configuration/Examples/) for templates.

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
