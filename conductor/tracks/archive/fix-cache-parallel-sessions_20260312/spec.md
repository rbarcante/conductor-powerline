# Specification: Fix cache parallel sessions - thundering herd

> **Type:** bugfix
> **Track ID:** `fix-cache-parallel-sessions_20260312`

## Overview

Fix `conductor-powerline` cache behaviour so that percentage values no longer "hang" when multiple terminal sessions run the powerline in parallel. The root cause is a **thundering herd**: every session independently sees an expired TTL, concurrently calls the Anthropic usage API, gets rate-limited (429), and falls back to forever-stale cached data. A secondary issue is the absence of any caching for Conductor CLI workflow data, meaning every render spawns a Python subprocess.

## Background

`internal/oauth/filecache.go` implements a file-based cache keyed by workspace SHA-256 hash with a configurable TTL (default 30 s). The `FetchUsage()` function in `usage.go` checks the cache, and if stale, goes straight to the API — with no inter-process coordination. When N sessions exist, up to N concurrent API calls are made at TTL expiry. Additionally, `os.WriteFile` is not atomic; concurrent writes to the same cache file can produce truncated or corrupt JSON.

On the Conductor workflow side, `FetchWorkflowStatus` spawns a `python3 conductor_cli.py` process on **every single powerline render** with no caching at all.

## Requirements

### Functional Requirements

1. Only **one process** at a time may call the Anthropic usage API for a given workspace; all other concurrent processes must wait briefly and then read the result that the "winner" wrote.
2. Cache file writes must be **atomic** (write to a temp file, then `os.Rename`) to prevent corrupt reads.
3. Conductor CLI results must be **cached on disk** with a configurable TTL (default: same as `CacheTTL`) to avoid spawning a Python process on every render.
4. The lock must have a **timeout** (500 ms) so that if the lock holder crashes the system self-heals without blocking the prompt indefinitely.
5. All existing silent-failure / graceful-degradation behaviour must be preserved.

### Non-Functional Requirements

- Pure Go stdlib — no new external dependencies.
- `go test -race` must pass on all platforms (Linux, macOS, Windows).
- `gofmt -w` applied to all changed files before commit.
- Test coverage ≥ 80 % for modified packages.

## Acceptance Criteria

- With 5 parallel powerline sessions active, only **one** API call is made per TTL window (observable via debug logging or network capture).
- After TTL expiry, the percentage updates correctly in all sessions within TTL + 500 ms (lock wait budget).
- The Conductor workflow line does not spawn Python on every render; it reuses cached output within TTL.
- `go test ./...` passes, including with `-race`.

## Out of Scope

- Exponential backoff / retry logic for sustained 429s (tracked separately).
- OAuth token caching (separate concern).
- Windows-specific flock alternatives (lock file via `O_EXCL` is portable).

## Dependencies

- None identified

## References

- None
