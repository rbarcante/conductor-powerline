# OAuth Token Rotation on 429 Rate Limits — Specification

## Overview
Implement OAuth token rotation to bypass per-token rate limits on the usage API. When the API returns HTTP 429, refresh the OAuth token to obtain a fresh rate-limit window and retry the request, instead of just serving stale cached data.

## Background
The Anthropic usage API (`/api/oauth/usage`) enforces rate limits **per-access-token, not per-account**. A community workaround ([anthropics/claude-code#31021](https://github.com/anthropics/claude-code/issues/31021#issuecomment-4010584731)) discovered that calling the token refresh endpoint returns a new access token with a fresh rate-limit budget.

**Refresh endpoint:**
```
POST https://console.anthropic.com/v1/oauth/token
{"grant_type":"refresh_token","refresh_token":"...","client_id":"9d1c250a-e61b-44d9-88ed-5944d1962f5e"}
```

**Key constraint:** Refresh tokens are one-time use — both the new `access_token` and `refresh_token` must be persisted.

Currently, conductor-powerline's 429 handler (`usage.go:101-110`) only extends cache TTL and returns stale data. Claude Code's credential stores already contain `refreshToken` in their JSON blob, but it is not parsed.

## Functional Requirements

### FR-1: Parse refresh tokens from credential sources
- Expand `claudeAiOAuthEntry` struct in `credfile.go` to include `refreshToken` and `expiresAt`
- Extract `refreshToken` from macOS Keychain JSON blob
- Windows wincred and Linux secret-tool return plain tokens — `refreshToken` will be empty (graceful degradation)

### FR-2: Token refresh client
- New `RefreshOAuthToken(refreshToken string) (*TokenCredentials, error)` function
- POST to `console.anthropic.com/v1/oauth/token` with `grant_type=refresh_token`
- HTTP/1.1, 5s timeout, consistent with existing transport config
- Return typed `RefreshError` for 400/401 (invalid token) vs generic error for network failures

### FR-3: Rotated token storage
- Store refreshed tokens in `~/.cache/conductor-powerline/rotated-token.json` (our file, not Claude Code's credential stores)
- Atomic write pattern (temp + rename), file permissions 0600
- Check rotated token file first in token retrieval, before platform credential stores
- Auto-expire after 7 days (Claude Code may have rotated its own tokens)

### FR-4: Reactive rotation in FetchUsage
- On 429: if `refreshToken` available → refresh → store → retry API → return fresh data
- If refresh fails or no refresh token → fall back to existing cache TTL extension
- If retry after refresh also fails → fall back to cache TTL extension

### FR-5: Concurrency safety
- Dedicated rotation lock file (`rotated-token.json.lock`) prevents concurrent processes from consuming the same one-time refresh token
- If lock is held, skip rotation and fall back to cache extension

## Non-Functional Requirements

- **NFR-1**: Token rotation adds at most one extra HTTP round-trip (~200ms). Total runtime stays under 400ms for rotation path.
- **NFR-2**: No writes to Claude Code's credential stores (Keychain, wincred, secret-tool, credfile). We only read from them.
- **NFR-3**: Cross-platform: rotation works on macOS (Keychain JSON has refreshToken) and credfile users. Gracefully degrades to current behavior on Windows/Linux where refreshToken isn't available.
- **NFR-4**: Silent failure — rotation errors are debug-logged only, never surface to the user.

## Acceptance Criteria

- [ ] On 429, if refresh token is available, token is refreshed and API retried with fresh token
- [ ] Refreshed tokens are persisted to disk and reused on subsequent invocations
- [ ] If refresh fails or no refresh token, behavior is identical to current (cache TTL extension)
- [ ] Concurrent processes don't both attempt to consume the one-time refresh token
- [ ] All existing tests pass; new tests cover rotation happy path, failures, and edge cases
- [ ] `go test ./internal/oauth/... -cover` shows >80% coverage for changed files

## Out of Scope

- Proactive token refresh before expiry (we only refresh on 429)
- Writing back to Claude Code's credential stores
- Multiple token sources / token pools
- Token rotation for non-OAuth authentication methods
