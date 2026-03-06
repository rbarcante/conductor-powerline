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
}

// FetchUsage orchestrates usage data retrieval: checks cache first to avoid
// unnecessary API calls, only hitting the API when cache is stale or empty.
// Returns nil with error on first-run failure (no cache available).
func FetchUsage(fetcher UsageFetcher, cache UsageCache, workspaceKey string) (*UsageData, error) {
	// 1. Check cache first — avoid API call if data is fresh
	cached := cache.Get(workspaceKey)
	if cached != nil && !cached.IsStale {
		debug.Logf("usage", "cache hit (fresh): block=%.1f%% weekly=%.1f%%", cached.BlockPercentage, cached.WeeklyPercentage)
		return cached, nil
	}

	// 2. Cache is stale or empty — need to call API
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

	// 3. API failed — serve stale cached data if available
	if cached != nil {
		// Staleness is transient for this render cycle only — not persisted.
		// FileCache.Get() recomputes staleness from TTL on each read.
		cached.IsStale = true
		debug.Logf("usage", "serving stale cached data")
		return cached, nil
	}

	return nil, errors.New("oauth: API failed and no cached data available")
}
