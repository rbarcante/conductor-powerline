package oauth

import "errors"

// tokenGetter is the function used to get the OAuth token.
// It is a package-level variable to allow testing with mocks.
var tokenGetter = GetToken

// FetchUsage orchestrates usage data retrieval: gets token, calls API,
// caches on success, serves stale on failure. Returns nil with error
// on first-run failure (no cache available).
func FetchUsage(fetcher usageFetcher, cache *Cache) (*UsageData, error) {
	token, err := tokenGetter()
	if err != nil {
		return nil, err
	}

	data, err := fetcher.FetchUsageData(token)
	if err == nil {
		data.IsStale = false
		cache.Store(data)
		return data, nil
	}

	// API failed — try serving cached data
	cached := cache.Get()
	if cached != nil {
		cached.IsStale = true
		return cached, nil
	}

	// First run, no cache — return error
	return nil, errors.New("oauth: API failed and no cached data available")
}
