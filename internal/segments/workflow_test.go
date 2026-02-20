package segments

import (
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// testWorkflowData returns a WorkflowData fixture for testing.
func testWorkflowData() *WorkflowData {
	return &WorkflowData{
		Setup: WorkflowSetupInfo{IsValid: true, SetupComplete: true},
		Tracks: WorkflowTracksInfo{
			Tracks: []WorkflowTrackInfo{
				{
					Description: "Active feature",
					Status:      "in_progress",
					TrackID:     "active-feature_20260220",
					UpdatedAt:   "2026-02-20T10:00:00Z",
					Tasks:       WorkflowTaskSum{Completed: 5, InProgress: 1, Total: 10},
				},
				{
					Description: "Done track",
					Status:      "completed",
					TrackID:     "done-track_20260219",
					UpdatedAt:   "2026-02-19T08:00:00Z",
					Tasks:       WorkflowTaskSum{Completed: 8, Total: 8},
				},
			},
		},
	}
}

// --- WorkflowSetup ---

func TestWorkflowSetupComplete(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := testWorkflowData()
	seg := WorkflowSetup(data, theme)

	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	if seg.Name != "workflow_setup" {
		t.Errorf("expected name 'workflow_setup', got %q", seg.Name)
	}
	if seg.Text != "Setup 100%" {
		t.Errorf("expected 'Setup 100%%', got %q", seg.Text)
	}
	colors := theme.Segments["workflow_setup"]
	if seg.FG != colors.FG || seg.BG != colors.BG {
		t.Errorf("wrong colors: FG=%q BG=%q", seg.FG, seg.BG)
	}
}

func TestWorkflowSetupIncomplete(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := &WorkflowData{Setup: WorkflowSetupInfo{IsValid: true, SetupComplete: false}}
	seg := WorkflowSetup(data, theme)

	if seg.Text != "Setup --" {
		t.Errorf("expected 'Setup --', got %q", seg.Text)
	}
}

func TestWorkflowSetupNilData(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := WorkflowSetup(nil, theme)

	if !seg.Enabled {
		t.Error("expected segment enabled even with nil data")
	}
	if seg.Text != "Setup --" {
		t.Errorf("expected 'Setup --' for nil data, got %q", seg.Text)
	}
}

// --- WorkflowTrack ---

func TestWorkflowTrackWithActiveTrack(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := testWorkflowData()
	seg := WorkflowTrack(data, theme)

	if seg.Name != "workflow_track" {
		t.Errorf("expected name 'workflow_track', got %q", seg.Name)
	}
	if seg.Text != "Active feature" {
		t.Errorf("expected 'Active feature', got %q", seg.Text)
	}
}

func TestWorkflowTrackFallsBackToTrackID(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := &WorkflowData{
		Tracks: WorkflowTracksInfo{
			Tracks: []WorkflowTrackInfo{
				{Description: "", Status: "in_progress", TrackID: "my-track_20260220"},
			},
		},
	}
	seg := WorkflowTrack(data, theme)

	if seg.Text != "my-track_20260220" {
		t.Errorf("expected TrackID fallback, got %q", seg.Text)
	}
}

func TestWorkflowTrackNilData(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := WorkflowTrack(nil, theme)

	if seg.Text != "--" {
		t.Errorf("expected '--' for nil data, got %q", seg.Text)
	}
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
}

func TestWorkflowTrackNoTracks(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := &WorkflowData{Tracks: WorkflowTracksInfo{Tracks: []WorkflowTrackInfo{}}}
	seg := WorkflowTrack(data, theme)

	if seg.Text != "--" {
		t.Errorf("expected '--' for empty tracks, got %q", seg.Text)
	}
}

// --- WorkflowTasks ---

func TestWorkflowTasksWithActiveTrack(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := testWorkflowData()
	seg := WorkflowTasks(data, true, theme)

	if seg.Name != "workflow_tasks" {
		t.Errorf("expected name 'workflow_tasks', got %q", seg.Name)
	}
	if seg.Text != "5/10" {
		t.Errorf("expected '5/10', got %q", seg.Text)
	}
}

func TestWorkflowTasksNilData(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := WorkflowTasks(nil, true, theme)

	if seg.Text != "--" {
		t.Errorf("expected '--' for nil data, got %q", seg.Text)
	}
}

func TestWorkflowTasksNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := testWorkflowData()
	seg := WorkflowTasks(data, false, theme)
	// nerdFonts flag doesn't change the text for tasks, just verify it doesn't panic
	if seg.Text != "5/10" {
		t.Errorf("expected '5/10', got %q", seg.Text)
	}
}

