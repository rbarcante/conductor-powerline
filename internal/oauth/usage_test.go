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
	stored  *UsageData
	touched bool
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

func (m *mockCache) Touch(key string) {
	m.touched = true
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

	data, err := FetchUsage(fetcher, cache, "workspace-key", nil)
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

func TestFetchUsageCacheHitFresh(t *testing.T) {
	apiCalled := false
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
	tokenGetter = func() (string, error) {
		apiCalled = true
		return "test-token", nil
	}

	// Fresh cache should return immediately without API call
	data, err := FetchUsage(fetcher, cache, "workspace-key", nil)
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	if data.BlockPercentage != 50.0 {
		t.Errorf("expected cached data 50.0, got %f", data.BlockPercentage)
	}
	if apiCalled {
		t.Error("expected no API call when cache is fresh")
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
			IsStale:         true, // Simulate expired TTL
		},
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache, "workspace-key", nil)
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

	data, err := FetchUsage(fetcher, cache, "workspace-key", nil)
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

	data, err := FetchUsage(fetcher, cache, "workspace-key", nil)
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

	data, err := FetchUsage(fetcher, fc, "global-usage", nil)
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	if data.BlockPercentage != 75.0 {
		t.Errorf("expected block 75.0, got %f", data.BlockPercentage)
	}

	// Verify it persisted to disk via a new FileCache instance
	fc2 := NewFileCache(dir, 1*time.Minute)
	cached := fc2.Get("global-usage")
	if cached == nil {
		t.Fatal("expected cached data from second FileCache instance")
	}
	if cached.BlockPercentage != 75.0 {
		t.Errorf("expected cached block 75.0, got %f", cached.BlockPercentage)
	}
}

func TestFetchUsageLockContention(t *testing.T) {
	dir := t.TempDir()
	lock := NewCacheLock(dir, 15*time.Second)

	// Hold the lock to simulate another process
	if !lock.TryLock() {
		t.Fatal("expected to acquire lock")
	}
	defer lock.Unlock()

	fetcher := &mockFetcher{
		data: &UsageData{BlockPercentage: 99.0},
	}
	cache := &mockCache{
		stored: &UsageData{
			BlockPercentage: 50.0,
			IsStale:         true,
		},
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	apiCalled := false
	tokenGetter = func() (string, error) {
		apiCalled = true
		return "test-token", nil
	}

	// Create a second lock instance for FetchUsage
	lock2 := NewCacheLock(dir, 15*time.Second)
	data, err := FetchUsage(fetcher, cache, "workspace-key", lock2)
	if err != nil {
		t.Fatalf("expected stale data, got error: %v", err)
	}
	if data.BlockPercentage != 50.0 {
		t.Errorf("expected stale cached data 50.0, got %f", data.BlockPercentage)
	}
	if apiCalled {
		t.Error("expected no API call during lock contention")
	}
	if !cache.touched {
		t.Error("expected Touch to be called on contention")
	}
}

func TestFetchUsageLockAcquired(t *testing.T) {
	dir := t.TempDir()

	fetcher := &mockFetcher{
		data: &UsageData{BlockPercentage: 80.0},
	}
	cache := &mockCache{
		stored: &UsageData{
			BlockPercentage: 50.0,
			IsStale:         true,
		},
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	lock := NewCacheLock(dir, 15*time.Second)
	data, err := FetchUsage(fetcher, cache, "workspace-key", lock)
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	if data.BlockPercentage != 80.0 {
		t.Errorf("expected fresh API data 80.0, got %f", data.BlockPercentage)
	}
	if data.IsStale {
		t.Error("expected fresh data from API")
	}

	// Lock should be released
	lock2 := NewCacheLock(dir, 15*time.Second)
	if !lock2.TryLock() {
		t.Error("expected lock to be released after FetchUsage")
	}
	lock2.Unlock()
}
