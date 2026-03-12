# Implementation Plan: Address code review findings in oauth package

## Phase 1: Extract helpers and constants (no API changes)

- [ ] Task 1.1: Write test for `mapAPIResponse` helper — verify it produces correct `UsageData` from `apiResponse`
- [ ] Task 1.2: Extract `mapAPIResponse(*apiResponse) *UsageData` from `FetchUsageData` in `client.go`
- [ ] Task 1.3: Extract `const minRateLimitBackoff = 60 * time.Second` in `usage.go`
- [ ] Task 1.4: Extract `handleRateLimitBackoff(rle *RateLimitError, cached *UsageData, cache LockableCache) *UsageData` in `usage.go`
- [ ] Task 1.5: Fix cached-data mutation — copy `*cached` before setting `IsStale` in rate-limit backoff path (usage.go line 67)
- [ ] Task 1.6: Fix `Content-Type` → `Accept` header on GET request in `client.go`
- [ ] Task 1.7: Verify — `go build ./...` && `go test ./...`
- [ ] Task 1.8: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Deduplicate credential parsing

- [ ] Task 2.1: Write test for `extractTokenFromCredentialJSON` — JSON with claudeAiOauth, legacy format, raw token prefix, malformed JSON, empty token
- [ ] Task 2.2: Implement `extractTokenFromCredentialJSON(data []byte) (string, error)` in `credfile.go` (shared parsing logic)
- [ ] Task 2.3: Refactor `getCredfileToken` to use `extractTokenFromCredentialJSON`
- [ ] Task 2.4: Refactor `getKeychainToken` to use `extractTokenFromCredentialJSON`
- [ ] Task 2.5: Verify all existing keychain and credfile tests still pass
- [ ] Task 2.6: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: API surface cleanup

- [ ] Task 3.1: Remove `workspaceKey` parameter from `FetchUsage` signature
- [ ] Task 3.2: Update call site in `main.go`
- [ ] Task 3.3: Update all test call sites in `usage_test.go`
- [ ] Task 3.4: Make `cleanup()` probabilistic — add counter, run on ~1-in-10 Store calls
- [ ] Task 3.5: Add test for probabilistic cleanup behavior
- [ ] Task 3.6: Verify — `go build ./...` && `go test ./...` && `gofmt -w` all changed files
- [ ] Task 3.7: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)

## Files to modify

| File | Change |
|------|--------|
| `internal/oauth/client.go` | Extract `mapAPIResponse`, fix `Accept` header |
| `internal/oauth/usage.go` | Extract `handleRateLimitBackoff`, `minRateLimitBackoff` const, copy cached before mutation, remove `workspaceKey` param |
| `internal/oauth/credfile.go` | Add `extractTokenFromCredentialJSON`, refactor `getCredfileToken` |
| `internal/oauth/keychain.go` | Refactor `getKeychainToken` to use shared parser |
| `internal/oauth/filecache.go` | Make `cleanup()` probabilistic |
| `main.go` | Update `FetchUsage` call (remove unused arg) |
| `internal/oauth/client_test.go` | Add `TestMapAPIResponse` |
| `internal/oauth/usage_test.go` | Update `FetchUsage` calls, add backoff helper test |
| `internal/oauth/credfile_test.go` | Add `TestExtractTokenFromCredentialJSON` |
| `internal/oauth/filecache_test.go` | Add probabilistic cleanup test |

## Verification

1. `go build ./...`
2. `go test ./...`
3. `go test -cover ./internal/oauth/` — verify ≥80%
4. `gofmt -l ./internal/oauth/ ./main.go` — no output
5. `echo '{"model":"claude-opus-4-6"}' | CONDUCTOR_DEBUG=1 conductor-powerline` — verify usage still works
