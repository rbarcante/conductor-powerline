package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/rbarcante/conductor-powerline/internal/segments"
)

// inTmux reports whether the process is running inside tmux.
// Used to skip OSC 8 hyperlinks which Claude Code doesn't forward in tmux.
var inTmux = os.Getenv("TMUX") != ""

// minCompactTextLen is the minimum characters a segment can be truncated to.
const minCompactTextLen = 3

// overheadPerSeg is the character overhead per segment: 1 space left + 1 space right + 1 separator.
const overheadPerSeg = 3

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

// osc8Open emits the OSC 8 hyperlink opening escape for the given URL
// with underline enabled to visually indicate a clickable link.
func osc8Open(url string) string {
	return fmt.Sprintf("\033]8;;%s\033\\\033[4m", url)
}

// osc8CloseStr emits the OSC 8 hyperlink closing escape and disables underline.
func osc8CloseStr() string {
	return "\033[24m\033]8;;\033\\"
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

	var texts []string
	if compact {
		texts = compactTexts(active, termWidth)
	}

	var b strings.Builder

	for i, seg := range active {
		text := seg.Text
		if compact {
			text = texts[i]
		}

		if seg.Link != "" && !inTmux {
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

		if seg.Link != "" && !inTmux {
			b.WriteString(osc8CloseStr())
		}

		// In tmux, append the URL as plain text since Claude Code
		// doesn't forward OSC 8 hyperlinks inside tmux.
		if seg.Link != "" && inTmux {
			b.WriteString(fmt.Sprintf(" \033[38;5;244m%s\033[0m", seg.Link))
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
		if seg.Link != "" && !inTmux {
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

		if seg.Link != "" && !inTmux {
			b.WriteString(osc8CloseStr())
		}

		if seg.Link != "" && inTmux {
			b.WriteString(fmt.Sprintf(" \033[38;5;244m%s\033[0m", seg.Link))
		}
	}

	// Reset at the end
	b.WriteString("\033[0m")
	return b.String()
}

// compactTexts calculates per-segment max text lengths proportional to each
// segment's original text length, so the total rendered width fits within
// termWidth. Each segment gets at least minCompactTextLen characters.
// Proportional shares are rounded down, so the result may slightly undershoot
// the available width; remainder characters are distributed to the longest segments.
func compactTexts(segs []segments.Segment, termWidth int) []string {
	n := len(segs)
	result := make([]string, n)
	lengths := make([]int, n)

	// The last segment has no trailing separator, so overhead is slightly less,
	// but we use a uniform estimate for simplicity (slightly conservative).
	totalOverhead := n * overheadPerSeg

	// Available width for text only. When termWidth is very small (less than
	// overhead alone), clamp to the minimum floor — this is best-effort and
	// the result may still exceed termWidth.
	availableTextWidth := termWidth - totalOverhead
	clamped := false
	if availableTextWidth < n*minCompactTextLen {
		availableTextWidth = n * minCompactTextLen
		clamped = true
	}

	// Calculate total original text length in a single pass
	totalTextLen := 0
	for i, s := range segs {
		lengths[i] = len([]rune(s.Text))
		totalTextLen += lengths[i]
	}

	// If everything already fits (and we didn't clamp), no truncation needed
	if !clamped && totalTextLen <= availableTextWidth {
		for i, s := range segs {
			result[i] = s.Text
		}
		return result
	}

	// Proportionally allocate available width, then distribute remainder
	allocated := 0
	for i := range segs {
		maxLen := (lengths[i] * availableTextWidth) / totalTextLen
		if maxLen < minCompactTextLen {
			maxLen = minCompactTextLen
		}
		lengths[i] = maxLen
		allocated += maxLen
	}

	// Distribute leftover characters to longest segments first
	remainder := availableTextWidth - allocated
	for remainder > 0 {
		best := -1
		for i := range segs {
			origLen := len([]rune(segs[i].Text))
			if lengths[i] < origLen && (best < 0 || lengths[i] > lengths[best]) {
				best = i
			}
		}
		if best < 0 {
			break
		}
		lengths[best]++
		remainder--
	}

	for i, s := range segs {
		result[i] = truncate(s.Text, lengths[i])
	}

	return result
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
		totalLen += len([]rune(s.Text)) + overheadPerSeg
	}
	return totalLen > termWidth
}

func truncate(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	if maxLen == 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}
