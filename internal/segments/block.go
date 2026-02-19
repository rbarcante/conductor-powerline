package segments

import (
	"fmt"
	"time"

	"github.com/rbarcante/conductor-powerline/internal/oauth"
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

// Block returns a segment displaying the 5-hour block usage percentage and countdown.
// Color intensity changes based on usage thresholds: normal (<70%), warning (70-90%), critical (>90%).
func Block(data *oauth.UsageData, theme themes.Theme) Segment {
	colors := theme.Segments["block"]

	if data == nil {
		return Segment{
			Name:    "block",
			Text:    "--",
			FG:      colors.FG,
			BG:      colors.BG,
			Enabled: true,
		}
	}

	// Select color based on usage threshold
	switch {
	case data.BlockPercentage >= 90:
		colors = theme.Segments["critical"]
	case data.BlockPercentage >= 70:
		colors = theme.Segments["warning"]
	}

	// Format countdown
	remaining := time.Until(data.BlockResetTime)
	countdown := formatCountdown(remaining)

	text := fmt.Sprintf("%.0f%% %s", data.BlockPercentage, countdown)
	if data.IsStale {
		text += " ~"
	}

	return Segment{
		Name:    "block",
		Text:    text,
		FG:      colors.FG,
		BG:      colors.BG,
		Enabled: true,
	}
}

// formatCountdown formats a duration as a compact countdown string (e.g., "2h13m").
func formatCountdown(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
