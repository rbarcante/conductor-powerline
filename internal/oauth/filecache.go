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

const lockPollInterval = 50 * time.Millisecond

// staleLockAge is the age after which a lock file is considered orphaned and can
// be safely removed. This handles the case where a process crashes or is killed
// without releasing its lock, which would otherwise block all future refreshes.
const staleLockAge = 10 * time.Second

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
	if err := fc.atomicWrite(path, b); err != nil {
		debug.Logf("filecache", "write error: %v", err)
	}

	fc.cleanup()
}

// atomicWrite writes data to a temp file then renames it to path, ensuring readers
// never see a partial write. The temp file is cleaned up on any error.
func (fc *FileCache) atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(fc.dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o600); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
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

// lockPath returns the path for the lock file associated with the given key.
func (fc *FileCache) lockPath(key string) string {
	return fc.keyPath(key) + ".lock"
}

// TryLock attempts to acquire an exclusive lock for the given key using an
// O_EXCL file creation — atomic on both POSIX and Windows NTFS.
// Returns (true, releaseFn) on success; (false, nil) if already locked.
// The caller must call releaseFn to release the lock.
//
// If an existing lock file is older than staleLockAge, it is treated as orphaned
// (the holder likely crashed) and removed before retrying.
func (fc *FileCache) TryLock(key string) (bool, func()) {
	if err := os.MkdirAll(fc.dir, 0o700); err != nil {
		return false, nil
	}
	lockFile := fc.lockPath(key)
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		// Lock file exists — check if it's orphaned (stale).
		info, statErr := os.Stat(lockFile)
		if statErr == nil && time.Since(info.ModTime()) > staleLockAge {
			debug.Logf("filecache", "removing stale lock file (age %v)", time.Since(info.ModTime()))
			os.Remove(lockFile)
			// Retry once after removing the stale lock.
			f, err = os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL, 0o600)
			if err != nil {
				return false, nil
			}
		} else {
			return false, nil
		}
	}
	f.Close()
	release := func() {
		os.Remove(lockFile)
	}
	return true, release
}

// WaitForUnlock polls every lockPollInterval until the lock file for key
// disappears or the timeout elapses. Returns true if the lock was released
// before the timeout, false if the timeout expired first.
func (fc *FileCache) WaitForUnlock(key string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	lockFile := fc.lockPath(key)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(lockFile); os.IsNotExist(err) {
			return true
		}
		time.Sleep(lockPollInterval)
	}
	// Final check after sleep
	_, err := os.Stat(lockFile)
	return os.IsNotExist(err)
}
