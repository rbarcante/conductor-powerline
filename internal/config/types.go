// Package config handles loading and merging configuration for conductor-powerline.
package config

// Config is the top-level configuration structure.
type Config struct {
	Display      DisplayConfig            `json:"display"`
	Segments     map[string]SegmentConfig `json:"segments"`
	Theme        string                   `json:"theme"`
	SegmentOrder []string                 `json:"segmentOrder"`
}

// DisplayConfig controls rendering behavior.
type DisplayConfig struct {
	NerdFonts    bool `json:"nerdFonts"`
	CompactWidth int  `json:"compactWidth"`
}

// SegmentConfig controls an individual segment's behavior.
type SegmentConfig struct {
	Enabled bool `json:"enabled"`
}
