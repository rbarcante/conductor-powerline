# Plan: MVP Core

## Phase 1: Project Foundation

- [ ] Task: Initialize Go module (`go.mod` with `github.com/rbarcante/conductor-powerline`, Go 1.23)
- [ ] Task: Create `internal/config/types.go` — define Config, DisplayConfig, SegmentConfig structs with JSON tags
- [ ] Task: Create `internal/config/config_test.go` — tests for default config, file loading, deep merge, missing file handling
- [ ] Task: Create `internal/config/config.go` — implement config loading with defaults, file discovery, deep merge
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Stdin & Theme System

- [ ] Task: Create `internal/hook/hook_test.go` — tests for stdin JSON parsing (valid, empty, malformed, missing fields)
- [ ] Task: Create `internal/hook/hook.go` — parse stdin hook data, extract model/workspace/context
- [ ] Task: Create `internal/themes/themes_test.go` — tests for all 6 themes, color lookup, fallback to dark
- [ ] Task: Create `internal/themes/themes.go` — define 6 themes (dark, light, nord, gruvbox, tokyo-night, rose-pine) with segment color maps
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Segments

- [ ] Task: Create `internal/segments/types.go` — define Segment struct (Name, Text, FG, BG, Enabled)
- [ ] Task: Create `internal/segments/directory_test.go` — tests for directory name extraction from paths
- [ ] Task: Create `internal/segments/directory.go` — extract repo/dir name from workspace path or cwd
- [ ] Task: Create `internal/segments/git_test.go` — tests for branch detection, dirty state, git unavailable
- [ ] Task: Create `internal/segments/git.go` — run git commands for branch and dirty state
- [ ] Task: Create `internal/segments/model_test.go` — tests for model ID to friendly name mapping
- [ ] Task: Create `internal/segments/model.go` — map model identifiers to display names
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)

## Phase 4: Renderer & Integration

- [ ] Task: Create `internal/render/symbols.go` — define powerline glyphs and text fallback constants
- [ ] Task: Create `internal/render/renderer_test.go` — tests for ANSI output, segment ordering, compact mode, empty segments, no trailing newline
- [ ] Task: Create `internal/render/renderer.go` — build ANSI-colored powerline string from ordered segments
- [ ] Task: Create `main.go` — orchestrate stdin→config→theme→segments→render→stdout pipeline
- [ ] Task: Create `main_test.go` — integration test: pipe stdin JSON, verify stdout output format
- [ ] Task: Conductor - User Manual Verification 'Phase 4' (Protocol in workflow.md)
