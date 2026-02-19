// Package segments provides individual powerline segment providers.
package segments

// Segment represents a single rendered powerline segment.
type Segment struct {
	Name    string
	Text    string
	FG      string
	BG      string
	Enabled bool
}
