package oauth

import (
	"errors"
	"os"
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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) { return nil, errors.New("no token") }

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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

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

// TestTryLock_StaleLockRecovery verifies that an orphaned lock file older than
// staleLockAge is automatically removed, allowing a new process to acquire the lock.
func TestTryLock_StaleLockRecovery(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	// Acquire and "orphan" the lock by not calling the release function.
	ok, _ := fc.TryLock("test-key")
	if !ok {
		t.Fatal("expected to acquire initial lock")
	}

	// A second TryLock should fail — the lock is fresh.
	ok2, _ := fc.TryLock("test-key")
	if ok2 {
		t.Fatal("expected second TryLock to fail while lock is fresh")
	}

	// Backdate the lock file to make it appear stale.
	lockFile := fc.lockPath("test-key")
	staleTime := time.Now().Add(-(staleLockAge + 1*time.Second))
	if err := os.Chtimes(lockFile, staleTime, staleTime); err != nil {
		t.Fatalf("failed to backdate lock file: %v", err)
	}

	// Now TryLock should succeed — the stale lock is auto-removed.
	ok3, release3 := fc.TryLock("test-key")
	if !ok3 {
		t.Fatal("expected TryLock to succeed after stale lock recovery")
	}
	release3()
}

// TestFetchUsage_RateLimitExtendsCacheTTL verifies that when the API returns
// a 429 RateLimitError and stale cache exists, FetchUsage re-stores the stale
// data with a refreshed timestamp (extending the TTL) to prevent immediate retry.
func TestFetchUsage_RateLimitExtendsCacheTTL(t *testing.T) {
	staleData := &UsageData{
		BlockPercentage: 42.0,
		FetchedAt:       time.Now().Add(-10 * time.Minute),
		IsStale:         true,
	}
	fetcher := &mockFetcher{
		err: &RateLimitError{RetryAfter: 30 * time.Second, Body: "rate limited"},
	}
	cache := &mockCache{stored: staleData}

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected stale data on rate limit, got error: %v", err)
	}
	if data.BlockPercentage != 42.0 {
		t.Errorf("expected cached block 42.0, got %f", data.BlockPercentage)
	}
	if !data.IsStale {
		t.Error("expected data marked as stale")
	}
	// Verify cache was re-stored with refreshed timestamp.
	if cache.stored == nil {
		t.Fatal("expected cache.Store to have been called")
	}
	if time.Since(cache.stored.FetchedAt) > 2*time.Second {
		t.Errorf("expected refreshed FetchedAt (recent), got %v ago", time.Since(cache.stored.FetchedAt))
	}
}

// TestFetchUsage_RateLimitNoCacheFallsThrough verifies that a 429 without
// any cached data falls through to the generic error path.
func TestFetchUsage_RateLimitNoCacheFallsThrough(t *testing.T) {
	fetcher := &mockFetcher{
		err: &RateLimitError{RetryAfter: 10 * time.Second},
	}
	cache := &mockCache{}

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err == nil {
		t.Error("expected error when rate limited with no cache")
	}
	if data != nil {
		t.Errorf("expected nil data, got %+v", data)
	}
}

// --- Token rotation tests ---

// mockCredentialWriter prevents write-back to real credential stores in tests.
func mockCredentialWriter(t *testing.T) func() {
	t.Helper()
	orig := credentialWriter
	credentialWriter = func(creds *TokenCredentials) {}
	return func() { credentialWriter = orig }
}

func TestFetchUsage_429_RotationSuccess(t *testing.T) {
	defer mockCredentialWriter(t)()
	dir := t.TempDir()
	origDir := rotatedTokenDir
	defer func() { rotatedTokenDir = origDir }()
	rotatedTokenDir = dir

	callCount := 0
	fetcher := &mockFetcher{}
	// Override FetchUsageData to return 429 on first call, success on second
	dynamicFetcher := &dynamicMockFetcher{
		fn: func(token string) (*UsageData, error) {
			callCount++
			if callCount == 1 {
				return nil, &RateLimitError{RetryAfter: 30 * time.Second}
			}
			// Second call with rotated token
			if token != "new-access-token" {
				t.Errorf("expected retry with new-access-token, got %q", token)
			}
			return &UsageData{BlockPercentage: 88.0, WeeklyPercentage: 44.0}, nil
		},
	}
	_ = fetcher

	cache := &mockCache{
		stored: &UsageData{BlockPercentage: 42.0, FetchedAt: time.Now().Add(-10 * time.Minute), IsStale: true},
	}

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "old-token", RefreshToken: "old-refresh"}, nil
	}

	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = func(refreshToken string) (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "new-access-token", RefreshToken: "new-refresh"}, nil
	}

	data, err := FetchUsage(dynamicFetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected data, got error: %v", err)
	}
	if data.BlockPercentage != 88.0 {
		t.Errorf("expected 88.0 from rotated token, got %f", data.BlockPercentage)
	}
	if data.IsStale {
		t.Error("expected fresh data after rotation")
	}
}

