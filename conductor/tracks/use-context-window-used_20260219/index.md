# Track: Use `context_window.used_percentage`

**ID:** `use-context-window-used_20260219`
**Type:** refactor
**Status:** planned
**Branch:** `refactor/context-window-used-percentage`
**Created:** 2026-02-19

## Summary

Refactor `internal/hook/hook.go` to parse and prefer the pre-calculated `context_window.used_percentage` field from Claude Code's statusline JSON, eliminating the manual token-count derivation while preserving it as a fallback.

## Files

- [spec.md](./spec.md) — Requirements and acceptance criteria
- [plan.md](./plan.md) — Implementation plan
- [decisions.md](./decisions.md) — Architecture decisions
