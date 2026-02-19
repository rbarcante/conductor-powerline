# Specification: Detect Conductor Plugin and Suggest Installation

## Overview

Add a new `conductor` segment to the powerline statusline that detects whether the `claude-conductor` plugin (by `rbarcante`) is installed — either as a local plugin or via the Claude Code marketplace. The segment always displays: showing a checkmark when installed, or a suggestion prompt with a clickable link when not found.

## Functional Requirements

1. **Plugin Detection** — Check `~/.claude/` directories for the presence of the `claude-conductor` plugin:
   - Scan `~/.claude/plugins/` for local plugin installs (look for directories/files matching `claude-conductor` or `rbarcante/claude-conductor`)
   - Scan `~/.claude/marketplace/` (or equivalent marketplace directory) for marketplace installs
   - Detection must be fast (file-system stat only, no network calls)

2. **New `conductor` Segment** — A new segment provider in `internal/segments/conductor.go`:
   - When installed: display `✓ Conductor` (or Nerd Font equivalent) with a "success" color from the active theme
   - When NOT installed: display `⚡ Get Conductor` with an "attention" color from the active theme, wrapped in an **OSC 8 terminal hyperlink** pointing to `https://github.com/rbarcante/claude-conductor`
   - OSC 8 format: `\033]8;;URL\033\\TEXT\033]8;;\033\\` — gracefully ignored by terminals that don't support it
   - Segment name: `"conductor"`

3. **Segment Order Integration** — Add `"conductor"` to the default `segmentOrder` in config defaults, positioned after `model` and before `block`

4. **Configurable** — Follows existing segment config pattern:
   - Can be disabled via `{"segments": {"conductor": {"enabled": false}}}`
   - Respects the same enable/disable pattern as all other segments

5. **Theme Integration** — Use existing theme colors:
   - Installed state: use a green/success-type color from the theme palette
   - Not-installed state: use a yellow/warning-type color from the theme palette

## Non-Functional Requirements

- Detection must add < 5ms to startup time (filesystem stat only)
- Zero external dependencies (consistent with project philosophy)
- Cross-platform: must work on macOS, Linux, and Windows (`~/.claude/` path resolution via `os.UserHomeDir()`)
- Follow existing segment patterns for testability (dependency injection for filesystem checks)

## Acceptance Criteria

- [ ] `conductor` segment appears in the statusline by default
- [ ] Shows installed status (`✓ Conductor`) when plugin is found in `~/.claude/plugins/` or `~/.claude/marketplace/`
- [ ] Shows suggestion (`⚡ Get Conductor`) when plugin is not found
- [ ] "Get Conductor" text includes an OSC 8 hyperlink to `https://github.com/rbarcante/claude-conductor`
- [ ] Hyperlink degrades gracefully in terminals without OSC 8 support (text still displays normally)
- [ ] Segment is configurable (can be disabled)
- [ ] Theme colors are applied correctly for both states
- [ ] Works on macOS, Linux, and Windows
- [ ] Unit tests cover: installed detection, not-installed detection, config disable, theme colors
- [ ] Code coverage > 80%

## Out of Scope

- Automatic installation of the conductor plugin
- Network-based checks (e.g., querying GitHub for the plugin)
- Deep validation of plugin integrity (just presence detection)
- Clickable links or interactive elements beyond OSC 8 (stdout-only statusline)
