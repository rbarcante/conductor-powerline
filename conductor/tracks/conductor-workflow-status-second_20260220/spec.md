# Specification: Conductor Workflow Status Second Line

## Overview

Add a second powerline line that displays real-time Conductor workflow status when the Conductor plugin is active in the current project. The line uses matching powerline segment style (same arrows, padding, and theme colors as line 1) and is only visible when ConductorActive status is detected.

## Data Source

Execute `conductor_cli.py --json status` and parse the JSON output. The CLI provides:
- **Setup validity** and completion percentage
- **Active track** name/ID and task progress (completed/total)
- **Overall progress** across all tracks (completed/total tasks, percentage)

## Segments (left-to-right)

| Segment | Key | Example | Color Source |
|---------|-----|---------|-------------|
| Setup status | `workflow_setup` | `Setup 100%` | Theme key: `workflow_setup` |
| Current track | `workflow_track` | `auth-flow` | Theme key: `workflow_track` |
| Track tasks | `workflow_tasks` | `12/35` | Theme key: `workflow_tasks` |
| Overall tracks | `workflow_overall` | `9/9 tracks` | Theme key: `workflow_overall` |

## Functional Requirements

1. **CLI Execution**: Run `conductor_cli.py --json status` via `os/exec`, parse JSON response
2. **Visibility**: Only render line 2 when ConductorActive (plugin installed + `conductor/` dir exists)
3. **Active Track Detection**: Show the first `in_progress` track; if none, show the most recently updated track
4. **Theme Integration**: Add 4 new color keys to all 6 themes (`workflow_setup`, `workflow_track`, `workflow_tasks`, `workflow_overall`), following each theme's palette
5. **Timeout**: CLI call respects `apiTimeout` config (default 5s); on failure/timeout, silently skip line 2
6. **Config**: New `"conductor_workflow"` segment config to enable/disable the entire second line
7. **Compact mode**: Apply same compact truncation rules as line 1
8. **Output**: Append `\n` after line 1, then render line 2 with the same `Render()` function
9. **Debug Logging**: All operations emit debug logs via `debug.Logf()` following existing conventions — CLI execution start/result, parse outcomes, segment build counts, visibility decisions
10. **Activation Criteria**: Line 2 is rendered only when ALL conditions are met:
    - ConductorActive status detected (plugin in registry + conductor/ dir in project)
    - CLI execution succeeds (exit code 0, valid JSON)
    - `conductor_workflow` segment is enabled in config (default: true)

## Non-Functional Requirements

- Sub-200ms target maintained (CLI call runs concurrently with API usage fetch)
- Zero external dependencies (uses `os/exec` + stdlib JSON parsing)
- Silent failure — if CLI not found or fails, skip line 2 entirely

## Acceptance Criteria

- [ ] Second line renders only when Conductor is active
- [ ] All 4 segments display correct data from CLI JSON
- [ ] All 6 themes have matching workflow color keys
- [ ] Second line hidden when CLI fails, times out, or Conductor inactive
- [ ] Configurable via `segments.conductor_workflow.enabled`
- [ ] Debug logs emitted for CLI execution, parsing, segment building, and visibility decisions
- [ ] Tests cover: CLI parsing, segment building, visibility logic, theme colors

## Out of Scope

- Real-time task updates (relies on CLI snapshot at render time)
- Interactive track switching from the statusline
- Custom segment ordering for line 2
