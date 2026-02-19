# Plan: API Usage Integration

## Phase 1: OAuth Token Retrieval [checkpoint: 84c27ae]

- [x] Task: Create `internal/oauth/oauth_test.go` — tests for token retrieval orchestration (platform dispatch, fallback chain, all-fail returns error) [ccf22d9]
- [x] Task: Create `internal/oauth/oauth.go` — `GetToken()` function that dispatches to platform-specific retriever based on `runtime.GOOS`, falls through to credential file [ccf22d9]
- [x] Task: Create `internal/oauth/keychain_test.go` — tests for macOS Keychain retrieval (success, not found, command error) using exec mock [ccf22d9]
- [x] Task: Create `internal/oauth/keychain.go` — macOS `security find-generic-password` token retrieval [ccf22d9]
- [x] Task: Create `internal/oauth/wincred_test.go` — tests for Windows Credential Manager retrieval [ccf22d9]
- [x] Task: Create `internal/oauth/wincred.go` — Windows credential retrieval via `wincred` [ccf22d9]
- [x] Task: Create `internal/oauth/secretool_test.go` — tests for Linux secret-tool retrieval [ccf22d9]
- [x] Task: Create `internal/oauth/secretool.go` — Linux `secret-tool lookup` token retrieval [ccf22d9]
- [x] Task: Create `internal/oauth/credfile_test.go` — tests for credential file fallback (valid JSON, missing file, malformed JSON) [ccf22d9]
- [x] Task: Create `internal/oauth/credfile.go` — read token from `~/.claude/.credentials.json` [ccf22d9]
- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: API Client & Caching [checkpoint: b956569]

- [x] Task: Create `internal/oauth/client_test.go` — tests for API client (successful response parsing, timeout, HTTP errors, malformed JSON) [54d301b]
- [x] Task: Create `internal/oauth/client.go` — HTTP client to call Anthropic usage endpoint, parse response into structured usage data [54d301b]
- [x] Task: Create `internal/oauth/cache_test.go` — tests for cache (store/retrieve, TTL expiry, stale indicator, empty cache returns nil) [54d301b]
- [x] Task: Create `internal/oauth/cache.go` — in-memory cache for usage data with configurable TTL and stale tracking [54d301b]
- [x] Task: Create `internal/oauth/usage.go` — `FetchUsage()` function that combines client + cache: try API, cache on success, serve stale on failure [54d301b]
- [x] Task: Create `internal/oauth/usage_test.go` — tests for FetchUsage orchestration (fresh fetch, cache hit, stale fallback, first-run placeholder) [54d301b]
- [x] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Usage Segments & Trends [checkpoint: 6dc9c9b]

- [x] Task: Create `internal/segments/block_test.go` — tests for 5-hour block segment (percentage display, countdown format, color thresholds, nil data shows `--`) [4e1d1da]
- [x] Task: Create `internal/segments/block.go` — 5-hour block usage segment with percentage, countdown, and theme-aware color intensity [4e1d1da]
- [x] Task: Create `internal/segments/weekly_test.go` — tests for 7-day rolling segment (percentage, Opus/Sonnet breakdown, week progress, nil data) [4e1d1da]
- [x] Task: Create `internal/segments/weekly.go` — 7-day rolling usage segment with smart mode breakdown [4e1d1da]
- [x] Task: Create `internal/segments/trend_test.go` — tests for trend indicator (increasing, decreasing, stable within ±2%, no previous data) [4e1d1da]
- [x] Task: Create `internal/segments/trend.go` — trend arrow logic comparing current vs previous usage values [4e1d1da]
- [x] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)

## Phase 4: Integration & Config Update

- [x] Task: Update `internal/config/types.go` — add `APITimeout`, `CacheTTL`, `TrendThreshold` fields to Config; add `block` and `weekly` to SegmentConfig [8917fbf]
- [x] Task: Update `internal/config/config.go` — add defaults for new config fields; update default `segmentOrder` to include `block` and `weekly` [8917fbf]
- [x] Task: Update `internal/config/config_test.go` — tests for new config fields, defaults, and merge behavior [8917fbf]
- [x] Task: Update `main.go` — wire OAuth token retrieval, API usage fetch (parallelized with git via `sync.WaitGroup`), register `block` and `weekly` segment builders [8917fbf]
- [x] Task: Update `main_test.go` — integration tests for full pipeline with usage segments (mock HTTP, mock exec) [8917fbf]
- [x] Task: Conductor - User Manual Verification 'Phase 4' (Protocol in workflow.md)
