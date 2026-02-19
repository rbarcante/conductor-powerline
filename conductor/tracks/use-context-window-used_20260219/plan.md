# Plan: Use `context_window.used_percentage` from Claude Code Statusline JSON

**Track ID:** `use-context-window-used_20260219`
**Type:** refactor
**Branch:** `refactor/context-window-used-percentage`

---

## Phase 1: Extend `ContextWindow` Struct and Update `ContextPercent()`

### Tasks

- [ ] Task: Write failing tests for `used_percentage` pre-calculated path
  - Add `TestContextPercentUsedPercentageField` — JSON with `used_percentage: 42.7`, expect `ContextPercent() == 43`
  - Add `TestContextPercentUsedPercentageNull` — JSON with `used_percentage: null`, expect fallback to manual calc
  - Add `TestContextPercentUsedPercentageZero` — JSON with `used_percentage: 0`, expect `ContextPercent() == 0`
  - Add `TestContextPercentPrecalcNoCurrentUsage` — JSON with only `used_percentage` (no `current_usage`), expect it to work
  - Add `TestContextPercentRemainingPercentageParsed` — verify `RemainingPercentage` field parses correctly
  - Run `go test ./internal/hook/...` and confirm failures

- [ ] Task: Add `UsedPercentage` and `RemainingPercentage` fields to `ContextWindow` struct
  - In `internal/hook/hook.go`, add `UsedPercentage *float64 \`json:"used_percentage"\`` to `ContextWindow`
  - Add `RemainingPercentage *float64 \`json:"remaining_percentage"\`` to `ContextWindow`

- [ ] Task: Update `ContextPercent()` to prefer pre-calculated value
  - When `d.contextWindow.UsedPercentage != nil`, return `int(math.Round(*d.contextWindow.UsedPercentage))`
  - Otherwise fall back to existing manual calculation
  - Keep existing guard: return `-1` if `ContextWindowSize == 0` and manual path used

- [ ] Task: Run tests and confirm green
  - `go test ./internal/hook/...`
  - Verify all new and existing tests pass

- [ ] Task: Run full test suite and coverage check
  - `go test ./... -coverprofile=coverage.out`
  - `go tool cover -func=coverage.out | grep hook`
  - Confirm hook package coverage ≥ 80%

- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

---

## Phase 2: Checkpoint and Cleanup

### Tasks

- [ ] Task: Verify no regressions in dependent code
  - `go vet ./...`
  - `go build ./...`
  - Confirm `main.go` and `segments/context.go` need no changes

- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
