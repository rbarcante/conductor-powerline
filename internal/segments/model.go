package segments

import (
	"strings"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// modelNames maps known model ID prefixes to friendly display names.
var modelNames = map[string]string{
	"claude-opus-4-6":    "Opus 4.6",
	"claude-sonnet-4-6":  "Sonnet 4.6",
	"claude-sonnet-4-5":  "Sonnet 4.5",
	"claude-haiku-4-5":   "Haiku 4.5",
	"claude-opus-4-5":    "Opus 4.5",
	"claude-sonnet-4-0":  "Sonnet 4",
	"claude-haiku-3-5":   "Haiku 3.5",
	"claude-sonnet-3-5":  "Sonnet 3.5",
}

// Model returns a segment displaying the friendly name of the active Claude model.
// Returns a disabled segment if the model ID is empty.
func Model(modelID string, theme themes.Theme) Segment {
	colors := theme.Segments["model"]

	if modelID == "" {
		return Segment{Name: "model", Enabled: false}
	}

	name := resolveFriendlyName(modelID)

	return Segment{
		Name:    "model",
		Text:    name,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

func resolveFriendlyName(modelID string) string {
	// Try exact match first
	if name, ok := modelNames[modelID]; ok {
		return name
	}

	// Try prefix match (handles dated model IDs like claude-haiku-4-5-20251001)
	for prefix, name := range modelNames {
		if strings.HasPrefix(modelID, prefix) {
			return name
		}
	}

	return modelID
}
