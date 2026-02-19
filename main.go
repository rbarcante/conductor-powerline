// conductor-powerline is a fast powerline statusline for Claude Code.
// It reads hook JSON from stdin, loads configuration, builds segments,
// and renders ANSI-colored output to stdout.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rbarcante/conductor-powerline/internal/config"
	"github.com/rbarcante/conductor-powerline/internal/hook"
	"github.com/rbarcante/conductor-powerline/internal/render"
	"github.com/rbarcante/conductor-powerline/internal/segments"
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

func main() {
	if err := run(); err != nil {
		// Silent exit — statusline must never crash or produce stderr noise
		os.Exit(0)
	}
}

func run() error {
	// 1. Parse stdin hook data
	hookData, err := hook.Parse(os.Stdin)
	if err != nil {
		return err
	}

	// 2. Load config (project → user → defaults)
	projectCfg := filepath.Join(".", ".conductor-powerline.json")
	userCfg := ""
	if home, err := os.UserHomeDir(); err == nil {
		userCfg = filepath.Join(home, ".claude", "conductor-powerline.json")
	}
	cfg := config.Load(projectCfg, userCfg)

	// 3. Resolve theme
	theme, _ := themes.Get(cfg.Theme)

	// 4. Build segments in configured order
	segs := buildSegments(cfg, hookData, theme)

	// 5. Render and output (no trailing newline)
	output := render.Render(segs, cfg.Display.NerdFonts, cfg.Display.CompactWidth)
	fmt.Print(output)

	return nil
}

func buildSegments(cfg config.Config, hookData hook.Data, theme themes.Theme) []segments.Segment {
	builders := map[string]func() segments.Segment{
		"directory": func() segments.Segment {
			return segments.Directory(hookData.Workspace, theme)
		},
		"git": func() segments.Segment {
			return segments.Git(theme)
		},
		"model": func() segments.Segment {
			return segments.Model(hookData.Model, theme)
		},
	}

	var result []segments.Segment
	for _, name := range cfg.SegmentOrder {
		segCfg, hasCfg := cfg.Segments[name]
		if hasCfg && !segCfg.Enabled {
			continue
		}
		builder, ok := builders[name]
		if !ok {
			continue
		}
		result = append(result, builder())
	}
	return result
}
