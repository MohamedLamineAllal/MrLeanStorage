# Issue: Size Discrepancy between `scan` and `clean`

## Problem Definition
The `mls scan` command and `mls clean` command report different total sizes for the files to be cleaned. In some cases, `scan` reports ~90GB while `clean` reports ~50GB.

## Root Cause Analysis
Two main issues contribute to this discrepancy:

1.  **Overlapping Cleanup Targets & Double Counting**: In `Scanner.Scan`, when a target's type is set to `both`, the scanner checks if the directory itself is stale. If it is, it adds the directory to the results and calculates its total size. However, it *then* continues to walk the directory and adds individual old files to the results as well. This leads to double-counting in the `scan` summary.
2.  **Inaccurate Directory Size Calculation in Cleaner**: The `Cleaner.Clean` method uses `os.Stat(path).Size()` to calculate the freed space. For directories, `os.Stat().Size()` returns the size of the directory entry (typically 64B to 4KB) rather than the cumulative size of its contents. This causes the `clean` summary to significantly under-report the actual space freed when directories are deleted.

## Fix Strategy
1.  **Scanner Refinement**: Modify `Scanner.Scan` to ensure that if a directory is marked for deletion (as a whole), its contents are not processed further. This prevents overlapping paths in the results and avoids double-counting.
2.  **Cleaner Size Calculation**: Update `Cleaner.Clean` to use a recursive size calculation for directories (similar to `Scanner.getDirSize`) to accurately report the space that will be freed.
3.  **Result Deduplication**: Ensure `TargetProcessor` handles paths consistently.

## Implementation Details
- Refactored `Scanner.Scan` to return early for a path if it's already added as a stale folder.
- Updated `Scanner.getDirSize` to respect ignore patterns (like `.DS_Store`), ensuring consistency with the files that will actually be processed.
- Added a `getDirSize` utility to `Cleaner` to correctly calculate directory sizes for the final summary.

## Note on Remaining Minor Differences
A small difference (e.g., ~100MB out of 50GB) may still appear between `scan` and `clean` totals. This is because:
1.  **Ignore Patterns**: `Scanner` respects global ignore patterns (e.g., `.DS_Store`) when calculating directory sizes.
2.  **Cleaner vs. OS**: `Cleaner` reports the size of what it *intends* to delete, while the actual space freed is determined by the OS filesystem (which includes metadata, block sizes, and hidden files that `mls` might ignore).
3.  **Filesystem Latency**: Files might be modified or deleted by the system between the `scan` and `clean` phases.
