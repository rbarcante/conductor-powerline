package segments

import (
	"fmt"

	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// Context returns a segment displaying the context window usage percentage.
// Returns a disabled segment if percent is -1 (missing data).
// Icon and color change dynamically based on usage thresholds.
func Context(percent int, nerdFonts bool, theme themes.Theme) Segment {
	if percent < 0 {
		return Segment{Name: "context", Enabled: false}
	}

	// Select color based on usage threshold
	var colorKey string
	switch {
	case percent > 80:
		colorKey = "critical"
	case percent >= 50:
		colorKey = "warning"
	default:
		colorKey = "context"
	}
	colors := theme.Segments[colorKey]

	// Select icon based on threshold
	var icon string
	if nerdFonts {
		switch {
		case percent > 80:
			icon = "●"
		case percent >= 50:
			icon = "◐"
		default:
			icon = "○"
		}
	} else {
		icon = "CTX"
	}

	text := fmt.Sprintf("%s %d%%", icon, percent)

	return Segment{
		Name:    "context",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}
