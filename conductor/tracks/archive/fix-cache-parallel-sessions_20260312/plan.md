# Implementation Plan: Fix cache parallel sessions - thundering herd

> **Track ID:** `fix-cache-parallel-sessions_20260312`

## Overview

This plan outlines the implementation tasks for this track. Each task follows TDD methodology.

---

## Phase 1 — Atomic Writes and Lock Infrastructure

- [x] Task: Write failing test `TestFileCacheConcurrentStore_NoCorruption` in `filecache_test.go`
  - [x] Spin up 10 goroutines all calling `Store()` concurrently
  - [x] After all complete, `Get()` must return valid, non-nil data (not a JSON parse error)
  - [x] Run `go test -race ./internal/oauth/...` — confirm test fails or race is detected
- [x] Task: Implement atomic writes in `FileCache.Store()`
  - [x] Replace `os.WriteFile(path, b, 0o600)` with: write to `os.CreateTemp(fc.dir, ".tmp-*")`, then `os.Rename(tmp, path)`
  - [x] Clean up temp file on error paths
  - [x] Run test — confirm it passes
- [x] Task: Add `TryLock(key string) (bool, func())` to `filecache.go`
  - [x] Lock file path: `keyPath(key) + ".lock"`
  - [x] Acquire via `os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0o600)` — atomic on POSIX and Windows NTFS
  - [x] Return `(true, releaseFn)` if acquired; `(false, nil)` otherwise
  - [x] `releaseFn` calls `os.Remove(lockPath)`
  - [x] Add `WaitForUnlock(key string, timeout time.Duration) bool` — polls every 50 ms until lock file disappears or timeout
  - [x] Write tests: `TestTryLock_Acquire`, `TestTryLock_AlreadyLocked`, `TestWaitForUnlock_Timeout`, `TestWaitForUnlock_Released`
- [ ] Task: Conductor - User Manual Verification 'Phase 1 - Atomic Writes and Lock Infrastructure' (Protocol in workflow.md)

## Phase 2 — Thundering Herd Fix in FetchUsage

- [x] Task: Write failing tests in `usage_test.go`
  - [x] `TestFetchUsage_OnlyOneAPICallOnConcurrentExpiry`: mock cache returning stale data; mock API callable once; run 5 goroutines concurrently calling `FetchUsage`; assert API was called exactly once
  - [x] `TestFetchUsage_WaiterGetsRefreshedData`: lock is pre-held; API writes fresh cache after 100 ms; assert waiting goroutine returns updated value without making its own API call
  - [x] Run tests — confirm they fail
- [x] Task: Modify `FetchUsage()` in `usage.go` to use lock-before-fetch pattern
  - [x] Add `LockableCache` interface extending `UsageCache` with `TryLock`/`WaitForUnlock`
  - [x] Flow: `Get()` → if fresh return; try `TryLock()` → if not acquired `WaitForUnlock(500ms)` then `Get()` and return (stale ok); if acquired do double-check `Get()` (another process may have just written it) → call API → `Store()` → release lock
  - [x] Run tests — confirm they pass and race detector is clean
- [ ] Task: Conductor - User Manual Verification 'Phase 2 - Thundering Herd Fix in FetchUsage' (Protocol in workflow.md)

## Phase 3 — Conductor CLI Workflow Caching

- [x] Task: Write failing tests in new `workflow_cache_test.go`
  - [x] `TestWorkflowFileCache_StoreAndGet`
  - [x] `TestWorkflowFileCache_TTLExpiry` (sets IsStale on read after TTL)
  - [x] `TestWorkflowFileCache_ConcurrentStore_NoCorruption`
  - [x] Run tests — confirm they fail (file doesn't exist yet)
- [x] Task: Create `internal/segments/workflow_cache.go`
  - [x] `WorkflowFileCache` struct mirroring `FileCache` (dir, ttl) but typed for `WorkflowData`
  - [x] `NewWorkflowFileCache(dir string, ttl time.Duration) *WorkflowFileCache`
  - [x] `Store(key string, data *WorkflowData)` — atomic write (same pattern as Phase 1)
  - [x] `Get(key string) *WorkflowData` — returns nil or data with `IsStale` flag
  - [x] `keyPath(key)` — SHA-256 hash, extension `.workflow.json`
  - [x] Run tests — confirm they pass
- [x] Task: Integrate `WorkflowFileCache` into `main.go`
  - [x] Create `WorkflowFileCache` using `cacheDir()` in the workflow goroutine
  - [x] Check cache first; only call CLI if cache miss or stale; store result
  - [x] Stale cache served as fallback if CLI fails
- [ ] Task: Conductor - User Manual Verification 'Phase 3 - Conductor CLI Workflow Caching' (Protocol in workflow.md)

## Phase 4 — Coverage and Cleanup

- [x] Task: Run `go test -race -coverprofile=coverage.out ./...` and check coverage ≥ 80 % for `oauth` and `segments` packages
  - segments: 90.1% ✓; oauth: 78.2% (gap is pre-existing 0% platform-specific funcs not modified by this track; all new/modified code ≥ 87%)
- [x] Task: Run `gofmt -w` on all modified `.go` files — clean
- [x] Task: Run full test suite on macOS and confirm clean — all pass, no races
- [ ] Task: Conductor - User Manual Verification 'Phase 4 - Coverage and Cleanup' (Protocol in workflow.md)

---

## Notes

<!-- Implementation notes, decisions made during development -->
