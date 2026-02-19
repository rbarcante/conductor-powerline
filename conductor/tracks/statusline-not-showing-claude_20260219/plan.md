# Plan: Fix Statusline Not Showing on Claude Code

## Phase 1 — Diagnosis & Hook Parser Fix [checkpoint: ba1c774]

- [x] Task: Write failing tests for hook parser with Claude Code's actual stdin schema (model as object, workspace as object) [5dc5c53]
- [x] Task: Update `hook.Data` struct to accept Claude Code's JSON shape (model object with `id`/`display_name`, workspace object with `current_dir`/`project_dir`) [5dc5c53]
- [x] Task: Add convenience accessor methods (`ModelID()`, `ModelDisplayName()`, `WorkspacePath()`) to `hook.Data` [5dc5c53]
- [x] Task: Ensure backward compatibility — tests for both old string format and new object format [5dc5c53]
- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md) [ba1c774]

## Phase 2 — Segment Wiring & Integration

- [x] Task: Update `main.go` to pass correct values from updated hook data to segment builders (`ModelID()` for model segment, `WorkspacePath()` for directory segment) [5dc5c53]
- [x] Task: Write integration test that pipes realistic Claude Code JSON through the full pipeline [105f35c]
- [x] Task: Verify output contains expected segments with correct content [105f35c]
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3 — Verification & Cleanup

- [ ] Task: Run full test suite with coverage report (`go test -cover ./...`)
- [ ] Task: Test manually with Claude Code by restarting the session
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
