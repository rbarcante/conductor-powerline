# conductor-powerline

[![CI](https://github.com/rbarcante/conductor-powerline/actions/workflows/ci.yml/badge.svg)](https://github.com/rbarcante/conductor-powerline/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/rbarcante/conductor-powerline)](https://github.com/rbarcante/conductor-powerline/releases/latest)

A fast, zero-dependency Go CLI that renders a powerline-style statusline for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

- Model info, git branch, directory, API usage (5h block / 7d rolling), context window
- 6 built-in themes — dark, light, nord, gruvbox, tokyo-night, rose-pine
- Nerd Font glyphs (with plain-text fallback)
- macOS Keychain, Linux secret-tool, Windows Credential Manager
- Silent failure — never crashes or pollutes your shell

## Prerequisites

- [Go 1.25+](https://go.dev/dl/) — needed to install from source
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) — the statusline hooks into its Notification event
- A [Nerd Font](https://www.nerdfonts.com/) — optional, falls back to plain text

## Quick start

```bash
go install github.com/rbarcante/conductor-powerline@latest
```

Add to your Claude Code settings (`.claude/settings.json`):

```json
{
  "hooks": {
    "Notification": [
      {
        "type": "command",
        "command": "echo '$CLAUDE_NOTIFICATION' | conductor-powerline"
      }
    ]
  }
}
```

Test it:

```bash
echo '{"model":"claude-sonnet-4-20250514"}' | conductor-powerline
```

## Segments

| Segment | Description |
|---------|-------------|
| `directory` | Current project/directory name |
| `git` | Branch name with dirty-state indicator |
| `model` | Active Claude model (Opus, Sonnet, Haiku) |
| `block` | 5-hour block usage percentage and time remaining |
| `weekly` | 7-day rolling usage percentage |
| `context` | Context window usage with threshold colors |
| `conductor` | Conductor plugin status / "Try Conductor" hyperlink |

## Configuration

Loaded in order (later overrides earlier):

1. Built-in defaults
2. User config: `~/.claude/conductor-powerline.json`
3. Project config: `./.conductor-powerline.json`

```json
{
  "theme": "nord",
  "display": {
    "nerdFonts": true,
    "compactWidth": 100
  },
  "segments": {
    "directory": { "enabled": true },
    "git": { "enabled": true },
    "model": { "enabled": true },
    "block": { "enabled": true },
    "weekly": { "enabled": false },
    "context": { "enabled": true },
    "conductor": { "enabled": true }
  },
  "segmentOrder": ["directory", "git", "model", "block", "weekly", "context", "conductor"],
  "apiTimeout": "5s",
  "cacheTTL": "30s",
  "trendThreshold": 2.0
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `theme` | string | `"dark"` | Color theme name |
| `display.nerdFonts` | bool | `true` | Use Nerd Font glyphs |
| `display.compactWidth` | int | `100` | Truncate segments when total width exceeds this |
| `segments.<name>.enabled` | bool | `true` | Enable/disable individual segments |
| `segmentOrder` | []string | *(all)* | Order of segments left-to-right |
| `apiTimeout` | duration | `"5s"` | HTTP timeout for usage API |
| `cacheTTL` | duration | `"30s"` | Cache lifetime for API responses |
| `trendThreshold` | float | `2.0` | Percentage change threshold for trend arrows |

Themes: `dark` `light` `nord` `gruvbox` `tokyo-night` `rose-pine`

## tmux

Works inside tmux. For OSC 8 hyperlink support (tmux 3.1+), add to `.tmux.conf`:

```
set -as terminal-features ",*:hyperlinks"
```

> **Note:** Hyperlinks in the conductor segment are not clickable inside tmux due to a [Claude Code limitation](https://github.com/anthropics/claude-code/issues/27047). The URL is shown as plain text instead.

## Development

```bash
make test          # Run all tests
make test-coverage # Generate HTML coverage report
make lint          # Run golangci-lint
make fmt           # Format code
make vet           # Run go vet
```

Or build from source:

```bash
git clone https://github.com/rbarcante/conductor-powerline.git
cd conductor-powerline
make build
```

## License

[MIT](LICENSE)
