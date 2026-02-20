package oauth

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// fileCacheEntry is the on-disk JSON structure for cached usage data.
type fileCacheEntry struct {
	Data     UsageData `json:"data"`
	StoredAt time.Time `json:"stored_at"`
	TTL      string    `json:"ttl"`
}

// FileCache persists usage data to disk, keyed by workspace path hash.
// Each workspace gets its own JSON file under the cache directory.
type FileCache struct {
	dir string
	ttl time.Duration
}

// NewFileCache creates a file-based cache rooted at dir with the given TTL.
// If dir does not exist, it will be created on the first Store call.
func NewFileCache(dir string, ttl time.Duration) *FileCache {
	return &FileCache{dir: dir, ttl: ttl}
}

// Store writes usage data to a JSON file keyed by the workspace identifier.
// Silently returns on any I/O error (graceful degradation).
func (fc *FileCache) Store(key string, data *UsageData) {
	if err := os.MkdirAll(fc.dir, 0o700); err != nil {
		debug.Logf("filecache", "cannot create cache dir: %v", err)
		return
	}

	entry := fileCacheEntry{
		Data:     *data,
		StoredAt: time.Now(),
		TTL:      fc.ttl.String(),
	}

	b, err := json.Marshal(entry)
	if err != nil {
		debug.Logf("filecache", "marshal error: %v", err)
		return
	}

	path := fc.keyPath(key)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		debug.Logf("filecache", "write error: %v", err)
	}

	fc.cleanup()
}

// cleanupMaxAge is the maximum age for cache files before they are removed.
const cleanupMaxAge = 7 * 24 * time.Hour

// cleanup removes cache files not modified in the last 7 days.
func (fc *FileCache) cleanup() {
	entries, err := os.ReadDir(fc.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if time.Since(info.ModTime()) > cleanupMaxAge {
			_ = os.Remove(filepath.Join(fc.dir, e.Name()))
		}
	}
}

// Get reads cached usage data for the given key. Returns nil if the file
// does not exist or cannot be read. Sets IsStale if the TTL has expired.
func (fc *FileCache) Get(key string) *UsageData {
	path := fc.keyPath(key)

	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var entry fileCacheEntry
	if err := json.Unmarshal(b, &entry); err != nil {
		debug.Logf("filecache", "unmarshal error for %s: %v", key, err)
		return nil
	}

	result := entry.Data
	if time.Since(entry.StoredAt) > fc.ttl {
		result.IsStale = true
	}
	return &result
}

// keyPath returns the file path for a given workspace key.
// The key is hashed with SHA-256 to produce a safe, fixed-length filename.
func (fc *FileCache) keyPath(key string) string {
	h := sha256.Sum256([]byte(key))
	return filepath.Join(fc.dir, fmt.Sprintf("%x.json", h))
}
