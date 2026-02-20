package segments

import (
	"strings"
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestConductorActive(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorActive, true, theme)

	if seg.Enabled {
		t.Error("expected segment disabled for ConductorActive (shown on line 2 instead)")
	}
	if seg.Name != "conductor" {
		t.Errorf("expected name 'conductor', got %q", seg.Name)
	}
}

func TestConductorActiveNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorActive, false, theme)

	if seg.Enabled {
		t.Error("expected segment disabled for ConductorActive regardless of nerd fonts")
	}
}

func TestConductorInstalled(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorInstalled, true, theme)

	if seg.Enabled {
		t.Error("expected segment disabled for ConductorInstalled (plugin present, no project setup)")
	}
	if seg.Name != "conductor" {
		t.Errorf("expected name 'conductor', got %q", seg.Name)
	}
}

func TestConductorInstalledNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorInstalled, false, theme)

	if seg.Enabled {
		t.Error("expected segment disabled for ConductorInstalled regardless of nerd fonts")
	}
}

func TestConductorMarketplace(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorMarketplace, true, theme)

	if !strings.Contains(seg.Text, "Install Conductor") {
		t.Errorf("expected 'Install Conductor', got %q", seg.Text)
	}
	colors := theme.Segments["conductor_missing"]
	if seg.FG != colors.FG || seg.BG != colors.BG {
		t.Errorf("expected conductor_missing colors, got FG=%q BG=%q", seg.FG, seg.BG)
	}
}

func TestConductorNone(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorNone, true, theme)

	if !strings.Contains(seg.Text, "Try Conductor") {
		t.Errorf("expected 'Try Conductor', got %q", seg.Text)
	}
	if seg.Link != "https://github.com/rbarcante/claude-conductor" {
		t.Errorf("expected Link to conductor URL, got %q", seg.Link)
	}
	colors := theme.Segments["conductor_missing"]
	if seg.FG != colors.FG || seg.BG != colors.BG {
		t.Errorf("expected conductor_missing colors, got FG=%q BG=%q", seg.FG, seg.BG)
	}
}

func TestConductorNoneNoNerdFonts(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Conductor(ConductorNone, false, theme)

	if seg.Text != "Try Conductor" {
		t.Errorf("expected 'Try Conductor', got %q", seg.Text)
	}
	if seg.Link == "" {
		t.Error("expected Link to be set for none state")
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

			// Active and Installed are disabled (no line-1 segment)
			for _, status := range []ConductorStatus{ConductorActive, ConductorInstalled} {
				seg := Conductor(status, true, theme)
				if seg.Enabled {
					t.Errorf("status %d: expected disabled for theme %q", status, name)
				}
			}

			// Marketplace and None use conductor_missing colors
			for _, status := range []ConductorStatus{ConductorMarketplace, ConductorNone} {
				seg := Conductor(status, true, theme)
				colors := theme.Segments["conductor_missing"]
				if seg.FG != colors.FG || seg.BG != colors.BG {
					t.Errorf("status %d: wrong colors for theme %q", status, name)
				}
			}
		})
	}
}
