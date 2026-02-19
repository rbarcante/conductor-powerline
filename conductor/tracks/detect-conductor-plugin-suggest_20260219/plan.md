# Plan: Detect Conductor Plugin and Suggest Installation

## Phase 1: Conductor Segment Core

- [ ] Task: Add `conductor` theme colors to all 6 themes in `internal/themes/themes.go`
  - Add `"conductor"` and `"conductor_missing"` segment color entries to each theme
  - `conductor`: green/success tones matching theme aesthetic
  - `conductor_missing`: yellow/warning tones matching theme aesthetic
- [ ] Task: Write failing tests for `Conductor()` segment function in `internal/segments/conductor_test.go`
  - Test: returns enabled segment with `✓ Conductor` text when plugin detected
  - Test: returns enabled segment with `⚡ Get Conductor` text + OSC 8 hyperlink when plugin not detected
  - Test: uses correct theme colors for installed vs missing states
  - Test: Nerd Font vs text fallback for icons
- [ ] Task: Implement `Conductor()` segment function in `internal/segments/conductor.go`
  - Accept a `detected bool`, `nerdFonts bool`, and `theme themes.Theme` parameter
  - When detected: render `✓ Conductor` with `conductor` theme colors
  - When not detected: render `⚡ Get Conductor` wrapped in OSC 8 hyperlink (`https://github.com/rbarcante/claude-conductor`) with `conductor_missing` colors
  - OSC 8 format: `\033]8;;URL\033\\TEXT\033]8;;\033\\`
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Plugin Detection Logic

- [ ] Task: Write failing tests for plugin detection in `internal/segments/conductor_detect_test.go`
  - Test: returns true when `~/.claude/plugins/claude-conductor/` directory exists
  - Test: returns true when `~/.claude/marketplace/claude-conductor/` directory exists (or equivalent)
  - Test: returns false when neither directory exists
  - Test: works cross-platform via `os.UserHomeDir()`
  - Use a test helper that creates/removes temp directories to simulate `~/.claude/`
- [ ] Task: Implement `DetectConductorPlugin()` function in `internal/segments/conductor_detect.go`
  - Check `~/.claude/plugins/` for directories matching `claude-conductor`
  - Check `~/.claude/marketplace/` for directories matching `claude-conductor`
  - Return `bool` — true if found in either location
  - Accept an optional base dir parameter for testability (dependency injection)
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Integration with Main Pipeline

- [ ] Task: Add `"conductor"` to default config in `internal/config/config.go`
  - Add `"conductor": {Enabled: true}` to `DefaultConfig().Segments`
  - Insert `"conductor"` into `SegmentOrder` after `"model"` and before `"block"`
- [ ] Task: Write failing test for conductor segment integration in main builder
- [ ] Task: Wire `conductor` segment into `buildSegments()` in `main.go`
  - Add `"conductor"` case to the `builders` map
  - Call `DetectConductorPlugin()` and pass result to `Conductor()` segment
  - Pass `cfg.Display.NerdFontsEnabled()` for icon selection
- [ ] Task: Update existing config tests to reflect new default segment
- [ ] Task: Run full test suite and verify > 80% coverage
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