// --- WorkflowOverall ---

func TestWorkflowOverallWithTracks(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := testWorkflowData()
	seg := WorkflowOverall(data, true, theme)

	if seg.Name != "workflow_overall" {
		t.Errorf("expected name 'workflow_overall', got %q", seg.Name)
	}
	// 1 completed out of 2 total
	if seg.Text != "1/2 tracks" {
		t.Errorf("expected '1/2 tracks', got %q", seg.Text)
	}
}

func TestWorkflowOverallAllCompleted(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := &WorkflowData{
		Tracks: WorkflowTracksInfo{
			Tracks: []WorkflowTrackInfo{
				{Status: "completed"},
				{Status: "completed"},
				{Status: "completed"},
			},
		},
	}
	seg := WorkflowOverall(data, true, theme)

	if seg.Text != "3/3 tracks" {
		t.Errorf("expected '3/3 tracks', got %q", seg.Text)
	}
}

func TestWorkflowOverallNilData(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := WorkflowOverall(nil, true, theme)

	if seg.Text != "--" {
		t.Errorf("expected '--' for nil data, got %q", seg.Text)
	}
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
}

func TestWorkflowOverallEmptyTracks(t *testing.T) {
	theme, _ := themes.Get("dark")
	data := &WorkflowData{Tracks: WorkflowTracksInfo{Tracks: []WorkflowTrackInfo{}}}
	seg := WorkflowOverall(data, true, theme)

	if seg.Text != "--" {
		t.Errorf("expected '--' for empty tracks, got %q", seg.Text)
	}
}

// --- selectActiveTrack ---

func TestSelectActiveTrackPrefersInProgress(t *testing.T) {
	data := testWorkflowData()
	track := selectActiveTrack(data)
	if track == nil {
		t.Fatal("expected non-nil track")
	}
	if track.Status != "in_progress" {
		t.Errorf("expected in_progress track, got %q", track.Status)
	}
	if track.TrackID != "active-feature_20260220" {
		t.Errorf("expected active-feature_20260220, got %q", track.TrackID)
	}
}

func TestSelectActiveTrackFallsBackToMostRecent(t *testing.T) {
	data := &WorkflowData{
		Tracks: WorkflowTracksInfo{
			Tracks: []WorkflowTrackInfo{
				{TrackID: "older", Status: "completed", UpdatedAt: "2026-02-18T00:00:00Z"},
				{TrackID: "newer", Status: "completed", UpdatedAt: "2026-02-20T00:00:00Z"},
			},
		},
	}
	track := selectActiveTrack(data)
	if track == nil {
		t.Fatal("expected non-nil track")
	}
	if track.TrackID != "newer" {
		t.Errorf("expected 'newer' track, got %q", track.TrackID)
	}
}

func TestSelectActiveTrackNilData(t *testing.T) {
	track := selectActiveTrack(nil)
	if track != nil {
		t.Error("expected nil for nil data")
	}
}

func TestSelectActiveTrackEmptyTracks(t *testing.T) {
	data := &WorkflowData{Tracks: WorkflowTracksInfo{Tracks: []WorkflowTrackInfo{}}}
	track := selectActiveTrack(data)
	if track != nil {
		t.Error("expected nil for empty tracks")
	}
}

// --- Theme coverage ---

func TestWorkflowSegmentsAllThemes(t *testing.T) {
	themeNames := []string{"dark", "light", "nord", "gruvbox", "tokyo-night", "rose-pine"}
	data := testWorkflowData()

	for _, name := range themeNames {
		t.Run(name, func(t *testing.T) {
			theme, ok := themes.Get(name)
			if !ok {
				t.Fatalf("theme %q not found", name)
			}

			// Each segment function must not panic and must return an enabled segment
			segs := []Segment{
				WorkflowSetup(data, theme),
				WorkflowTrack(data, theme),
				WorkflowTasks(data, true, theme),
				WorkflowOverall(data, true, theme),
			}
			names := []string{"workflow_setup", "workflow_track", "workflow_tasks", "workflow_overall"}
			for i, seg := range segs {
				if !seg.Enabled {
					t.Errorf("%s: expected enabled", names[i])
				}
				if seg.FG == "" || seg.BG == "" {
					t.Errorf("%s: expected non-empty colors, got FG=%q BG=%q", names[i], seg.FG, seg.BG)
				}
			}
		})
	}
}
