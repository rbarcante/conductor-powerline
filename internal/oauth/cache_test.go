package oauth

import (
	"testing"
	"time"
)

func TestCacheStoreAndRetrieve(t *testing.T) {
	c := NewCache(1 * time.Minute)

	data := &UsageData{
		BlockPercentage:  50.0,
		WeeklyPercentage: 30.0,
		FetchedAt:        time.Now(),
	}
	c.Store(data)

	got := c.Get()
	if got == nil {
		t.Fatal("expected cached data, got nil")
	}
	if got.BlockPercentage != 50.0 {
		t.Errorf("expected block 50.0, got %f", got.BlockPercentage)
	}
	if got.IsStale {
		t.Error("expected fresh data, got stale")
	}
}

func TestCacheTTLExpiry(t *testing.T) {
	c := NewCache(50 * time.Millisecond)

	data := &UsageData{
		BlockPercentage: 50.0,
		FetchedAt:       time.Now(),
	}
	c.Store(data)

	// Data should be fresh initially
	got := c.Get()
	if got == nil || got.IsStale {
		t.Error("expected fresh data immediately after store")
	}

	// Wait for TTL to expire
	time.Sleep(60 * time.Millisecond)

	got = c.Get()
	if got == nil {
		t.Fatal("expected stale data, got nil")
	}
	if !got.IsStale {
		t.Error("expected stale data after TTL expiry")
	}
}

func TestCacheEmptyReturnsNil(t *testing.T) {
	c := NewCache(1 * time.Minute)

	got := c.Get()
	if got != nil {
		t.Errorf("expected nil from empty cache, got %+v", got)
	}
}

func TestCacheStaleIndicator(t *testing.T) {
	c := NewCache(10 * time.Millisecond)

	data := &UsageData{
		BlockPercentage: 75.0,
		FetchedAt:       time.Now(),
	}
	c.Store(data)

	time.Sleep(20 * time.Millisecond)

	got := c.Get()
	if got == nil {
		t.Fatal("expected stale data, got nil")
	}
	if !got.IsStale {
		t.Error("expected IsStale=true after TTL expiry")
	}
	if got.BlockPercentage != 75.0 {
		t.Errorf("expected original data preserved, got %f", got.BlockPercentage)
	}
}
