// Package render builds ANSI-colored powerline output from segments.
package render

// Powerline separator glyphs (Nerd Font).
const (
	SeparatorNerd = "\ue0b0" // Right-pointing arrow (left-side segments)
	SeparatorText = "|"      // Fallback for non-Nerd Font terminals

	SeparatorLeftNerd = "\ue0b2" // Left-pointing arrow (right-side segments)
	SeparatorLeftText = "|"      // Fallback for non-Nerd Font terminals
)
