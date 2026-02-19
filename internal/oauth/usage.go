package oauth

import (
	"errors"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// tokenGetter is the function used to get the OAuth token.
// It is a package-level variable to allow testing with mocks.
var tokenGetter = GetToken

// FetchUsage orchestrates usage data retrieval: gets token, calls API,
// caches on success, serves stale on failure. Returns nil with error
// on first-run failure (no cache available).
func FetchUsage(fetcher UsageFetcher, cache *Cache) (*UsageData, error) {
	debug.Logf("usage", "fetching token...")
	token, err := tokenGetter()
	if err != nil {
		debug.Logf("usage", "token retrieval failed: %v", err)
		return nil, err
	}
	debug.Logf("usage", "token retrieved, calling API...")
	data, err := fetcher.FetchUsageData(token)
	if err == nil {
		data.IsStale = false
		cache.Store(data)
		debug.Logf("usage", "API success: block=%.1f%% weekly=%.1f%%", data.BlockPercentage, data.WeeklyPercentage)
		return data, nil
	}
	debug.Logf("usage", "API call failed: %v", err)

	// API failed — try serving cached data
	cached := cache.Get()
	if cached != nil {
		cached.IsStale = true
		debug.Logf("usage", "serving stale cached data (block=%.1f%% weekly=%.1f%%)", cached.BlockPercentage, cached.WeeklyPercentage)
		return cached, nil
	}

	// First run, no cache — return error
	debug.Logf("usage", "no cached data available — returning error")
	return nil, errors.New("oauth: API failed and no cached data available")
}
