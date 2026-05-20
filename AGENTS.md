# MrLeanStorage (mls) — Agent Guidelines & Instructions

Welcome to the `mls` development agent guide. This codebase is highly optimized for performance, safety, and background execution. To ensure stability and avoid common development regressions, you **MUST** read and adhere to all instructions in this document without exception.

---

## 1. Agentic Workflow (Non-Negotiable Mode of Operation)

Every prompt interaction must strictly follow the **Research -> Strategy -> Execution -> Validation** lifecycle. You must take your time, work step-by-step, and never rush modifications.

### Step 1: Context Absorption & Resumability
- **At the start of every interaction**, read `MEMORY.md` to identify the current phase, progress, and pending tasks.
- Run `git log -n 5 --oneline` to synchronize with recent commits.
- Check the git status to avoid modifying staged files.

### Step 2: Strict Logging Protocols
- **Prompt Logging:** Log every significant user prompt into [Prompts.log](file:///Users/mohamedlamineallal/repos/MacosLeanStorage/Prompts.log) chronologically.
- **Action Logging:** For every modification, analysis, or test run, log a detailed action record in [ACTIONS.log](file:///Users/mohamedlamineallal/repos/MacosLeanStorage/ACTIONS.log). Include specific context so developers can understand the history at a glance.
- **Memory Updates:** Keep [MEMORY.md](file:///Users/mohamedlamineallal/repos/MacosLeanStorage/MEMORY.md) in sync as progress is made, detailing completed/pending milestones and design decisions.

### Step 3: Strategic Planning & Analysis-First
- **Do not modify files blindly.** If the user asks for analysis, **only read files and write the analysis**—never modify source files or make structural changes under an analysis request.
- **Zero-Regression Safeguard:** Never delete, disable, or alter existing CLI commands, utilities, or configuration options unless explicitly requested. Always review git diffs carefully.
- **Staging Protection:** Respect the user's workspace state. Do not run commands that reset or destroy staged changes or wipe out working directory progress.
- **Documentation Updates & Preservation:** When updating documentation, never delete or remove useful information or context. Just add to, refine, or enhance the language. If you believe some information has become completely obsolete and should be removed, do not delete it silently. Instead, explicitly prompt the Developer to review the proposed deletion and provide them with clear, detailed reasoning for why that specific detail should be removed.

### Step 4: Verification & Testing
- **Test Before Complete:** You must build the binaries (`go build -o mls main.go`) and run the test suite (`go test -race ./...`) after every change.
- **Race Detector:** Always use the `-race` flag when running tests. Concurrency is critical to `mls`, and all race conditions must be detected and resolved immediately.
- **Incremental Commits:** Commit your changes in clean, descriptive chunks matching conventional commit guidelines (e.g. `feat: ...`, `fix: ...`). Run tests before committing.

---

## 2. Technical Stack & Conventions

`mls` is a high-performance cleanup daemon and CLI tool written in Go.

- **Stack:** Go 1.26+, Cobra (CLI Framework), Viper (Configuration Manager), Zap (Structured Logging), and Testify (Testing Assertions).
- **Concurrency:** Uses goroutine worker pools in both the scanner (`internal/scanner/scanner.go`) and cleaner (`internal/cleaner/cleaner.go`) to maximize filesystem I/O performance on macOS.
- **Traversals:** Prefers high-performance, single-pass `os.ReadDir` traversals over slow recursive `filepath.Walk` operations.
- **Safety First:** All cleaning operations default to **dry-run mode**. Deletion requires an explicit `--force` or `--confirm` flag to prevent accidental data loss.
- **Hot Reloading:** The server handles runtime configuration reloading upon receiving `SIGHUP` signals.
- **Storage Persistence:** No `/tmp` storage. All transient directories, databases, and logs are persisted inside local directories resolved using `utils.GetAppCacheDir()` (resolving to `~/Library/Caches/mls`).
- **Sleep-Wake Resiliency:** Uses a background missed-task recovery ticker running every 30 minutes in `cmd/serve.go` to catch up on cron executions missed during macOS sleep cycles.
- **Cron Scheduling:** The scheduler expects standard 6-field cron expressions (granularity includes seconds: `Second Minute Hour DayOfMonth Month DayOfWeek`).

---

## 3. Go Code Style & Quality

- **GoDoc Compliance:** Every public and private function, type, method, and package-level variable **MUST** be fully documented with clear, GoDoc-compliant comments explaining both the "what" (behavior) and the "why" (intent).
- **Formatting:** Code must be formatted using standard `gofmt` or `gofumpt` (run `gofumpt -l -w .`).
- **Error Handling:** Always handle errors gracefully. Wrap errors with detailed structural context using `fmt.Errorf("context: %w", err)`.
- **Conventions:** Follow the Uber Go Style Guide as a structural baseline. Keep names concise and camelCase.

---

## 4. CI/CD & GoReleaser Guidelines

When modifying `.goreleaser.yaml` or `.github/workflows/release.yml`, follow these strict guidelines to avoid pipeline failures:
- **Search First:** If you are unsure of the GoReleaser syntax or deprecations, search the web or consult official documentation first.
- **Deprecations:** `brews` block is deprecated in newer GoReleaser versions (v2+). Use the `homebrew_casks` block to configure Homebrew Tap distributions (`homebrew-mls`), even for standard CLI tools.
- **Quarantine Hook:** macOS binaries distributed through custom taps must include a post-install hook in the cask block to strip the quarantine attribute:
  ```ruby
  hooks:
    post:
      install: |
        if OS.mac?
          system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/mls"]
        end
  ```
- **Permissions:** Releasing requires write access to repository contents. Ensure the workflow provides `contents: write` permissions, and use the `HOMEBREW_TAP_GITHUB_TOKEN` secret for Tap repository write access to avoid 403 API errors.

---

## 5. Development Shortcuts & Commands

Make sure to run these commands frequently during development:
```bash
# Build the binary locally
go build -o mls main.go

# Run the test suite with race detection enabled
go test -race ./...

# Format the Go source code
gofumpt -l -w .

# Run static linter
golangci-lint run
```
