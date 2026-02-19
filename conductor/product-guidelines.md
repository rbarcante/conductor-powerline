# Product Guidelines

## Tone & Voice

- **Technical and concise** — minimal prose, maximum clarity
- Documentation reads like a well-written Go project: direct, factual, example-driven
- Error messages are actionable: state what happened, what the user can do
- No emojis in code or logs; Unicode powerline glyphs are for rendering only

## Naming Conventions

- **Binary name:** `conductor-powerline`
- **Config file:** `.conductor-powerline.json`
- **Go module:** `github.com/rbarcante/conductor-powerline`
- **Config search path:** `.conductor-powerline.json` (project) → `~/.claude/conductor-powerline.json` (user) → defaults

## UX Principles

- **Silent by default** — no output to stderr unless debug mode is enabled; statusline output goes to stdout only
- **Graceful degradation** — if OAuth fails, show `--` placeholders; if git is unavailable, skip the segment; never crash
- **Fast startup** — target sub-200ms; parallelize API calls and git commands
- **Respect terminal capabilities** — detect Nerd Font support via config, fall back to text separators

## Output Format

- Single-line stdout output with ANSI escape codes for colors
- Powerline-style segments with directional separators (`` or `|` fallback)
- Compact mode auto-activates below configurable terminal width threshold
- No trailing newline in output (Claude Code statusline expectation)

## Configuration Philosophy

- **Zero config works** — sensible defaults for everything
- **Progressive customization** — config file overrides only what you specify, deep-merged with defaults
- **Project-level overrides** — per-repo config takes precedence over user-level

## Error Handling

- API errors: use cached data if available, show `--` if not, log to debug
- Git errors: silently omit the git segment
- Config errors: log warning to debug, fall back to defaults
- Never panic in production; recover and degrade gracefully

## Versioning & Distribution

- Semantic versioning (semver)
- Primary distribution: `go install github.com/rbarcante/conductor-powerline@latest`
- Runtime usage: `go run github.com/rbarcante/conductor-powerline@latest`
- GitHub Releases with prebuilt binaries for macOS (arm64, amd64), Linux (amd64), Windows (amd64)
