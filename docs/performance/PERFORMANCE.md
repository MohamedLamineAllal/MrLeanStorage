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

## Impacts
- **CLI Efficiency**: Scanning performance will be significantly faster on multi-core systems.
- **Stability**: Resource usage is now bounded, reducing the risk of crashes when scanning massive directory structures.
