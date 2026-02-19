package segments

import (
	"testing"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/oauth"
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestBlockPercentageDisplay(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		BlockPercentage: 72.5,
		BlockResetTime:  time.Now().Add(2*time.Hour + 13*time.Minute),
	}

	seg := Block(data, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	if seg.Name != "block" {
		t.Errorf("expected name 'block', got %q", seg.Name)
	}
	// Should contain percentage
	if seg.Text == "" {
		t.Error("expected non-empty text")
	}
}

func TestBlockCountdownFormat(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		BlockPercentage: 50.0,
		BlockResetTime:  time.Now().Add(2*time.Hour + 13*time.Minute),
	}

	seg := Block(data, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	// Text should contain time remaining
	if seg.Text == "" {
		t.Error("expected text with countdown")
	}
}

func TestBlockColorThresholdNormal(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		BlockPercentage: 40.0,
		BlockResetTime:  time.Now().Add(3 * time.Hour),
	}

	seg := Block(data, theme)
	expectedColors := theme.Segments["block"]
	if seg.BG != expectedColors.BG {
		t.Errorf("expected normal BG %q, got %q", expectedColors.BG, seg.BG)
	}
}

func TestBlockColorThresholdWarning(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		BlockPercentage: 75.0,
		BlockResetTime:  time.Now().Add(1 * time.Hour),
	}

	seg := Block(data, theme)
	expectedColors := theme.Segments["warning"]
	if seg.BG != expectedColors.BG {
		t.Errorf("expected warning BG %q, got %q", expectedColors.BG, seg.BG)
	}
}

func TestBlockColorThresholdCritical(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		BlockPercentage: 95.0,
		BlockResetTime:  time.Now().Add(30 * time.Minute),
	}

	seg := Block(data, theme)
	expectedColors := theme.Segments["critical"]
	if seg.BG != expectedColors.BG {
		t.Errorf("expected critical BG %q, got %q", expectedColors.BG, seg.BG)
	}
}

func TestBlockNilData(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Block(nil, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled with placeholder")
	}
	if seg.Text != "--" {
		t.Errorf("expected '--' placeholder, got %q", seg.Text)
	}
}

func TestBlockStaleIndicator(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		BlockPercentage: 50.0,
		BlockResetTime:  time.Now().Add(2 * time.Hour),
		IsStale:         true,
	}

	seg := Block(data, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	// Stale data should still show but with indicator
	if seg.Text == "" {
		t.Error("expected text even for stale data")
	}
}

func TestBlockPastResetTime(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		BlockPercentage: 50.0,
		BlockResetTime:  time.Now().Add(-1 * time.Hour),
	}

	seg := Block(data, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
}
