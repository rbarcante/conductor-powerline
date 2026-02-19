package segments

import "math"

// trendStableThreshold is the minimum percentage change to register as a trend.
// This is a fixed default; the configurable TrendThreshold in Config is intended
// for higher-level callers that may adjust sensitivity.
const trendStableThreshold = 2.0

// TrendArrow returns a directional arrow comparing current vs previous usage.
// Returns "↑" for increasing, "↓" for decreasing, "→" for stable (within ±trendStableThreshold).
// Returns empty string if previous is negative (no previous data available).
func TrendArrow(current, previous float64) string {
	if previous < 0 {
		return ""
	}

	diff := current - previous
	if math.Abs(diff) <= trendStableThreshold {
		return "\u2192" // →
	}
	if diff > 0 {
		return "\u2191" // ↑
	}
	return "\u2193" // ↓
}
