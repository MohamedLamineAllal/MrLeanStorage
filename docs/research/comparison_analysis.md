# Comparative Analysis: MrLeanStorage (mls)

To understand how **MrLeanStorage (mls)** fits into the landscape of disk cleanup tools, it is helpful to categorize the available alternatives. `mls` is unique because it is a configurable, daemonized, and extensible system utility rather than a "one-click" GUI cleaner.

## 1. Similar Tools

### GUI-Based (User-Friendly, Interactive)
* **DaisyDisk / GrandPerspective / OmniDiskSweeper:** These visualize disk space as maps or trees. They are best for finding *what* is taking up space (large files, forgotten folders) rather than *automating* the cleanup of ephemeral caches.
* **CleanMyMac X / CCleaner:** These are "all-in-one" proprietary suites. They feature one-click scanning, malware removal, and uninstallation tools.
    * **Comparison:** Unlike `mls`, these are black boxes. They handle the "what to delete" logic internally, whereas `mls` puts that control in a user-maintained YAML file.

### Command-Line & Scripting (Automated, Power-User)
* **`rm`, `find` + `xargs`:** The classic Unix approach.
    * **Comparison:** You can replicate `mls` functionality with a complex shell script or a `crontab` job. However, `mls` provides a structured, multi-platform safety layer (dry-runs, configuration file validation, and logging) that raw scripts lack.
* **BleachBit:** An open-source cleaner with both CLI and GUI interfaces. It uses defined "cleaners" to target specific application caches and system junk.
    * **Comparison:** BleachBit is more "opinionated" and pre-packaged with rules for specific applications. `mls` is more of a "general-purpose engine" where you define the paths and staleness thresholds yourself.

### Advanced Infrastructure (System Administrators)
* **`tmpwatch` / `tmpreaper` (Linux):** These are classic Linux utilities that delete files in directories that have not been accessed for a certain period.
    * **Comparison:** `mls` is essentially the modern, cross-platform, and more feature-rich evolution of these tools. It adds background daemonization (launchd/systemd support), command execution (not just deletion), and more granular configuration (YAML).

## 2. How They Compare to `mls`

| Feature | `mls` | GUI Cleaners (e.g., CleanMyMac) | Simple Scripts (`find`/`cron`) |
| :--- | :--- | :--- | :--- |
| **Control** | High (YAML config) | Low (Black box) | Total |
| **Automation** | Native Daemon (Background) | Limited/Proprietary | Manual (`cron`) |
| **Safety** | High (Dry-run, Logging) | Moderate | Low (Easy to break) |
| **Performance** | High (Go + Concurrency) | Moderate | High (Native tools) |
| **Extensibility**| High (Execute arbitrary commands) | Limited | High |

## 3. Which are "Better"?

"Better" depends entirely on your persona:

* **You want "Set and Forget" convenience:** GUI tools like **CleanMyMac X** are "better." They are designed to be intuitive for non-technical users and require zero configuration.
* **You want absolute transparency and auditability:** **`mls`** is likely better. It allows you to see *exactly* what files will be deleted before they are, and it doesn't run proprietary code on your system.
* **You are a sysadmin managing hundreds of machines:** **`tmpwatch` or custom Ansible-driven scripts** might be preferred, as `mls` requires an installed binary and configuration file on each node.
* **You prioritize performance and automation:** `mls` shines because it leverages Go's concurrency, supports hot-reloading configurations without daemon restarts, and allows for both file cleaning and arbitrary system command execution (like `pnpm prune`).

**Summary:** If you are a developer or power user who likes to define *exactly* how your system is cleaned—and wants that cleanup to happen reliably in the background without needing a heavy GUI application—`mls` is likely the superior choice.
