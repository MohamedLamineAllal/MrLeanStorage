# Performance Architecture

## Overview
To improve the performance of the scanning operation, especially when dealing with large numbers of targets or deep directory trees, I have implemented a worker pool pattern for parallel scanning.

## Design Choices
- **Worker Pool Pattern**: Instead of scanning targets sequentially, I have introduced a dispatcher/worker pattern using Go routines and channels.
- **Why Worker Pool**:
    - **Resource Management**: Limits the number of concurrent I/O operations, preventing the tool from overwhelming the OS or hitting file descriptor limits.
    - **Concurrency**: Leverages Go's lightweight routines for efficient parallel processing of multiple target paths.
    - **Backpressure**: Prevents memory spikes by controlling the ingestion of jobs.

## Implementation Details
1. **Job Queue**: A channel `jobs := make(chan scanner.Target, len(targets))` handles target distribution.
2. **Result Aggregation**: A buffered channel `results := make(chan scanner.Result, len(targets))` aggregates results from workers.
3. **Workers**: A pool of workers (configurable, currently set to a reasonable number based on CPU cores) consumes the jobs and performs the scanning.
4. **Synchronization**: `sync.WaitGroup` is used to ensure all workers finish before closing result channels and summarizing findings.

## Directory Traversal & Globbing Strategy

### How We Walk the Tree
The `mls` tool uses a combination of pattern-based globbing and recursive depth-first traversal to identify stale files.

1.  **Globbing Phase**:
    - We use `github.com/bmatcuk/doublestar/v4` to parse target paths.
    - This allows for powerful patterns like `**/Cache` or `.../Service Worker/**`.
    - Globbing handles the high-level path expansion. Once the glob matches the initial directory structure, the scanner takes over.

2.  **Recursive Traversal (`walkFiles`)**:
    - For each matched path, `mls` recursively crawls the subdirectories.
    - **Optimization - Fast Exit**: During the scan, if a folder is marked as "stale" (based on its mtime or its contents' staleness), `mls` treats it as a single deleteable unit. This prevents the scanner from wasting I/O by walking thousands of files inside a directory that will be deleted anyway.
    - **Ignore Pattern Enforcement**: At every step of the walk, the scanner checks the entry name against `ignorePatterns` (e.g., `.DS_Store`, `.git`, `.fseventsd`). If an entry is ignored, the scanner skips it and any of its children entirely.

### Why This Strategy Is Optimal
- **I/O Efficiency**: By using `os.ReadDir`, we fetch directory entries (names and types) in a single system call rather than performing `Lstat` on every single file, significantly reducing kernel overhead.
- **Safety**:
    - **Permission Handling**: The `walkFiles` implementation explicitly catches `os.IsPermission(err)` and treats it as a partial scan rather than a failure, ensuring we clean what we can even if some subfolders are protected.
    - **Staleness Logic**: The recursive `checkStaleness` function ensures that a directory is only marked "stale" if *all* its contents are stale, preventing the premature deletion of active, recent files.
- **Accuracy**: By separating the globbing phase (high-level path discovery) from the recursive traversal (deep content staleness analysis), we guarantee that all files covered by a glob are visited, while the staleness logic keeps the actual cleanup safe.
