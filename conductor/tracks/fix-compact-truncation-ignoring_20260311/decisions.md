# Decisions: Fix Compact Truncation Ignoring CompactWidth Config

## ADR-001: Proportional Truncation Strategy

**Date:** 2026-03-11
**Status:** Accepted

**Context:** When compact mode triggers, segments need to be truncated to fit within `CompactWidth`. The previous approach used a hardcoded 12-character limit for all segments.

**Decision:** Use proportional truncation where each segment gets a share of the available character budget proportional to its original text length. Minimum floor of 3 characters per segment.

**Consequences:**
- Longer segments get more space than shorter ones
- All segments remain readable (minimum 3 chars)
- The total rendered width approximates the configured `CompactWidth`
