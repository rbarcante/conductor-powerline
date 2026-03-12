# Specification: Address code review findings in oauth package

## Overview

Refactor the `internal/oauth` package to address findings from the code review of `feature/token-rotation` vs `develop`. The changes improve maintainability, remove dead API surface, eliminate magic numbers, fix incorrect HTTP semantics, and reduce shared-state mutation risks.

## Background

A code review identified 2 high-severity (function length) and 6 medium-severity issues across `client.go`, `usage.go`, `credfile.go`, `keychain.go`, and `filecache.go`. These are correctness-preserving refactors — no behavioral changes.

## Functional Requirements

1. **FR-1: Extract response mapping helper from `FetchUsageData`** — Move the `apiResponse` → `UsageData` mapping (client.go ~lines 141-168) into a `mapAPIResponse(*apiResponse) *UsageData` helper to bring the function under 50 lines.

2. **FR-2: Extract rate-limit backoff helper from `FetchUsage`** — Move the 429 handling block (usage.go ~lines 116-130) into `handleRateLimitBackoff(rle *RateLimitError, cached *UsageData, cache LockableCache) *UsageData` to reduce nesting and function length.

3. **FR-3: Extract named constant for minimum backoff** — Replace `60 * time.Second` magic number with `const minRateLimitBackoff = 60 * time.Second`.

4. **FR-4: Fix Content-Type on GET request** — Replace `Content-Type: application/json` with `Accept: application/json` on the GET request in `FetchUsageData`.

5. **FR-5: Deduplicate credential JSON parsing** — Extract `extractTokenFromCredentialJSON(data []byte) (string, error)` shared between `getKeychainToken` and `getCredfileToken`.

6. **FR-6: Copy cached data before mutation** — In the rate-limit backoff path (usage.go line 67), copy `*cached` before mutating `IsStale` to avoid shared-state side effects.

7. **FR-7: Remove unused `workspaceKey` parameter** — Change `FetchUsage(fetcher, cache, workspaceKey)` to `FetchUsage(fetcher, cache)` and update all call sites.

8. **FR-8: Make `cleanup()` probabilistic** — Run cache cleanup on ~1-in-10 `Store()` calls instead of every call to reduce I/O on the hot path.

## Non-Functional Requirements

- Zero behavioral changes — all existing tests must continue to pass
- Maintain >80% code coverage
- `gofmt -w` on all changed files
- `go build ./...` and `go test ./...` must pass

## Acceptance Criteria

- [ ] No function in `client.go` or `usage.go` exceeds 50 lines
- [ ] No nesting deeper than 3 levels in `usage.go`
- [ ] No magic numbers in rate-limit handling
- [ ] GET request uses `Accept` header, not `Content-Type`
- [ ] Credential JSON parsing exists in exactly one place
- [ ] `cached` pointer is never mutated in-place in `FetchUsage`
- [ ] `FetchUsage` signature has no unused parameters
- [ ] Cache cleanup runs probabilistically, not on every Store
- [ ] All existing tests pass, coverage ≥80%

## Out of Scope

- Test boilerplate reduction (low-severity, can be done opportunistically)
- `main.go` `run()` length (orchestration function, acceptable)
- `atomicWrite` documentation (low-severity)
- `defaultCredfilePath` empty-string edge case (low-severity)
- `anthropic-beta` header constant extraction (low-severity)
