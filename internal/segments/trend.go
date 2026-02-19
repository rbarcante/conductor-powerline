package segments

import "math"

// TrendArrow returns a directional arrow comparing current vs previous usage.
// Returns "↑" for increasing, "↓" for decreasing, "→" for stable (±2% threshold).
// Returns empty string if previous is negative (no previous data available).
func TrendArrow(current, previous float64) string {
	if previous < 0 {
		return ""
	}

	diff := current - previous
	if math.Abs(diff) <= 2.0 {
		return "\u2192" // →
	}
	if diff > 0 {
		return "\u2191" // ↑
	}
	return "\u2193" // ↓
}
