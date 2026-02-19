# Spec: Fix Cross-Project Cache Contamination

## Overview

The statusline displays stale data from other projects/repositories because:
1. **Git segment** (`internal/segments/git.go:44-50`) runs `git` commands against the process's CWD instead of the workspace path provided via stdin hook data.
2. **Directory segment** (`internal/segments/directory.go:28-33`) falls back to `os.Getwd()` when workspace is empty — correct behavior, but the git segment has no equivalent workspace awareness.
3. **In-memory cache** (`internal/oauth/cache.go`) is created and destroyed on every invocation, making the TTL dead code. Usage data should persist to disk across invocations, keyed per project.

## Functional Requirements

1. **FR-1: Git segment uses workspace path** — `segments.Git()` must accept a workspace path parameter and run git commands with `git -C <workspace>` so it reports the correct branch/dirty state for the Claude Code project, not the shell's CWD.
2. **FR-2: File-based usage cache** — Replace the in-memory `oauth.Cache` with a file-based cache stored in `$XDG_CACHE_HOME/conductor-powerline/` (defaulting to `~/.cache/conductor-powerline/`). Cache files are keyed by a hash of the workspace path.
3. **FR-3: Only usage data is cached** — The cache must store only `UsageData` (block %, weekly %, reset times, trend data). Context window, directory, git branch, and model data must never be persisted.
4. **FR-4: Cache auto-cleanup** — On each run, remove cache files not accessed in 7+ days to prevent unbounded disk growth.

## Non-Functional Requirements

- **NFR-1:** No new external dependencies (pure stdlib).
- **NFR-2:** Cache read/write must not measurably impact startup (<5ms overhead).
- **NFR-3:** Graceful degradation: if cache dir is unwritable, fall back to no caching (API-only).
- **NFR-4:** >80% test coverage on new/modified code.

## Acceptance Criteria

- [ ] Running conductor-powerline from directory A while Claude Code is in directory B shows B's git branch, not A's.
- [ ] Usage data survives across invocations (second run within TTL serves cached data without API call).
- [ ] Cache files are created under `~/.cache/conductor-powerline/` (or `$XDG_CACHE_HOME`).
- [ ] Stale cache files (>7 days) are cleaned up automatically.
- [ ] Context window, directory, model, and git data are never written to disk.

## Out of Scope

- Migration of existing in-memory cache consumers (the current cache is effectively unused)
- Multi-user cache sharing
- Cache encryption
