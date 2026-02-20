# Plan: Fix "Try Conductor" OSC 8 Hyperlink in tmux

## Phase 1: Write Failing Tests (Red) [checkpoint: b48b669]

- [x] Task: Add unit test for `osc8Open()` — assert it returns standard OSC 8 format `\033]8;;URL\033\\` (no DCS wrapping) [b887514]
- [x] Task: Add unit test for `osc8CloseStr()` — assert it returns `\033]8;;\033\\` (no DCS wrapping) [b887514]
- [x] Task: Add unit test for `Render()` with a segment that has a `Link` field — assert output contains OSC 8 open and close sequences wrapping the segment [b887514]
- [x] Task: Add unit test for `RenderRight()` with a segment that has a `Link` field — assert output contains OSC 8 open and close sequences wrapping the segment [b887514]
- [x] Task: Run tests and confirm new tests fail (existing DCS passthrough logic causes mismatch when `$TMUX` is set) [b887514]
- [x] Task: Conductor - User Manual Verification 'Phase 1' [b48b669]

## Phase 2: Fix OSC 8 Implementation (Green) [checkpoint: c61a597]

- [x] Task: Remove `inTmux` variable from `renderer.go` [c61a597]
- [x] Task: Simplify `osc8Open()` to always return `fmt.Sprintf("\033]8;;%s\033\\", url)` [c61a597]
- [x] Task: Simplify `osc8CloseStr()` to always return `"\033]8;;\033\\"` [c61a597]
- [x] Task: Remove `"os"` import from `renderer.go` if no longer used [c61a597]
- [x] Task: Run all tests and confirm they pass [c61a597]
- [x] Task: Conductor - User Manual Verification 'Phase 2' [verified]

## Phase 3: Verify and Clean Up

- [x] Task: Run `go vet ./...` and `go test ./...` to confirm no regressions [verified]
- [x] Task: Verify code coverage for `internal/render/` meets >80% [95.5%]
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
