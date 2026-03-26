# Specification: Fix Compact Truncation Ignoring CompactWidth Config

## Overview

Fix the compact mode truncation bug where segment text is always truncated to a hardcoded 12 characters (`maxCompactTextLen = 12`) regardless of the `CompactWidth` configuration value. When `CompactWidth` is set to a large value (e.g., 200), segments should not be truncated if the total width fits. When truncation IS needed, segments should be proportionally shrunk to fit within the configured width.

## Bug Description

- **Root Cause:** `renderer.go` line 15 defines `const maxCompactTextLen = 12`. Line 69 uses this constant to truncate all segment text when compact mode triggers, ignoring the actual `CompactWidth` config value.
- **Expected:** `CompactWidth` controls both *whether* to compact AND *how much* to truncate.
- **Actual:** `CompactWidth` only controls whether to compact; truncation is always to 12 chars.

## Functional Requirements

1. **FR-1:** When total segment width fits within `CompactWidth`, no truncation occurs (existing behavior is correct here via `shouldCompact()`).
2. **FR-2:** When total segment width exceeds `CompactWidth`, segments are proportionally truncated so the rendered line fits approximately within `CompactWidth` characters.
3. **FR-3:** Remove the hardcoded `maxCompactTextLen = 12` constant. Truncation limits must be derived from `CompactWidth` and the number/size of active segments.
4. **FR-4:** Each segment gets a share of the available width proportional to its original text length relative to the total text length.
5. **FR-5:** Minimum truncation length per segment should be 3 characters (to always show at least some text + ellipsis).

## Non-Functional Requirements

- No new dependencies
- Maintain existing test coverage targets (>80%)
- No performance regression in rendering

## Acceptance Criteria

- [ ] Setting `CompactWidth: 200` with a short powerline shows full untruncated text
- [ ] Setting `CompactWidth: 40` with a wide powerline proportionally truncates segments
- [ ] No segment is truncated below 3 characters
- [ ] Existing compact mode tests updated to reflect new behavior
- [ ] `maxCompactTextLen` constant is removed

## Out of Scope

- Right-side segments (already no compact mode)
- Second-line (conductor workflow) rendering
- Configuration schema changes (CompactWidth field already exists)