func TestFetchUsage_429_RotationRetryFails(t *testing.T) {
	defer mockCredentialWriter(t)()
	dir := t.TempDir()
	origDir := rotatedTokenDir
	defer func() { rotatedTokenDir = origDir }()
	rotatedTokenDir = dir

	callCount := 0
	dynamicFetcher := &dynamicMockFetcher{
		fn: func(token string) (*UsageData, error) {
			callCount++
			// Both calls return 429
			return nil, &RateLimitError{RetryAfter: 30 * time.Second}
		},
	}

	staleData := &UsageData{BlockPercentage: 42.0, FetchedAt: time.Now().Add(-10 * time.Minute), IsStale: true}
	cache := &mockCache{stored: staleData}

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "old-token", RefreshToken: "old-refresh"}, nil
	}

	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = func(refreshToken string) (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "new-token", RefreshToken: "new-refresh"}, nil
	}

	data, err := FetchUsage(dynamicFetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected stale data, got error: %v", err)
	}
	if !data.IsStale {
		t.Error("expected stale data when rotation retry fails")
	}
	if data.BlockPercentage != 42.0 {
		t.Errorf("expected cache fallback 42.0, got %f", data.BlockPercentage)
	}
}

func TestFetchUsage_429_RefreshFails(t *testing.T) {
	defer mockCredentialWriter(t)()
	dir := t.TempDir()
	origDir := rotatedTokenDir
	defer func() { rotatedTokenDir = origDir }()
	rotatedTokenDir = dir

	fetcher := &mockFetcher{
		err: &RateLimitError{RetryAfter: 30 * time.Second},
	}

	staleData := &UsageData{BlockPercentage: 42.0, FetchedAt: time.Now().Add(-10 * time.Minute), IsStale: true}
	cache := &mockCache{stored: staleData}

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "old-token", RefreshToken: "old-refresh"}, nil
	}

	origRefresher := tokenRefresher
	defer func() { tokenRefresher = origRefresher }()
	tokenRefresher = func(refreshToken string) (*TokenCredentials, error) {
		return nil, &RefreshError{StatusCode: 400, Body: "invalid_grant"}
	}

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected stale data, got error: %v", err)
	}
	if !data.IsStale {
		t.Error("expected stale data when refresh fails")
	}
}

func TestFetchUsage_429_NoRefreshToken(t *testing.T) {
	fetcher := &mockFetcher{
		err: &RateLimitError{RetryAfter: 30 * time.Second},
	}

	staleData := &UsageData{BlockPercentage: 42.0, FetchedAt: time.Now().Add(-10 * time.Minute), IsStale: true}
	cache := &mockCache{stored: staleData}

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "old-token"}, nil // No refresh token
	}

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected stale data, got error: %v", err)
	}
	if !data.IsStale {
		t.Error("expected stale data when no refresh token")
	}
}

func TestFetchUsage_429_RotationLockHeld(t *testing.T) {
	dir := t.TempDir()
	origDir := rotatedTokenDir
	defer func() { rotatedTokenDir = origDir }()
	rotatedTokenDir = dir

	// Pre-acquire the rotation lock
	acquired, release := TryRotationLock(dir)
	if !acquired {
		t.Fatal("expected to acquire lock")
	}
	defer release()

	fetcher := &mockFetcher{
		err: &RateLimitError{RetryAfter: 30 * time.Second},
	}

	staleData := &UsageData{BlockPercentage: 42.0, FetchedAt: time.Now().Add(-10 * time.Minute), IsStale: true}
	cache := &mockCache{stored: staleData}

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "old-token", RefreshToken: "old-refresh"}, nil
	}

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err != nil {
		t.Fatalf("expected stale data, got error: %v", err)
	}
	if !data.IsStale {
		t.Error("expected stale data when rotation lock held")
	}
}

func TestFetchUsage_429_NoCacheNoRefreshToken(t *testing.T) {
	fetcher := &mockFetcher{
		err: &RateLimitError{RetryAfter: 10 * time.Second},
	}
	cache := &mockCache{} // No cached data

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "old-token"}, nil // No refresh token
	}

	data, err := FetchUsage(fetcher, cache, "workspace-key")
	if err == nil {
		t.Error("expected error when rate limited with no cache and no refresh token")
	}
	if data != nil {
		t.Errorf("expected nil data, got %+v", data)
	}
}

// dynamicMockFetcher allows different responses per call.
type dynamicMockFetcher struct {
	fn func(token string) (*UsageData, error)
}

func (d *dynamicMockFetcher) FetchUsageData(token string) (*UsageData, error) {
	return d.fn(token)
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

	origCredentialsGetter := credentialsGetter
	defer func() { credentialsGetter = origCredentialsGetter }()
	credentialsGetter = func() (*TokenCredentials, error) {
		return &TokenCredentials{AccessToken: "test-token"}, nil
	}

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
