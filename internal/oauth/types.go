package oauth

import "time"

// TokenCredentials holds an OAuth access token and optional refresh token.
// Platform credential stores that support refresh tokens (macOS Keychain, credfile)
// populate both fields; others (wincred, secret-tool) only set AccessToken.
type TokenCredentials struct {
	AccessToken  string
	RefreshToken string
}

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
