# Implementation Plan: Fix project config not loading from workspace path

## Phase 1: Fix project config path resolution (TDD)

- [ ] Task 1.1: Write failing test — project config loaded from workspace path
  - [ ] Add test in `main_test.go` or `config_test.go` that provides a workspace path with a `.conductor-powerline.json` containing custom `compactWidth` and verifies it's applied
- [ ] Task 1.2: Write failing test — fallback to CWD when workspace is empty
  - [ ] Verify that when workspace path is empty, the config is loaded from CWD (current behavior)
- [ ] Task 1.3: Fix `main.go` config loading
  - [ ] Replace `filepath.Join(".", ".conductor-powerline.json")` with workspace-aware path resolution
  - [ ] Use `hookData.WorkspacePath()` with `os.Getwd()` fallback, matching the pattern at lines 60-64
  - [ ] Add debug log showing resolved project config path
- [ ] Task 1.4: Run tests — verify all pass (green)
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Key Files
- `main.go:46` — bug location (project config path)
- `internal/config/config.go` — `Load()`, `LoadFromFile()`, `MergeConfig()`
- `internal/config/config_test.go` — config unit tests
- `main_test.go` — integration tests (if exists)
