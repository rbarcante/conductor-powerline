# Spec: Statusline Not Showing on Claude Code

## Overview

The conductor-powerline statusline renders blank in Claude Code despite `go run .` working standalone. The root cause is a mismatch between the Claude Code statusline stdin JSON schema and the hook parser's expected types. Claude Code sends `model` as an object (`{"id": "...", "display_name": "..."}`) but the parser expects a plain string. Similarly, `workspace` is sent as an object (`{"current_dir": "...", "project_dir": "..."}`) but parsed as a string. This causes `json.Unmarshal` to fail silently (returning zero-value Data), resulting in empty segments.

## Functional Requirements

1. **FR-1: Update hook parser to match Claude Code stdin schema** — The `hook.Data` struct must accept the actual JSON shape Claude Code sends: `model` as an object with `id` and `display_name` fields, `workspace` as an object with `current_dir` and `project_dir` fields.

2. **FR-2: Map parsed fields to segment inputs** — Extract `model.display_name` (or `model.id` as fallback) for the model segment. Extract `workspace.project_dir` (or `workspace.current_dir`) for the directory segment.

3. **FR-3: Parse additional useful fields** — Accept and expose `cwd`, `session_id`, `version`, `cost`, and `context_window` from the stdin JSON for future segment use.

4. **FR-4: Handle backward compatibility** — If `model` or `workspace` arrive as plain strings (standalone testing), gracefully handle both shapes.

## Non-Functional Requirements

1. **NFR-1:** Startup time must remain under 200ms
2. **NFR-2:** No external dependencies — use stdlib only
3. **NFR-3:** Never crash or produce stderr output; degrade gracefully on parse errors

## Acceptance Criteria

- [ ] Claude Code displays the powerline statusline with visible segments (directory, git, model at minimum)
- [ ] The model segment shows the friendly model name (e.g., "Opus" not "claude-opus-4-6")
- [ ] The directory segment shows the project directory name
- [ ] Running `echo '{}' | go run .` still produces output (empty input graceful handling)
- [ ] All existing tests pass; new tests cover the updated hook parser
- [ ] Test coverage >80% for modified packages

## Out of Scope

- Adding new segments for `cost` or `context_window` (future track)
- Switching from `go run` to a prebuilt binary
- Changes to the rendering engine or theme system
