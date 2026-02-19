# Product Definition

## Initial Concept

A Claude Code powerline/statusline tool written entirely in Go. following Go idioms and conventions. Installable via `go run github.com/rbarcante/conductor-powerline@latest`. Displays Claude Code API usage (5-hour block, 7-day rolling), git branch info, active model, and more — targeting Claude Code power users who take AI-assisted coding seriously.

## Vision

A fast, zero-dependency powerline statusline for Claude Code — written in idiomatic Go — that gives developers instant visibility into their API usage, active model, and git context, installable with a single `go run` command.

## Problem Statement

Claude Code users on Pro/Team/Enterprise plans have usage limits (5-hour blocks and 7-day rolling windows) but no built-in way to see how much they've consumed at a glance. Switching between the dashboard and their terminal breaks flow. Developers need a lightweight, always-visible statusline that shows usage, model, and git context without leaving their coding environment.

## Target Users

- **Claude Code power users** on Pro, Team, or Enterprise plans who code daily with Claude
- **Developers who value terminal aesthetics** and use powerline-style prompts (Nerd Fonts, custom shells)
- **Go ecosystem users** who prefer `go install`/`go run` over `npm`/`npx` for CLI tools

## Success Criteria

- Full segment, theme, and config support
- Single-binary distribution — no runtime dependencies
- Startup time under 200ms (Go advantage over Node.js cold start)
- Cross-platform: macOS, Linux, Windows
- Installable via `go run github.com/rbarcante/conductor-powerline@latest`
- Comprehensive test coverage (80%+)

## Core Features

1. **5-Hour Block Usage** — Shows current utilization percentage and time remaining until reset
2. **7-Day Rolling Usage** — Weekly usage with week-progress indicator and smart mode (Opus/Sonnet breakdown)
3. **Git Integration** — Current branch name with dirty-state indicator
4. **Model Display** — Active Claude model (Opus, Sonnet, Haiku) with friendly names
5. **Directory/Repo Name** — Current project name segment
6. **Theming** — 6 built-in themes (dark, light, nord, gruvbox, tokyo-night, rose-pine) with ANSI color rendering
7. **Configuration** — JSON config file support (project-level and user-level) with deep merge
8. **Powerline Rendering** — Nerd Font glyphs with text fallback, compact mode for narrow terminals
9. **Cross-Platform OAuth** — Token retrieval from macOS Keychain, Windows Credential Manager, Linux secret-tool, and credential file fallback
10. **Usage Trends** — Directional arrows showing usage change since last poll
11. **Configurable Segment Order** — Users can reorder segments via config
12. **Stdin Hook Data** — Reads Claude Code hook JSON from stdin for model/workspace context

## Non-Goals

- GUI or TUI application — this is a single-line stdout tool
- Replacing Claude Code's built-in features — only augments the statusline
- Supporting non-Claude AI providers
- Historical usage analytics or data persistence beyond trend arrows

## Constraints

- Must work as a Claude Code statusline command (stdout, no interactive input)
- Must read OAuth tokens from platform credential stores (no hardcoded tokens)
- Must handle API failures gracefully (show cached/fallback data, never crash)
- Go module path: `github.com/rbarcante/conductor-powerline`
- Single `main` package for `go run` compatibility
