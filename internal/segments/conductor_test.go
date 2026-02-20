package segments

import (
	"strings"
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestConductorActive(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorActive, true, theme)

	if !seg.Enabled {
		t.Error("expected segment enabled")
	}
	if seg.Name != "conductor" {
		t.Errorf("expected name 'conductor', got %q", seg.Name)
	}
	if !strings.Contains(seg.Text, "✓") || !strings.Contains(seg.Text, "Conductor") {
		t.Errorf("expected '✓ Conductor', got %q", seg.Text)
	}
	if strings.Contains(seg.Text, "\033]8;;") {
		t.Error("expected no OSC 8 hyperlink for active state")
	}
	colors := theme.Segments["conductor"]
	if seg.FG != colors.FG || seg.BG != colors.BG {
		t.Errorf("expected conductor colors, got FG=%q BG=%q", seg.FG, seg.BG)
	}
}

func TestConductorActiveNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorActive, false, theme)

	if !strings.Contains(seg.Text, "OK Conductor") {
		t.Errorf("expected 'OK Conductor' without nerd fonts, got %q", seg.Text)
	}
}

func TestConductorInstalled(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorInstalled, true, theme)

	if !strings.Contains(seg.Text, "Setup Conductor") {
		t.Errorf("expected 'Setup Conductor', got %q", seg.Text)
	}
	if strings.Contains(seg.Text, "\033]8;;") {
		t.Error("expected no OSC 8 hyperlink for installed state")
	}
	colors := theme.Segments["conductor_missing"]
	if seg.FG != colors.FG || seg.BG != colors.BG {
		t.Errorf("expected conductor_missing colors, got FG=%q BG=%q", seg.FG, seg.BG)
	}
}

func TestConductorInstalledNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorInstalled, false, theme)

	if !strings.Contains(seg.Text, "Setup Conductor") {
		t.Errorf("expected 'Setup Conductor', got %q", seg.Text)
	}
	if strings.Contains(seg.Text, "⚡") {
		t.Error("expected no lightning icon without nerd fonts")
	}
}

func TestConductorMarketplace(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorMarketplace, true, theme)

	if !strings.Contains(seg.Text, "Install Conductor") {
		t.Errorf("expected 'Install Conductor', got %q", seg.Text)
	}
	if strings.Contains(seg.Text, "\033]8;;") {
		t.Error("expected no OSC 8 hyperlink for marketplace state")
	}
	colors := theme.Segments["conductor_missing"]
	if seg.FG != colors.FG || seg.BG != colors.BG {
		t.Errorf("expected conductor_missing colors, got FG=%q BG=%q", seg.FG, seg.BG)
	}
}

func TestConductorNone(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorNone, true, theme)

	if !strings.Contains(seg.Text, "\033]8;;https://github.com/rbarcante/claude-conductor") {
		t.Errorf("expected OSC 8 hyperlink for none state, got %q", seg.Text)
	}
	if !strings.Contains(seg.Text, "Try Conductor") {
		t.Errorf("expected 'Try Conductor', got %q", seg.Text)
	}
	if seg.VisualText == "" {
		t.Error("expected VisualText to be set for OSC 8 link")
	}
	colors := theme.Segments["conductor_missing"]
	if seg.FG != colors.FG || seg.BG != colors.BG {
		t.Errorf("expected conductor_missing colors, got FG=%q BG=%q", seg.FG, seg.BG)
	}
}

func TestConductorNoneNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorNone, false, theme)

	if !strings.Contains(seg.Text, "Try Conductor") {
		t.Errorf("expected 'Try Conductor', got %q", seg.Text)
	}
	if strings.Contains(seg.Text, "⚡") {
		t.Error("expected no lightning icon without nerd fonts")
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

			// Active uses conductor colors
			seg := Conductor(ConductorActive, true, theme)
			colors := theme.Segments["conductor"]
			if seg.FG != colors.FG || seg.BG != colors.BG {
				t.Errorf("active: wrong colors for theme %q", name)
			}

			// All non-active states use conductor_missing colors
			for _, status := range []ConductorStatus{ConductorInstalled, ConductorMarketplace, ConductorNone} {
				seg = Conductor(status, true, theme)
				colors = theme.Segments["conductor_missing"]
				if seg.FG != colors.FG || seg.BG != colors.BG {
					t.Errorf("status %d: wrong colors for theme %q", status, name)
				}
			}
		})
	}
}
