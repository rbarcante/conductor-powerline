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
	"github.com/rbarcante/conductor-powerline/internal/debug"
	"github.com/rbarcante/conductor-powerline/internal/hook"
	"github.com/rbarcante/conductor-powerline/internal/oauth"
	"github.com/rbarcante/conductor-powerline/internal/render"
	"github.com/rbarcante/conductor-powerline/internal/segments"
	"github.com/rbarcante/conductor-powerline/internal/themes"
)

const anthropicUsageURL = "https://api.anthropic.com/api/oauth/usage"

func main() {
	debug.Init()
	if err := run(); err != nil {
		debug.Logf("main", "run error: %v", err)
		// Deliberate os.Exit(0) on error: a statusline tool must never return a
		// non-zero exit code or produce stderr noise, as that would break the
		// shell prompt rendering. Silent failure is the correct behavior here.
		os.Exit(0)
	}
}

func run() error {
	debug.Logf("main", "starting conductor-powerline")

	// 1. Parse stdin hook data
	hookData, err := hook.Parse(os.Stdin)
	if err != nil {
		return err
	}
	debug.Logf("main", "hook parsed: model=%s workspace=%s", hookData.ModelID(), hookData.WorkspacePath())

	// 2. Load config (project → user → defaults)
	projectCfg := filepath.Join(".", ".conductor-powerline.json")
	userCfg := ""
	if home, err := os.UserHomeDir(); err == nil {
		userCfg = filepath.Join(home, ".claude", "conductor-powerline.json")
	}
	cfg := config.Load(projectCfg, userCfg)
	debug.Logf("main", "config loaded: theme=%s segments=%v timeout=%v cacheTTL=%v", cfg.Theme, cfg.SegmentOrder, cfg.APITimeout.Duration, cfg.CacheTTL.Duration)

	// 3. Resolve theme
	theme, _ := themes.Get(cfg.Theme)

	// 4. Fetch usage data in parallel with segment building
	var usageData *oauth.UsageData
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		client := oauth.NewClient(anthropicUsageURL, cfg.APITimeout.Duration)
		cache := oauth.NewCache(cfg.CacheTTL.Duration)
		data, err := oauth.FetchUsage(client, cache)
		if err == nil {
			usageData = data
		} else {
			debug.Logf("main", "usage fetch failed: %v", err)
		}
		// On error, usageData remains nil → segments show "--" placeholder
	}()

	wg.Wait()

	if usageData != nil {
		debug.Logf("main", "usage data available: block=%.1f%% weekly=%.1f%% stale=%v", usageData.BlockPercentage, usageData.WeeklyPercentage, usageData.IsStale)
	} else {
		debug.Logf("main", "usage data is nil — segments will show '--'")
	}

	// 5. Build segments in configured order
	segs := buildSegments(cfg, hookData, theme, usageData)
	debug.Logf("main", "built %d segments", len(segs))

	// 6. Build right-side segments (context window)
	rightSegs := buildRightSegments(cfg, hookData, theme)
	debug.Logf("main", "built %d right segments", len(rightSegs))

	// 7. Render and output (no trailing newline)
	output := render.Render(segs, cfg.Display.NerdFontsEnabled(), cfg.Display.CompactWidth)
	rightOutput := render.RenderRight(rightSegs, cfg.Display.NerdFontsEnabled())
	fmt.Print(output + rightOutput)

	return nil
}

// rightSideSegments lists segment names that render on the right side.
var rightSideSegments = map[string]bool{
	"context": true,
}

func buildSegments(cfg config.Config, hookData hook.Data, theme themes.Theme, usageData *oauth.UsageData) []segments.Segment {
	builders := map[string]func() segments.Segment{
		"directory": func() segments.Segment {
			return segments.Directory(hookData.WorkspacePath(), theme)
		},
		"git": func() segments.Segment {
			return segments.Git(hookData.WorkspacePath(), theme)
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
		if rightSideSegments[name] {
			continue // Right-side segments rendered separately
		}
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

func buildRightSegments(cfg config.Config, hookData hook.Data, theme themes.Theme) []segments.Segment {
	segCfg, hasCfg := cfg.Segments["context"]
	if hasCfg && !segCfg.Enabled {
		return nil
	}

	seg := segments.Context(hookData.ContextPercent(), cfg.Display.NerdFontsEnabled(), theme)
	if !seg.Enabled {
		return nil
	}
	return []segments.Segment{seg}
}
