# Plan: Context Window Percentage Segment

## Phase 1: Hook Data — Context Window Parsing [checkpoint: 21313d8]

- [x] Task: Add `ContextWindow` struct and fields to `hook.Data` for `context_window` JSON parsing [54b6c80]
- [x] Task: Add `ContextPercent()` method to `hook.Data` that calculates the percentage [6763b7c]
- [x] Task: Write tests for context window parsing — valid data, missing data, zero values, partial fields [6763b7c]
- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Context Segment Provider [checkpoint: 1ee3146]

- [x] Task: Create `internal/segments/context.go` with `Context()` function returning a `Segment` [ecc6819]
- [x] Task: Implement dynamic icon selection (○ < 50%, ◐ 50-80%, ● > 80%) with text fallback [ecc6819]
- [x] Task: Implement dynamic color selection using theme threshold keys (`context`, `context-warning`, `context-critical`) [ecc6819]
- [x] Task: Write tests — all threshold boundaries, zero percent, 100%, missing data, nerd font vs text mode [ecc6819]
- [x] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Theme Updates [checkpoint: 6dd604b]

- [x] Task: Add `context`, `context-warning`, `context-critical` color entries to all 6 themes [f93ec02]
- [x] Task: Write/update theme tests to verify new entries exist [f93ec02]
- [x] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)

## Phase 4: Right-Side Rendering [checkpoint: 0e49ab1]

- [x] Task: Add `RenderRight()` function to `internal/render/renderer.go` using left-pointing arrow separators [0f5b978]
- [x] Task: Add left arrow symbol to `symbols.go` [0f5b978]
- [x] Task: Write tests — single right segment, empty input, nerd font vs text fallback [0f5b978]
- [x] Task: Conductor - User Manual Verification 'Phase 4' (Protocol in workflow.md)

## Phase 5: Integration — Main & Config

- [x] Task: Add `"context"` to default segment order in config defaults [8136830]
- [x] Task: Wire context segment in `main.go` `buildSegments()` — build separately, render on right side [8136830]
- [x] Task: Update `run()` to call `RenderRight()` after left-side `Render()` and concatenate output [8136830]
- [x] Task: Write integration-level test verifying end-to-end context segment output [8136830]
- [x] Task: Conductor - User Manual Verification 'Phase 5' (Protocol in workflow.md)
