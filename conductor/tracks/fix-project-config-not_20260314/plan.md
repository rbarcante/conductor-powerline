# Implementation Plan: Fix project config not loading from workspace path

## Phase 1: Fix project config path resolution (TDD)

- [x] Task 1.1: Write failing test — project config loaded from workspace path
  - [x] Add test `TestIntegrationProjectConfigFromWorkspacePath` in `main_test.go`
- [x] Task 1.2: Write failing test — fallback to CWD when workspace is empty
  - [x] Add test `TestIntegrationProjectConfigFallbackToCWD` in `main_test.go`
- [x] Task 1.3: Fix `main.go` config loading
  - [x] Replace `filepath.Join(".", ".conductor-powerline.json")` with workspace-aware path resolution
  - [x] Use `hookData.WorkspacePath()` with `os.Getwd()` fallback
  - [x] Add debug log showing resolved project config path
  - [x] Fix existing tests that relied on buggy CWD behavior (`TestIntegrationConductorSegmentDisabled`, `TestIntegrationWorkflowSecondLineDisabled`)
- [x] Task 1.4: Run tests — all pass (green)

## Key Files
- `main.go:46` — bug location (project config path) — FIXED
- `internal/config/config.go` — `Load()`, `LoadFromFile()`, `MergeConfig()`
- `internal/config/config_test.go` — config unit tests
- `main_test.go` — integration tests
