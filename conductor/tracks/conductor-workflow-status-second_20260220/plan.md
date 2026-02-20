# Implementation Plan: Conductor Workflow Status Second Line

## Phase 1: CLI Integration Layer
- [ ] Task: Create `internal/segments/workflow_cli.go` — define `WorkflowData` struct matching CLI JSON schema
- [ ] Task: Implement `FetchWorkflowStatus()` — execute `conductor_cli.py --json status`, parse JSON, return `WorkflowData`
- [ ] Task: Add timeout support using `context.WithTimeout` respecting `cfg.APITimeout`
- [ ] Task: Add debug logging for CLI execution (command, duration, success/failure)
- [ ] Task: Write tests `workflow_cli_test.go` — mock `os/exec` via function variable, test success/failure/timeout/malformed JSON
- [ ] Task: Conductor - User Manual Verification 'CLI Integration Layer' (Protocol in workflow.md)

## Phase 2: Workflow Segments
- [ ] Task: Create `internal/segments/workflow.go` — implement 4 segment builder functions:
  - `WorkflowSetup(data *WorkflowData, theme Theme) Segment` — `Setup 100%`
  - `WorkflowTrack(data *WorkflowData, theme Theme) Segment` — active track name/ID
  - `WorkflowTasks(data *WorkflowData, nerdFonts bool, theme Theme) Segment` — `12/35` for active track
  - `WorkflowOverall(data *WorkflowData, nerdFonts bool, theme Theme) Segment` — `9/9 tracks`
- [ ] Task: Implement active track selection logic (first `in_progress`, fallback to most recently updated)
- [ ] Task: Add debug logging for segment building decisions
- [ ] Task: Write tests `workflow_test.go` — test all 4 segments with various data states (active track, no tracks, all completed, nil data)
- [ ] Task: Conductor - User Manual Verification 'Workflow Segments' (Protocol in workflow.md)

## Phase 3: Theme Colors
- [ ] Task: Add 4 new color keys (`workflow_setup`, `workflow_track`, `workflow_tasks`, `workflow_overall`) to all 6 themes in `themes.go`
- [ ] Task: Update theme tests to validate new keys exist in all themes
- [ ] Task: Conductor - User Manual Verification 'Theme Colors' (Protocol in workflow.md)

## Phase 4: Config & Rendering Integration
- [ ] Task: Add `"conductor_workflow"` to `DefaultConfig()` segments map (enabled: true) and `SegmentOrder`
- [ ] Task: Update `main.go` — add concurrent CLI fetch alongside usage fetch, build line 2 segments, render with newline separator
- [ ] Task: Add activation criteria checks: ConductorActive + CLI success + config enabled; debug log visibility decision
- [ ] Task: Update config tests for new segment default
- [ ] Task: Write integration test in `main_test.go` — verify two-line output when Conductor is active
- [ ] Task: Conductor - User Manual Verification 'Config & Rendering Integration' (Protocol in workflow.md)

## Phase 5: Documentation & Polish
- [ ] Task: Update README.md with second line documentation and examples
- [ ] Task: Run full test suite, verify >80% coverage on new code
- [ ] Task: Run linter (`golangci-lint run`) and fix any issues
- [ ] Task: Conductor - User Manual Verification 'Documentation & Polish' (Protocol in workflow.md)
