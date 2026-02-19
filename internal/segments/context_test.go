package segments

import (
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestContextNormalRange(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(30, true, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	if seg.Name != "context" {
		t.Errorf("Name = %q, want %q", seg.Name, "context")
	}
	// Should use context (normal) colors
	expectedColors := theme.Segments["context"]
	if seg.BG != expectedColors.BG {
		t.Errorf("BG = %q, want %q (normal)", seg.BG, expectedColors.BG)
	}
	// Nerd font icon should be empty circle
	if seg.Text != "○ 30%" {
		t.Errorf("Text = %q, want %q", seg.Text, "○ 30%")
	}
}

func TestContextWarningRange(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(65, true, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	expectedColors := theme.Segments["context-warning"]
	if seg.BG != expectedColors.BG {
		t.Errorf("BG = %q, want %q (warning)", seg.BG, expectedColors.BG)
	}
	if seg.Text != "◐ 65%" {
		t.Errorf("Text = %q, want %q", seg.Text, "◐ 65%")
	}
}

func TestContextCriticalRange(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(85, true, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	expectedColors := theme.Segments["context-critical"]
	if seg.BG != expectedColors.BG {
		t.Errorf("BG = %q, want %q (critical)", seg.BG, expectedColors.BG)
	}
	if seg.Text != "● 85%" {
		t.Errorf("Text = %q, want %q", seg.Text, "● 85%")
	}
}

func TestContextBoundary50(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(50, true, theme)
	// 50% should be warning range
	expectedColors := theme.Segments["context-warning"]
	if seg.BG != expectedColors.BG {
		t.Errorf("BG = %q, want %q (warning at 50%%)", seg.BG, expectedColors.BG)
	}
	if seg.Text != "◐ 50%" {
		t.Errorf("Text = %q, want %q", seg.Text, "◐ 50%")
	}
}

func TestContextBoundary80(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(80, true, theme)
	// 80% should still be warning (> 80% is critical)
	expectedColors := theme.Segments["context-warning"]
	if seg.BG != expectedColors.BG {
		t.Errorf("BG = %q, want %q (warning at 80%%)", seg.BG, expectedColors.BG)
	}
}

func TestContextBoundary81(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(81, true, theme)
	// 81% should be critical
	expectedColors := theme.Segments["context-critical"]
	if seg.BG != expectedColors.BG {
		t.Errorf("BG = %q, want %q (critical at 81%%)", seg.BG, expectedColors.BG)
	}
}

func TestContextZeroPercent(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(0, true, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled at 0%")
	}
	if seg.Text != "○ 0%" {
		t.Errorf("Text = %q, want %q", seg.Text, "○ 0%")
	}
}

func TestContext100Percent(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(100, true, theme)
	if !seg.Enabled {
		t.Error("expected segment enabled at 100%")
	}
	expectedColors := theme.Segments["context-critical"]
	if seg.BG != expectedColors.BG {
		t.Errorf("BG = %q, want %q (critical at 100%%)", seg.BG, expectedColors.BG)
	}
	if seg.Text != "● 100%" {
		t.Errorf("Text = %q, want %q", seg.Text, "● 100%")
	}
}

func TestContextMissingData(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(-1, true, theme)
	if seg.Enabled {
		t.Error("expected segment disabled when percent is -1")
	}
}

func TestContextTextFallback(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(30, false, theme)
	if seg.Text != "CTX 30%" {
		t.Errorf("Text = %q, want %q (text fallback)", seg.Text, "CTX 30%")
	}
}

func TestContextTextFallbackWarning(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(65, false, theme)
	if seg.Text != "CTX 65%" {
		t.Errorf("Text = %q, want %q (text fallback warning)", seg.Text, "CTX 65%")
	}
}

func TestContextTextFallbackCritical(t *testing.T) {
	theme, _ := themes.Get("dark")

	seg := Context(90, false, theme)
	if seg.Text != "CTX 90%" {
		t.Errorf("Text = %q, want %q (text fallback critical)", seg.Text, "CTX 90%")
	}
}
