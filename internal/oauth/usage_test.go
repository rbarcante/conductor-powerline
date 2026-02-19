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

// mockCache implements UsageCache for testing.
type mockCache struct {
	stored *UsageData
}

func (m *mockCache) Store(key string, data *UsageData) {
	m.stored = data
}

func (m *mockCache) Get(key string) *UsageData {
	if m.stored == nil {
		return nil
	}
	result := *m.stored
	return &result
}

func TestFetchUsageFreshFetch(t *testing.T) {
	fetcher := &mockFetcher{
		data: &UsageData{
			BlockPercentage:  60.0,
			WeeklyPercentage: 40.0,
		},
	}
	cache := &mockCache{}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	if data.BlockPercentage != 60.0 {
		t.Errorf("expected block 60.0, got %f", data.BlockPercentage)
	}
	if data.IsStale {
		t.Error("expected fresh data")
	}
	// Verify data was stored in cache
	if cache.stored == nil {
		t.Error("expected data to be stored in cache")
	}
}

func TestFetchUsageCacheHit(t *testing.T) {
	fetcher := &mockFetcher{
		data: &UsageData{BlockPercentage: 99.0},
	}
	cache := &mockCache{
		stored: &UsageData{
			BlockPercentage: 50.0,
			FetchedAt:       time.Now(),
		},
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	// FetchUsage should try API and get fresh data
	data, err := FetchUsage(fetcher, cache, "workspace-key")
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
	cache := &mockCache{
		stored: &UsageData{
			BlockPercentage: 50.0,
			FetchedAt:       time.Now(),
		},
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache, "workspace-key")
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
	cache := &mockCache{}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err == nil {
		t.Error("expected error on first run with API failure")
	}
	if data != nil {
		t.Errorf("expected nil data, got %+v", data)
	}
}

func TestFetchUsageNoToken(t *testing.T) {
	fetcher := &mockFetcher{}
	cache := &mockCache{}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "", errors.New("no token") }

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err == nil {
		t.Error("expected error when no token available")
	}
	if data != nil {
		t.Errorf("expected nil data, got %+v", data)
	}
}

func TestFetchUsageWithFileCache(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	fetcher := &mockFetcher{
		data: &UsageData{
			BlockPercentage:  75.0,
			WeeklyPercentage: 55.0,
		},
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, fc, "/home/user/my-project")
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	if data.BlockPercentage != 75.0 {
		t.Errorf("expected block 75.0, got %f", data.BlockPercentage)
	}

	// Verify it persisted to disk via a new FileCache instance
	fc2 := NewFileCache(dir, 1*time.Minute)
	cached := fc2.Get("/home/user/my-project")
	if cached == nil {
		t.Fatal("expected cached data from second FileCache instance")
	}
	if cached.BlockPercentage != 75.0 {
		t.Errorf("expected cached block 75.0, got %f", cached.BlockPercentage)
	}
}
