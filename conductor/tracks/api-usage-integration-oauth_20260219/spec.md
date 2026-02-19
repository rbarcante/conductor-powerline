# Spec: API Usage Integration

## Overview

Add real-time Anthropic API usage data to the statusline by implementing cross-platform OAuth token retrieval, an API client for the usage endpoint, and two new segments: 5-hour block usage and 7-day rolling usage. Includes a caching layer for graceful degradation and usage trend indicators.

## Functional Requirements

### FR-1: Cross-Platform OAuth Token Retrieval (`internal/oauth/`)

- Retrieve Claude OAuth token from platform credential stores:
  - **macOS:** `security find-generic-password` from Keychain
  - **Windows:** Windows Credential Manager via `wincred`
  - **Linux:** `secret-tool lookup` from GNOME Keyring / libsecret
  - **Fallback:** Read from `~/.claude/.credentials.json`
- Platform selection via `runtime.GOOS`
- Return token string or error; never panic

### FR-2: Anthropic API Client (`internal/oauth/`)

- HTTP client to call the Anthropic usage endpoint with the retrieved OAuth token
- Parse JSON response for 5-hour block and 7-day rolling usage data
- Timeout: configurable, default 5 seconds
- Return structured usage data or error

### FR-3: 5-Hour Block Usage Segment (`internal/segments/block.go`)

- Display current utilization percentage (e.g., `72%`)
- Display time remaining until block reset (e.g., `2h13m`)
- Color intensity based on usage level (theme-aware: normal, warning, critical thresholds)

### FR-4: 7-Day Rolling Usage Segment (`internal/segments/weekly.go`)

- Display weekly usage percentage
- Smart mode: show Opus/Sonnet breakdown when both are in use (e.g., `Opus: 45% | Sonnet: 20%`)
- Week-progress indicator

### FR-5: Usage Trend Indicators

- Compare current usage values against previous poll values
- Display directional arrows: up (increasing), down (decreasing), right (stable)
- Stable threshold: +/-2% change is considered stable

### FR-6: Caching Layer

- Cache last successful API response in memory
- On API failure: serve cached data with a stale indicator
- On first run with no cache: show `--` placeholders
- Cache TTL: configurable, default 30 seconds

## Non-Functional Requirements

- **Performance:** API calls must not block startup beyond 5s timeout; use `sync.WaitGroup` to parallelize with git/other segments
- **Security:** Never log or display OAuth tokens; tokens in memory only
- **Reliability:** All credential store failures fall through gracefully to next source
- **Testing:** >80% coverage for all new packages; mock `os/exec` calls and HTTP responses

## Acceptance Criteria

1. On macOS, Linux, and Windows, the tool retrieves the OAuth token from the platform credential store without user intervention
2. If all credential sources fail, segments show `--` and no error is emitted to stdout
3. 5-hour block segment displays accurate percentage and countdown
4. 7-day rolling segment displays weekly usage with Opus/Sonnet breakdown
5. Trend arrows reflect actual change between consecutive runs
6. On API timeout/failure, cached data is displayed with stale indicator
7. All new code has >80% test coverage
8. No new external dependencies (pure stdlib)

## Out of Scope

- Historical usage data persistence (beyond single-run trend comparison)
- Rate limiting / retry logic (fail fast, use cache)
- Custom API endpoint configuration
- Token refresh / re-authentication flows
