// Package config handles loading and merging configuration for conductor-powerline.
package config

import "time"

// Config is the top-level configuration structure.
type Config struct {
	Display        DisplayConfig            `json:"display"`
	Segments       map[string]SegmentConfig `json:"segments"`
	Theme          string                   `json:"theme"`
	SegmentOrder   []string                 `json:"segmentOrder"`
	APITimeout     Duration                 `json:"apiTimeout"`
	CacheTTL       Duration                 `json:"cacheTTL"`
	TrendThreshold float64                  `json:"trendThreshold"`
}

// DisplayConfig controls rendering behavior.
type DisplayConfig struct {
	NerdFonts    *bool `json:"nerdFonts,omitempty"`
	CompactWidth int   `json:"compactWidth"`
}

// NerdFontsEnabled returns the effective NerdFonts value, defaulting to true if nil.
func (d DisplayConfig) NerdFontsEnabled() bool {
	if d.NerdFonts == nil {
		return true
	}
	return *d.NerdFonts
}

// boolPtr returns a pointer to the given bool value.
func boolPtr(b bool) *bool {
	return &b
}

// SegmentConfig controls an individual segment's behavior.
type SegmentConfig struct {
	Enabled bool `json:"enabled"`
}

// Duration wraps time.Duration for JSON marshaling as a string (e.g., "5s", "30s").
type Duration struct {
	time.Duration
}

// UnmarshalJSON parses a JSON string like "5s" into a Duration.
func (d *Duration) UnmarshalJSON(b []byte) error {
	// Remove quotes
	s := string(b)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = dur
	return nil
}

// MarshalJSON encodes a Duration as a JSON string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Duration.String() + `"`), nil
}
