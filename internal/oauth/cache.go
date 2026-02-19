package oauth

import (
	"sync"
	"time"
)

// Cache stores usage data in memory with a configurable TTL.
type Cache struct {
	mu       sync.RWMutex
	data     *UsageData
	storedAt time.Time
	ttl      time.Duration
}

// NewCache creates a new cache with the given TTL duration.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl}
}

// Store saves usage data to the cache.
func (c *Cache) Store(data *UsageData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = data
	c.storedAt = time.Now()
}

// Get retrieves cached usage data. Returns nil if cache is empty.
// Sets IsStale=true if the data has exceeded the TTL.
func (c *Cache) Get() *UsageData {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil {
		return nil
	}

	// Copy to avoid mutating stored data
	result := *c.data
	if time.Since(c.storedAt) > c.ttl {
		result.IsStale = true
	}
	return &result
}
