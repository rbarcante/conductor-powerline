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
	// Colors defined as hex values converted via hexToAnsi256.
	// Unified warning/critical keys replace per-segment variants.
	"dark": {
		Name: "dark",
		Segments: map[string]SegmentColors{
			"directory":         {FG: "231", BG: "130"}, // #ffffff / #8b4513
			"git":               {FG: "231", BG: "237"}, // #ffffff / #404040
			"model":             {FG: "231", BG: "236"}, // #ffffff / #2d2d2d
			"block":             {FG: "153", BG: "235"}, // #87ceeb / #2a2a2a
			"weekly":            {FG: "157", BG: "234"}, // #98fb98 / #1a1a1a
			"opus":              {FG: "183", BG: "234"}, // #c792ea / #1a1a1a
			"sonnet":            {FG: "153", BG: "234"}, // #89ddff / #1a1a1a
			"context":           {FG: "153", BG: "235"}, // #87ceeb / #2a2a2a
			"warning":           {FG: "231", BG: "172"}, // #ffffff / #d75f00
			"critical":          {FG: "231", BG: "124"}, // #ffffff / #af0000
			"conductor":         {FG: "157", BG: "236"}, // #98fb98 / #2d2d2d
			"conductor_missing": {FG: "221", BG: "236"}, // #ffd700 / #2d2d2d
		},
	},
	"light": {
		Name: "light",
		Segments: map[string]SegmentColors{
			"directory":         {FG: "231", BG: "209"}, // #ffffff / #ff6b47
			"git":               {FG: "231", BG: "116"}, // #ffffff / #4fb3d9
			"model":             {FG: "16", BG: "153"},  // #000000 / #87ceeb
			"block":             {FG: "231", BG: "105"}, // #ffffff / #6366f1
			"weekly":            {FG: "231", BG: "43"},  // #ffffff / #10b981
			"opus":              {FG: "231", BG: "141"}, // #ffffff / #8b5cf6
			"sonnet":            {FG: "231", BG: "39"},  // #ffffff / #0ea5e9
			"context":           {FG: "231", BG: "105"}, // #ffffff / #6366f1
			"warning":           {FG: "16", BG: "214"},  // #000000 / #f59e0b
			"critical":          {FG: "231", BG: "203"}, // #ffffff / #ef4444
			"conductor":         {FG: "231", BG: "35"},  // #ffffff / #00af5f
			"conductor_missing": {FG: "16", BG: "220"},  // #000000 / #ffd700
		},
	},
	"nord": {
		Name: "nord",
		Segments: map[string]SegmentColors{
			"directory":         {FG: "189", BG: "60"},  // #d8dee9 / #434c5e
			"git":               {FG: "151", BG: "60"},  // #a3be8c / #3b4252
			"model":             {FG: "146", BG: "66"},  // #81a1c1 / #4c566a
			"block":             {FG: "146", BG: "60"},  // #81a1c1 / #3b4252
			"weekly":            {FG: "152", BG: "59"},  // #8fbcbb / #2e3440
			"opus":              {FG: "181", BG: "59"},  // #b48ead / #2e3440
			"sonnet":            {FG: "152", BG: "59"},  // #88c0d0 / #2e3440
			"context":           {FG: "146", BG: "60"},  // #81a1c1 / #3b4252
			"warning":           {FG: "59", BG: "180"},  // #2e3440 / #d08770
			"critical":          {FG: "231", BG: "174"}, // #eceff4 / #bf616a
			"conductor":         {FG: "151", BG: "66"},  // #a3be8c / #4c566a
			"conductor_missing": {FG: "59", BG: "180"},  // #2e3440 / #d08770
		},
	},
	"gruvbox": {
		Name: "gruvbox",
		Segments: map[string]SegmentColors{
			"directory":         {FG: "223", BG: "95"},  // #ebdbb2 / #504945
			"git":               {FG: "185", BG: "59"},  // #b8bb26 / #3c3836
			"model":             {FG: "145", BG: "102"}, // #83a598 / #665c54
			"block":             {FG: "145", BG: "59"},  // #83a598 / #3c3836
			"weekly":            {FG: "221", BG: "235"}, // #fabd2f / #282828
			"opus":              {FG: "181", BG: "235"}, // #d3869b / #282828
			"sonnet":            {FG: "150", BG: "235"}, // #8ec07c / #282828
			"context":           {FG: "145", BG: "59"},  // #83a598 / #3c3836
			"warning":           {FG: "235", BG: "179"}, // #282828 / #d79921
			"critical":          {FG: "223", BG: "167"}, // #ebdbb2 / #cc241d
			"conductor":         {FG: "150", BG: "59"},  // #8ec07c / #3c3836
			"conductor_missing": {FG: "235", BG: "179"}, // #282828 / #d79921
		},
	},
	"tokyo-night": {
		Name: "tokyo-night",
		Segments: map[string]SegmentColors{
			"directory":         {FG: "147", BG: "60"}, // #82aaff / #2f334d
			"git":               {FG: "193", BG: "59"}, // #c3e88d / #1e2030
			"model":             {FG: "219", BG: "23"}, // #fca7ea / #191b29
			"block":             {FG: "111", BG: "59"}, // #7aa2f7 / #2d3748
			"weekly":            {FG: "116", BG: "59"}, // #4fd6be / #1a202c
			"opus":              {FG: "183", BG: "59"}, // #bb9af7 / #1a202c
			"sonnet":            {FG: "117", BG: "59"}, // #7dcfff / #1a202c
			"context":           {FG: "111", BG: "59"}, // #7aa2f7 / #2d3748
			"warning":           {FG: "59", BG: "180"}, // #1a1b26 / #e0af68
			"critical":          {FG: "59", BG: "211"}, // #1a1b26 / #f7768e
			"conductor":         {FG: "193", BG: "23"}, // #c3e88d / #191b29
			"conductor_missing": {FG: "59", BG: "180"}, // #1a1b26 / #e0af68
		},
	},
	"rose-pine": {
		Name: "rose-pine",
		Segments: map[string]SegmentColors{
			"directory":         {FG: "183", BG: "59"}, // #c4a7e7 / #26233a
			"git":               {FG: "152", BG: "59"}, // #9ccfd8 / #1f1d2e
			"model":             {FG: "224", BG: "17"}, // #ebbcba / #191724
			"block":             {FG: "211", BG: "59"}, // #eb6f92 / #2a273f
			"weekly":            {FG: "152", BG: "59"}, // #9ccfd8 / #232136
			"opus":              {FG: "183", BG: "59"}, // #c4a7e7 / #232136
			"sonnet":            {FG: "67", BG: "59"},  // #31748f / #232136
			"context":           {FG: "152", BG: "59"}, // #9ccfd8 / #2a273f
			"warning":           {FG: "17", BG: "222"}, // #191724 / #f6c177
			"critical":          {FG: "17", BG: "211"}, // #191724 / #eb6f92
			"conductor":         {FG: "152", BG: "17"}, // #9ccfd8 / #191724
			"conductor_missing": {FG: "17", BG: "222"}, // #191724 / #f6c177
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
