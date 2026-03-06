// conductor-powerline is a fast powerline statusline for Claude Code.
// It reads hook JSON from stdin, loads configuration, builds segments,
// and renders ANSI-colored output to stdout.
package main

import (
	"context"
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

	// 4. Detect conductor status once (used for right segments and line 2 visibility)
	// Prefer hookData.WorkspacePath() (explicit project from Claude Code hook JSON) with
	// os.Getwd() as fallback for local config loading.
	cwd, _ := os.Getwd()
	workspace := hookData.WorkspacePath()
	if workspace == "" {
		workspace = cwd
	}
	conductorStatus := segments.DetectConductorStatus("", workspace)
	debug.Logf("main", "conductor status: %d (workspace=%s)", conductorStatus, workspace)

	// 5. Fetch usage data and workflow data concurrently
	var usageData *oauth.UsageData
	var workflowData *segments.WorkflowData
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		client := oauth.NewClient(anthropicUsageURL, cfg.APITimeout.Duration)
		cache := oauth.NewFileCache(cacheDir(), cfg.CacheTTL.Duration)
		workspace := hookData.WorkspacePath()
		data, err := oauth.FetchUsage(client, cache, workspace)
		if err == nil {
			usageData = data
		} else {
			debug.Logf("main", "usage fetch failed: %v", err)
		}
		// On error, usageData remains nil → segments show "--" placeholder
	}()

	// Fetch workflow data concurrently when conductor is active
	if conductorStatus == segments.ConductorActive {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), cfg.APITimeout.Duration)
			defer cancel()
			home, _ := os.UserHomeDir()
			data, err := segments.FetchWorkflowStatus(ctx, home, workspace)
			if err == nil {
				workflowData = data
			} else {
				debug.Logf("main", "workflow fetch failed: %v", err)
			}
		}()
	}

	wg.Wait()

	if usageData != nil {
		debug.Logf("main", "usage data available: block=%.1f%% weekly=%.1f%% stale=%v", usageData.BlockPercentage, usageData.WeeklyPercentage, usageData.IsStale)
	} else {
		debug.Logf("main", "usage data is nil — segments will show '--'")
	}

	// 6. Build line 1 segments in configured order
	segs := buildSegments(cfg, hookData, theme, usageData)
	debug.Logf("main", "built %d segments", len(segs))

	// 7. Build right-side segments (context window + conductor indicator)
	rightSegs := buildRightSegments(cfg, conductorStatus, theme, hookData)
	debug.Logf("main", "built %d right segments", len(rightSegs))

	// 8. Render line 1
	output := render.Render(segs, cfg.Display.NerdFontsEnabled(), cfg.Display.CompactWidth)
	rightOutput := render.RenderRight(rightSegs, cfg.Display.NerdFontsEnabled())

	// 9. Build and render line 2 (conductor workflow) when conditions are met
	workflowEnabled := true
	if wfCfg, ok := cfg.Segments["conductor_workflow"]; ok {
		workflowEnabled = wfCfg.Enabled
	}
	debug.Logf("main", "line2 conditions: conductorActive=%v workflowData=%v workflowEnabled=%v",
		conductorStatus == segments.ConductorActive, workflowData != nil, workflowEnabled)

	if conductorStatus == segments.ConductorActive && workflowData != nil && workflowEnabled {
		line2Segs := buildWorkflowSegments(workflowData, cfg, theme)
		debug.Logf("main", "built %d line2 workflow segments", len(line2Segs))
		line2Output := render.Render(line2Segs, cfg.Display.NerdFontsEnabled(), cfg.Display.CompactWidth)
		fmt.Print(output + rightOutput + "\n" + line2Output)
	} else {
		fmt.Print(output + rightOutput)
	}

	return nil
}

// rightSideSegments lists segment names that render on the right side of line 1.
var rightSideSegments = map[string]bool{
	"context":   true,
	"conductor": true,
}

// line2Segments lists segment names that belong to line 2 (not rendered on line 1).
var line2Segments = map[string]bool{
	"conductor_workflow": true,
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
		if rightSideSegments[name] || line2Segments[name] {
			continue // Rendered separately
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

// cacheDir returns the cache directory for conductor-powerline.
// Uses $XDG_CACHE_HOME/conductor-powerline if set, otherwise ~/.cache/conductor-powerline.
func cacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "conductor-powerline")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache", "conductor-powerline")
	}
	return filepath.Join(os.TempDir(), "conductor-powerline")
}

func buildRightSegments(cfg config.Config, conductorStatus segments.ConductorStatus, theme themes.Theme, hookData hook.Data) []segments.Segment {
	var result []segments.Segment

	// Context segment (leftmost right-side segment)
	ctxCfg, hasCfg := cfg.Segments["context"]
	if !hasCfg || ctxCfg.Enabled {
		seg := segments.Context(hookData.ContextPercent(), cfg.Display.NerdFontsEnabled(), theme)
		if seg.Enabled {
			result = append(result, seg)
		}
	}

	// Conductor segment (rightmost — after context)
	condCfg, hasCfg := cfg.Segments["conductor"]
	if !hasCfg || condCfg.Enabled {
		seg := segments.Conductor(conductorStatus, cfg.Display.NerdFontsEnabled(), theme)
		if seg.Enabled {
			result = append(result, seg)
		}
	}

	return result
}

// buildWorkflowSegments constructs the four line-2 workflow segments.
func buildWorkflowSegments(data *segments.WorkflowData, cfg config.Config, theme themes.Theme) []segments.Segment {
	nerdFonts := cfg.Display.NerdFontsEnabled()
	return []segments.Segment{
		segments.WorkflowSetup(data, theme),
		segments.WorkflowTrack(data, theme),
		segments.WorkflowTasks(data, nerdFonts, theme),
		segments.WorkflowOverall(data, nerdFonts, theme),
	}
}
