package segments

import (
	"strings"
	"testing"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/oauth"
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestWeeklyPercentageDisplay(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 45.0,
		WeekResetTime:    time.Now().Add(72 * time.Hour),
	}

	seg := Weekly(data, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	if seg.Name != "weekly" {
		t.Errorf("expected name 'weekly', got %q", seg.Name)
	}
	if !strings.Contains(seg.Text, "45%") {
		t.Errorf("expected text to contain '45%%', got %q", seg.Text)
	}
}

func TestWeeklyOpusSonnetBreakdown(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 65.0,
		OpusPercentage:   45.0,
		SonnetPercentage: 20.0,
		WeekResetTime:    time.Now().Add(48 * time.Hour),
	}

	seg := Weekly(data, theme)
	// Should show breakdown when both are in use
	if !strings.Contains(seg.Text, "O:45%") {
		t.Errorf("expected Opus breakdown in text, got %q", seg.Text)
	}
	if !strings.Contains(seg.Text, "S:20%") {
		t.Errorf("expected Sonnet breakdown in text, got %q", seg.Text)
	}
}

func TestWeeklyNoBreakdownSingleModel(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 30.0,
		OpusPercentage:   30.0,
		SonnetPercentage: 0.0,
		WeekResetTime:    time.Now().Add(96 * time.Hour),
	}

	seg := Weekly(data, theme)
	// Should not show breakdown when only one model is used
	if strings.Contains(seg.Text, "S:") {
		t.Errorf("expected no Sonnet breakdown for single model, got %q", seg.Text)
	}
}

func TestWeeklyNilData(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Weekly(nil, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled with placeholder")
	}
	if seg.Text != "--" {
		t.Errorf("expected '--' placeholder, got %q", seg.Text)
	}
}

func TestWeeklyStaleData(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 50.0,
		WeekResetTime:    time.Now().Add(48 * time.Hour),
		IsStale:          true,
	}

	seg := Weekly(data, theme)
	if !strings.Contains(seg.Text, "~") {
		t.Errorf("expected stale indicator '~' in text, got %q", seg.Text)
	}
}

func TestWeeklyWeekProgress(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 50.0,
		WeekResetTime:    time.Now().Add(4*24*time.Hour + 12*time.Hour),
	}

	seg := Weekly(data, theme)
	// 4d12h from now truncates to 4 days
	if !strings.Contains(seg.Text, "4d") {
		t.Errorf("expected '4d' day indicator in text, got %q", seg.Text)
	}
}

func TestWeeklyColorThresholdNormal(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 40.0,
		WeekResetTime:    time.Now().Add(5 * 24 * time.Hour),
	}

	seg := Weekly(data, theme)
	expectedColors := theme.Segments["weekly"]
	if seg.BG != expectedColors.BG {
		t.Errorf("expected normal BG %q, got %q", expectedColors.BG, seg.BG)
	}
}

func TestWeeklyColorThresholdWarning(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 75.0,
		WeekResetTime:    time.Now().Add(3 * 24 * time.Hour),
	}

	seg := Weekly(data, theme)
	expectedColors := theme.Segments["warning"]
	if seg.BG != expectedColors.BG {
		t.Errorf("expected warning BG %q, got %q", expectedColors.BG, seg.BG)
	}
}

func TestWeeklyColorThresholdCritical(t *testing.T) {
	theme, _ := themes.Get("dark")

	data := &oauth.UsageData{
		WeeklyPercentage: 95.0,
		WeekResetTime:    time.Now().Add(1 * 24 * time.Hour),
	}

	seg := Weekly(data, theme)
	expectedColors := theme.Segments["critical"]
	if seg.BG != expectedColors.BG {
		t.Errorf("expected critical BG %q, got %q", expectedColors.BG, seg.BG)
	}
}
