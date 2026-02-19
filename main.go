// conductor-powerline is a fast powerline statusline for Claude Code.
// It reads hook JSON from stdin, loads configuration, builds segments,
// and renders ANSI-colored output to stdout.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/rbarcante/conductor-powerline/internal/config"
	"github.com/rbarcante/conductor-powerline/internal/hook"
	"github.com/rbarcante/conductor-powerline/internal/oauth"
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

	// 4. Fetch usage data in parallel with segment building
	var usageData *oauth.UsageData
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		client := oauth.NewClient("https://api.anthropic.com/v1/usage", cfg.APITimeout.Duration)
		cache := oauth.NewCache(cfg.CacheTTL.Duration)
		data, err := oauth.FetchUsage(client, cache)
		if err == nil {
			usageData = data
		}
		// On error, usageData remains nil → segments show "--" placeholder
	}()

	wg.Wait()

	// 5. Build segments in configured order
	segs := buildSegments(cfg, hookData, theme, usageData)

	// 6. Render and output (no trailing newline)
	output := render.Render(segs, cfg.Display.NerdFonts, cfg.Display.CompactWidth)
	fmt.Print(output)

	return nil
}

func buildSegments(cfg config.Config, hookData hook.Data, theme themes.Theme, usageData *oauth.UsageData) []segments.Segment {
	builders := map[string]func() segments.Segment{
		"directory": func() segments.Segment {
			return segments.Directory(hookData.WorkspacePath(), theme)
		},
		"git": func() segments.Segment {
			return segments.Git(theme)
		},
		"model": func() segments.Segment {
			return segments.Model(hookData.ModelID(), theme)
		},
		"block": func() segments.Segment {
			return segments.Block(usageData, theme)
		},
		"weekly": func() segments.Segment {
			return segments.Weekly(usageData, theme)
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
