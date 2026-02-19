package oauth

import "time"

// UsageData holds parsed usage information from the Anthropic API.
type UsageData struct {
	// Block usage (5-hour window)
	BlockPercentage float64
	BlockResetTime  time.Time

	// Weekly usage (7-day rolling)
	WeeklyPercentage float64
	OpusPercentage   float64
	SonnetPercentage float64
	WeekResetTime    time.Time

	// Metadata
	IsStale   bool
	FetchedAt time.Time
}
