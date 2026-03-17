package oauth

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// CacheLock implements a cross-process, non-blocking file lock using
// O_CREATE|O_EXCL for atomic creation. It guards the usage API call
// so that only one conductor-powerline process calls the API at a time.
type CacheLock struct {
	path     string
	staleAge time.Duration
}

// NewCacheLock returns a CacheLock that writes a lock file in dir.
// staleAge is the duration after which a lock file is considered abandoned
// and can be removed by another process.
func NewCacheLock(dir string, staleAge time.Duration) *CacheLock {
	return &CacheLock{
		path:     filepath.Join(dir, ".lock"),
		staleAge: staleAge,
	}
}

// TryLock attempts to acquire the lock without blocking.
// Returns true if the lock was acquired, false otherwise.
// On ErrExist: checks if the lock is stale, removes it, and retries once.
func (cl *CacheLock) TryLock() bool {
	if cl == nil {
		return true
	}

	if err := os.MkdirAll(filepath.Dir(cl.path), 0o700); err != nil {
		debug.Logf("cachelock", "cannot create lock dir: %v", err)
		return true // degrade gracefully — proceed without lock
	}

	acquired, err := cl.tryCreate()
	if acquired {
		return true
	}

	if !os.IsExist(err) {
		debug.Logf("cachelock", "unexpected lock error: %v", err)
		return true // degrade gracefully
	}

	// Lock file exists — check if stale (Lstat avoids following symlinks)
	info, statErr := os.Lstat(cl.path)
	if statErr != nil {
		debug.Logf("cachelock", "cannot stat lock file: %v", statErr)
		return false
	}

	age := time.Since(info.ModTime())
	if age <= cl.staleAge {
		debug.Logf("cachelock", "lock held by another process (age %v)", age)
		return false
	}

	// Stale lock — remove and retry once.
	// NOTE: There is a small TOCTOU window between Remove and tryCreate where
	// another process can also detect the stale lock and acquire it. This is an
	// inherent limitation of file-based locking without OS-level advisory locks.
	// Worst case is a duplicate API call, not data corruption.
	debug.Logf("cachelock", "removing stale lock (age %v > %v)", age, cl.staleAge)
	_ = os.Remove(cl.path)

	acquired, _ = cl.tryCreate()
	if acquired {
		debug.Logf("cachelock", "acquired lock after stale removal")
		return true
	}

	debug.Logf("cachelock", "failed to acquire lock after stale removal (another process won the race)")
	return false
}

// tryCreate atomically creates the lock file. Returns true on success.
func (cl *CacheLock) tryCreate() (bool, error) {
	f, err := os.OpenFile(cl.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return false, err
	}
	// Write timestamp for debugging
	_, _ = fmt.Fprintf(f, "%d", time.Now().UnixNano())
	_ = f.Close()
	return true, nil
}

// Unlock removes the lock file. Errors are ignored (idempotent).
func (cl *CacheLock) Unlock() {
	if cl == nil {
		return
	}
	if err := os.Remove(cl.path); err != nil && !os.IsNotExist(err) {
		debug.Logf("cachelock", "unlock remove error: %v", err)
	}
}
