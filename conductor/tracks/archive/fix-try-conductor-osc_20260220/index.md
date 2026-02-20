# Track: Fix Try Conductor OSC 8 Hyperlink in tmux

**ID:** fix-try-conductor-osc_20260220
**Type:** bugfix
**Status:** planned
**Created:** 2026-02-20
**Branch:** bugfix/tmux-osc8-hyperlink

## Summary

The "Try Conductor" segment's OSC 8 hyperlink renders as non-clickable plain text inside tmux. The fix removes the DCS passthrough wrapping and emits standard OSC 8 sequences, which modern tmux (3.1+) handles natively.

## Files

- [spec.md](./spec.md) — Specification
- [plan.md](./plan.md) — Implementation plan
- [decisions.md](./decisions.md) — Architectural decisions
- [metadata.json](./metadata.json) — Track metadata
- [Code Review Report](./review.md) - Auto-generated review on track completion
