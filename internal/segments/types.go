// Package segments provides individual powerline segment providers.
package segments

// Segment represents a single rendered powerline segment.
type Segment struct {
	Name       string
	Text       string
	VisualText string // Optional: plain text for width calculation and compact display.
	// When set, VisualText is used instead of Text for measuring display width and
	// for compact-mode truncation. Text (which may contain escape sequences like
	// OSC 8 hyperlinks) is used for full rendering.
	FG      string
	BG      string
	Enabled bool
}
