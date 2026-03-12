package segments

import (
	"sync"
	"testing"
	"time"
)

func TestWorkflowFileCache_StoreAndGet(t *testing.T) {
	dir := t.TempDir()
	wc := NewWorkflowFileCache(dir, 1*time.Minute)

	data := &WorkflowData{
		Setup: WorkflowSetupInfo{IsValid: true, SetupComplete: true},
		Tracks: WorkflowTracksInfo{
			Tracks: []WorkflowTrackInfo{
				{TrackID: "track-a", Status: "in_progress", Description: "My Track"},
			},
		},
	}

	wc.Store("workspace-1", data)

	got := wc.Get("workspace-1")
	if got == nil {
		t.Fatal("expected cached WorkflowData, got nil")
	}
	if !got.Setup.SetupComplete {
		t.Error("expected SetupComplete true")
	}
	if len(got.Tracks.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(got.Tracks.Tracks))
	}
	if got.Tracks.Tracks[0].TrackID != "track-a" {
		t.Errorf("expected track-a, got %q", got.Tracks.Tracks[0].TrackID)
	}
	if got.IsStale {
		t.Error("expected fresh data")
	}
}

func TestWorkflowFileCache_GetMissReturnsNil(t *testing.T) {
	dir := t.TempDir()
	wc := NewWorkflowFileCache(dir, 1*time.Minute)

	got := wc.Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil for missing key, got %+v", got)
	}
}

func TestWorkflowFileCache_TTLExpiry(t *testing.T) {
	dir := t.TempDir()
	wc := NewWorkflowFileCache(dir, 50*time.Millisecond)

	data := &WorkflowData{Setup: WorkflowSetupInfo{IsValid: true}}
	wc.Store("ws", data)

	// Immediately fresh
	got := wc.Get("ws")
	if got == nil || got.IsStale {
		t.Error("expected fresh data immediately after store")
	}

	// After TTL
	time.Sleep(60 * time.Millisecond)
	got = wc.Get("ws")
	if got == nil {
		t.Fatal("expected stale data, got nil")
	}
	if !got.IsStale {
		t.Error("expected stale data after TTL expiry")
	}
}

func TestWorkflowFileCache_PersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()

	wc1 := NewWorkflowFileCache(dir, 1*time.Minute)
	data := &WorkflowData{Setup: WorkflowSetupInfo{SetupComplete: true}}
	wc1.Store("myws", data)

	wc2 := NewWorkflowFileCache(dir, 1*time.Minute)
	got := wc2.Get("myws")
	if got == nil {
		t.Fatal("expected data from second instance, got nil")
	}
	if !got.Setup.SetupComplete {
		t.Error("expected SetupComplete true from second instance")
	}
}

func TestWorkflowFileCache_ConcurrentStore_NoCorruption(t *testing.T) {
	dir := t.TempDir()
	wc := NewWorkflowFileCache(dir, 1*time.Minute)

	data := &WorkflowData{
		Setup: WorkflowSetupInfo{IsValid: true, SetupComplete: true},
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wc.Store("shared-key", data)
		}()
	}
	wg.Wait()

	got := wc.Get("shared-key")
	if got == nil {
		t.Fatal("concurrent Store corrupted the cache: Get returned nil")
	}
	if !got.Setup.SetupComplete {
		t.Error("expected SetupComplete true after concurrent writes")
	}
}

func TestWorkflowFileCache_UnwritableDir(t *testing.T) {
	wc := NewWorkflowFileCache("/nonexistent/path/that/does/not/exist", 1*time.Minute)

	data := &WorkflowData{Setup: WorkflowSetupInfo{IsValid: true}}
	// Should not panic
	wc.Store("key", data)

	// Should return nil gracefully
	got := wc.Get("key")
	if got != nil {
		t.Errorf("expected nil from unwritable cache, got %+v", got)
	}
}
