# Spec: MVP Core

## User Stories

1. **As a Claude Code user**, I want to see my current git branch and dirty state in the statusline so I know my repo context at a glance.
2. **As a Claude Code user**, I want to see the active Claude model (Opus, Sonnet, Haiku) so I know which model is responding.
3. **As a Claude Code user**, I want to see the project/directory name in the statusline.
4. **As a Claude Code user**, I want to choose from multiple color themes to match my terminal aesthetic.
5. **As a Claude Code user**, I want to configure the statusline via a JSON file (toggle segments, change theme, reorder segments).
6. **As a developer**, I want to install and run this with `go run github.com/rbarcante/conductor-powerline@latest`.

## Acceptance Criteria

### Go Module & Entry Point
- `go.mod` with module path `github.com/rbarcante/conductor-powerline`
- `main.go` that reads stdin, loads config, builds segments, renders output to stdout
- Exits cleanly with no output on error (statusline must never crash)

### Configuration System (`internal/config/`)
- `types.go`: Config struct with all fields (display, segments, theme, segment order)
- `config.go`: Load from `.conductor-powerline.json` (project) → `~/.claude/conductor-powerline.json` (user) → defaults
- Deep merge: user config overrides only specified fields
- `config_test.go`: Tests for loading, merging, defaults

### Stdin Hook Data Parser (`internal/hook/`)
- Parse JSON from stdin: extract `model`, `workspace`, `context` fields
- Handle empty stdin, malformed JSON, missing fields gracefully
- `hook_test.go`: Tests for parsing, edge cases

### Segments (`internal/segments/`)
- **directory.go**: Extract repo/directory name from workspace path or cwd
- **git.go**: Run `git rev-parse --abbrev-ref HEAD` for branch, `git status --porcelain` for dirty state
- **model.go**: Map model identifiers to friendly names (e.g., `claude-opus-4-6` → `Opus 4.6`)
- Each segment returns a `Segment` struct with `Text`, `Color` (from theme)
- Each segment can be disabled via config
- `*_test.go` for each segment

### Theme System (`internal/themes/`)
- 6 themes: dark, light, nord, gruvbox, tokyo-night, rose-pine
- Each theme defines fg/bg colors for: directory, git, model, block, weekly, opus, sonnet, warning, critical
- `themes.go`: Theme registry, lookup by name, fallback to dark
- `themes_test.go`: All themes defined, color values valid

### Powerline Renderer (`internal/render/`)
- `symbols.go`: Nerd Font powerline separator (``) and text fallback (`|`)
- `renderer.go`: Takes ordered `[]Segment`, produces ANSI-colored powerline string
- Compact mode: truncate long segment text when terminal width below threshold
- No trailing newline
- `renderer_test.go`: Rendering with/without Nerd Fonts, compact mode, empty segments

### Integration
- `main.go` orchestrates: read stdin → load config → resolve theme → build segments → render → stdout
- Segment order configurable via `segmentOrder` config field
- Default order: `["directory", "git", "model"]`
