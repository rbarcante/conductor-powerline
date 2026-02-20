# conductor-powerline

[![CI](https://github.com/rbarcante/conductor-powerline/actions/workflows/ci.yml/badge.svg)](https://github.com/rbarcante/conductor-powerline/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/rbarcante/conductor-powerline)](https://github.com/rbarcante/conductor-powerline/releases/latest)

A fast, zero-dependency Go CLI that renders a powerline-style statusline for [Claude Code](https://docs.anthropic.com/en/docs/claude-code). It displays model info, git status, API usage (block/weekly), and context window usage with Nerd Font glyphs and configurable color themes.

## Features

- **Powerline rendering** with Nerd Font arrow separators (or plain-text fallback)
- **Live API usage** via Claude Code's OAuth token (macOS Keychain, Linux secret-tool, Windows Credential Manager)
- **6 built-in themes**: dark, light, nord, gruvbox, tokyo-night, rose-pine
- **Context window** usage indicator with threshold colors
- **Conductor plugin detection** with "Try Conductor" prompt and hyperlink
- **Configurable segments**: directory, git, model, block (5h), weekly (7d), context, conductor
- **Zero dependencies** outside the Go standard library
- **Silent failure** — never crashes or pollutes your shell

## Installation

```bash
go install github.com/rbarcante/conductor-powerline@latest
```

Or build from source:

```bash
git clone https://github.com/rbarcante/conductor-powerline.git
cd conductor-powerline
make build
```

## Usage

conductor-powerline reads Claude Code hook JSON from stdin and outputs ANSI-colored powerline text to stdout. Configure it as a Claude Code hook:

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

### Quick test

```bash
echo '{"model":"claude-sonnet-4-20250514"}' | conductor-powerline
```

## Segments

| Segment | Position | Description |
|---------|----------|-------------|
| `directory` | left | Current project/directory name |
| `git` | left | Branch name with dirty-state indicator |
| `model` | left | Active Claude model (Opus, Sonnet, Haiku) |
| `block` | left | 5-hour block usage percentage and time remaining |
| `weekly` | left | 7-day rolling usage percentage |
| `context` | right | Context window usage with threshold colors |
| `conductor` | right | Conductor plugin status with "Try Conductor" hyperlink |

### Hyperlinks

The `conductor` segment includes an OSC 8 hyperlink when displaying "Try Conductor". Outside tmux, the text is underlined and clickable. Inside tmux, the URL is shown as plain text instead, since Claude Code does not currently forward OSC 8 hyperlinks in tmux ([tracking issue](https://github.com/anthropics/claude-code/issues/27047)).

## Configuration

Configuration is loaded in order (later overrides earlier):

1. **Defaults** (built-in)
2. **User config**: `~/.claude/conductor-powerline.json`
3. **Project config**: `./.conductor-powerline.json`

### Example config

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

### Config fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `theme` | string | `"dark"` | Color theme name |
| `display.nerdFonts` | bool | `true` | Use Nerd Font glyphs |
| `display.compactWidth` | int | `100` | Truncate segments when total width exceeds this |
| `segments.<name>.enabled` | bool | `true` | Enable/disable individual segments |
| `segmentOrder` | []string | see below | Order of segments left-to-right |
| `apiTimeout` | duration | `"5s"` | HTTP timeout for usage API |
| `cacheTTL` | duration | `"30s"` | Cache lifetime for API responses |
| `trendThreshold` | float | `2.0` | Percentage change threshold for trend arrows |

Default segment order: `directory`, `git`, `model`, `block`, `weekly`, `context`, `conductor`

### Themes

Available themes: `dark`, `light`, `nord`, `gruvbox`, `tokyo-night`, `rose-pine`

## tmux

conductor-powerline works inside tmux. For best results:

- **tmux 3.1+** is required for OSC 8 hyperlink support
- Add to your `.tmux.conf`:
  ```
  set -as terminal-features ",*:hyperlinks"
  ```
- Restart tmux after adding the config (`tmux kill-server`)

**Note:** Hyperlinks in the conductor segment are currently not clickable when rendered through Claude Code's statusline inside tmux due to a Claude Code limitation. The URL is displayed as plain text instead. Outside tmux, hyperlinks are underlined and clickable.

## Development

```bash
make test          # Run all tests
make test-coverage # Generate HTML coverage report
make lint          # Run golangci-lint
make fmt           # Format code
make vet           # Run go vet
```

## License

[MIT](LICENSE)
