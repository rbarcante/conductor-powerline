# Decisions: conductorstatus-only-shown-first_20260220

## ADR-001: Completely hide conductor segment on line 1 when installed

**Date:** 2026-02-20

**Decision:** When conductor is installed (`ConductorInstalled`) or active (`ConductorActive`), the conductor segment is completely hidden from line 1 rather than showing a minimal indicator.

**Rationale:** Line 2 already shows conductor workflow status when active, making a line 1 indicator redundant. Hiding it reduces visual clutter. The line 1 conductor segment's purpose is promotional/CTA — only relevant when conductor is not yet installed.
