# OAuth Token Rotation on 429 Rate Limits — Implementation Plan

## Phase 1: TokenCredentials Type and Credential Parsing
- [x] Task: Define `TokenCredentials` struct in `internal/oauth/types.go`
  - [ ] Add struct with `AccessToken` and `RefreshToken` fields
- [x] Task: Expand `claudeAiOAuthEntry` in `credfile.go` to parse `refreshToken`
  - [ ] Add `RefreshToken string` and `ExpiresAt int64` to struct
  - [ ] Add `getCredfileCredentials() (*TokenCredentials, error)` function
  - [ ] Write tests: with/without refreshToken in JSON
- [x] Task: Extract `refreshToken` from Keychain JSON in `keychain.go`
  - [ ] Add `getKeychainCredentials() (*TokenCredentials, error)` function
  - [ ] Write tests: full JSON blob, raw token string
- [x] Task: Add `GetCredentials()` in `oauth.go`
  - [ ] Try platform-specific credentials getter, fall back to credfile
  - [ ] wincred/secretool: wrap plain token in `TokenCredentials{AccessToken: token}`
  - [ ] Add package-level `credentialsGetter` var for testability
  - [ ] Write tests for dispatch logic
- [x] Task: Update `usage.go` to use `credentialsGetter` instead of `tokenGetter`
  - [ ] Store full `creds` so refresh token is available in 429 handler
  - [ ] Update existing tests that mock `tokenGetter`
- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Refresh Client
- [x] Task: Implement `RefreshOAuthToken()` in `client.go`
  - [ ] POST to `console.anthropic.com/v1/oauth/token` with JSON body
  - [ ] Parse response for new `access_token` + `refresh_token`
  - [ ] HTTP/1.1 transport, 5s timeout
  - [ ] Define `RefreshError` type for 400/401 responses
  - [ ] Add package-level `tokenRefresher` var for testability
- [x] Task: Write comprehensive tests using `httptest`
  - [ ] Success case: valid refresh → new tokens
  - [ ] Invalid refresh token (400) → `RefreshError`
  - [ ] Network error → generic error
  - [ ] Malformed response → error
  - [ ] Verify request body and headers
- [x] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Rotated Token Storage
- [x] Task: Create `internal/oauth/rotatedtoken.go`
  - [ ] Define `rotatedTokenEntry` struct (access_token, refresh_token, rotated_at)
  - [ ] `LoadRotatedToken(cacheDir string) (*TokenCredentials, error)` — read + parse + expiry check (7d)
  - [ ] `StoreRotatedToken(cacheDir string, creds *TokenCredentials) error` — atomic write, 0600 permissions
  - [ ] Dedicated rotation lock: `TryRotationLock(cacheDir) (acquired, release)`
- [x] Task: Write tests in `rotatedtoken_test.go`
  - [ ] Load: no file, valid file, corrupt file, expired file (>7d)
  - [ ] Store: atomic write, correct permissions
  - [ ] Lock: concurrent access, stale lock cleanup
- [x] Task: Integrate with `GetCredentials()` — check rotated token first
  - [ ] Add `SetRotatedTokenDir(dir string)` setter
  - [ ] Modify `GetCredentials()` to call `LoadRotatedToken` before platform stores
- [x] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)

## Phase 4: Wire Token Rotation into FetchUsage
- [x] Task: Modify 429 handler in `usage.go`
  - [ ] On 429 + refresh token available: acquire rotation lock → refresh → store → retry
  - [ ] If refresh fails or lock held: fall back to cache TTL extension
  - [ ] If retry after refresh also 429s: fall back to cache TTL extension
  - [ ] Add debug logging for rotation flow
- [x] Task: Write tests for all rotation paths
  - [ ] 429 → refresh succeeds → retry succeeds → fresh data returned
  - [ ] 429 → refresh succeeds → retry fails → cache TTL extension
  - [ ] 429 → refresh fails → cache TTL extension
  - [ ] 429 → no refresh token → cache TTL extension (backward compat)
  - [ ] 429 → rotation lock held → skip rotation → cache TTL extension
  - [ ] 429 → no cache, no refresh token → error
- [x] Task: Wire `cacheDir` in `main.go`
  - [ ] Call `oauth.SetRotatedTokenDir(cacheDir())` before `oauth.FetchUsage()`
- [x] Task: Conductor - User Manual Verification 'Phase 4' (Protocol in workflow.md)

## Phase 5: Edge Cases, Cleanup, and Coverage
- [x] Task: Handle rotated token file in cache cleanup
  - [ ] Include in existing `cleanup()` or add 7-day age check
- [x] Task: Windows path handling
  - [ ] Verify all new file paths use `filepath.Join`
  - [ ] Test with `USERPROFILE` env var
- [x] Task: Run full test suite and verify coverage
  - [ ] `go test ./... -cover` passes
  - [ ] Changed files >80% coverage
  - [ ] `gofmt -w` on all changed files
- [x] Task: Conductor - User Manual Verification 'Phase 5' (Protocol in workflow.md)
