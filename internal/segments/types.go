// Package segments provides individual powerline segment providers.
package segments

// Segment represents a single rendered powerline segment.
type Segment struct {
	Name    string
	Text    string
	Link    string // Optional: URL for OSC 8 hyperlink wrapping the entire segment.
	FG      string
	BG      string
	Enabled bool
}
