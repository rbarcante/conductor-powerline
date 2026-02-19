# conductor-powerline

A fast, zero-dependency Go CLI that renders a powerline-style statusline for [Claude Code](https://docs.anthropic.com/en/docs/claude-code). It displays model info, git status, API usage (block/weekly), and context window usage with Nerd Font glyphs and configurable color themes.

## Features

- **Powerline rendering** with Nerd Font arrow separators (or plain-text fallback)
- **Live API usage** via Claude Code's OAuth token (macOS Keychain, Linux secret-tool, Windows Credential Manager)
- **6 built-in themes**: dark, light, nord, gruvbox, tokyo-night, rose-pine
- **Context window** usage indicator with threshold colors
- **Configurable segments**: directory, git, model, block (5h), weekly (7d), context
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
    "weekly": { "enabled": false }
  },
  "segmentOrder": ["directory", "git", "model", "block", "context"],
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
| `segmentOrder` | []string | all | Order of segments left-to-right |
| `apiTimeout` | duration | `"5s"` | HTTP timeout for usage API |
| `cacheTTL` | duration | `"30s"` | Cache lifetime for API responses |
| `trendThreshold` | float | `2.0` | Percentage change threshold for trend arrows |

### Themes

Available themes: `dark`, `light`, `nord`, `gruvbox`, `tokyo-night`, `rose-pine`

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
