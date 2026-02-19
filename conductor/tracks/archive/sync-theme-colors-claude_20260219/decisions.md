# Architectural Decision Records

## ADR-001: Unified Warning/Critical Color Keys

**Date:** 2026-02-19
**Status:** Accepted

### Context

The current `themes.go` uses per-segment warning/critical color keys (e.g., `block-warning`, `block-critical`, `context-warning`, `context-critical`).

### Decision

Adopt unified `warning`/`critical` keys shared across all segments.

### Rationale

- Reduces duplication (2 keys instead of 4+ per theme)
- Simpler to maintain and extend
- Warning/critical colors are meant to convey urgency levels, not segment identity

### Consequences

- `block.go` and `context.go` must be updated to reference `warning`/`critical` instead of segment-specific keys
- All corresponding tests must be updated
