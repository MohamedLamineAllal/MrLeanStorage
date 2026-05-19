# Concurrency, Parallelism, and Goroutines in Go

This document explains how the `mls` tool leverages Go's concurrency model to achieve high-performance filesystem scanning.

## 1. The Core Concept: Goroutines vs. OS Threads

### Traditional Multi-threading (OS Threads)
In languages like C++ or Java, "threading" typically refers to **OS Threads**. These are managed directly by the Operating System.
- **Heavyweight**: Each thread requires significant memory (stack size, often 1MB+) and context switching between threads is expensive (CPU registers, state saving/restoring).
- **Limited**: Creating thousands of OS threads will quickly exhaust system resources.

### Go Goroutines
Goroutines are **User-space Threads** managed by the Go runtime, not the OS.
- **Lightweight**: They start with a tiny stack (a few KB) that grows and shrinks as needed.
- **Fast Switching**: Context switching a goroutine is much faster than an OS thread because it involves saving/restoring fewer registers in user-space.

## 2. Are Goroutines "Real Parallelism"?

**Yes.** 

Go uses an **M:N Scheduler** (often called the "Go runtime scheduler"). 
- It maps `M` goroutines onto `N` OS threads.
- If you have 8 CPU cores, the Go runtime will generally attempt to keep 8 OS threads active (the `GOMAXPROCS` setting, which defaults to the number of CPU cores).
- **The "Magic"**: When a goroutine performs a blocking operation (like waiting for I/O from the disk), the Go scheduler moves other ready goroutines to a different, non-blocked OS thread, keeping your CPU cores busy.

So, while goroutines are designed for concurrency (managing many tasks at once, like async I/O), they **achieve true parallelism** by distributing those tasks across all available CPU cores when those tasks are computationally intensive or when they can run in parallel.

## 3. How the Worker Pool Works in `mls`

In the `TargetProcessor`, we use this model to parallelize the scanning of different cleanup targets:

```go
numWorkers := runtime.NumCPU() // Get the number of logical cores
jobs := make(chan config.TargetConfig, len(targets))
// ...
for i := 0; i < numWorkers; i++ {
    go func() { /* Worker goroutine logic */ }()
}
```

1. **Alignment with Cores**: By setting `numWorkers := runtime.NumCPU()`, we ensure that we don't create an arbitrary amount of CPU contention. We match the number of workers to the number of physical/logical cores.
2. **Channel-based Distribution**: We feed `targets` into a `jobs` channel. Multiple goroutines (workers) wait on this channel.
3. **True Parallelism**: Because these targets involve intensive file system I/O (which is often buffered and handled by the kernel) and CPU-bound staleness calculations (mtime checks, dir size recursion), having one worker per core ensures that all cores are working on different branches of the filesystem simultaneously.

## 4. Summary

| Feature | OS Threads | Goroutines |
| :--- | :--- | :--- |
| **Manager** | Operating System | Go Runtime |
| **Memory** | Heavy (MBs) | Lightweight (KBs) |
| **Performance** | Slower context switch | Fast context switch |
| **Parallelism** | True Parallelism | True Parallelism (when managed by runtime) |

By using a **Worker Pool** aligned with `runtime.NumCPU()`, `mls` effectively saturates the available CPU capacity for scanning, turning a traditionally sequential I/O task into a parallelized powerhouse.
