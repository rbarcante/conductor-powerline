# Code Review Report

**Branch:** `feature/api-usage-integration` vs `origin/develop`
**Generated:** 2026-02-19
**Track:** API Usage Integration

---

## Summary

| Metric | Value |
|--------|-------|
| Files Changed | 29 |
| Lines Added | +1676 |
| Lines Removed | -33 |
| **Code Quality** | 🔴 High: 3 \| 🟡 Medium: 9 \| 🟢 Low: 8 |
| **Security** | 🔴 High: 4 \| 🟡 Medium: 3 \| 🟢 Low: 3 |
| **Test Coverage** | 🔴 High: 3 \| 🟡 Medium: 5 \| 🟢 Low: 3 |

### Package Coverage

| Package | Coverage | Status |
|---------|----------|--------|
| internal/render | 100.0% | ✅ |
| internal/themes | 100.0% | ✅ |
| internal/config | 92.2% | ✅ |
| internal/segments | 92.3% | ✅ |
| internal/hook | 88.9% | ✅ |
| internal/oauth | 77.1% | ⚠️ Below 80% |
| main (root) | 0.0% | 🔴 Subprocess tests |

---

## Code Quality

### High Severity

1. **Cache is ephemeral per-process** (`main.go:46`) — `oauth.Cache` is allocated fresh every invocation and discarded on exit. The stale-fallback path in `FetchUsage` can never trigger across invocations. Either remove the in-memory cache or replace with a persistent backing store.

2. **Goroutine+WaitGroup is a no-op** (`main.go:46`) — `wg.Wait()` is called immediately after launching the single goroutine, making this synchronous with extra overhead. No actual parallelism occurs since segment building happens after `wg.Wait()`.

3. **`getWincredToken` depends on third-party PowerShell module** (`wincred.go:15`) — `Get-StoredCredential` requires the `CredentialManager` module which is not included in Windows by default.

### Medium Severity

1. **Silent timestamp parse errors** (`client.go:74`) — `time.Parse` errors discarded with `_`, leading to zero-time reset values and nonsensical countdowns.
2. **No response body size limit** (`client.go:64`) — `io.ReadAll` with no `io.LimitReader` is an unbounded allocation risk.
3. **NerdFonts merge bug** (`config.go:61`) — Setting only `compactWidth` silently disables NerdFonts because `bool` zero-value is `false` and merge can't distinguish "not set" from "explicitly false".
4. **TrendThreshold config field wired but never used** (`trend.go:14`) — Hardcoded `2.0` in `TrendArrow` ignores `cfg.TrendThreshold`. Function needs a threshold parameter.
5. **`themes.Get` bool return always true** (`themes.go:96`) — Second return is never `false`, making signature misleading.
6. **`tokenGetter` mutable global** (`usage.go:7`) — Package-level var for test injection; should be a function parameter.
7. **Days-remaining truncation** (`weekly.go:35`) — Integer division suppresses day indicator for sub-24-hour periods.
8. **`FetchUsage` return contract unclear** (`usage.go:12`) — Three distinct return states not documented.
9. **Zero-value sentinel ambiguity** (`config.go:90`) — `TrendThreshold != 0` check means user cannot set threshold to 0.

### Low Severity

- `Duration.UnmarshalJSON` manual quote stripping vs proper `json.Unmarshal` into string
- `IsStale` field conflates TTL expiry vs API failure
- Shallow cache copy assumption undocumented
- Missing empty-output tests for wincred and secretool
- Block test assertions too weak (non-empty vs exact format)
- `Duration.MarshalJSON` untested

---

## Security Analysis

### High Severity

1. **PowerShell PATH resolution** (`wincred.go:31`) — `exec.Command("powershell")` resolves via PATH; compromised PATH can redirect to malicious binary. Use absolute path or native Win32 API.

2. **PowerShell module injection** (`wincred.go:15`) — `Get-StoredCredential` is third-party; an attacker who plants a function with that name in a PowerShell profile gets code execution.

3. **Unbounded response body** (`client.go:64`) — `io.ReadAll(resp.Body)` with no size limit. Fix: `io.LimitReader(resp.Body, 1<<20)`.

4. **No HTTPS enforcement** (`client.go:23`) — `baseURL` not validated as HTTPS. Token could be sent in plaintext if user config sets non-HTTPS URL.

### Medium Severity

1. **Credential file permission not checked** (`credfile.go:22`) — `~/.claude/.credentials.json` read without verifying 0600 permissions. Other local users could read the token.
2. **Token value not sanitized before header** (`client.go:52`) — No CRLF check on token before setting Authorization header.
3. **Cache not persistent** (`main.go:53`) — Token read from credential store and sent over network on every prompt. Cross-invocation caching would reduce exposure surface.

### Low Severity

- PowerShell plaintext token in stdout buffer
- 401 vs 5xx not distinguished (revoked token shows stale data)
- Keychain runner variadic args architectural risk for future dynamic args

---

## Test Coverage

### Missing Tests

1. **`main.go` functions** — `run()` and `buildSegments()` at 0.0%. Subprocess integration tests don't instrument source.
2. **`Duration.MarshalJSON()`** at 0.0%. No round-trip serialization test.
3. **`oauth.go` unknown platform branch** — `runtime.GOOS` non-darwin/windows/linux path untested.

### Insufficient Coverage

1. `internal/oauth` at 77.1% (below 80% threshold) — Real command runners intentionally bypassed
2. `client.go` `io.ReadAll` error path uncovered
3. `secretool.go` and `wincred.go` missing empty-output tests
4. `config.go` `LoadFromFile` permission-denied path uncovered
5. `weekly.go` past-reset-time boundary untested
6. `block_test.go` stale indicator assertion too weak

---

## Recommendations

**Priority Actions (address before merging):**
1. Add `io.LimitReader` to `client.go` response body read
2. Validate `baseURL` is HTTPS in `NewClient`
3. Pass `TrendThreshold` config value to `TrendArrow` function
4. Fix or remove the ephemeral cache / sync.WaitGroup no-op in `main.go`
5. Add credential file permission check (0600)
6. Add missing empty-output tests for wincred and secretool

**Suggested Improvements:**
1. Replace PowerShell-based Windows credential retrieval with native API
2. Handle `time.Parse` errors in `client.go` instead of discarding
3. Distinguish 401 from 5xx in API error handling
4. Use `*bool` for NerdFonts to fix merge semantics
5. Add `buildSegments` unit tests
6. Add `Duration.MarshalJSON` round-trip test

---

*Auto-review generated by `/conductor:implement` on track completion*
