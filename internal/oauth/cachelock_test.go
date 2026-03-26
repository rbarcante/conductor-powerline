package oauth

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCacheLockAcquireAndRelease(t *testing.T) {
	dir := t.TempDir()
	cl := NewCacheLock(dir, 15*time.Second)

	if !cl.TryLock() {
		t.Fatal("expected to acquire lock")
	}

	// Lock file should exist
	if _, err := os.Stat(filepath.Join(dir, ".lock")); os.IsNotExist(err) {
		t.Fatal("expected lock file to exist")
	}

	cl.Unlock()

	// Lock file should be gone
	if _, err := os.Stat(filepath.Join(dir, ".lock")); !os.IsNotExist(err) {
		t.Fatal("expected lock file to be removed after unlock")
	}
}

func TestCacheLockContention(t *testing.T) {
	dir := t.TempDir()
	cl1 := NewCacheLock(dir, 15*time.Second)
	cl2 := NewCacheLock(dir, 15*time.Second)

	if !cl1.TryLock() {
		t.Fatal("first lock should succeed")
	}
	defer cl1.Unlock()

	if cl2.TryLock() {
		t.Fatal("second lock should fail (contention)")
	}
}

func TestCacheLockStaleRemoval(t *testing.T) {
	dir := t.TempDir()

	// Create a lock file and backdate it
	lockPath := filepath.Join(dir, ".lock")
	if err := os.WriteFile(lockPath, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-30 * time.Second)
	_ = os.Chtimes(lockPath, oldTime, oldTime)

	cl := NewCacheLock(dir, 15*time.Second)

	if !cl.TryLock() {
		t.Fatal("expected to acquire lock after stale removal")
	}
	defer cl.Unlock()
}

func TestCacheLockUnlockIdempotent(t *testing.T) {
	dir := t.TempDir()
	cl := NewCacheLock(dir, 15*time.Second)

	if !cl.TryLock() {
		t.Fatal("expected to acquire lock")
	}

	// Multiple unlocks should not panic
	cl.Unlock()
	cl.Unlock()
	cl.Unlock()
}

func TestCacheLockUnwritableDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unwritable directory semantics are not reliable on Windows")
	}

	cl := NewCacheLock("/nonexistent/path/that/does/not/exist", 15*time.Second)

	// Should degrade gracefully (return true — proceed without lock)
	if !cl.TryLock() {
		t.Fatal("expected graceful degradation (true) on unwritable dir")
	}
}

func TestCacheLockNilSafe(t *testing.T) {
	var cl *CacheLock

	// nil lock should return true (no-op)
	if !cl.TryLock() {
		t.Fatal("nil lock TryLock should return true")
	}

	// nil unlock should not panic
	cl.Unlock()
}
