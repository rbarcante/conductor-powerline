# Spec: Use `context_window.used_percentage` from Claude Code Statusline JSON

## Overview

Claude Code's statusline JSON provides a pre-calculated `context_window.used_percentage` field (and `remaining_percentage`) that is authoritative and already validated by Claude Code itself. Currently, `hook.go` manually derives the percentage from raw token counts (`input_tokens + cache_creation_input_tokens + cache_read_input_tokens`) divided by `context_window_size`. This refactor eliminates the manual calculation by preferring the pre-calculated field, keeping the manual path as a fallback for older/incomplete payloads.

**Reference:** https://code.claude.com/docs/en/statusline.md — `context_window.used_percentage` is described as "Pre-calculated percentage of context window used".

## Functional Requirements

1. **Parse `used_percentage`:** The `ContextWindow` struct in `internal/hook/hook.go` must parse the `context_window.used_percentage` JSON field. The field is a nullable float (`*float64`) because it may be `null` before the first API call.
2. **Parse `remaining_percentage`:** Similarly, parse `context_window.remaining_percentage` (`*float64`, nullable).
3. **Prefer pre-calculated value:** `ContextPercent()` must return `int(math.Round(*used_percentage))` when `used_percentage` is non-nil, bypassing the manual token-count calculation.
4. **Fallback to manual calculation:** When `used_percentage` is `nil` (older Claude Code versions or early-session state), `ContextPercent()` falls back to the existing `current_usage`-based calculation — but only when `ContextWindowSize > 0`.
5. **Return -1 for fully absent data:** When both `used_percentage` is nil and `current_usage`/`context_window_size` provide no usable data, `ContextPercent()` returns `-1` (unchanged sentinel).

## Non-Functional Requirements

- **No breaking changes to public API** — `ContextPercent() int` signature is unchanged.
- **Zero new dependencies** — pure stdlib, no external packages.
- **All existing tests must continue to pass** — fallback path is preserved.
- **>80% code coverage** maintained.

## Acceptance Criteria

- [ ] `ContextWindow` struct has `UsedPercentage *float64` and `RemainingPercentage *float64` fields.
- [ ] `ContextPercent()` returns the pre-calculated value (rounded) when `used_percentage` is present and non-nil.
- [ ] `ContextPercent()` falls back to manual calculation when `used_percentage` is nil.
- [ ] `ContextPercent()` returns -1 when both paths yield no data.
- [ ] New tests cover: pre-calculated path, null value (fallback), and absence of both paths.
- [ ] All existing `hook_test.go` tests pass without modification.
- [ ] `go test ./...` passes with no failures.

## Out of Scope

- Changes to `segments/context.go` — the `Context(percent int, ...)` function signature is unchanged.
- Changes to `main.go` call sites — `ContextPercent()` return type/semantics are unchanged.
- Exposing `RemainingPercentage` as a new segment — future track if desired.
- Any UI/display changes.
