package segments

import (
	"testing"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func TestModelKnownIDs(t *testing.T) {
	theme, _ := themes.Get("dark")

	tests := []struct {
		modelID  string
		expected string
	}{
		{"claude-opus-4-6", "Opus 4.6"},
		{"claude-sonnet-4-6", "Sonnet 4.6"},
		{"claude-haiku-4-5-20251001", "Haiku 4.5"},
		{"claude-sonnet-4-5-20250514", "Sonnet 4.5"},
	}

	for _, tt := range tests {
		t.Run(tt.modelID, func(t *testing.T) {
			seg := Model(tt.modelID, theme)
			if seg.Text != tt.expected {
				t.Errorf("Model(%q) = %q, want %q", tt.modelID, seg.Text, tt.expected)
			}
			if !seg.Enabled {
				t.Error("expected segment enabled")
			}
			if seg.Name != "model" {
				t.Errorf("expected name 'model', got %q", seg.Name)
			}
		})
	}
}

func TestModelEmptyID(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Model("", theme)

	if seg.Enabled {
		t.Error("expected segment disabled for empty model ID")
	}
}

func TestModelUnknownID(t *testing.T) {
	theme, _ := themes.Get("dark")
	seg := Model("some-unknown-model", theme)

	if !seg.Enabled {
		t.Error("expected segment enabled for unknown model")
	}
	if seg.Text != "some-unknown-model" {
		t.Errorf("expected raw model ID as text, got %q", seg.Text)
	}
}
