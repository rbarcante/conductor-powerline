package render

import (
	"strings"
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/segments"
)

func TestRenderWithNerdFonts(t *testing.T) {
	segs := []segments.Segment{
		{Name: "directory", Text: "my-project", FG: "15", BG: "236", Enabled: true},
		{Name: "git", Text: "\ue0a0 main", FG: "15", BG: "22", Enabled: true},
	}

	out := Render(segs, true, 120)

	if strings.HasSuffix(out, "\n") {
		t.Error("output must not have trailing newline")
	}
	if !strings.Contains(out, "my-project") {
		t.Error("expected 'my-project' in output")
	}
	if !strings.Contains(out, "main") {
		t.Error("expected 'main' in output")
	}
	// Should contain ANSI escape sequences
	if !strings.Contains(out, "\033[") {
		t.Error("expected ANSI escape codes in output")
	}
	// Should contain nerd font separator
	if !strings.Contains(out, SeparatorNerd) {
		t.Error("expected nerd font separator in output")
	}
}

func TestRenderWithTextFallback(t *testing.T) {
	segs := []segments.Segment{
		{Name: "directory", Text: "my-project", FG: "15", BG: "236", Enabled: true},
		{Name: "git", Text: "main", FG: "15", BG: "22", Enabled: true},
	}

	out := Render(segs, false, 120)

	if strings.Contains(out, SeparatorNerd) {
		t.Error("should not contain nerd font separator in text mode")
	}
	if !strings.Contains(out, SeparatorText) {
		t.Error("expected text separator in output")
	}
}

func TestRenderEmptySegments(t *testing.T) {
	out := Render(nil, true, 120)
	if out != "" {
		t.Errorf("expected empty output for nil segments, got %q", out)
	}

	out = Render([]segments.Segment{}, true, 120)
	if out != "" {
		t.Errorf("expected empty output for empty segments, got %q", out)
	}
}

func TestRenderDisabledSegments(t *testing.T) {
	segs := []segments.Segment{
		{Name: "directory", Text: "my-project", FG: "15", BG: "236", Enabled: true},
		{Name: "git", Text: "main", FG: "15", BG: "22", Enabled: false},
		{Name: "model", Text: "Opus 4.6", FG: "15", BG: "57", Enabled: true},
	}

	out := Render(segs, true, 120)

	if strings.Contains(out, "main") {
		t.Error("disabled segment 'git' should not appear in output")
	}
	if !strings.Contains(out, "my-project") {
		t.Error("expected 'my-project' in output")
	}
	if !strings.Contains(out, "Opus 4.6") {
		t.Error("expected 'Opus 4.6' in output")
	}
}

func TestRenderCompactMode(t *testing.T) {
	segs := []segments.Segment{
		{Name: "directory", Text: "a-very-long-project-name-here", FG: "15", BG: "236", Enabled: true},
		{Name: "model", Text: "Opus 4.6", FG: "15", BG: "57", Enabled: true},
	}

	// Render at narrow width — compact mode should truncate
	out := Render(segs, true, 30)

	if strings.Contains(out, "a-very-long-project-name-here") {
		t.Error("expected text to be truncated in compact mode")
	}
}

func TestRenderNoTrailingNewline(t *testing.T) {
	segs := []segments.Segment{
		{Name: "directory", Text: "test", FG: "15", BG: "236", Enabled: true},
	}

	out := Render(segs, true, 120)

	if strings.HasSuffix(out, "\n") {
		t.Error("output must not have trailing newline")
	}
}

// --- Tests for RenderRight (left-pointing arrow separators) ---

func TestRenderRightSingleSegment(t *testing.T) {
	segs := []segments.Segment{
		{Name: "context", Text: "○ 30%", FG: "231", BG: "36", Enabled: true},
	}

	out := RenderRight(segs, true)
	if !strings.Contains(out, "○ 30%") {
		t.Error("expected '○ 30%' in output")
	}
	if !strings.Contains(out, SeparatorLeftNerd) {
		t.Error("expected left-pointing nerd font separator")
	}
	if !strings.Contains(out, "\033[") {
		t.Error("expected ANSI escape codes")
	}
}

func TestRenderRightEmpty(t *testing.T) {
	out := RenderRight(nil, true)
	if out != "" {
		t.Errorf("expected empty output for nil segments, got %q", out)
	}

	out = RenderRight([]segments.Segment{}, true)
	if out != "" {
		t.Errorf("expected empty output for empty segments, got %q", out)
	}
}

func TestRenderRightDisabledSkipped(t *testing.T) {
	segs := []segments.Segment{
		{Name: "context", Text: "○ 30%", FG: "231", BG: "36", Enabled: false},
	}

	out := RenderRight(segs, true)
	if out != "" {
		t.Errorf("expected empty output for disabled segments, got %q", out)
	}
}

func TestRenderRightTextFallback(t *testing.T) {
	segs := []segments.Segment{
		{Name: "context", Text: "CTX 30%", FG: "231", BG: "36", Enabled: true},
	}

	out := RenderRight(segs, false)
	if !strings.Contains(out, "CTX 30%") {
		t.Error("expected 'CTX 30%' in output")
	}
	if strings.Contains(out, SeparatorLeftNerd) {
		t.Error("should not contain nerd font separator in text mode")
	}
	// Single segment has no separator; ANSI codes still present
	if !strings.Contains(out, "\033[") {
		t.Error("expected ANSI escape codes")
	}
}

func TestLeftArrowSymbolsDefined(t *testing.T) {
	if SeparatorLeftNerd == "" {
		t.Error("SeparatorLeftNerd must be defined")
	}
	if SeparatorLeftText == "" {
		t.Error("SeparatorLeftText must be defined")
	}
}

func TestRenderSegmentOrder(t *testing.T) {
	segs := []segments.Segment{
		{Name: "model", Text: "Opus", FG: "15", BG: "57", Enabled: true},
		{Name: "directory", Text: "proj", FG: "15", BG: "236", Enabled: true},
	}

	out := Render(segs, false, 120)

	modelIdx := strings.Index(out, "Opus")
	dirIdx := strings.Index(out, "proj")

	if modelIdx > dirIdx {
		t.Error("expected segments rendered in order: model before directory")
	}
}
