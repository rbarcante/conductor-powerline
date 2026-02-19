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
	// Colors ported from conductor-powerline hex values via hexToAnsi256 conversion
	"dark": {
		Name: "dark",
		Segments: map[string]SegmentColors{
			"directory":      {FG: "231", BG: "130"}, // #8b4513 / #ffffff
			"git":            {FG: "231", BG: "237"}, // #404040 / #ffffff
			"model":          {FG: "231", BG: "236"}, // #2d2d2d / #ffffff
			"block":          {FG: "231", BG: "28"},  // green normal
			"block-warning":  {FG: "231", BG: "208"}, // orange warning
			"block-critical": {FG: "231", BG: "196"}, // red critical
			"weekly":         {FG: "231", BG: "62"},  // purple
		},
	},
	"light": {
		Name: "light",
		Segments: map[string]SegmentColors{
			"directory":      {FG: "231", BG: "209"}, // #ff6b47 / #ffffff
			"git":            {FG: "231", BG: "116"}, // #4fb3d9 / #ffffff
			"model":          {FG: "16", BG: "153"},  // #87ceeb / #000000
			"block":          {FG: "231", BG: "34"},  // green normal
			"block-warning":  {FG: "231", BG: "214"}, // orange warning
			"block-critical": {FG: "231", BG: "196"}, // red critical
			"weekly":         {FG: "231", BG: "135"}, // purple
		},
	},
	"nord": {
		Name: "nord",
		Segments: map[string]SegmentColors{
			"directory":      {FG: "189", BG: "60"},  // #434c5e / #d8dee9
			"git":            {FG: "151", BG: "60"},  // #3b4252 / #a3be8c
			"model":          {FG: "146", BG: "66"},  // #4c566a / #81a1c1
			"block":          {FG: "151", BG: "60"},  // green normal
			"block-warning":  {FG: "223", BG: "172"}, // yellow warning
			"block-critical": {FG: "231", BG: "131"}, // red critical
			"weekly":         {FG: "146", BG: "60"},  // blue
		},
	},
	"gruvbox": {
		Name: "gruvbox",
		Segments: map[string]SegmentColors{
			"directory":      {FG: "223", BG: "95"},  // #504945 / #ebdbb2
			"git":            {FG: "185", BG: "59"},  // #3c3836 / #b8bb26
			"model":          {FG: "145", BG: "102"}, // #665c54 / #83a598
			"block":          {FG: "185", BG: "59"},  // green normal
			"block-warning":  {FG: "223", BG: "172"}, // yellow warning
			"block-critical": {FG: "223", BG: "124"}, // red critical
			"weekly":         {FG: "145", BG: "102"}, // blue
		},
	},
	"tokyo-night": {
		Name: "tokyo-night",
		Segments: map[string]SegmentColors{
			"directory":      {FG: "147", BG: "60"},  // #2f334d / #82aaff
			"git":            {FG: "193", BG: "59"},  // #1e2030 / #c3e88d
			"model":          {FG: "219", BG: "23"},  // #191b29 / #fca7ea
			"block":          {FG: "193", BG: "59"},  // green normal
			"block-warning":  {FG: "223", BG: "172"}, // orange warning
			"block-critical": {FG: "231", BG: "197"}, // red critical
			"weekly":         {FG: "147", BG: "60"},  // blue
		},
	},
	"rose-pine": {
		Name: "rose-pine",
		Segments: map[string]SegmentColors{
			"directory":      {FG: "183", BG: "59"},  // #26233a / #c4a7e7
			"git":            {FG: "152", BG: "59"},  // #1f1d2e / #9ccfd8
			"model":          {FG: "224", BG: "17"},  // #191724 / #ebbcba
			"block":          {FG: "152", BG: "59"},  // teal normal
			"block-warning":  {FG: "223", BG: "172"}, // gold warning
			"block-critical": {FG: "231", BG: "131"}, // love critical
			"weekly":         {FG: "183", BG: "59"},  // iris
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
