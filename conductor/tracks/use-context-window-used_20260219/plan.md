# Plan: Use `context_window.used_percentage` from Claude Code Statusline JSON

**Track ID:** `use-context-window-used_20260219`
**Type:** refactor
**Branch:** `refactor/context-window-used-percentage`

---

## Phase 1: Extend `ContextWindow` Struct and Update `ContextPercent()` [checkpoint: 1b26133]

### Tasks

- [x] Task: Write failing tests for `used_percentage` pre-calculated path [b64aec7]
  - Add `TestContextPercentUsedPercentageField` — JSON with `used_percentage: 42.7`, expect `ContextPercent() == 43`
  - Add `TestContextPercentUsedPercentageNull` — JSON with `used_percentage: null`, expect fallback to manual calc
  - Add `TestContextPercentUsedPercentageZero` — JSON with `used_percentage: 0`, expect `ContextPercent() == 0`
  - Add `TestContextPercentPrecalcNoCurrentUsage` — JSON with only `used_percentage` (no `current_usage`), expect it to work
  - Add `TestContextPercentRemainingPercentageParsed` — verify `RemainingPercentage` field parses correctly
  - Run `go test ./internal/hook/...` and confirm failures

- [x] Task: Add `UsedPercentage` and `RemainingPercentage` fields to `ContextWindow` struct [adb2a1a]
  - In `internal/hook/hook.go`, add `UsedPercentage *float64 \`json:"used_percentage"\`` to `ContextWindow`
  - Add `RemainingPercentage *float64 \`json:"remaining_percentage"\`` to `ContextWindow`

- [x] Task: Update `ContextPercent()` to prefer pre-calculated value [560b6cd]
  - When `d.contextWindow.UsedPercentage != nil`, return `int(math.Round(*d.contextWindow.UsedPercentage))`
  - Otherwise fall back to existing manual calculation
  - Keep existing guard: return `-1` if `ContextWindowSize == 0` and manual path used

- [x] Task: Run tests and confirm green [560b6cd]
  - `go test ./internal/hook/...`
  - Verify all new and existing tests pass

- [x] Task: Run full test suite and coverage check [560b6cd]
  - `go test ./... -coverprofile=coverage.out`
  - `go tool cover -func=coverage.out | grep hook`
  - Confirm hook package coverage ≥ 80%

- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

---

## Phase 2: Checkpoint and Cleanup [checkpoint: 0c6cb65]

### Tasks

- [x] Task: Verify no regressions in dependent code [1b26133]
  - `go vet ./...`
  - `go build ./...`
  - Confirm `main.go` and `segments/context.go` need no changes

- [x] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)
