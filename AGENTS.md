# MacosLeanStorage (mls)

Make sure to read all the bellow and all instruction bellow. Make sure to adhere to them. Check Agentic workflow and Memory logic. Code style ...

## Instructions

Bellow is the mode of operation that you need to follow. Make sure you don't miss any step.

### Agentic Workflow

- **Resumability**: At the start of every interaction, the agent MUST read `MEMORY.md` to understand the current phase, progress, and context. As well as git commits.
- **Prompt Logging**: The agent MUST log every significant user prompt into `Prompts.log` to maintain a history of directives and intent.
- **Log done actions**: For every action you do, log it down to `ACTIONS.log`. That would help me and any other dev to see what was done by the AI. As a full history. Not to be read and used by the Agent. Unless there an issue (commits, prompts, MEMORY not done, logged synced forgotten), and that would help.
- **Contextual Awareness**: The agent should always reference `MEMORY.md` before proposing new actions to ensure continuity.
- **Incremental Updates**: Every significant step or decision must be recorded in `MEMORY.md`.
- **Response Documentation**: When creating a substantial response, analysis, guide, recommendation, or decision-support note, write it as a Markdown file under `docs` using a clear well named subfolder structure and organization. Keep response files organized by topic, use descriptive filenames, and reference the created file in the final chat response.
- **Git Integration**: All changes and actions should be committed with descriptive messages following conventional commits.
- **Documentation**: Maintain and keep Project Documentation up to date. Document all functionalities, features, .... As well as all decisions, architecture choices, analysis and research.

### Memory Logic

- `MEMORY.md` serves as the source of truth for the project's state.
- It includes:
    - Current Phase.
    - Summary of completed tasks.
    - Pending tasks for the current phase.
    - Brainstormed items, analysis results, and decisions (as they are generated).

### Coding
- Always make best analysis and decision making, pick the best choices, best practices, adhere to high quality of code, architecture, design and choices.
- Take as many steps and time to do things the right way and at the highest level.

### Make sure
- Make sure you don't miss anything from the mode of operation above and bellow.
- Make sure to update docs/configuration/examples/default.yml when you update the default configuration, it should always be kept up to date.

## Project Description
High-performance, safe, and efficient storage cleanup tool for macOS.

We want to build a CLI tool and Daemon using Go Lang (for better memory usage). That clean safe to remove files and directories that consume storage. That are taken by many applications. Notably Arc, google chrome, and many browsers, electron applications, Discord, vscode, cursor, antigravity, OpenAIAtlas ...

## Technical Stack & Architecture

- **Main Technologies**: Go (Golang), Cobra (CLI), Viper (Config), Zap (Logging).
- **Architecture**:
    - `cmd/`: CLI entry points and command definitions.
    - `internal/scanner/`: Concurrent directory traversal and file analysis.
    - `internal/cleaner/`: Logic for safe file deletion (with dry-run support).
    - `internal/config/`: Configuration management using YAML and environment variables.
    - `internal/scheduler/`: Logic for periodic background cleanup tasks.

### Development Conventions

- **Concurrency**: Use Go routines and worker pools for filesystem I/O to maximize performance without overwhelming OS limits.
- **Safety First**: All cleanup operations must default to **dry-run mode**. Deletion requires an explicit `--force` or `--confirm` flag.
- **I/O Efficiency**: Prefer `os.ReadDir` over `filepath.Walk` for directory traversal.
- **Logging**: Use structured logging with `zap` for performance and searchability.
- **Multi-Profile Handling**: Logic should explicitly account for multi-profile applications (e.g., Arc/VSCode) by iterating through `User Data` subdirectories.
- **Safety Thresholds**:
    - **Safe Anytime**: Rebuildable data like `CachedData` or `DerivedData`.
    - **Safe after 3+ Days**: System temp files and general logs.
    - **Safe after 7+ Days**: Browser `CacheStorage` and VSCode `workspaceStorage`.

### Execution and Context
- Make sure you tackle deeply and efficiently all of the request of the prompt, take your time, do things in multiple steps, as much as it's required. Stop when needed and resume till you finish everything.

## Building and Running

### Prerequisites
- Go 1.26.2 or later.
- Homebrew (for system dependencies).

### Build & Dev Commands
```bash
go build -o mls main.go
./mls --help
go test ./...
golangci-lint run
gofumpt -l -w .
```
## Tech Stack
- **Language:** Go 1.26+
- **CLI Framework:** Cobra
- **Configuration:** Viper
- **Logging:** Uber-zap
- **Testing:** Standard library `testing` with `testify` for assertions.

## Code Style
- Follow standard Go conventions (Uber Go Style Guide as a reference).
- Use `gofmt` for formatting.
- **Naming:**
    - Use CamelCase for public symbols.
    - Keep names concise but descriptive.
- **Error Handling:**
    - Always handle errors.
    - Wrap errors with context where helpful using `fmt.Errorf("context: %w", err)`.
- **Concurrency:**
    - Use goroutines and channels for parallel scanning.
    - Ensure thread safety in the `Cleaner` and `Scanner`.
- **Logging:**
    - Use `zap.Logger` for structured logging.
    - Log levels: `Info` for general progress, `Debug` for detailed info, `Error` for failures.
- **Documentation:**
    - All public functions and types must have comments.
    - Maintain `ARCHITECTURE.md` for high-level design.

## Development Workflow
- **Research -> Strategy -> Execution** lifecycle.
- **Plan -> Act -> Validate** for each sub-task.
- Incremental commits: One feature/fix per commit.
- Run tests before committing.
