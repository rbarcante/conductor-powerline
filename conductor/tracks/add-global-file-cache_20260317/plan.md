# Implementation Plan: Add global file cache with locking to avoid redundant API calls

## Phase 1: CacheLock type
- [ ] Task 1.1: Create `internal/oauth/cachelock.go`
  - [ ] `CacheLock` struct with `path string` and `staleAge time.Duration`
  - [ ] `NewCacheLock(dir, staleAge)` constructor (path = `filepath.Join(dir, ".lock")`)
  - [ ] `TryLock() bool` — atomic create via `O_CREATE|O_EXCL`, write timestamp, close handle. On `ErrExist`: check stale age, remove + retry once if stale.
  - [ ] `Unlock()` — `os.Remove`, ignore errors
- [ ] Task 1.2: Create `internal/oauth/cachelock_test.go`
  - [ ] `TestCacheLockAcquireAndRelease`
  - [ ] `TestCacheLockContention` (second TryLock returns false)
  - [ ] `TestCacheLockStaleRemoval`
  - [ ] `TestCacheLockUnlockIdempotent`
  - [ ] `TestCacheLockUnwritableDir`
- [ ] Task 1.3: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: FileCache Touch + cleanup guard
- [ ] Task 2.1: Add `Touch(key string)` to `UsageCache` interface in `usage.go`
- [ ] Task 2.2: Implement `Touch` on `FileCache` in `filecache.go` — re-read entry, update `StoredAt` to now, write back
- [ ] Task 2.3: Update `cleanup()` to skip `.lock` file
- [ ] Task 2.4: Add `TestFileCacheTouch` to `filecache_test.go`
- [ ] Task 2.5: Update `mockCache` in `usage_test.go` to implement `Touch`
- [ ] Task 2.6: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: FetchUsage flow change
- [ ] Task 3.1: Change `FetchUsage` signature to accept `lock *CacheLock`
  - New flow: cache check → lock acquire → token + API → store → unlock
  - On lock contention: return stale + Touch
  - On lock=nil: behave as before (for tests)
- [ ] Task 3.2: Update all existing `FetchUsage` calls in `usage_test.go` to pass `nil` lock
- [ ] Task 3.3: Add new tests:
  - [ ] `TestFetchUsageCacheFresh` — fresh cache skips API
  - [ ] `TestFetchUsageLockContention` — stale data returned, Touch called
  - [ ] `TestFetchUsageLockAcquired` — API called, lock released
- [ ] Task 3.4: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)

## Phase 4: Config + wiring
- [ ] Task 4.1: Change default `CacheTTL` from 30s to 60s in `config.go`
- [ ] Task 4.2: Wire `CacheLock` in `main.go` — create lock, pass to `FetchUsage`
- [ ] Task 4.3: Run `go test ./...` and `gofmt -w` on all changed files
- [ ] Task 4.4: Integration smoke test
- [ ] Task 4.5: Conductor - User Manual Verification 'Phase 4' (Protocol in workflow.md)

## Key Files

| File | Change |
|------|--------|
| `internal/oauth/cachelock.go` | **New** — CacheLock type |
| `internal/oauth/cachelock_test.go` | **New** — lock tests |
| `internal/oauth/filecache.go` | Add `Touch()`, update `cleanup()` |
| `internal/oauth/filecache_test.go` | Add Touch test |
| `internal/oauth/usage.go` | Add `Touch` to interface, change `FetchUsage` flow + signature |
| `internal/oauth/usage_test.go` | Update mock, add lock-aware tests |
| `internal/config/config.go` | Default CacheTTL 30s → 60s |
| `main.go` | Create CacheLock, pass to FetchUsage |

## Verification
1. `gofmt -w` on all changed `.go` files
2. `go test ./...` — all tests pass
3. Integration smoke test: run two concurrent invocations, verify only one API call occurs (check debug logs with `CONDUCTOR_DEBUG=1`)
4. Verify lock file is created and cleaned up under `~/.cache/conductor-powerline/.lock`
