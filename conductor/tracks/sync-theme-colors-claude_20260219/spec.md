# Spec: Sync Theme Colors

## Overview

Update all theme color definitions in `conductor-powerline` to strictly match the colors defined in `themes/index.ts`. This includes restructuring the color model from per-segment warning/critical variants to unified warning/critical colors, adding new `opus` and `sonnet` segment colors, and converting all hex values through the same `hexToAnsi256` algorithm.

## Functional Requirements

1. **Color Sync**: All 6 themes (dark, light, nord, gruvbox, tokyo-night, rose-pine) must use the exact hex-to-ANSI256 converted values from `conductor-powerline`
2. **Unified Warning/Critical**: Replace `block-warning`, `block-critical`, `context-warning`, `context-critical` segment keys with unified `warning` and `critical` keys per theme
3. **New Segments**: Add `opus`, `sonnet` segment colors to all themes
4. **Consumer Update**: Update `block.go` and `context.go` (and their tests) to use the unified `warning`/`critical` keys instead of per-segment variants

## Non-Functional Requirements

- Zero visual regression for directory, git, model segments (values stay the same where unchanged)
- All existing tests must be updated and pass
- >80% code coverage maintained

## Acceptance Criteria

- [ ] Running the powerline with each theme produces colors matching conductor-powerline output
- [ ] `themes.go` has unified `warning`/`critical` keys (no `block-warning`, `context-warning`, etc.)
- [ ] `themes.go` includes `opus`, `sonnet`, `block`, `weekly`, `context` with dark-bg/colored-fg style from conductor-powerline
- [ ] `block.go` and `context.go` use `theme.Segments["warning"]` and `theme.Segments["critical"]`
- [ ] All tests pass

## Out of Scope

- Adding new rendering logic for opus/sonnet model differentiation
- Changing the powerline arrow/separator rendering
- Adding new themes beyond the existing 6

## Reference

- Source of truth: `../themes/index.ts`
