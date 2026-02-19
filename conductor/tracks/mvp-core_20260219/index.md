# Track: MVP Core

## Overview

Set up the Go module foundation and implement core segments (directory, git, model) with powerline rendering and theming. This delivers a working statusline that can be integrated with Claude Code — without API calls yet.

## Scope

- Go module initialization (`go.mod`, `main.go`)
- Configuration system (types, loader, defaults, deep merge)
- Powerline renderer (ANSI colors, Nerd Font glyphs, text fallback, compact mode)
- Theme system (6 built-in themes)
- Stdin hook data parser
- Directory segment
- Git segment (branch + dirty state)
- Model segment
- Configurable segment ordering
- End-to-end integration: stdin → segments → render → stdout

## Documents

- [Specification](./spec.md) - User stories and acceptance criteria
- [Implementation Plan](./plan.md) - Task breakdown and progress
- [Code Review Report](./review.md) - Auto-generated review on track completion

## Out of Scope

- OAuth token retrieval
- API calls to Anthropic usage endpoint
- Block usage segment
- Weekly usage segment
- Usage trend tracking
