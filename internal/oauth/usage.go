package oauth

import (
	"errors"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// lockTimeout is the maximum time a non-lock-holding caller will wait for the
// lock holder to finish before giving up and returning whatever is in cache.
const lockTimeout = 500 * time.Millisecond

// minRateLimitBackoff is the minimum backoff duration after a 429 response.
const minRateLimitBackoff = 60 * time.Second

// globalCacheKey is the single cache key for usage data. Usage is account-level,
// not per-workspace, so all sessions share one cache entry.
const globalCacheKey = "__global_usage__"

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

// tokenGetter retrieves the OAuth access token. Package-level variable for testability.
var tokenGetter = GetToken

// FetchUsage retrieves usage data using a cache-first, lock-guarded strategy:
//
//  1. Return immediately if cache has fresh (non-stale) data.
//  2. Try to acquire the advisory lock for the global cache key.
//     - If the lock is already held (another session is refreshing):
//     wait up to lockTimeout, then return whatever is in cache (stale ok).
//     - If the lock is acquired:
//     double-check the cache (another process may have just refreshed it),
//     then call the API, store the result, and release the lock.
//
// All error paths preserve the existing silent-failure / graceful-degradation
// behaviour: stale data is preferred over an error wherever possible.
func FetchUsage(fetcher UsageFetcher, cache LockableCache) (*UsageData, error) {
	// 1. Cache-first: serve fresh data without touching the network.
	cached := cache.Get(globalCacheKey)
	if cached != nil && !cached.IsStale {
		debug.Logf("usage", "cache hit (fresh): block=%.1f%% weekly=%.1f%%", cached.BlockPercentage, cached.WeeklyPercentage)
		return cached, nil
	}

	// 1b. Respect rate-limit backoff: if we were recently 429'd, serve stale data
	// instead of hammering the API again before the server's Retry-After window.
	if cached != nil && !cached.RateLimitedUntil.IsZero() && time.Now().Before(cached.RateLimitedUntil) {
		debug.Logf("usage", "rate-limit backoff active (until %v), serving stale data", cached.RateLimitedUntil.Format("15:04:05"))
		result := *cached
		result.IsStale = true
		return &result, nil
	}

	// 2. Try to become the one process that refreshes the cache.
	acquired, release := cache.TryLock(globalCacheKey)
	if !acquired {
		// Another process is already refreshing — wait briefly then serve cache.
		debug.Logf("usage", "lock busy, waiting up to %v for another session to refresh", lockTimeout)
		cache.WaitForUnlock(globalCacheKey, lockTimeout)
		result := cache.Get(globalCacheKey)
		if result == nil {
			return nil, errors.New("oauth: no cached data available after waiting for lock")
		}
		debug.Logf("usage", "serving data after wait (stale=%v): block=%.1f%%", result.IsStale, result.BlockPercentage)
		return result, nil
	}
	defer release()

	return fetchUnderLock(fetcher, cache)
}

// fetchUnderLock performs the API fetch while holding the advisory lock.
// It double-checks the cache, retrieves a token, calls the API, and handles
// rate-limit backoff and stale fallback on any error path.
func fetchUnderLock(fetcher UsageFetcher, cache LockableCache) (*UsageData, error) {
	// Double-check: another process may have refreshed between step 1 and lock acquire.
	cached := cache.Get(globalCacheKey)
	if cached != nil && !cached.IsStale {
		debug.Logf("usage", "cache fresh after lock acquire (double-check hit): block=%.1f%%", cached.BlockPercentage)
		return cached, nil
	}

	// Fetch from the API.
	debug.Logf("usage", "fetching token...")
	token, err := tokenGetter()
	if err != nil {
		debug.Logf("usage", "token retrieval failed: %v", err)
		if cached != nil {
			result := *cached
			result.IsStale = true
			return &result, nil
		}
		return nil, err
	}
	debug.Logf("usage", "token retrieved, calling API...")

	data, err := fetcher.FetchUsageData(token)
	if err == nil {
		data.IsStale = false
		cache.Store(globalCacheKey, data)
		debug.Logf("usage", "API success: block=%.1f%% weekly=%.1f%%", data.BlockPercentage, data.WeeklyPercentage)
		return data, nil
	}
	debug.Logf("usage", "API call failed: %v", err)

	// On 429: back off for RetryAfter duration.
	var rle *RateLimitError
	if errors.As(err, &rle) {
		if result := handleRateLimitBackoff(rle, cached, cache); result != nil {
			return result, nil
		}
	}

	// API failed — serve stale or error.
	if cached != nil {
		result := *cached
		result.IsStale = true
		debug.Logf("usage", "serving stale cached data after API failure (block=%.1f%%)", result.BlockPercentage)
		return &result, nil
	}
	debug.Logf("usage", "no cached data available — returning error")
	return nil, errors.New("oauth: API failed and no cached data available")
}

// handleRateLimitBackoff processes a 429 rate-limit response by copying the
// cached data (to avoid mutating the shared pointer), setting backoff fields,
// and re-storing it. Returns nil if there is no cached data to fall back on.
func handleRateLimitBackoff(rle *RateLimitError, cached *UsageData, cache LockableCache) *UsageData {
	if cached == nil {
		return nil
	}
	backoff := rle.RetryAfter
	if backoff < minRateLimitBackoff {
		backoff = minRateLimitBackoff
	}
	now := time.Now()
	result := *cached
	result.IsStale = true
	result.FetchedAt = now
	result.RateLimitedUntil = now.Add(backoff)
	cache.Store(globalCacheKey, &result)
	debug.Logf("usage", "rate limited — backing off %v (block=%.1f%%)", backoff, result.BlockPercentage)
	return &result
}
