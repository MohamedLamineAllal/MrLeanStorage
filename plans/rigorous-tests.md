# Rigorous Test Plan for Engine and CommandHandler

## 1. Engine Testing Coverage
- [x] **New()**: Verify initialization of `scanner` and `cleaner`.
- **Scan()**:
    - [x] Sequential vs Parallel target scans (validate hook timing).
    - [x] Hook order: `OnTargetScanStart` -> `OnTargetScanEnd` for each target.
    - [x] Error handling: What happens if `scanner.Scan` fails for one target? (Should it continue or abort?)
    - [x] Empty target list handling.
    - [x] Partial/Full results verification.
- **Clean()**:
    - [x] Verify concurrency safety (run with `-race`).
    - [x] Hook order: `OnFileCleaned` and `OnTargetCleaned` callback sequence.
    - [x] Handle skipped targets (when `len(res.Files) == 0`).
    - [x] Verify `ResultAggregator` aggregates unique paths correctly across parallel threads.
- **ProcessCommands()**:
    - [x] Verify command dispatch order.
    - [x] Hook order: `BeforeHandleCommand`, `BeforeExecutingCommand`, `AfterExecutingCommand`, `AfterHandleCommand`.
    - [x] Verify dry-run logic (ensure commands are NOT executed if `DryRun == true`).
- **ScanAndClean()**:
    - [x] End-to-end flow integration test.

## 2. CommandHandler Testing Coverage
- **Handle()**:
    - [x] Verify `scheduler` interaction (`ShouldRunCommand` check).
    - [x] Verify `UpdateCommandRunTime` behavior.
    - [x] Verify hook sequence and error propagation.
- **ExecuteCommand()**:
    - [x] Dry-run mode verification (returns `nil` error).
    - [x] Real execution simulation (mocking `exec.Command` if possible or using `/bin/true`/`/bin/false`).
    - [x] Error handling for failed command execution.

## 3. Implementation Status
- [x] Step 1: Create a mock interface for `Cleaner` and `Scanner` to isolate engine logic.
- [x] Step 2: Implement Test Cases in phases (Phase A: Hook sequence, Phase B: Concurrency/Race tests, Phase C: Error/Edge cases).
- [x] Step 3: Run all with `go test -race`.
