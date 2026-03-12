# Code Review: Fix cache parallel sessions - thundering herd

**Track:** Fix cache parallel sessions thundering herd
**Branch:** fix/cache-parallel-sessions
**Base:** develop
**Files changed:** 8 Go files (+651 lines / -18 lines)

---

## Summary

| Category | Critical | High | Medium | Low |
|----------|----------|------|--------|-----|
| Code Quality | 0 | 4 | 5 | 3 |
| Security | 0 | 0 | 0 | 6 |
| Test Coverage | 0 | 2 | 5 | 3 |
| **Total** | **0** | **6** | **10** | **12** |

No blocking issues. ✅ No critical or security high/medium findings.

---

## Code Quality

### High

**`atomicWrite` duplicated across two packages**
`internal/oauth/filecache.go:66` and `internal/segments/workflow_cache.go:91` contain 25 lines of identical logic. Future changes to atomic write semantics must be applied in two places.
→ Extract to a shared `internal/atomicio` or `internal/filecache` package.

**Structural duplication: `WorkflowFileCache` mirrors `FileCache`**
`workflowCacheEntry`, `Store`, `Get`, `keyPath`, `atomicWrite` replicate the entire `FileCache` pattern with only the data type changed.
→ A generic `FileCache[T any]` (Go 1.21+ generics) would collapse both into one implementation.

**`tokenGetter` mutable package-level variable**
`internal/oauth/usage.go:16` — concurrent tests that swap `tokenGetter` can race each other when run with `go test -race -parallel`.
→ Pass the token getter as a parameter or in an options struct to make dependencies explicit.

**`TryLock` returns `(false, nil)` on `MkdirAll` failure**
`internal/oauth/filecache.go:155` — callers cannot distinguish "lock already held" from "I/O failure". The waiter then blocks on `WaitForUnlock` which will never resolve, stalling the prompt for `lockTimeout` (500 ms).
→ Return a third `error` value: `TryLock(key string) (bool, func(), error)`.

### Medium

**`WaitForUnlock` return value discarded**
`internal/oauth/usage.go:62` — timeout is not distinguished from successful unlock; falls through silently to `cache.Get` either way.
→ Log or handle the timeout case: `if !cache.WaitForUnlock(...) { debug.Logf(...) }`.

**`atomicWrite` `Chmod` after `Close`**
`filecache.go:82` / `workflow_cache.go:107` — `os.CreateTemp` on Go 1.16+ already creates files at mode `0o600`, so the `Chmod` is redundant. If kept for defense-in-depth, call `tmp.Chmod(0o600)` before `tmp.Close()` to avoid the brief permission window.

**`WaitForUnlock` sleeps before first check**
`filecache.go:178` — the loop sleeps 50 ms unconditionally before checking if the lock is gone. Restructure to stat-first, sleep-second to eliminate one poll interval on the fast path.

**Usage goroutine re-derives `workspace`**
`main.go:78` — the usage goroutine calls `hookData.WorkspacePath()` again instead of capturing the outer `workspace` variable (which has the `os.Getwd()` fallback). When `WorkspacePath()` returns empty, the goroutine passes `""` to `FetchUsage` as the cache key.
→ Remove the local re-declaration and close over the outer `workspace`.

**On-disk `ttl` field is dead data**
`workflowCacheEntry.TTL` and `fileCacheEntry.TTL` are written to disk but never parsed back on `Get`. Document as human-inspection-only or remove.

### Low

- `cleanup()` runs on every `Store` (full `ReadDir` scan each shell render). Gate with probabilistic check or a last-run sentinel.
- `FetchUsage` doc comment doesn't document when stale data can be returned with `nil` error.
- `cacheDir()` three-tier fallback logic is non-obvious — add a brief comment.

---

## Security

No critical, high, or medium findings. All 6 findings are low-severity / informational.

**Most actionable:**
`atomicWrite` calls `os.Chmod` after `tmp.Close()` — a narrow window where the temp file exists at the wrong permissions. On Go 1.16+ `os.CreateTemp` defaults to `0600`, so the window is theoretical. Either remove `os.Chmod` (relying on Go's guarantee) or move it to before `Close`.

**Reliability (not security):**
Lock files are never expired. If the lock-holding process is `SIGKILL`-ed mid-API-call, the `.lock` file is never removed. All subsequent renders for that workspace spin for 500 ms before self-healing. Consider treating lock files older than ~5 s as orphaned and removing them.

---

## Test Coverage

| Package | Coverage | Target |
|---------|----------|--------|
| `internal/oauth` | 78.2% | 80% |
| `internal/segments` | 90.1% | 80% ✅ |

**Gap in `internal/oauth` (1.8% below target):** entirely from pre-existing 0% platform-specific functions (`runKeychainCommand`, `runSecretoolCommand`, `runWincredCommand`, `defaultCredfilePath`) not modified by this track. All new/modified code is ≥ 87%.

**Primary gaps in new code:**
- `atomicWrite` error paths (Write/Close/Chmod/Rename failures) — 44.4% in both packages; requires OS-level mocking or filesystem test helpers.
- `json.Marshal`/`Unmarshal` failure paths in `Store`/`Get`.
- `FindConductorCLI` at 63.6% (pre-existing, not modified by this track).

---

## Recommendations

1. **Fix `workspace` capture in usage goroutine** (main.go:78) — correctness bug, low risk but should be addressed before merge.
2. **Log `WaitForUnlock` timeout** — makes the 500 ms stall visible in debug output.
3. **Dedup `atomicWrite`** — not urgent but reduces future drift risk.
4. **Orphaned lock cleanup** — add mtime check in `WaitForUnlock` to self-heal crashed holders faster.
