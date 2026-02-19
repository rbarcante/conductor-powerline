package render

import (
	"fmt"
	"strings"

	"github.com/rbarcante/conductor-powerline/internal/segments"
)

const maxCompactTextLen = 12

// Render produces an ANSI-colored powerline string from ordered segments.
// It skips disabled segments, applies compact mode below the given terminal width,
// and returns a string with no trailing newline.
func Render(segs []segments.Segment, nerdFonts bool, termWidth int) string {
	active := filterEnabled(segs)
	if len(active) == 0 {
		return ""
	}

	compact := shouldCompact(active, termWidth)
	sep := SeparatorNerd
	if !nerdFonts {
		sep = SeparatorText
	}

	var b strings.Builder

	for i, seg := range active {
		text := seg.Text
		if compact {
			text = truncate(text, maxCompactTextLen)
		}

		if nerdFonts {
			// Segment body: fg on bg
			b.WriteString(fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm %s ", seg.FG, seg.BG, text))

			// Separator: bg of current as fg, bg of next (or reset)
			if i < len(active)-1 {
				next := active[i+1]
				b.WriteString(fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm%s", seg.BG, next.BG, sep))
			} else {
				b.WriteString(fmt.Sprintf("\033[0m\033[38;5;%sm%s\033[0m", seg.BG, sep))
			}
		} else {
			b.WriteString(fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm %s \033[0m", seg.FG, seg.BG, text))
			if i < len(active)-1 {
				b.WriteString(sep)
			}
		}
	}

	return b.String()
}

func filterEnabled(segs []segments.Segment) []segments.Segment {
	var result []segments.Segment
	for _, s := range segs {
		if s.Enabled {
			result = append(result, s)
		}
	}
	return result
}

func shouldCompact(segs []segments.Segment, termWidth int) bool {
	totalLen := 0
	for _, s := range segs {
		totalLen += len([]rune(s.Text)) + 3 // text + padding + separator
	}
	return totalLen > termWidth
}

func truncate(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen-1]) + "…"
}
