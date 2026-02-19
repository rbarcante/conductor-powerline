package oauth

import (
	"errors"
	"testing"
	"time"
)

// mockFetcher implements UsageFetcher for testing.
type mockFetcher struct {
	data *UsageData
	err  error
}

func (m *mockFetcher) FetchUsageData(token string) (*UsageData, error) {
	return m.data, m.err
}

func TestFetchUsageFreshFetch(t *testing.T) {
	fetcher := &mockFetcher{
		data: &UsageData{
			BlockPercentage:  60.0,
			WeeklyPercentage: 40.0,
		},
	}
	cache := NewCache(1 * time.Minute)

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache)
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	if data.BlockPercentage != 60.0 {
		t.Errorf("expected block 60.0, got %f", data.BlockPercentage)
	}
	if data.IsStale {
		t.Error("expected fresh data")
	}
}

func TestFetchUsageCacheHit(t *testing.T) {
	fetcher := &mockFetcher{
		data: &UsageData{BlockPercentage: 99.0},
	}
	cache := NewCache(1 * time.Minute)
	cache.Store(&UsageData{
		BlockPercentage: 50.0,
		FetchedAt:       time.Now(),
	})

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	// FetchUsage should try API and get fresh data
	data, err := FetchUsage(fetcher, cache)
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	// Should have fetched new data from API
	if data.BlockPercentage != 99.0 {
		t.Errorf("expected fresh API data 99.0, got %f", data.BlockPercentage)
	}
}

func TestFetchUsageStaleFallback(t *testing.T) {
	fetcher := &mockFetcher{
		err: errors.New("api error"),
	}
	cache := NewCache(10 * time.Millisecond)
	cache.Store(&UsageData{
		BlockPercentage: 50.0,
		FetchedAt:       time.Now(),
	})

	// Wait for cache to become stale
	time.Sleep(20 * time.Millisecond)

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache)
	if err != nil {
		t.Fatalf("expected stale data, got error: %v", err)
	}
	if !data.IsStale {
		t.Error("expected stale indicator")
	}
	if data.BlockPercentage != 50.0 {
		t.Errorf("expected cached data, got %f", data.BlockPercentage)
	}
}

func TestFetchUsageFirstRunPlaceholder(t *testing.T) {
	fetcher := &mockFetcher{
		err: errors.New("api error"),
	}
	cache := NewCache(1 * time.Minute)

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache)
	if err == nil {
		t.Error("expected error on first run with API failure")
	}
	if data != nil {
		t.Errorf("expected nil data, got %+v", data)
	}
}

func TestFetchUsageNoToken(t *testing.T) {
	fetcher := &mockFetcher{}
	cache := NewCache(1 * time.Minute)

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "", errors.New("no token") }

	data, err := FetchUsage(fetcher, cache)
	if err == nil {
		t.Error("expected error when no token available")
	}
	if data != nil {
		t.Errorf("expected nil data, got %+v", data)
	}
}
