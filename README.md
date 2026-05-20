# MrLeanStorage (mls)

`mls` (MrLeanStorage) is a high-performance cleaning tool with a low memory footprint, written in Go, designed to safely and efficiently reclaim disk space. Out of the box, `mls` comes with a default configuration that makes it incredibly easy to get started, and is highly extensible so you can easily adapt and extend it to fit your custom cleanup needs. You can also explore the configuration examples provided in this repository to customize your rules.

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

# Reveal configuration location in Finder
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

---

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

---

## 🖥️ Multi-platform Support

`mls` is designed to be cross-platform and should work on macOS, Linux, and Windows. However, we are currently focusing our development efforts primarily on macOS. With time, we plan to improve and expand support for other platforms.

Please note that the background agent management commands (`mls agent ...`) are currently supported **only on macOS**. We will update this section as support for background services on other platforms is implemented.

The rest of the commands should work on all platforms:

- `mls scan`: Scans targets for files and directories to clean based on your configuration.
- `mls clean`: Scan and deletes files and directories identified during the scan.
- `mls serve`: Starts the background scheduler loop to perform automated cleanup. You can use it with CLI on any platform, you can set it up as a daemon, or start when the system start.
- `mls config open`: Opens the configuration file in your default system editor.
  - (Works only on MacOS, we will update this for cross platform)
- `mls config reveal`: Reveals the configuration file location in your file explorer.
  - (Works only on MacOS, we will update this for cross platform)
- `mls config reload`: Signals the running `mls serve` daemon to reload its configuration.
  - (Works only on Macos, we will update this for cross platform)
  - Stop `mls serve` and start it again to reload.

If you don't want to wait for the Daemon support on other platforms you can setup yours, with `mls serve`. Ask `gemini` or `gpt` for how to set up a daemon on linux or windows for `mls serve` command. `mls serve` will handle the rest for you.

---

## 🛠️ Configuration Example

The `~/.MrLeanStorage.yaml` configuration uses simple and flexible YAML format:

```yaml
# Global safety switch. If true, no files are ever deleted.
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
