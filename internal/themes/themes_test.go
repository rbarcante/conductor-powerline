package themes

import "testing"

var expectedThemes = []string{"dark", "light", "nord", "gruvbox", "tokyo-night", "rose-pine"}

func TestAllThemesDefined(t *testing.T) {
	for _, name := range expectedThemes {
		theme, ok := Get(name)
		if !ok {
			t.Errorf("theme %q not found in registry", name)
			continue
		}
		if theme.Name != name {
			t.Errorf("theme name mismatch: expected %q, got %q", name, theme.Name)
		}
	}
}

func TestThemeSegmentColors(t *testing.T) {
	requiredSegments := []string{"directory", "git", "model", "block", "block-warning", "block-critical", "weekly"}

	for _, name := range expectedThemes {
		theme, ok := Get(name)
		if !ok {
			t.Fatalf("theme %q not found", name)
		}
		for _, seg := range requiredSegments {
			colors, ok := theme.Segments[seg]
			if !ok {
				t.Errorf("theme %q missing segment colors for %q", name, seg)
				continue
			}
			if colors.FG == "" {
				t.Errorf("theme %q segment %q has empty FG color", name, seg)
			}
			if colors.BG == "" {
				t.Errorf("theme %q segment %q has empty BG color", name, seg)
			}
		}
	}
}

func TestFallbackToDark(t *testing.T) {
	theme, ok := Get("nonexistent-theme")
	if !ok {
		t.Fatal("fallback should return a theme")
	}
	if theme.Name != "dark" {
		t.Errorf("expected fallback to 'dark', got %q", theme.Name)
	}
}

func TestGetExact(t *testing.T) {
	theme, ok := Get("nord")
	if !ok {
		t.Fatal("expected to find 'nord' theme")
	}
	if theme.Name != "nord" {
		t.Errorf("expected 'nord', got %q", theme.Name)
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) != len(expectedThemes) {
		t.Fatalf("expected %d themes, got %d", len(expectedThemes), len(names))
	}
	// Check all expected themes are present
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, expected := range expectedThemes {
		if !nameSet[expected] {
			t.Errorf("theme %q missing from Names()", expected)
		}
	}
}
