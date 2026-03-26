package oauth

import (
	"errors"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// tokenGetter is the function used to get the OAuth token.
// It is a package-level variable to allow testing with mocks.
var tokenGetter = GetToken

// UsageCache defines the interface for caching usage data.
// Both the in-memory Cache and file-based FileCache satisfy this interface.
type UsageCache interface {
	Store(key string, data *UsageData)
	Get(key string) *UsageData
	Touch(key string)
}

// FetchUsage orchestrates usage data retrieval with cache-first semantics
// and cross-process lock coordination.
//
// Flow:
//  1. Check cache — if fresh, return immediately (no API call).
//  2. Try to acquire lock — if contention, return stale data + Touch.
//  3. Get token, call API, store result, unlock.
//  4. On API failure, serve stale cache.
//
// When lock is nil, locking is skipped (useful for tests).
func FetchUsage(fetcher UsageFetcher, cache UsageCache, workspaceKey string, lock *CacheLock) (*UsageData, error) {
	// 1. Cache-first: return immediately if fresh
	cached := cache.Get(workspaceKey)
	if cached != nil && !cached.IsStale {
		debug.Logf("usage", "cache hit (fresh): block=%.1f%% weekly=%.1f%%", cached.BlockPercentage, cached.WeeklyPercentage)
		return cached, nil
	}

	// 2. Cache is stale or missing — try to acquire lock
	if lock != nil {
		if !lock.TryLock() {
			// Another process is calling the API — return stale data if available
			debug.Logf("usage", "lock contention — another process is fetching")
			if cached != nil {
				cache.Touch(workspaceKey)
				debug.Logf("usage", "serving stale data + touch (block=%.1f%% weekly=%.1f%%)", cached.BlockPercentage, cached.WeeklyPercentage)
				return cached, nil
			}
			// Cold start: no cached data to serve, proceed without lock.
			// Multiple concurrent first-run processes may all hit the API —
			// acceptable tradeoff since blocking would delay the first render.
			debug.Logf("usage", "lock contention but no cached data — proceeding without lock")
		} else {
			defer lock.Unlock()
		}
	}

	// 3. Fetch token and call API
	debug.Logf("usage", "fetching token...")
	token, err := tokenGetter()
	if err != nil {
		debug.Logf("usage", "token retrieval failed: %v", err)
		if cached != nil {
			// Staleness is transient for this render cycle only — not persisted.
			// FileCache.Get() recomputes staleness from TTL on each read.
			cached.IsStale = true
			return cached, nil
		}
		return nil, err
	}

	data, err := fetcher.FetchUsageData(token)
	if err == nil {
		data.IsStale = false
		cache.Store(workspaceKey, data)
		debug.Logf("usage", "API success: block=%.1f%% weekly=%.1f%%", data.BlockPercentage, data.WeeklyPercentage)
		return data, nil
	}
	debug.Logf("usage", "API call failed: %v", err)

	// 4. API failed — try serving cached data
	if cached != nil {
		// Staleness is transient for this render cycle only — not persisted.
		// FileCache.Get() recomputes staleness from TTL on each read.
		cached.IsStale = true
		debug.Logf("usage", "serving stale cached data")
		return cached, nil
	}

	return nil, errors.New("oauth: API failed and no cached data available")
}
