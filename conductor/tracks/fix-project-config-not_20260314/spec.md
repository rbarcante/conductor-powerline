# Specification: Fix project config not loading from workspace path

## Overview
The project-level configuration file (`.conductor-powerline.json`) is not being loaded from the correct directory. The current implementation resolves the project config path relative to the current working directory (CWD) instead of the workspace path provided by the Claude Code hook JSON. This causes `compactWidth` and all other project-level config settings to be silently ignored whenever CWD differs from the project directory.

## Background
The Claude Code hook system invokes `conductor-powerline` as a subprocess, passing workspace metadata via stdin JSON. The `hookData.WorkspacePath()` method returns the authoritative project path. Other parts of the codebase (conductor detection, directory/git segments) already use this workspace path correctly — only config loading is inconsistent.

The bug was introduced in the original implementation. It was partially masked because the previous compact mode used a hardcoded 12-character truncation that ignored `CompactWidth` entirely. Now that PR #9 fixed compact mode to respect `CompactWidth`, the config-loading bug is exposed: users set `compactWidth` in their project config but it has no effect.

## Functional Requirements
- **FR-1**: Project config must be loaded from `<workspacePath>/.conductor-powerline.json` when `hookData.WorkspacePath()` is non-empty
- **FR-2**: When `hookData.WorkspacePath()` is empty, fall back to `os.Getwd()` for the project config path (preserving current behavior)
- **FR-3**: User-level config path (`~/.claude/conductor-powerline.json`) must remain unchanged
- **FR-4**: Config merge order must remain: defaults → user config → project config

## Non-Functional Requirements
- No new dependencies
- Silent failure behavior preserved (missing config file is not an error)
- Debug logging must include the resolved project config path

## Acceptance Criteria
- AC-1: A `.conductor-powerline.json` in the workspace directory (from hook JSON) is loaded and applied
- AC-2: When workspace path is empty, CWD-based config loading still works
- AC-3: User-level config continues to work independently
- AC-4: All existing tests pass; new tests cover workspace-path config loading

## Out of Scope
- Changing the user-level config path
- Adding config file watching or hot-reloading
- Changing the config merge priority order
