package segments

import (
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

const conductorURL = "https://github.com/rbarcante/claude-conductor"

// Conductor returns a segment reflecting the conductor plugin status.
//
// States:
//   - ConductorActive:      "✓ Conductor"         (green — fully set up)
//   - ConductorInstalled:   "⚡ Setup Conductor"   (yellow — plugin installed, project needs setup)
//   - ConductorMarketplace: "⚡ Install Conductor"  (yellow — marketplace present, plugin not installed)
//   - ConductorNone:        "⚡ Try Conductor"      (yellow — nothing installed)
func Conductor(status ConductorStatus, nerdFonts bool, theme themes.Theme) Segment {
	switch status {
	case ConductorActive:
		colors := theme.Segments["conductor"]
		text := "✓ Conductor"
		if !nerdFonts {
			text = "OK Conductor"
		}
		return Segment{
			Name:    "conductor",
			Text:    text,
			FG:      colors.FG,
			BG:      colors.BG,
			Enabled: true,
		}

	case ConductorInstalled:
		colors := theme.Segments["conductor_missing"]
		label := "⚡ Setup Conductor"
		if !nerdFonts {
			label = "Setup Conductor"
		}
		return Segment{
			Name:    "conductor",
			Text:    label,
			FG:      colors.FG,
			BG:      colors.BG,
			Enabled: true,
		}

	case ConductorMarketplace:
		colors := theme.Segments["conductor_missing"]
		label := "⚡ Install Conductor"
		if !nerdFonts {
			label = "Install Conductor"
		}
		return Segment{
			Name:    "conductor",
			Text:    label,
			FG:      colors.FG,
			BG:      colors.BG,
			Enabled: true,
		}

	default: // ConductorNone
		colors := theme.Segments["conductor_missing"]
		label := "⚡ Try Conductor"
		if !nerdFonts {
			label = "Try Conductor"
		}
		return Segment{
			Name:    "conductor",
			Text:    label,
			Link:    conductorURL,
			FG:      colors.FG,
			BG:      colors.BG,
			Enabled: true,
		}
	}
}
