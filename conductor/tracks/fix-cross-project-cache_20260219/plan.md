# Plan: Fix Cross-Project Cache Contamination

## Phase 1: Fix Git Segment Workspace Isolation `[status: complete]` [checkpoint: f6d80b5]

- [x] Task 1.1: Add workspace parameter to `Git()` function signature [29b6503]
  - Update `segments.Git(theme)` → `segments.Git(workspace string, theme themes.Theme)`
  - Update `runGitCommand` to accept an optional `-C <workspace>` prefix
  - TDD: Write tests for git commands with explicit workspace path
- [x] Task 1.2: Wire workspace into `Git()` call in `main.go` [29b6503]
  - Update `buildSegments()` to pass `hookData.WorkspacePath()` to `segments.Git()`
  - TDD: Test that `buildSegments` passes workspace correctly
- [x] Task 1.3: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Implement File-Based Usage Cache `[status: complete]`

- [x] Task 2.1: Create `internal/oauth/filecache.go` with file-based cache [11ac6a6]
  - Implement `FileCache` struct with `Store(key string, data *UsageData)` and `Get(key string) *UsageData`
  - Cache location: `$XDG_CACHE_HOME/conductor-powerline/` (default `~/.cache/conductor-powerline/`)
  - Key: SHA-256 hash of workspace path → filename
  - File format: JSON with `{data: UsageData, stored_at: time, ttl: duration}`
  - Graceful fallback: if dir is unwritable, return nil (no-cache mode)
  - TDD: Red-green tests for Store, Get, TTL expiry, unwritable dir fallback
- [x] Task 2.2: Implement cache auto-cleanup [35e38af]
  - On each `Store()` call, scan cache dir and remove files with mtime > 7 days
  - TDD: Test cleanup removes old files, keeps recent ones
- [x] Task 2.3: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Integration — Replace In-Memory Cache with File Cache `[status: pending]`

- [ ] Task 3.1: Update `oauth.FetchUsage()` to use `FileCache`
  - Replace `Cache` parameter with `FileCache` (or introduce interface)
  - Pass workspace-derived key to cache operations
  - TDD: Test FetchUsage with file-based cache (cache hit, cache miss, stale data)
- [ ] Task 3.2: Update `main.go` to instantiate `FileCache` and pass workspace key
  - Remove in-memory `oauth.NewCache()` call
  - Create `FileCache` once, pass `hookData.WorkspacePath()` as cache key source
  - TDD: Verify main.go wiring
- [ ] Task 3.3: Remove old `internal/oauth/cache.go` and its tests
  - Delete in-memory cache code (now fully replaced)
  - Ensure no imports reference old cache
- [ ] Task 3.4: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
