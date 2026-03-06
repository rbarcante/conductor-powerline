package segments

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// sampleStatusJSON is a realistic conductor_cli.py --json status response.
const sampleStatusJSON = `{
  "success": true,
  "data": {
    "setup": {
      "is_valid": true,
      "setup_complete": true,
      "last_setup_step": "3.3_initial_track_generated",
      "missing_required": []
    },
    "tracks": {
      "tracks": [
        {
          "description": "My active track",
          "status": "in_progress",
          "track_id": "my-active-track_20260220",
          "updated_at": "2026-02-20T10:00:00Z",
          "tasks": {"completed": 5, "in_progress": 1, "total": 10}
        },
        {
          "description": "Old completed track",
          "status": "completed",
          "track_id": "old-track_20260219",
          "updated_at": "2026-02-19T08:00:00Z",
          "tasks": {"completed": 8, "in_progress": 0, "total": 8}
        }
      ]
    }
  }
}`

// withMockExec replaces execCommandFunc for the duration of a test.
func withMockExec(t *testing.T, jsonOutput string) {
	t.Helper()
	orig := execCommandFunc
	t.Cleanup(func() { execCommandFunc = orig })

	tmpFile := filepath.Join(t.TempDir(), "output.json")
	if err := os.WriteFile(tmpFile, []byte(jsonOutput), 0644); err != nil {
		t.Fatalf("failed to write mock output: %v", err)
	}
	execCommandFunc = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "cat", tmpFile)
	}
}

func TestFetchWorkflowStatusSuccess(t *testing.T) {
	withMockExec(t, sampleStatusJSON)

	// Create a fake home dir with the conductor_cli.py path so FindConductorCLI succeeds
	fakeHome := makeFakeCLI(t)

	ctx := context.Background()
	data, err := FetchWorkflowStatus(ctx, fakeHome, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil WorkflowData")
	}
	if !data.Setup.SetupComplete {
		t.Error("expected setup_complete true")
	}
	if !data.Setup.IsValid {
		t.Error("expected is_valid true")
	}
	if len(data.Tracks.Tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(data.Tracks.Tracks))
	}
	if data.Tracks.Tracks[0].Status != "in_progress" {
		t.Errorf("expected first track in_progress, got %q", data.Tracks.Tracks[0].Status)
	}
	if data.Tracks.Tracks[0].Tasks.Completed != 5 {
		t.Errorf("expected completed=5, got %d", data.Tracks.Tracks[0].Tasks.Completed)
	}
}

func TestFetchWorkflowStatusCLINotFound(t *testing.T) {
	ctx := context.Background()
	// Use a non-existent home dir so FindConductorCLI returns ""
	_, err := FetchWorkflowStatus(ctx, t.TempDir(), t.TempDir())
	if err == nil {
		t.Error("expected error when CLI not found")
	}
}

func TestFetchWorkflowStatusCLIFailure(t *testing.T) {
	fakeHome := makeFakeCLI(t)
	orig := execCommandFunc
	t.Cleanup(func() { execCommandFunc = orig })
	execCommandFunc = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// Command that exits non-zero
		return exec.CommandContext(ctx, "false")
	}

	ctx := context.Background()
	_, err := FetchWorkflowStatus(ctx, fakeHome, t.TempDir())
	if err == nil {
		t.Error("expected error when CLI exits non-zero")
	}
}

func TestFetchWorkflowStatusTimeout(t *testing.T) {
	fakeHome := makeFakeCLI(t)
	orig := execCommandFunc
	t.Cleanup(func() { execCommandFunc = orig })
	execCommandFunc = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// Command that hangs longer than the context
		return exec.CommandContext(ctx, "sleep", "10")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := FetchWorkflowStatus(ctx, fakeHome, t.TempDir())
	if err == nil {
		t.Error("expected error on timeout")
	}
}

func TestFetchWorkflowStatusMalformedJSON(t *testing.T) {
	withMockExec(t, `{invalid json`)
	fakeHome := makeFakeCLI(t)

	ctx := context.Background()
	_, err := FetchWorkflowStatus(ctx, fakeHome, t.TempDir())
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestFetchWorkflowStatusSuccessFalse(t *testing.T) {
	withMockExec(t, `{"success":false,"data":{}}`)
	fakeHome := makeFakeCLI(t)

	ctx := context.Background()
	_, err := FetchWorkflowStatus(ctx, fakeHome, t.TempDir())
	if err == nil {
		t.Error("expected error when success=false")
	}
}

func TestFindConductorCLINotFound(t *testing.T) {
	result := FindConductorCLI(t.TempDir())
	if result != "" {
		t.Errorf("expected empty string for non-existent CLI, got %q", result)
	}
}

func TestFindConductorCLIFound(t *testing.T) {
	fakeHome := makeFakeCLI(t)
	result := FindConductorCLI(fakeHome)
	if result == "" {
		t.Error("expected to find conductor_cli.py")
	}
	if filepath.Base(result) != "conductor_cli.py" {
		t.Errorf("expected conductor_cli.py, got %q", filepath.Base(result))
	}
}

// makeFakeCLI creates a fake conductor_cli.py in a temp home dir and returns the home path.
func makeFakeCLI(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	cliDir := filepath.Join(home, ".claude", "plugins", "cache", "claude-conductor", "conductor", "1.0.0", "scripts")
	if err := os.MkdirAll(cliDir, 0755); err != nil {
		t.Fatalf("failed to create CLI dir: %v", err)
	}
	cliPath := filepath.Join(cliDir, "conductor_cli.py")
	if err := os.WriteFile(cliPath, []byte("#!/usr/bin/env python3\nprint('{}')"), 0755); err != nil {
		t.Fatalf("failed to create fake CLI: %v", err)
	}
	return home
}
