# Track: Statusline Not Showing on Claude Code

- **ID:** statusline-not-showing-claude_20260219
- **Type:** bugfix
- **Status:** planned
- **Priority:** high
- **Created:** 2026-02-19
- **Branch:** bugfix/statusline-not-showing

## Summary

The conductor-powerline statusline renders blank in Claude Code despite `go run .` working standalone. The root cause is a mismatch between the Claude Code statusline stdin JSON schema and the hook parser's expected types.

## Files

- [Specification](spec.md)
- [Plan](plan.md)
- [Decisions](decisions.md)
