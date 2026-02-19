# Plan: Sync Theme Colors

## Phase 1: Update Theme Definitions

- [x] Task: Convert conductor-powerline hex colors to ANSI 256 codes for all 6 themes
- [x] Task: Restructure `themes.go` — replace per-segment warning/critical keys (`block-warning`, `block-critical`, `context-warning`, `context-critical`) with unified `warning` and `critical` keys
- [x] Task: Add new segment color keys (`opus`, `sonnet`) to all 6 themes with correct conductor-powerline values
- [x] Task: Update `block`, `weekly`, `context` segment colors to use dark-bg/colored-fg style matching conductor-powerline
- [x] Task: Update `themes_test.go` to validate new key structure (unified warning/critical, new opus/sonnet keys)
- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Update Segment Consumers

- [x] Task: Update `internal/segments/block.go` to use `theme.Segments["warning"]` and `theme.Segments["critical"]` instead of `block-warning`/`block-critical`
- [x] Task: Update `internal/segments/context.go` to use `theme.Segments["warning"]` and `theme.Segments["critical"]` instead of `context-warning`/`context-critical`
- [x] Task: Update `internal/segments/block_test.go` to reference unified warning/critical keys
- [x] Task: Update `internal/segments/context_test.go` to reference unified warning/critical keys
- [x] Task: Run full test suite and verify >80% coverage
- [x] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
