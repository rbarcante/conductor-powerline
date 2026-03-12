package oauth

import (
	"errors"
	"sync"
	"sync/atomic"
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

// countingFetcher counts API calls and is safe for concurrent use.
type countingFetcher struct {
	data      *UsageData
	err       error
	callCount *int64
}

func (c *countingFetcher) FetchUsageData(token string) (*UsageData, error) {
	atomic.AddInt64(c.callCount, 1)
	return c.data, c.err
}

// mockCache implements LockableCache for testing.
// TryLock uses an internal mutex — always acquires (no inter-process locking needed for unit tests).
type mockCache struct {
	mu     sync.Mutex
	stored *UsageData
	locked bool
}

func (m *mockCache) Store(key string, data *UsageData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stored = data
}

func (m *mockCache) Get(key string) *UsageData {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stored == nil {
		return nil
	}
	result := *m.stored
	return &result
}

func (m *mockCache) TryLock(key string) (bool, func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locked {
		return false, nil
	}
	m.locked = true
	return true, func() {
		m.mu.Lock()
		m.locked = false
		m.mu.Unlock()
	}
}

func (m *mockCache) WaitForUnlock(key string, timeout time.Duration) bool {
	return true
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
	// With cache-first design, fresh cached data is returned without calling the API.
	fetcher := &mockFetcher{
		data: &UsageData{BlockPercentage: 99.0},
	}
	cache := &mockCache{
		stored: &UsageData{
			BlockPercentage: 50.0,
			FetchedAt:       time.Now(),
			IsStale:         false, // explicitly fresh
		},
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	// Cache-first: should return cached value (50.0), not API value (99.0)
	if data.BlockPercentage != 50.0 {
		t.Errorf("expected cached data 50.0 (cache-first), got %f", data.BlockPercentage)
	}
	if data.IsStale {
		t.Error("expected fresh data")
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
			IsStale:         true, // stale so cache-first check triggers API attempt
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

// TestFetchUsage_OnlyOneAPICallOnConcurrentExpiry verifies that when multiple goroutines
// concurrently encounter a stale cache, only one makes an API call and the others wait.
func TestFetchUsage_OnlyOneAPICallOnConcurrentExpiry(t *testing.T) {
	dir := t.TempDir()
	const shortTTL = 100 * time.Millisecond

	// Pre-seed with data that will be stale after the TTL elapses.
	fc := NewFileCache(dir, shortTTL)
	fc.Store("ws-key", &UsageData{BlockPercentage: 10.0, FetchedAt: time.Now()})
	time.Sleep(150 * time.Millisecond) // wait for TTL to expire

	var apiCalls int64
	fetcher := &countingFetcher{
		data:      &UsageData{BlockPercentage: 75.0},
		callCount: &apiCalls,
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			FetchUsage(fetcher, fc, "ws-key") //nolint:errcheck
		}()
	}
	wg.Wait()

	got := atomic.LoadInt64(&apiCalls)
	if got != 1 {
		t.Errorf("expected exactly 1 API call (thundering herd prevention), got %d", got)
	}
}

// TestFetchUsage_WaiterGetsRefreshedData verifies that a goroutine waiting for the lock
// reads the data written by the lock holder rather than making its own API call.
func TestFetchUsage_WaiterGetsRefreshedData(t *testing.T) {
	dir := t.TempDir()
	const shortTTL = 100 * time.Millisecond

	// Pre-seed stale data using the same TTL that FetchUsage will use.
	fc := NewFileCache(dir, shortTTL)
	fc.Store("ws-key", &UsageData{BlockPercentage: 10.0, FetchedAt: time.Now()})
	time.Sleep(150 * time.Millisecond) // wait for TTL to expire

	// Pre-acquire the lock to simulate another process refreshing
	ok, release := fc.TryLock("ws-key")
	if !ok {
		t.Fatal("expected to acquire lock for test setup")
	}

	var apiCalls int64
	fetcher := &countingFetcher{
		data:      &UsageData{BlockPercentage: 99.0},
		callCount: &apiCalls,
	}

	origTokenGetter := tokenGetter
	defer func() { tokenGetter = origTokenGetter }()
	tokenGetter = func() (string, error) { return "test-token", nil }

	// After 80ms: write fresh data then release lock (simulates lock holder completing).
	// Store is done before release so the waiter sees the refreshed data on Get().
	go func() {
		time.Sleep(80 * time.Millisecond)
		fc.Store("ws-key", &UsageData{BlockPercentage: 99.0, FetchedAt: time.Now()})
		release()
	}()

	data, err := FetchUsage(fetcher, fc, "ws-key")
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}

	// Waiter should have read the data written by the lock holder
	if data.BlockPercentage != 99.0 {
		t.Errorf("expected 99.0 from lock holder's write, got %f", data.BlockPercentage)
	}

	// Waiter must NOT have called the API itself
	if got := atomic.LoadInt64(&apiCalls); got != 0 {
		t.Errorf("expected 0 API calls (waiter reads cache), got %d", got)
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
