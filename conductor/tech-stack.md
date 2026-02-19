# Tech Stack

## Primary Language

- **Go 1.23+** — latest stable release, leveraging modern stdlib features

## Frameworks

- None — pure Go standard library, no external frameworks

## Database

- None — no persistence layer; all data is fetched at runtime or cached in-memory

## Infrastructure

- **Distribution:** GitHub Releases with prebuilt binaries (goreleaser or manual)
- **CI/CD:** GitHub Actions — lint, test, build, release
- **Registry:** `pkg.go.dev` via Go module proxy

## Key Libraries

| Library | Purpose |
|---------|---------|
| `net/http` | OAuth API calls to Anthropic usage endpoint |
| `encoding/json` | Config file parsing, API response parsing, stdin hook data |
| `os/exec` | Git commands (branch, status), platform credential retrieval |
| `fmt` | ANSI escape code output for powerline rendering |
| `testing` | Standard library test framework |
| `os` | File system access, environment variables, stdin |
| `runtime` | Platform detection (GOOS) for credential store selection |
| `sync` | Concurrent API/git calls with WaitGroup |
| `time` | Block reset countdown, polling intervals |

## Architecture Decisions

### Decision 1: Pure stdlib, zero external dependencies

**Date:** 2026-02-19

**Decision:** No external dependencies — everything built on Go stdlib.

**Rationale:** Keeps the binary small, eliminates supply chain risk, ensures fast compilation, and aligns with Go philosophy of batteries-included stdlib. The tool's scope (HTTP calls, JSON, exec, ANSI output) is well-served by stdlib alone.

### Decision 2: Single main package for `go run` compatibility

**Date:** 2026-02-19

**Decision:** Entry point in `main` package at repo root with internal packages for organization.

**Rationale:** `go run github.com/rbarcante/conductor-powerline@latest` requires the module root to be a `main` package. Internal structure uses Go's `internal/` convention for clean separation without exposing public API.

### Decision 3: Platform-specific credential access via os/exec

**Date:** 2026-02-19

**Decision:** Use `os/exec` to call platform credential tools (`security` on macOS, `wincred` on Windows, `secret-tool` on Linux) rather than CGo bindings.

**Rationale:** Avoids CGo complexity and cross-compilation issues. The credential commands are stable, well-documented system utilities. Fallback to `~/.claude/.credentials.json` file covers edge cases.

---

## Development Environment

### Prerequisites

- Go 1.23 or later
- Git
- Nerd Font (optional, for powerline glyph rendering)
- Claude Code with active Pro/Team/Enterprise subscription (for OAuth token)

### Setup

```bash
# Clone and run
git clone https://github.com/rbarcante/conductor-powerline.git
cd conductor-powerline
go run .

# Run tests
go test ./...

# Build binary
go build -o conductor-powerline .

# Install globally
go install github.com/rbarcante/conductor-powerline@latest
```

### Project Structure

```
conductor-powerline/
├── main.go                     # Entry point — orchestrates segments and renders output
├── go.mod                      # Module definition
├── internal/
│   ├── config/                 # Config loading, defaults, deep merge
│   │   ├── config.go
│   │   ├── config_test.go
│   │   └── types.go
│   ├── segments/               # Individual segment providers
│   │   ├── block.go            # 5-hour block usage
│   │   ├── weekly.go           # 7-day rolling usage
│   │   ├── git.go              # Git branch + dirty state
│   │   ├── model.go            # Active Claude model
│   │   ├── directory.go        # Repo/directory name
│   │   └── *_test.go
│   ├── oauth/                  # OAuth token retrieval + API client
│   │   ├── oauth.go
│   │   ├── keychain.go         # macOS
│   │   ├── wincred.go          # Windows
│   │   ├── secretool.go        # Linux
│   │   └── *_test.go
│   ├── render/                 # Powerline rendering engine
│   │   ├── renderer.go
│   │   ├── renderer_test.go
│   │   └── symbols.go
│   ├── themes/                 # Color themes
│   │   ├── themes.go
│   │   └── themes_test.go
│   └── hook/                   # Stdin hook data parser
│       ├── hook.go
│       └── hook_test.go
└── conductor/                  # Conductor project management files
```
