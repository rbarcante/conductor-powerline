package segments

import (
	"fmt"

	"github.com/rbarcante/conductor-powerline/internal/debug"
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// WorkflowSetup returns a segment showing conductor setup completion status.
// Displays "Setup 100%" when setup is complete and valid, otherwise "Setup --".
func WorkflowSetup(data *WorkflowData, theme themes.Theme) Segment {
	colors := theme.Segments["workflow_setup"]
	text := "Setup --"
	if data != nil && data.Setup.SetupComplete && data.Setup.IsValid {
		text = "Setup 100%"
	}
	debug.Logf("workflow", "setup segment: %q", text)
	return Segment{
		Name:    "workflow_setup",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

// WorkflowTrack returns a segment showing the active track name or ID.
// Active track is the first in_progress track, or the most recently updated track.
func WorkflowTrack(data *WorkflowData, theme themes.Theme) Segment {
	colors := theme.Segments["workflow_track"]
	text := "--"

	if data != nil {
		track := selectActiveTrack(data)
		if track != nil {
			text = track.Description
			if text == "" {
				text = track.TrackID
			}
		}
	}

	debug.Logf("workflow", "track segment: %q", text)
	return Segment{
		Name:    "workflow_track",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

// WorkflowTasks returns a segment showing completed/total task counts for the active track.
func WorkflowTasks(data *WorkflowData, nerdFonts bool, theme themes.Theme) Segment {
	colors := theme.Segments["workflow_tasks"]
	text := "--"

	if data != nil {
		track := selectActiveTrack(data)
		if track != nil {
			text = fmt.Sprintf("%d/%d", track.Tasks.Completed, track.Tasks.Total)
		}
	}

	debug.Logf("workflow", "tasks segment: %q", text)
	return Segment{
		Name:    "workflow_tasks",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

// WorkflowOverall returns a segment showing the count of completed vs total tracks.
func WorkflowOverall(data *WorkflowData, nerdFonts bool, theme themes.Theme) Segment {
	colors := theme.Segments["workflow_overall"]
	text := "--"

	if data != nil && len(data.Tracks.Tracks) > 0 {
		completed := 0
		for _, t := range data.Tracks.Tracks {
			if t.Status == "completed" {
				completed++
			}
		}
		text = fmt.Sprintf("%d/%d tracks", completed, len(data.Tracks.Tracks))
	}

	debug.Logf("workflow", "overall segment: %q", text)
	return Segment{
		Name:    "workflow_overall",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

// selectActiveTrack returns the track to highlight:
// first in_progress track, or the most recently updated track as fallback.
func selectActiveTrack(data *WorkflowData) *WorkflowTrackInfo {
	if data == nil || len(data.Tracks.Tracks) == 0 {
		return nil
	}

	// Prefer the first in_progress track
	for i := range data.Tracks.Tracks {
		if data.Tracks.Tracks[i].Status == "in_progress" {
			debug.Logf("workflow", "active track (in_progress): %q", data.Tracks.Tracks[i].TrackID)
			return &data.Tracks.Tracks[i]
		}
	}

	// Fallback: track with the most recent UpdatedAt (lexicographic — ISO 8601 sorts correctly)
	var best *WorkflowTrackInfo
	for i := range data.Tracks.Tracks {
		t := &data.Tracks.Tracks[i]
		if best == nil || t.UpdatedAt > best.UpdatedAt {
			best = t
		}
	}

	if best != nil {
		debug.Logf("workflow", "active track (most recent): %q", best.TrackID)
	}
	return best
}
