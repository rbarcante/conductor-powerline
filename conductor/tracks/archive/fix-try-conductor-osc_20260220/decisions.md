# Decisions: fix-try-conductor-osc_20260220

## ADR-001: Remove DCS Passthrough in Favor of Native OSC 8

**Date:** 2026-02-20
**Status:** Accepted

**Context:** The renderer wraps OSC 8 hyperlink sequences in tmux DCS passthrough (`\033Ptmux;...\033\\`) when `$TMUX` is set. This fails because tmux 3.3a+ defaults `allow-passthrough` to `off`, silently dropping the sequences. Meanwhile, tmux 3.1+ natively supports OSC 8 and correctly forwards hyperlinks to the outer terminal without needing DCS passthrough.

**Decision:** Remove all DCS passthrough logic and the `inTmux` detection variable. Always emit standard OSC 8 sequences. This is simpler, correct for all modern tmux versions, and works identically for non-tmux terminals.

**Consequences:**
- Positive: Links work in tmux when output is rendered directly (requires `terminal-features hyperlinks` in `.tmux.conf`)
- Positive: Simpler code — no branching based on environment
- Negative: tmux versions older than 3.1 (released 2020) will not render hyperlinks, but they degrade gracefully to plain text
- Negative: Links are NOT clickable when rendered through Claude Code's statusline inside tmux — this is a Claude Code rendering limitation, not a conductor-powerline issue

## ADR-002: tmux Requires `terminal-features hyperlinks` Configuration

**Date:** 2026-02-20
**Status:** Accepted

**Context:** During manual verification, standard OSC 8 sequences did not produce clickable links in tmux until `set -as terminal-features ",*:hyperlinks"` was added to `.tmux.conf` and the tmux server was fully restarted (`tmux kill-server`). Additionally, links rendered through Claude Code's statusline are not clickable inside tmux, even though the same binary output is clickable when run directly in a tmux pane.

**Decision:** Accept that tmux users need the `terminal-features hyperlinks` config line. The Claude Code statusline rendering issue is out of scope — it is a limitation of how Claude Code re-renders stdout output inside tmux.

**Consequences:**
- Positive: Standard OSC 8 is the correct long-term approach
- Negative: tmux users need one config line for hyperlinks to work
- Negative: Claude Code statusline rendering in tmux remains a known limitation
