# Plan: MVP Core

## Phase 1: Project Foundation [checkpoint: 98e2d15]

- [x] Task: Initialize Go module (`go.mod` with `github.com/rbarcante/conductor-powerline`, Go 1.23) [a782b5d]
- [x] Task: Create `internal/config/types.go` — define Config, DisplayConfig, SegmentConfig structs with JSON tags [80a1a93]
- [x] Task: Create `internal/config/config_test.go` — tests for default config, file loading, deep merge, missing file handling [fc5dc75]
- [x] Task: Create `internal/config/config.go` — implement config loading with defaults, file discovery, deep merge [fc5dc75]
- [x] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Stdin & Theme System [checkpoint: 0254ffd]

- [x] Task: Create `internal/hook/hook_test.go` — tests for stdin JSON parsing (valid, empty, malformed, missing fields) [e5ec24e]
- [x] Task: Create `internal/hook/hook.go` — parse stdin hook data, extract model/workspace/context [e5ec24e]
- [x] Task: Create `internal/themes/themes_test.go` — tests for all 6 themes, color lookup, fallback to dark [b66642d]
- [x] Task: Create `internal/themes/themes.go` — define 6 themes (dark, light, nord, gruvbox, tokyo-night, rose-pine) with segment color maps [b66642d]
- [x] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Segments [checkpoint: c17aa0f]

- [x] Task: Create `internal/segments/types.go` — define Segment struct (Name, Text, FG, BG, Enabled) [cf6b40c]
- [x] Task: Create `internal/segments/directory_test.go` — tests for directory name extraction from paths [2571adb]
- [x] Task: Create `internal/segments/directory.go` — extract repo/dir name from workspace path or cwd [2571adb]
- [x] Task: Create `internal/segments/git_test.go` — tests for branch detection, dirty state, git unavailable [05a7468]
- [x] Task: Create `internal/segments/git.go` — run git commands for branch and dirty state [05a7468]
- [x] Task: Create `internal/segments/model_test.go` — tests for model ID to friendly name mapping [f53f2fb]
- [x] Task: Create `internal/segments/model.go` — map model identifiers to display names [f53f2fb]
- [x] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)

## Phase 4: Renderer & Integration [checkpoint: 706e5a6]

- [x] Task: Create `internal/render/symbols.go` — define powerline glyphs and text fallback constants [6595268]
- [x] Task: Create `internal/render/renderer_test.go` — tests for ANSI output, segment ordering, compact mode, empty segments, no trailing newline [d198728]
- [x] Task: Create `internal/render/renderer.go` — build ANSI-colored powerline string from ordered segments [d198728]
- [x] Task: Create `main.go` — orchestrate stdin→config→theme→segments→render→stdout pipeline [f785498]
- [x] Task: Create `main_test.go` — integration test: pipe stdin JSON, verify stdout output format [f785498]
- [x] Task: Conductor - User Manual Verification 'Phase 4' (Protocol in workflow.md)
