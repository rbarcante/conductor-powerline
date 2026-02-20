# Decisions: fix-try-conductor-osc_20260220

## ADR-001: Remove DCS Passthrough in Favor of Native OSC 8

**Date:** 2026-02-20
**Status:** Accepted

**Context:** The renderer wraps OSC 8 hyperlink sequences in tmux DCS passthrough (`\033Ptmux;...\033\\`) when `$TMUX` is set. This fails because tmux 3.3a+ defaults `allow-passthrough` to `off`, silently dropping the sequences. Meanwhile, tmux 3.1+ natively supports OSC 8 and correctly forwards hyperlinks to the outer terminal without needing DCS passthrough.

**Decision:** Remove all DCS passthrough logic and the `inTmux` detection variable. Always emit standard OSC 8 sequences. This is simpler, correct for all modern tmux versions, and works identically for non-tmux terminals.

**Consequences:**
- Positive: Links work in tmux without requiring user configuration (`allow-passthrough on`)
- Positive: Simpler code — no branching based on environment
- Negative: tmux versions older than 3.1 (released 2020) will not render hyperlinks, but they degrade gracefully to plain text
