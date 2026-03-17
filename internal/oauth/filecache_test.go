package oauth

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFileCacheStoreAndGet(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	data := &UsageData{
		BlockPercentage:  60.0,
		WeeklyPercentage: 40.0,
		FetchedAt:        time.Now(),
	}

	fc.Store("project-a", data)

	got := fc.Get("project-a")
	if got == nil {
		t.Fatal("expected cached data, got nil")
	}
	if got.BlockPercentage != 60.0 {
		t.Errorf("expected block 60.0, got %f", got.BlockPercentage)
	}
	if got.WeeklyPercentage != 40.0 {
		t.Errorf("expected weekly 40.0, got %f", got.WeeklyPercentage)
	}
	if got.IsStale {
		t.Error("expected fresh data, got stale")
	}
}

func TestFileCacheKeyIsolation(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	dataA := &UsageData{BlockPercentage: 10.0, FetchedAt: time.Now()}
	dataB := &UsageData{BlockPercentage: 90.0, FetchedAt: time.Now()}

	fc.Store("project-a", dataA)
	fc.Store("project-b", dataB)

	gotA := fc.Get("project-a")
	gotB := fc.Get("project-b")

	if gotA == nil || gotB == nil {
		t.Fatal("expected data for both keys")
	}
	if gotA.BlockPercentage != 10.0 {
		t.Errorf("expected project-a block 10.0, got %f", gotA.BlockPercentage)
	}
	if gotB.BlockPercentage != 90.0 {
		t.Errorf("expected project-b block 90.0, got %f", gotB.BlockPercentage)
	}
}

func TestFileCacheTTLExpiry(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 50*time.Millisecond)

	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}
	fc.Store("test-key", data)

	// Fresh immediately
	got := fc.Get("test-key")
	if got == nil || got.IsStale {
		t.Error("expected fresh data immediately after store")
	}

	// Wait for TTL
	time.Sleep(60 * time.Millisecond)

	got = fc.Get("test-key")
	if got == nil {
		t.Fatal("expected stale data, got nil")
	}
	if !got.IsStale {
		t.Error("expected stale data after TTL expiry")
	}
}

func TestFileCacheGetMissReturnsNil(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	got := fc.Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil for missing key, got %+v", got)
	}
}

func TestFileCacheUnwritableDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unwritable directory semantics are not reliable on Windows")
	}
	fc := NewFileCache("/nonexistent/path/that/does/not/exist", 1*time.Minute)

	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}
	// Should not panic
	fc.Store("test", data)

	// Should return nil gracefully
	got := fc.Get("test")
	if got != nil {
		t.Errorf("expected nil from unwritable cache, got %+v", got)
	}
}

func TestFileCacheFileCreatedOnDisk(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	data := &UsageData{BlockPercentage: 75.0, FetchedAt: time.Now()}
	fc.Store("my-workspace", data)

	// Verify a file exists in the cache directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read cache dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected cache file on disk, found none")
	}
}

func TestFileCacheHashedFilename(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}
	fc.Store("/home/user/very/long/workspace/path", data)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read cache dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 cache file, got %d", len(entries))
	}

	// Filename should be a hex hash, not the raw path
	name := entries[0].Name()
	if filepath.Ext(name) != ".json" {
		t.Errorf("expected .json extension, got %q", name)
	}
	// SHA-256 hex = 64 chars + .json = 69 chars
	if len(name) != 69 {
		t.Errorf("expected 69-char filename (sha256.json), got %d: %q", len(name), name)
	}
}

func TestFileCacheCleanupRemovesOldFiles(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}

	// Store 3 entries
	fc.Store("old-project-1", data)
	fc.Store("old-project-2", data)
	fc.Store("recent-project", data)

	// Backdate the first two files by 8 days
	entries, _ := os.ReadDir(dir)
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	for i, e := range entries {
		if i < 2 {
			path := filepath.Join(dir, e.Name())
			_ = os.Chtimes(path, oldTime, oldTime)
		}
	}

	// Advance counter so the next Store triggers cleanup (runs every 10th call)
	fc.storeCount.Store(int64(cleanupInterval - 1))

	// Trigger cleanup via Store
	fc.Store("trigger-cleanup", data)

	// Should have 2 files left: recent-project + trigger-cleanup
	entries, _ = os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("expected 2 files after cleanup, got %d", len(entries))
	}
}

func TestFileCacheCleanupKeepsRecentFiles(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}

	// Store 3 entries — all recent
	fc.Store("project-1", data)
	fc.Store("project-2", data)
	fc.Store("project-3", data)

	// Advance counter so the next Store triggers cleanup
	fc.storeCount.Store(int64(cleanupInterval - 1))

	// Trigger cleanup
	fc.Store("project-4", data)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 4 {
		t.Errorf("expected 4 files (all recent), got %d", len(entries))
	}
}

