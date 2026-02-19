# Plan: Fix Statusline Not Showing on Claude Code

## Phase 1 — Diagnosis & Hook Parser Fix

- [ ] Task: Write failing tests for hook parser with Claude Code's actual stdin schema (model as object, workspace as object)
- [ ] Task: Update `hook.Data` struct to accept Claude Code's JSON shape (model object with `id`/`display_name`, workspace object with `current_dir`/`project_dir`)
- [ ] Task: Add convenience accessor methods (`ModelID()`, `ModelDisplayName()`, `WorkspacePath()`) to `hook.Data`
- [ ] Task: Ensure backward compatibility — tests for both old string format and new object format
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2 — Segment Wiring & Integration

- [ ] Task: Update `main.go` to pass correct values from updated hook data to segment builders (`ModelID()` for model segment, `WorkspacePath()` for directory segment)
- [ ] Task: Write integration test that pipes realistic Claude Code JSON through the full pipeline
- [ ] Task: Verify output contains expected segments with correct content
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3 — Verification & Cleanup

- [ ] Task: Run full test suite with coverage report (`go test -cover ./...`)
- [ ] Task: Test manually with Claude Code by restarting the session
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
