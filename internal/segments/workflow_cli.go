package segments

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/debug"
)

// WorkflowData holds the parsed output from conductor_cli.py --json status.
type WorkflowData struct {
	Setup  WorkflowSetupInfo  `json:"setup"`
	Tracks WorkflowTracksInfo `json:"tracks"`
}

// WorkflowSetupInfo reflects the setup validity and completion state.
type WorkflowSetupInfo struct {
	IsValid       bool `json:"is_valid"`
	SetupComplete bool `json:"setup_complete"`
}

// WorkflowTracksInfo holds the list of tracks.
type WorkflowTracksInfo struct {
	Tracks []WorkflowTrackInfo `json:"tracks"`
}

// WorkflowTrackInfo represents a single track entry from the CLI.
type WorkflowTrackInfo struct {
	Description string          `json:"description"`
	Status      string          `json:"status"`
	TrackID     string          `json:"track_id"`
	UpdatedAt   string          `json:"updated_at"`
	Tasks       WorkflowTaskSum `json:"tasks"`
}

// WorkflowTaskSum holds completed/total task counts for a track.
type WorkflowTaskSum struct {
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
	Total      int `json:"total"`
}

// cliStatusResponse is the top-level wrapper for conductor_cli.py --json status output.
type cliStatusResponse struct {
	Success bool         `json:"success"`
	Data    WorkflowData `json:"data"`
}

// execCommandFunc creates an exec.Cmd; overridable in tests.
var execCommandFunc = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// FindConductorCLI locates conductor_cli.py in the Claude plugin cache directory.
// homeDir is the user's home directory; pass "" to use os.UserHomeDir().
func FindConductorCLI(homeDir string) string {
	if homeDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		homeDir = home
	}
	pattern := filepath.Join(homeDir, ".claude", "plugins", "cache", "claude-conductor", "conductor", "*", "scripts", "conductor_cli.py")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		debug.Logf("workflow_cli", "conductor_cli.py not found at pattern %s", pattern)
		return ""
	}
	// Use the last match (highest version, since glob returns sorted paths)
	return matches[len(matches)-1]
}

// FetchWorkflowStatus executes conductor_cli.py --json status and returns parsed data.
// Returns nil, error if the CLI is not found, fails, times out, or returns malformed JSON.
func FetchWorkflowStatus(ctx context.Context, homeDir string) (*WorkflowData, error) {
	start := time.Now()

	cliPath := FindConductorCLI(homeDir)
	if cliPath == "" {
		return nil, fmt.Errorf("conductor_cli.py not found")
	}

	debug.Logf("workflow_cli", "executing: python3 %s --json status", cliPath)

	cmd := execCommandFunc(ctx, "python3", cliPath, "--json", "status")
	out, err := cmd.Output()
	duration := time.Since(start)
	if err != nil {
		debug.Logf("workflow_cli", "CLI execution failed after %v: %v", duration, err)
		return nil, fmt.Errorf("CLI execution failed: %w", err)
	}

	debug.Logf("workflow_cli", "CLI succeeded in %v, output=%d bytes", duration, len(out))

	var resp cliStatusResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		debug.Logf("workflow_cli", "JSON parse failed: %v", err)
		return nil, fmt.Errorf("JSON parse failed: %w", err)
	}

	if !resp.Success {
		debug.Logf("workflow_cli", "CLI returned success=false")
		return nil, fmt.Errorf("CLI returned success=false")
	}

	debug.Logf("workflow_cli", "parsed: setup_complete=%v tracks=%d",
		resp.Data.Setup.SetupComplete, len(resp.Data.Tracks.Tracks))
	return &resp.Data, nil
}
