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

- [ ] Task: Write failing tests in `usage_test.go`
  - [ ] `TestFetchUsage_OnlyOneAPICallOnConcurrentExpiry`: mock cache returning stale data; mock API callable once; run 5 goroutines concurrently calling `FetchUsage`; assert API was called exactly once
  - [ ] `TestFetchUsage_WaiterGetsRefreshedData`: lock is pre-held; API writes fresh cache after 100 ms; assert waiting goroutine returns updated value without making its own API call
  - [ ] Run tests — confirm they fail
- [ ] Task: Modify `FetchUsage()` in `usage.go` to use lock-before-fetch pattern
  - [ ] Add `UsageLockCache` interface extension (or add `TryLock`/`WaitForUnlock` methods to the `UsageCache` interface) — evaluate cleanest approach; keep interface minimal
  - [ ] Flow: `Get()` → if fresh return; try `TryLock()` → if not acquired `WaitForUnlock(500ms)` then `Get()` and return (stale ok); if acquired do double-check `Get()` (another process may have just written it) → call API → `Store()` → release lock
  - [ ] Run tests — confirm they pass and race detector is clean
- [ ] Task: Conductor - User Manual Verification 'Phase 2 - Thundering Herd Fix in FetchUsage' (Protocol in workflow.md)

## Phase 3 — Conductor CLI Workflow Caching

- [ ] Task: Write failing tests in new `workflow_cache_test.go`
  - [ ] `TestWorkflowFileCache_StoreAndGet`
  - [ ] `TestWorkflowFileCache_TTLExpiry` (sets IsStale on read after TTL)
  - [ ] `TestWorkflowFileCache_ConcurrentStore_NoCorruption`
  - [ ] Run tests — confirm they fail (file doesn't exist yet)
- [ ] Task: Create `internal/segments/workflow_cache.go`
  - [ ] `WorkflowFileCache` struct mirroring `FileCache` (dir, ttl) but typed for `WorkflowData`
  - [ ] `NewWorkflowFileCache(dir string, ttl time.Duration) *WorkflowFileCache`
  - [ ] `Store(key string, data *WorkflowData)` — atomic write + lock (same pattern as Phase 1)
  - [ ] `Get(key string) *WorkflowData` — returns nil or data with `IsStale` flag
  - [ ] `keyPath(key)` — SHA-256 hash, extension `.workflow.json`
  - [ ] Run tests — confirm they pass
- [ ] Task: Integrate `WorkflowFileCache` into `main.go`
  - [ ] Create `WorkflowFileCache` alongside `FileCache` using `cacheDir()`
  - [ ] Wrap `FetchWorkflowStatus` call: check cache first; only call CLI if cache miss or stale; store result
  - [ ] Add integration test or manual verification via debug log
- [ ] Task: Conductor - User Manual Verification 'Phase 3 - Conductor CLI Workflow Caching' (Protocol in workflow.md)

## Phase 4 — Coverage and Cleanup

- [ ] Task: Run `go test -race -coverprofile=coverage.out ./...` and check coverage ≥ 80 % for `oauth` and `segments` packages
- [ ] Task: Run `gofmt -w` on all modified `.go` files
- [ ] Task: Run full test suite on macOS and confirm clean
- [ ] Task: Conductor - User Manual Verification 'Phase 4 - Coverage and Cleanup' (Protocol in workflow.md)

---

## Notes

<!-- Implementation notes, decisions made during development -->
