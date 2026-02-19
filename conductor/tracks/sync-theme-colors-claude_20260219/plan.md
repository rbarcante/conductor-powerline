# Plan: Sync Theme Colors

## Phase 1: Update Theme Definitions

- [ ] Task: Convert conductor-powerline hex colors to ANSI 256 codes for all 6 themes
- [ ] Task: Restructure `themes.go` — replace per-segment warning/critical keys (`block-warning`, `block-critical`, `context-warning`, `context-critical`) with unified `warning` and `critical` keys
- [ ] Task: Add new segment color keys (`opus`, `sonnet`) to all 6 themes with correct conductor-powerline values
- [ ] Task: Update `block`, `weekly`, `context` segment colors to use dark-bg/colored-fg style matching conductor-powerline
- [ ] Task: Update `themes_test.go` to validate new key structure (unified warning/critical, new opus/sonnet keys)
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Update Segment Consumers

- [ ] Task: Update `internal/segments/block.go` to use `theme.Segments["warning"]` and `theme.Segments["critical"]` instead of `block-warning`/`block-critical`
- [ ] Task: Update `internal/segments/context.go` to use `theme.Segments["warning"]` and `theme.Segments["critical"]` instead of `context-warning`/`context-critical`
- [ ] Task: Update `internal/segments/block_test.go` to reference unified warning/critical keys
- [ ] Task: Update `internal/segments/context_test.go` to reference unified warning/critical keys
- [ ] Task: Run full test suite and verify >80% coverage
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
