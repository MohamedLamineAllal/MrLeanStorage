# Analysis: Cleanup Estimation Discrepancy

I have analyzed the `Cleaner.Clean` and `Cleaner.getDirSize` logic. Here is a breakdown of potential areas where the "missing" size could be occurring:

### 1. Error Handling in Clean()
In the worker loop inside `Clean()`:

```go
if err != nil {
    if !os.IsNotExist(err) {
        c.logger.Debug("Failed to stat path", ...)
    }
    continue // <--- Silent skip
}
```

If `os.Stat(path)` fails for any reason (permission issues, race conditions, or temporary filesystem locks), the cleaner skips the file entirely without adding it to `resChan`. Because `resChan` is not updated for skipped files, they are never counted in `freedSpace`, even if they were counted in the initial `Scanner` estimate.

### 2. Error Handling in getDirSize()
Inside `getDirSize()`:

```go
entries, err := os.ReadDir(path)
if err != nil {
    return 0, err
}
// ...
info, err := entry.Info()
if err != nil {
    continue // <--- Silent skip
}
```

If a file or subdirectory cannot be read, it is simply ignored. If the `Scanner` was able to read these files (e.g., due to different timing, privilege elevation, or cache state), but the `Cleaner` encounters a locked or inaccessible file, the `Cleaner` will calculate a size of `0` for that branch.

### 3. Concurrency Channel Buffering
The `resChan` is buffered with `len(paths)`:

```go
resChan := make(chan result, len(paths))
```

This is safe. However, if multiple goroutines are working, and for some reason a goroutine terminates or panics (not caught), that work would be lost. Your current code handles errors and skips, but doesn't log failures to the `resChan`.

### 4. Logic Gap: The "Stale" definition mismatch
The `Scanner` defines "stale" based on `threshold`. 
The `Cleaner` does not care about the `threshold`. It receives a list of paths from the `Scanner` and deletes **everything** in that list. If the `Scanner` provides a folder path because the folder is stale, the `Cleaner` deletes the whole folder. If the folder contents changed between scan and clean, `getDirSize` will be running on a different set of files than the `Scanner` used during its initial estimate.

### Synthesis of the Discrepancy
The most likely culprit is files that exist during the scan but are locked, removed, or inaccessible when the cleaner runs `getDirSize`.

*   **Scanner** sees 40GB.
*   **Cleaner** starts running. It hits a large directory that is currently in use (e.g., Chrome is open). `os.Stat` or `os.ReadDir` on that directory fails (or returns partial results). The Cleaner logs a debug error (which you might not see in standard output) and skips the directory.
*   The **Cleaner** finishes. It reports 24GB. The missing 16GB were never "freed" because the cleaner couldn't safely access or remove them at that moment.
