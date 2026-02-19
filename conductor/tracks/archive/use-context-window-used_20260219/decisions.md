# Architecture Decisions

## ADR-001: Prefer pre-calculated `used_percentage`, keep manual fallback

**Date:** 2026-02-19

**Decision:** When `context_window.used_percentage` is present and non-nil in the hook JSON, use it directly. Otherwise fall back to the existing manual calculation from `current_usage` token counts.

**Rationale:** The pre-calculated field is authoritative (computed by Claude Code itself using the same formula), simpler, and more robust. Keeping the fallback ensures compatibility with older Claude Code versions or early-session state where `used_percentage` may be null.

**Formula alignment:** Per docs, `used_percentage` uses `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` (not `output_tokens`), which matches the existing manual calculation — so switching produces identical results when both are present.
