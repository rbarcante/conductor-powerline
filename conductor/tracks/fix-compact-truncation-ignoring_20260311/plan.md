# Plan: Fix Compact Truncation Ignoring CompactWidth Config

## Phase 1: Fix Compact Truncation Logic

- [x] Task 1.1: Write failing tests for proportional truncation
  - [ ] Sub-task: Test that segments are NOT truncated when total width < CompactWidth
  - [ ] Sub-task: Test that segments are proportionally truncated when total width > CompactWidth
  - [ ] Sub-task: Test minimum truncation floor of 3 characters per segment
  - [ ] Sub-task: Test edge case with single segment exceeding CompactWidth

- [x] Task 1.2: Implement proportional truncation in renderer
  - [ ] Sub-task: Remove `maxCompactTextLen` constant
  - [ ] Sub-task: Create `compactTexts(segs []segments.Segment, termWidth int) []string` function that calculates per-segment max lengths proportionally
  - [ ] Sub-task: Update `Render()` to use `compactTexts()` instead of per-segment `truncate(text, maxCompactTextLen)`
  - [ ] Sub-task: Keep `truncate()` helper (still useful) but remove hardcoded constant

- [x] Task 1.3: Update existing compact mode tests
  - [ ] Sub-task: Update `TestRenderCompactMode` to verify proportional behavior
  - [ ] Sub-task: Ensure all existing tests still pass with new logic

- [x] Task 1.4: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)
