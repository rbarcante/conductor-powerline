# Plan: Fix "Try Conductor" OSC 8 Hyperlink in tmux

## Phase 1: Write Failing Tests (Red)

- [ ] Task: Add unit test for `osc8Open()` — assert it returns standard OSC 8 format `\033]8;;URL\033\\` (no DCS wrapping)
- [ ] Task: Add unit test for `osc8CloseStr()` — assert it returns `\033]8;;\033\\` (no DCS wrapping)
- [ ] Task: Add unit test for `Render()` with a segment that has a `Link` field — assert output contains OSC 8 open and close sequences wrapping the segment
- [ ] Task: Add unit test for `RenderRight()` with a segment that has a `Link` field — assert output contains OSC 8 open and close sequences wrapping the segment
- [ ] Task: Run tests and confirm new tests fail (existing DCS passthrough logic causes mismatch when `$TMUX` is set)
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Fix OSC 8 Implementation (Green)

- [ ] Task: Remove `inTmux` variable from `renderer.go`
- [ ] Task: Simplify `osc8Open()` to always return `fmt.Sprintf("\033]8;;%s\033\\", url)`
- [ ] Task: Simplify `osc8CloseStr()` to always return `"\033]8;;\033\\"`
- [ ] Task: Remove `"os"` import from `renderer.go` if no longer used
- [ ] Task: Run all tests and confirm they pass
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Verify and Clean Up

- [ ] Task: Run `go vet ./...` and `go test ./...` to confirm no regressions
- [ ] Task: Verify code coverage for `internal/render/` meets >80%
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
