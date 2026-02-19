package segments

import (
	"fmt"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

const conductorURL = "https://github.com/rbarcante/claude-conductor"

// osc8Link wraps text in an OSC 8 terminal hyperlink.
// Terminals that do not support OSC 8 display the text without the link.
func osc8Link(url, text string) string {
	return fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", url, text)
}

// Conductor returns a segment indicating whether the claude-conductor plugin is installed.
// When detected is true, it shows a success indicator with conductor theme colors.
// When detected is false, it shows a prompt with an OSC 8 hyperlink and conductor_missing colors.
func Conductor(detected bool, nerdFonts bool, theme themes.Theme) Segment {
	if detected {
		colors := theme.Segments["conductor"]
		var text string
		if nerdFonts {
			text = "✓ Conductor"
		} else {
			text = "OK Conductor"
		}
		return Segment{
			Name:    "conductor",
			Text:    text,
			FG:      colors.FG,
			BG:      colors.BG,
			Enabled: true,
		}
	}

	colors := theme.Segments["conductor_missing"]
	var label string
	if nerdFonts {
		label = "⚡ Get Conductor"
	} else {
		label = "Get Conductor"
	}
	text := osc8Link(conductorURL, label)
	return Segment{
		Name:    "conductor",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}