func TestFileCacheCleanupProbabilistic(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}

	// Store an entry and backdate it to be old enough for cleanup
	fc.Store("old-project", data)
	entries, _ := os.ReadDir(dir)
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	for _, e := range entries {
		_ = os.Chtimes(filepath.Join(dir, e.Name()), oldTime, oldTime)
	}

	// Store 8 more times — none should trigger cleanup (counter starts at 1, needs 10)
	for i := 0; i < 8; i++ {
		fc.Store("new-project", data)
	}

	// Old file should still exist (cleanup hasn't run yet)
	got := fc.Get("old-project")
	if got == nil {
		t.Error("expected old file to survive — cleanup should not have run yet")
	}

	// The 10th Store (total) should trigger cleanup
	fc.Store("tenth-store", data)

	// Old file should now be cleaned up
	got = fc.Get("old-project")
	if got != nil {
		t.Error("expected old file to be cleaned up after 10th Store call")
	}
}

func TestFileCacheConcurrentStore_NoCorruption(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	data := &UsageData{BlockPercentage: 77.0, FetchedAt: time.Now()}

	// 10 goroutines all write to the same key concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fc.Store("shared-key", data)
		}()
	}
	wg.Wait()

	// After all writes, Get must return valid (non-nil) data
	got := fc.Get("shared-key")
	if got == nil {
		t.Fatal("concurrent Store corrupted the cache file: Get returned nil")
	}
	if got.BlockPercentage != 77.0 {
		t.Errorf("expected block 77.0, got %f", got.BlockPercentage)
	}
}

func TestTryLock_Acquire(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	ok, release := fc.TryLock("mykey")
	if !ok {
		t.Fatal("expected to acquire lock, got false")
	}
	if release == nil {
		t.Fatal("expected non-nil release function")
	}
	release()

	// After release, should be acquirable again
	ok2, release2 := fc.TryLock("mykey")
	if !ok2 {
		t.Fatal("expected to re-acquire lock after release")
	}
	release2()
}

func TestTryLock_AlreadyLocked(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	ok, release := fc.TryLock("mykey")
	if !ok {
		t.Fatal("expected first acquire to succeed")
	}
	defer release()

	// Second acquire must fail while first is held
	ok2, release2 := fc.TryLock("mykey")
	if ok2 {
		release2()
		t.Fatal("expected second acquire to fail while lock is held")
	}
	if release2 != nil {
		t.Error("expected nil release function when lock not acquired")
	}
}

func TestWaitForUnlock_Released(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	ok, release := fc.TryLock("mykey")
	if !ok {
		t.Fatal("expected to acquire lock")
	}

	// Release the lock after a short delay
	go func() {
		time.Sleep(80 * time.Millisecond)
		release()
	}()

	done := fc.WaitForUnlock("mykey", 500*time.Millisecond)
	if !done {
		t.Error("expected WaitForUnlock to return true (lock was released)")
	}
}

func TestWaitForUnlock_Timeout(t *testing.T) {
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	ok, release := fc.TryLock("mykey")
	if !ok {
		t.Fatal("expected to acquire lock")
	}
	defer release()

	// Lock is never released — should timeout
	done := fc.WaitForUnlock("mykey", 120*time.Millisecond)
	if done {
		t.Error("expected WaitForUnlock to return false (timeout)")
	}
}

func TestAtomicWrite_ReadOnlyDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod semantics differ on Windows")
	}
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	// Store data successfully first, then make dir read-only.
	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}
	fc.Store("before-readonly", data)

	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatalf("failed to chmod dir: %v", err)
	}
	defer func() { _ = os.Chmod(dir, 0o700) }() // restore so TempDir cleanup works

	// Store should silently fail (graceful degradation).
	fc.Store("after-readonly", data)

	// The new key should not be retrievable.
	got := fc.Get("after-readonly")
	if got != nil {
		t.Errorf("expected nil from read-only dir store, got %+v", got)
	}
}

func TestAtomicWrite_TempFileCleanedUpOnError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod semantics differ on Windows")
	}
	dir := t.TempDir()
	fc := NewFileCache(dir, 1*time.Minute)

	// Store some data so the cache dir exists, then make destination unwritable
	// by removing write permission from the target file path's directory won't work
	// for rename. Instead, create a subdirectory as the rename target to force error.
	data := &UsageData{BlockPercentage: 50.0, FetchedAt: time.Now()}

	// Create a directory at the target file path to make Rename fail.
	targetPath := fc.keyPath("conflict-key")
	if err := os.MkdirAll(targetPath, 0o700); err != nil {
		t.Fatalf("failed to create conflict dir: %v", err)
	}

	// Store should fail (rename onto a directory fails) but not leave .tmp-* files.
	fc.Store("conflict-key", data)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp-") {
			t.Errorf("temp file %q was not cleaned up after error", e.Name())
		}
	}
}

func TestFileCachePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	// Instance 1: Store data
	fc1 := NewFileCache(dir, 1*time.Minute)
	data := &UsageData{BlockPercentage: 42.0, FetchedAt: time.Now()}
	fc1.Store("my-project", data)

	// Instance 2: Read data (simulating a new process invocation)
	fc2 := NewFileCache(dir, 1*time.Minute)
	got := fc2.Get("my-project")
	if got == nil {
		t.Fatal("expected data from second instance, got nil")
	}
	if got.BlockPercentage != 42.0 {
		t.Errorf("expected block 42.0, got %f", got.BlockPercentage)
	}
}
