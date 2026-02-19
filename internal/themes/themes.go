// Package themes provides color theme definitions for powerline segments.
package themes

import "sort"

// SegmentColors holds foreground and background ANSI color codes for a segment.
type SegmentColors struct {
	FG string
	BG string
}

// Theme defines a named set of segment colors.
type Theme struct {
	Name     string
	Segments map[string]SegmentColors
}

var registry = map[string]Theme{
	"dark": {
		Name: "dark",
		Segments: map[string]SegmentColors{
			"directory": {FG: "15", BG: "236"},
			"git":       {FG: "15", BG: "22"},
			"model":     {FG: "15", BG: "57"},
		},
	},
	"light": {
		Name: "light",
		Segments: map[string]SegmentColors{
			"directory": {FG: "0", BG: "254"},
			"git":       {FG: "0", BG: "120"},
			"model":     {FG: "15", BG: "99"},
		},
	},
	"nord": {
		Name: "nord",
		Segments: map[string]SegmentColors{
			"directory": {FG: "15", BG: "60"},
			"git":       {FG: "15", BG: "71"},
			"model":     {FG: "15", BG: "110"},
		},
	},
	"gruvbox": {
		Name: "gruvbox",
		Segments: map[string]SegmentColors{
			"directory": {FG: "223", BG: "239"},
			"git":       {FG: "223", BG: "100"},
			"model":     {FG: "223", BG: "124"},
		},
	},
	"tokyo-night": {
		Name: "tokyo-night",
		Segments: map[string]SegmentColors{
			"directory": {FG: "189", BG: "236"},
			"git":       {FG: "189", BG: "29"},
			"model":     {FG: "189", BG: "62"},
		},
	},
	"rose-pine": {
		Name: "rose-pine",
		Segments: map[string]SegmentColors{
			"directory": {FG: "189", BG: "238"},
			"git":       {FG: "189", BG: "96"},
			"model":     {FG: "189", BG: "132"},
		},
	},
}

// Get returns the theme with the given name. If the theme is not found,
// it returns the "dark" theme as a fallback.
func Get(name string) (Theme, bool) {
	theme, ok := registry[name]
	if !ok {
		return registry["dark"], true
	}
	return theme, true
}

// Names returns a sorted list of all available theme names.
func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
