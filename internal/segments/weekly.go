package segments

import (
	"fmt"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/oauth"
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// Weekly returns a segment displaying the 7-day rolling usage with optional Opus/Sonnet breakdown.
func Weekly(data *oauth.UsageData, theme themes.Theme) Segment {
	colors := theme.Segments["weekly"]

	if data == nil {
		return Segment{
			Name:    "weekly",
			Text:    "--",
			FG:      colors.FG,
			BG:      colors.BG,
			Enabled: true,
		}
	}

	var text string

	// Smart mode: show breakdown when both Opus and Sonnet are in use
	if data.OpusPercentage > 0 && data.SonnetPercentage > 0 {
		text = fmt.Sprintf("O:%.0f%% S:%.0f%%", data.OpusPercentage, data.SonnetPercentage)
	} else {
		text = fmt.Sprintf("%.0f%%", data.WeeklyPercentage)
	}

	// Add week progress indicator
	daysLeft := int(time.Until(data.WeekResetTime).Hours() / 24)
	if daysLeft > 0 {
		text += fmt.Sprintf(" %dd", daysLeft)
	}

	if data.IsStale {
		text += " ~"
	}

	return Segment{
		Name:    "weekly",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}
