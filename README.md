# MrLeanStorage (mls)

`mls` (MrLeanStorage) is a high-performance cleaning tool with a low memory footprint, written in Go, designed to safely and efficiently reclaim disk space. Cleaning is driven entirely by an easy-to-use configuration file. `mls` comes out of the box with a sensible default configuration file that makes it incredibly easy to get started, and it is highly extensible so you can easily add or update targets to fit your own specific needs. Additionally, `mls` goes beyond simple file deletion by providing the ability to execute custom system commands (such as package manager pruning) directly within the cleanup cycle, making it a comprehensive and powerful cleanup solution. You can also explore the configuration examples provided in this repository for advanced setups.

---

## ⚡ Key Features

- **🚀 Concurrency Engine**: Uses high-performance goroutine worker pools in both the scanner (`os.ReadDir` based) and cleaner to scan and delete files simultaneously, maximizing macOS SSD throughput.
- **🔄 Zero-Downtime Hot Reloads**: Supports instant configuration reloading via `SIGHUP` signal without terminating the running background daemon.
- **🛌 Missed-Task Recovery Ticker**: Background daemon runs an automatic missed-task recovery ticker every 30 minutes to catch up on cleanup runs that were missed while your Mac was asleep.
- **📦 Application Cache Migration**: State files (e.g., last-run logs) are safely persisted under the persistent local cache (`~/Library/Caches/mls`) rather than the volatile `/tmp` directory.
- **🛡️ Dry-Run Safety**: Defaults to a strict dry-run mode so you can preview exactly which files will be deleted before taking any destructive action.
- **🤖 launchd Background Integration**: Complete agent management CLI to install, start, stop, restart, and inspect background services seamlessly on macOS.

---

## 📖 User Guide

For detailed explanations of all features, custom scheduling options, command execution targets, and tips for safe cleanup, consult our comprehensive [User Guide](./docs/USER_GUIDE.md).

---

## 📦 Installation

For full multi-platform instructions, see the detailed [Installation Guide](./docs/INSTALL.md).

### macOS (Homebrew Cask) — Recommended

`mls` is distributed as a Homebrew Cask via a custom Tap for seamless macOS installation and automatic quarantine bypass:

```bash
# Add our custom Tap
brew tap MohamedLamineAllal/mls

# Install mls
brew install mls
```

*Note: Homebrew will automatically map this to our Cask distribution. If you encounter any checksum errors from outdated Formula caches, resolve them by running `brew update && brew tap --repair` first.*

### Linux & Windows

For Debian/Ubuntu (`.deb`), RedHat/Fedora (`.rpm`), or Windows manual installations, please refer to the [Installation Guide](./docs/INSTALL.md#2-linux-pre-built-packages-manual-binary).

---

## 🚀 Usage

### 1. Initialize & Open Configuration

On first run, `mls` automatically creates a default configuration file at `~/.MrLeanStorage.yaml`.

```bash
# Open configuration in your default editor
mls config open

# Reveal configuration location in your system's file manager (Finder / File Explorer)
mls config reveal
```

### 2. Scan for Old Files

Analyze the configured targets and view matched files and sizes:

```bash
mls scan
```

Use the verbose flag to output all matches beyond the default summary:

```bash
mls scan -v
```

### 3. Clean Files (Dry Run & Confirmation)

Dry run to preview deletions:

```bash
mls clean
```

Execute the real cleanup by disabling dry-run:

```bash
mls clean --dry-run=false
```

### 4. Running Cleaning on a schedule

As per configuration, default is running daily

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

---

## ⏰ Background Automation

Manage the background daemon seamlessly using standard system controls (launchd on macOS, systemd on Linux, and Scheduled Tasks on Windows) built right into the CLI:

```bash
# Install the background agent
mls agent install

# Start the background service
mls agent start

# Check background daemon status
mls agent status

# View background agent logs
mls agent log          # Show last 20 lines
mls agent log --live   # Stream logs in real-time
mls agent log --path   # Show log file path

# Restart / Hot reload configuration
mls agent restart

# Stop the background service
mls agent stop

# Uninstall the background agent completely
mls agent uninstall
```

> [!WARNING]
> **Active Deletion Mode Enabled by Default in Background Automation**
>
> Starting `mls serve` or installing/starting the background agent (`mls agent install` / `mls agent start`) **will physically delete matched files** (running with `dry_run: false` regardless of global config settings). This ensures background automation performs actual cleanups.
>
> Always verify your target patterns using `mls scan` first before initiating background automation!

---

## 🖥️ Multi-platform Support

`mls` is designed to be cross-platform and works on macOS, Linux, and Windows.

- `mls scan`: Scans targets for files and directories to clean based on your configuration. (Cross-platform)
- `mls clean`: Scan and deletes files and directories identified during the scan. (Cross-platform)
- `mls agent`: Manages the background cleanup service. (Cross-platform: launchd on macOS, systemd on Linux, Scheduled Tasks on Windows)
- `mls serve`: Starts the background scheduler loop to perform automated cleanup. You can use it with CLI on any platform, you can set it up as a daemon, or start when the system starts. (Cross-platform)
- `mls config open`: Opens the configuration file in your default system editor. (Cross-platform)
- `mls config reveal`: Reveals the configuration file location in your system's file explorer (Finder, File Explorer, or parent folder on Linux). (Cross-platform)
- `mls config reload`: Signals all running `mls serve` daemons to reload their configuration. (Cross-platform)

---

## 🛠️ Configuration Example

The `~/.MrLeanStorage.yaml` configuration uses simple and flexible YAML format:

```yaml
# Global safety switch for manual CLI cleanups (e.g., mls clean). If true, manual runs will not delete files.
# Note: This is ignored by mls serve / mls agent background daemons, which always run in active deletion mode (dry_run: false).
dry_run: true

# Patterns to globally ignore during scanning and deep staleness checks
ignore_patterns:
  - ".DS_Store"
  - "._*"
  - ".Spotlight-V100"
  - ".Trashes"
  - ".fseventsd"

# Cron scheduling expression (supports 6-field standard with seconds)
# Format: Second Minute Hour DayOfMonth Month DayOfWeek
schedule: "0 0 0 * * *"

# Target directories to monitor and system commands to execute
targets:
  - name: "VSCode Caches"
    path: "~/Library/Caches/com.microsoft.VSCode"
    threshold_days: 7
    type: "file" # "file", "folder", or "both"
    safety_level: 1

  - name: "Chrome Caches"
    path: "~/Library/Caches/Google/Chrome/Default/Cache"
    threshold_days: 14
    type: "file"
    safety_level: 1

  - name: "PNPM Global Pruning"
    command: "pnpm store prune"
    interval_days: 7 # Run this command target once every 7 days
```

---

## 📂 Configuration Examples & Community Sharing

We provide default and advanced configuration templates to assist you:
- [Extensive Configuration](./docs/configuration/Examples/Extensive.yml)
  - Extensive and growing, we update it from time to time. (It's what the author is using for his own use case)
- [Default configuration](./docs/configuration/Examples/default.yml)
  - The Default configuration if you don't set yours (Updated with time)

We highly encourage and accept community-submitted configuration templates! If you have constructed a custom setup that you'd like to share with other users, please feel free to submit a [Pull Request (PR)](https://github.com/MohamedLamineAllal/MrLeanStorage/pulls). We are always happy to review, merge, and feature new configuration examples into this repository.

---

## 🧪 Testing

Run the test suite with race condition detection:

```bash
go test -race ./...
```

---

## 📄 License

MIT License. See [LICENSE](LICENSE) for details.
