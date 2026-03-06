# Spec: conductorStatus first-line visibility

## Overview

The conductor segment on line 1 (right side) currently always renders regardless of conductor status. This change restricts it to only render when conductor is **not installed** — i.e., `ConductorNone` or `ConductorMarketplace`. When conductor is installed (`ConductorInstalled`) or active (`ConductorActive`), the segment is hidden from line 1 entirely. The rationale: when conductor is active, line 2 already shows workflow status, making the line 1 indicator redundant.

## Functional Requirements

1. The `Conductor()` segment builder must return `Enabled: false` for `ConductorActive` and `ConductorInstalled` states
2. `ConductorNone` and `ConductorMarketplace` continue to render as before (promotional/CTA segments)
3. `buildRightSegments()` in `main.go` already skips disabled segments for context — ensure same logic applies to conductor

## Non-Functional Requirements

- No performance impact (logic change only)
- Cross-platform behavior unchanged

## Acceptance Criteria

- [ ] When `ConductorActive`: no conductor segment on line 1
- [ ] When `ConductorInstalled`: no conductor segment on line 1
- [ ] When `ConductorNone`: "Try Conductor" segment renders on line 1
- [ ] When `ConductorMarketplace`: "Install Conductor" segment renders on line 1
- [ ] Line 2 workflow behavior is unaffected
- [ ] Existing tests updated to reflect new behavior

## Out of Scope

- Changing line 2 workflow rendering
- Adding new segments or indicators
- Config-level toggle for this behavior
