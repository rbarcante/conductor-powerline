package segments

import (
	"strings"
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestConductorInstalled(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(true, true, theme)

	if !seg.Enabled {
		t.Error("expected segment enabled when plugin is detected")
	}
	if seg.Name != "conductor" {
		t.Errorf("expected name 'conductor', got %q", seg.Name)
	}
	if !strings.Contains(seg.Text, "Conductor") {
		t.Errorf("expected text to contain 'Conductor', got %q", seg.Text)
	}
	// Should not contain OSC 8 hyperlink when installed
	if strings.Contains(seg.Text, "\033]8;;") {
		t.Error("expected no OSC 8 hyperlink when plugin is installed")
	}
	// Should use conductor (success) colors
	colors := theme.Segments["conductor"]
	if seg.FG != colors.FG {
		t.Errorf("expected FG %q, got %q", colors.FG, seg.FG)
	}
	if seg.BG != colors.BG {
		t.Errorf("expected BG %q, got %q", colors.BG, seg.BG)
	}
}

func TestConductorNotInstalled(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(false, true, theme)

	if !seg.Enabled {
		t.Error("expected segment enabled when plugin is NOT detected (shows prompt)")
	}
	if seg.Name != "conductor" {
		t.Errorf("expected name 'conductor', got %q", seg.Name)
	}
	if !strings.Contains(seg.Text, "Conductor") {
		t.Errorf("expected text to contain 'Conductor', got %q", seg.Text)
	}
	// Should contain OSC 8 hyperlink when not installed
	if !strings.Contains(seg.Text, "\033]8;;https://github.com/rbarcante/claude-conductor") {
		t.Errorf("expected OSC 8 hyperlink in text, got %q", seg.Text)
	}
	// Should use conductor_missing (warning) colors
	colors := theme.Segments["conductor_missing"]
	if seg.FG != colors.FG {
		t.Errorf("expected FG %q, got %q", colors.FG, seg.FG)
	}
	if seg.BG != colors.BG {
		t.Errorf("expected BG %q, got %q", colors.BG, seg.BG)
	}
}

func TestConductorInstalledNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(true, true, theme)

	// With Nerd Fonts, should use checkmark icon
	if !strings.Contains(seg.Text, "✓") {
		t.Errorf("expected checkmark icon with nerd fonts, got %q", seg.Text)
	}
}

func TestConductorInstalledNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(true, false, theme)

	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	// Without Nerd Fonts, should still show Conductor text
	if !strings.Contains(seg.Text, "Conductor") {
		t.Errorf("expected text to contain 'Conductor', got %q", seg.Text)
	}
}

func TestConductorMissingNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(false, true, theme)

	// With Nerd Fonts, should use lightning bolt icon
	if !strings.Contains(seg.Text, "⚡") {
		t.Errorf("expected lightning icon with nerd fonts, got %q", seg.Text)
	}
}

func TestConductorMissingNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(false, false, theme)

	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	// Without Nerd Fonts, should still show prompt text with hyperlink
	if !strings.Contains(seg.Text, "\033]8;;https://github.com/rbarcante/claude-conductor") {
		t.Errorf("expected OSC 8 hyperlink in text, got %q", seg.Text)
	}
}

func TestConductorThemeColorsAllThemes(t *testing.T) {
	themeNames := []string{"dark", "light", "nord", "gruvbox", "tokyo-night", "rose-pine"}
	for _, name := range themeNames {
		t.Run(name, func(t *testing.T) {
			theme, ok := themes.Get(name)
			if !ok {
				t.Fatalf("theme %q not found", name)
			}

			// Installed state uses conductor colors
			seg := Conductor(true, true, theme)
			colors := theme.Segments["conductor"]
			if seg.FG != colors.FG || seg.BG != colors.BG {
				t.Errorf("installed: wrong colors for theme %q: got FG=%q BG=%q, want FG=%q BG=%q",
					name, seg.FG, seg.BG, colors.FG, colors.BG)
			}

			// Missing state uses conductor_missing colors
			seg = Conductor(false, true, theme)
			colors = theme.Segments["conductor_missing"]
			if seg.FG != colors.FG || seg.BG != colors.BG {
				t.Errorf("missing: wrong colors for theme %q: got FG=%q BG=%q, want FG=%q BG=%q",
					name, seg.FG, seg.BG, colors.FG, colors.BG)
			}
		})
	}
}
