# Spec: Fix "Try Conductor" OSC 8 Hyperlink in tmux

## Overview

The "Try Conductor" segment includes an OSC 8 hyperlink to `https://github.com/rbarcante/claude-conductor`. While this works correctly in iTerm2 (direct), the link renders as non-clickable plain text when running inside tmux 3.6a. The current code wraps OSC 8 sequences in tmux DCS passthrough (`\033Ptmux;...\033\\`), but this approach fails because `allow-passthrough` defaults to `off` in modern tmux (3.3a+). Meanwhile, tmux natively supports OSC 8 hyperlinks and will correctly handle raw OSC 8 sequences without DCS wrapping.

## Functional Requirements

1. **Remove tmux DCS passthrough wrapping** — Always emit standard OSC 8 sequences (`\033]8;;URL\033\\`) regardless of whether `$TMUX` is set. Modern tmux (3.1+) natively understands OSC 8 and forwards hyperlinks to the outer terminal.
2. **Remove `inTmux` detection** — The `TMUX` env var check and conditional branching in `osc8Open()`/`osc8CloseStr()` are no longer needed.
3. **Ensure the link is clickable in tmux** when using an OSC 8-capable outer terminal (iTerm2, Ghostty, etc.).

## Non-Functional Requirements

- No behavioral change for non-tmux terminals (they already receive plain OSC 8).
- Graceful degradation: terminals that don't support OSC 8 simply render the text without a link.
- All existing tests must continue to pass; new tests must cover the OSC 8 helpers.

## Acceptance Criteria

- [ ] `osc8Open(url)` always returns `\033]8;;URL\033\\` (no DCS wrapping)
- [ ] `osc8CloseStr()` always returns `\033]8;;\033\\` (no DCS wrapping)
- [ ] `inTmux` variable is removed from `renderer.go`
- [ ] "Try Conductor" is a clickable hyperlink in tmux 3.6a + iTerm2
- [ ] "Try Conductor" remains a clickable hyperlink in iTerm2 (without tmux)
- [ ] Unit tests cover `osc8Open` and `osc8CloseStr` output
- [ ] All existing renderer tests pass

## Out of Scope

- Supporting tmux versions older than 3.1 (no native OSC 8 support)
- Adding `allow-passthrough` documentation or tmux configuration guidance
- Adding OSC 8 links to other segments
