package oauth

import (
	"os"
	"path/filepath"
	"runtime"
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
			os.Chtimes(path, oldTime, oldTime)
		}
	}

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

	// Trigger cleanup
	fc.Store("project-4", data)

	entries, _ := os.ReadDir(dir)
	if len(entries) != 4 {
		t.Errorf("expected 4 files (all recent), got %d", len(entries))
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
