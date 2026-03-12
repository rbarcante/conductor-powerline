package oauth

import (
	"errors"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// lockTimeout is the maximum time a non-lock-holding caller will wait for the
// lock holder to finish before giving up and returning whatever is in cache.
const lockTimeout = 500 * time.Millisecond

// tokenGetter is the function used to get the OAuth token.
// Deprecated: Use credentialsGetter instead for access to refresh tokens.
var tokenGetter = GetToken

// UsageCache defines the interface for caching usage data.
// Both the in-memory Cache and file-based FileCache satisfy this interface.
type UsageCache interface {
	Store(key string, data *UsageData)
	Get(key string) *UsageData
}

// LockableCache extends UsageCache with inter-process advisory locking so that
// only one caller refreshes the cache at a time (thundering herd prevention).
type LockableCache interface {
	UsageCache
	// TryLock atomically creates a lock for key. Returns (true, releaseFn) on
	// success or (false, nil) if another holder already owns the lock.
	TryLock(key string) (bool, func())
	// WaitForUnlock blocks until the lock for key is released or timeout elapses.
	// Returns true if the lock was released before the timeout.
	WaitForUnlock(key string, timeout time.Duration) bool
}

// FetchUsage retrieves usage data using a cache-first, lock-guarded strategy:
//
//  1. Return immediately if cache has fresh (non-stale) data.
//  2. Try to acquire the advisory lock for this workspace key.
//     - If the lock is already held (another session is refreshing):
//     wait up to lockTimeout, then return whatever is in cache (stale ok).
//     - If the lock is acquired:
//     double-check the cache (another process may have just refreshed it),
//     then call the API, store the result, and release the lock.
//
// All error paths preserve the existing silent-failure / graceful-degradation
// behaviour: stale data is preferred over an error wherever possible.
func FetchUsage(fetcher UsageFetcher, cache LockableCache, workspaceKey string) (*UsageData, error) {
	// 1. Cache-first: serve fresh data without touching the network.
	cached := cache.Get(workspaceKey)
	if cached != nil && !cached.IsStale {
		debug.Logf("usage", "cache hit (fresh): block=%.1f%% weekly=%.1f%%", cached.BlockPercentage, cached.WeeklyPercentage)
		return cached, nil
	}

	// 2. Try to become the one process that refreshes the cache.
	acquired, release := cache.TryLock(workspaceKey)
	if !acquired {
		// Another process is already refreshing — wait briefly then serve cache.
		debug.Logf("usage", "lock busy, waiting up to %v for another session to refresh", lockTimeout)
		cache.WaitForUnlock(workspaceKey, lockTimeout)
		result := cache.Get(workspaceKey)
		if result == nil {
			return nil, errors.New("oauth: no cached data available after waiting for lock")
		}
		debug.Logf("usage", "serving data after wait (stale=%v): block=%.1f%%", result.IsStale, result.BlockPercentage)
		return result, nil
	}
	defer release()

	// 3. Double-check: another process may have refreshed between step 1 and 2.
	cached = cache.Get(workspaceKey)
	if cached != nil && !cached.IsStale {
		debug.Logf("usage", "cache fresh after lock acquire (double-check hit): block=%.1f%%", cached.BlockPercentage)
		return cached, nil
	}

	// 4. We hold the lock — fetch from the API.
	debug.Logf("usage", "fetching credentials...")
	creds, err := credentialsGetter()
	if err != nil {
		debug.Logf("usage", "credential retrieval failed: %v", err)
		if cached != nil {
			cached.IsStale = true
			return cached, nil
		}
		return nil, err
	}
	debug.Logf("usage", "credentials retrieved (hasRefresh=%v), calling API...", creds.RefreshToken != "")

	data, err := fetcher.FetchUsageData(creds.AccessToken)
	if err == nil {
		data.IsStale = false
		cache.Store(workspaceKey, data)
		debug.Logf("usage", "API success: block=%.1f%% weekly=%.1f%%", data.BlockPercentage, data.WeeklyPercentage)
		return data, nil
	}
	debug.Logf("usage", "API call failed: %v", err)

	// On 429: attempt token rotation if refresh token is available.
	var rle *RateLimitError
	if errors.As(err, &rle) {
		if rotatedData := tryTokenRotation(fetcher, cache, workspaceKey, creds); rotatedData != nil {
			return rotatedData, nil
		}
		// Rotation failed or unavailable — extend cache TTL as fallback.
		if cached != nil {
			cached.IsStale = true
			cached.FetchedAt = time.Now()
			cache.Store(workspaceKey, cached)
			debug.Logf("usage", "rate limited — extended cache TTL (block=%.1f%%)", cached.BlockPercentage)
			return cached, nil
		}
	}

	// API failed — serve stale or error.
	if cached != nil {
		cached.IsStale = true
		debug.Logf("usage", "serving stale cached data after API failure (block=%.1f%%)", cached.BlockPercentage)
		return cached, nil
	}
	debug.Logf("usage", "no cached data available — returning error")
	return nil, errors.New("oauth: API failed and no cached data available")
}

// tryTokenRotation attempts to refresh the OAuth token and retry the API call.
// Returns fresh UsageData on success, nil on any failure.
func tryTokenRotation(fetcher UsageFetcher, cache LockableCache, workspaceKey string, creds *TokenCredentials) *UsageData {
	if creds.RefreshToken == "" {
		debug.Logf("usage", "no refresh token available — skipping rotation")
		return nil
	}

	if rotatedTokenDir == "" {
		debug.Logf("usage", "rotated token dir not set — skipping rotation")
		return nil
	}

	// Acquire rotation lock to prevent concurrent refresh token consumption.
	acquired, release := TryRotationLock(rotatedTokenDir)
	if !acquired {
		debug.Logf("usage", "rotation lock held — skipping rotation")
		return nil
	}
	defer release()

	debug.Logf("usage", "attempting token rotation...")
	newCreds, err := tokenRefresher(creds.RefreshToken)
	if err != nil {
		debug.Logf("usage", "token refresh failed: %v", err)
		return nil
	}
	debug.Logf("usage", "token refresh succeeded, storing rotated token...")

	// Persist the new tokens to our cache file
	if err := StoreRotatedToken(rotatedTokenDir, newCreds); err != nil {
		debug.Logf("usage", "failed to store rotated token: %v", err)
	}

	// Write back to Claude Code's credential stores so it keeps working
	credentialWriter(newCreds)

	// Retry the API with the new token
	debug.Logf("usage", "retrying API with rotated token...")
	data, err := fetcher.FetchUsageData(newCreds.AccessToken)
	if err != nil {
		debug.Logf("usage", "retry after rotation failed: %v", err)
		return nil
	}

	data.IsStale = false
	cache.Store(workspaceKey, data)
	debug.Logf("usage", "rotation success: block=%.1f%% weekly=%.1f%%", data.BlockPercentage, data.WeeklyPercentage)
	return data
}
