# Decisions

## ADR-001: CLI Output Parsing for Workflow Data

**Date:** 2026-02-20
**Status:** Accepted

**Context:** Need to source Conductor workflow data (setup status, track progress, task counts) for the second powerline line.

**Decision:** Execute `conductor_cli.py --json status` via `os/exec` and parse the JSON output, rather than reading filesystem files directly.

**Rationale:** The CLI already aggregates all necessary data into a well-structured JSON response. This avoids duplicating the parsing logic and ensures consistency with the Conductor plugin's own status reporting.

**Consequences:** Adds a subprocess execution dependency. Mitigated by timeout support and silent failure (skip line 2 if CLI fails).

## ADR-002: Visibility Gated by ConductorActive Status

**Date:** 2026-02-20
**Status:** Accepted

**Context:** The second line only makes sense when Conductor is fully set up in the current project.

**Decision:** Only render line 2 when ALL conditions are met: ConductorActive status detected, CLI execution succeeds, and conductor_workflow segment is enabled in config.

**Rationale:** Keeps the statusline minimal for non-Conductor projects. Consistent with the product guideline of "silent by default."
