package render

import (
	"fmt"
	"strings"

	"github.com/rbarcante/conductor-powerline/internal/segments"
)

const maxCompactTextLen = 12

// ANSI 256-color escape code helpers.
// Format: \033[38;5;{n}m = set foreground to color n
//         \033[48;5;{n}m = set background to color n
//         \033[0m        = reset all attributes

// ansi256 formats text with 256-color ANSI foreground and background.
func ansi256(fg, bg, text string) string {
	return fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm %s ", fg, bg, text)
}

// ansiSep formats a powerline separator where the previous segment's bg
// becomes the foreground and the next segment's bg is the background.
func ansiSep(fg, bg, sep string) string {
	return fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm%s", fg, bg, sep)
}

// ansiReset returns a reset sequence followed by a colored separator.
func ansiResetSep(fg, sep string) string {
	return fmt.Sprintf("\033[0m\033[38;5;%sm%s\033[0m", fg, sep)
}

// osc8Open emits the OSC 8 hyperlink opening escape for the given URL.
// Modern tmux (3.1+) natively understands OSC 8 — no DCS passthrough needed.
func osc8Open(url string) string {
	return fmt.Sprintf("\033]8;;%s\033\\", url)
}

// osc8CloseStr emits the OSC 8 hyperlink closing escape.
func osc8CloseStr() string {
	return "\033]8;;\033\\"
}

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

		if seg.Link != "" {
			b.WriteString(osc8Open(seg.Link))
		}

		if nerdFonts {
			b.WriteString(ansi256(seg.FG, seg.BG, text))
			if i < len(active)-1 {
				next := active[i+1]
				b.WriteString(ansiSep(seg.BG, next.BG, sep))
			} else {
				b.WriteString(ansiResetSep(seg.BG, sep))
			}
		} else {
			b.WriteString(fmt.Sprintf("\033[38;5;%sm\033[48;5;%sm %s \033[0m", seg.FG, seg.BG, text))
			if i < len(active)-1 {
				b.WriteString(sep)
			}
		}

		if seg.Link != "" {
			b.WriteString(osc8CloseStr())
		}
	}

	return b.String()
}

// RenderRight produces an ANSI-colored powerline string for right-side segments
// using left-pointing arrow separators. No compact mode for right segments.
func RenderRight(segs []segments.Segment, nerdFonts bool) string {
	active := filterEnabled(segs)
	if len(active) == 0 {
		return ""
	}

	sep := SeparatorLeftNerd
	if !nerdFonts {
		sep = SeparatorLeftText
	}

	var b strings.Builder

	for i, seg := range active {
		if seg.Link != "" {
			b.WriteString(osc8Open(seg.Link))
		}

		if nerdFonts {
			if i == 0 {
				b.WriteString(fmt.Sprintf("\033[38;5;%sm%s", seg.BG, sep))
			} else {
				prev := active[i-1]
				b.WriteString(ansiSep(seg.BG, prev.BG, sep))
			}
			b.WriteString(ansi256(seg.FG, seg.BG, seg.Text))
		} else {
			if i > 0 {
				b.WriteString(sep)
			}
			b.WriteString(ansi256(seg.FG, seg.BG, seg.Text))
		}

		if seg.Link != "" {
			b.WriteString(osc8CloseStr())
		}
	}

	// Reset at the end
	b.WriteString("\033[0m")
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
