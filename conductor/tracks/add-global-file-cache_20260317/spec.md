# Specification: Add global file cache with locking to avoid redundant API calls

## Overview
Add a global file lock to the existing `~/.cache/conductor-powerline/` cache so that multiple concurrent conductor-powerline processes share cached API responses and avoid redundant API calls. The cache must enforce a 60-second TTL.

## Background
conductor-powerline runs as a statusline tool invoked by Claude Code hooks. Multiple terminal sessions can trigger concurrent invocations, each independently calling the Anthropic usage API. This wastes API quota and adds latency. The existing `FileCache` stores per-workspace usage data but has no locking, so concurrent processes may all hit the API simultaneously.

## Functional Requirements
1. **Cache-first flow**: Check cache freshness before making any API call. If cache is fresh (within TTL), return immediately.
2. **Global lock**: A single lock file (`~/.cache/conductor-powerline/.lock`) guards API calls across all processes and workspaces.
3. **Non-blocking lock acquisition**: `TryLock()` returns immediately — never blocks.
4. **Stale fallback on contention**: If the lock is held by another process and stale cache data exists, return stale data immediately.
5. **Touch on contention**: When a process encounters a lock, it resets the cache entry's `StoredAt` timestamp so it won't retry until the full TTL (60s) elapses.
6. **Stale lock detection**: Lock files older than `APITimeout + 10s` (default 15s) are considered stale and automatically removed.
7. **Default TTL**: Change from 30s to 60s.

## Non-Functional Requirements
- Cross-platform: macOS, Linux, Windows (uses `O_CREATE|O_EXCL` — supported on all three).
- Zero external dependencies (stdlib only).
- Silent failure: lock errors degrade gracefully (log to debug, proceed without lock).
- Lock file is excluded from the 7-day cache cleanup.

## Acceptance Criteria
- [ ] Fresh cache returns immediately without API call
- [ ] Only one process calls the API when multiple run concurrently
- [ ] Locked-out processes return stale data and don't retry until TTL expires
- [ ] Stale lock files (>15s) are automatically cleaned up
- [ ] Default CacheTTL is 60 seconds
- [ ] All tests pass on macOS, Linux, and Windows
- [ ] No external dependencies added

## Out of Scope
- Per-workspace locks (single global lock is sufficient)
- Caching workflow/conductor CLI responses (only usage API)
- Distributed/network-aware locking
