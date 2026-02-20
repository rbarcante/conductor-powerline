# Code Review Report

**Branch:** `feature/detect-conductor-plugin` vs `develop`
**Generated:** 2026-02-19
**Track:** Detect conductor plugin and suggest installation

---

## Summary

| Metric | Value |
|--------|-------|
| Files Changed | 12 |
| Lines Added | +470 |
| Lines Removed | -78 |
| **Findings** | 🔴 High: 1 \| 🟡 Medium: 7 \| 🟢 Low: 6 |

---

## Code Quality

### High Severity

**`internal/render/renderer.go`** — Duplicated `VisualText` fallback logic
The pattern `visualText := s.Text; if s.VisualText != "" { visualText = s.VisualText }` appears twice in the same file (compact rendering loop and `shouldCompact`). This is duplicated logic tied to a non-obvious field contract.
**Suggestion:** Add a `func (s Segment) DisplayText() string` method on `Segment` in `types.go` that encapsulates this fallback. Both call sites become `s.DisplayText()`.

### Medium Severity

- **`internal/segments/conductor.go:9`** — `conductorURL` is unexported but encodes a policy decision. If the URL changes, there's no way for callers to observe it.
  **Suggestion:** Export as `ConductorURL` constant.

- **`internal/segments/conductor.go:20`** — `Conductor()` mixes label selection (nerdFonts branching) with struct construction. Both branches duplicate the nerdFonts conditional.
  **Suggestion:** Extract `conductorLabel(detected, nerdFonts bool) string` helper.

- **`internal/segments/conductor.go:23`** — Theme colors accessed via untyped string keys — a typo produces a silent zero-value with no error.
  **Suggestion:** Add a typed accessor on `themes.Theme` that returns `(SegmentColors, bool)`.

- **`internal/segments/types.go:7`** — Multi-line inline comment on `VisualText` is unconventional Go style; won't render correctly in godoc.
  **Suggestion:** Use a single-line comment: `// Plain text for width calculation; may differ from Text when Text contains escape sequences.`

- **`main.go:101`** — `if !hasCfg || condCfg.Enabled` enables conductor by default (opt-out) with no documentation of this intent. Asymmetric with potential future segments.
  **Suggestion:** Add a comment explaining the intentional default-on behavior.

- **`main.go:98`** — Segment identity string `"conductor"` appears independently in `rightSideSegments`, `seg.Name`, and config keys — three sources of truth.
  **Suggestion:** Export `segments.ConductorName = "conductor"` and use it everywhere.

### Low Severity

- **`conductor.go:13`** — `osc8Link` is general-purpose but trapped in `conductor.go`, invisible to other segments.
  **Suggestion:** Move to `internal/segments/escape.go`.

- **`conductor_detect.go:16`** — Doc comment lists detection locations manually; can drift from the `locations` slice.
  **Suggestion:** Note that maintainers must update the comment when `locations` changes.

- **`main.go:98`** — Segment ordering in `buildRightSegments` enforced only by append order; comment describes intent but no enforcement.

- **`main.go:101`** — Default-on behavior performs filesystem stat calls on every invocation for all users.

---

## Security Analysis

### Critical/High Severity

No security vulnerabilities detected.

### Medium Severity

- **`internal/segments/conductor.go`** — `osc8Link(url, text string)` accepts an arbitrary URL with no validation. Currently safe (hardcoded constant), but if a caller ever passes user-controlled input, arbitrary terminal escape sequences could be injected.
  **Suggestion:** Add guard: `if strings.ContainsAny(url, "\x1b\x07") { return text }`. Also restrict to `https://` prefix.

### Low Severity

- **`main.go`** — Default-enabled conductor segment performs filesystem stat calls without explicit user opt-in. Minor config hygiene issue.

- **`conductor_detect.go`** — `baseDir` parameter has no validation against path traversal. Currently safe (always called with `""`), but the exported signature accepts arbitrary paths.
  **Suggestion:** Add `if baseDir != "" && !filepath.IsAbs(baseDir) { return false }`.

---

## Test Coverage

### Coverage Results

| Package | Coverage | Threshold | Status |
|---------|----------|-----------|--------|
| `internal/segments` | 94.4% | 80% | ✅ |
| `internal/themes` | 100% | 80% | ✅ |
| `internal/config` | 90.7% | 80% | ✅ |
| `internal/render` | 92.1% | 80% | ✅ |

### Missing Tests

- **`conductor_detect.go:17-20`** — `os.UserHomeDir()` error path not covered. The `TestDetectConductorPluginEmptyBaseUsesHomeDir` smoke test doesn't verify error handling behavior.

### Insufficient Coverage

- `renderer.go` compact mode: Unicode multi-byte edge cases for `truncate()` with mixed-width segments not explicitly tested.
- `config.go`: `MergeConfig` with nil/empty override Segments edge cases may be incomplete.

---

## Recommendations

**Priority Actions (address before merging):**
1. Add `func (s Segment) DisplayText() string` method to eliminate duplicated `VisualText` fallback in renderer — prevents future bugs when a third call site is added.
2. Add URL validation in `osc8Link` (escape char guard) — low effort, eliminates a latent injection surface.

**Suggested Improvements:**
1. Export `ConductorURL` and `ConductorName` constants for testability and single source of truth.
2. Extract `conductorLabel()` helper to simplify `Conductor()` function.
3. Move `osc8Link` to `internal/segments/escape.go` for reusability.
4. Add `filepath.IsAbs` guard in `DetectConductorPlugin` for `baseDir` input validation.
5. Add a test for the `os.UserHomeDir()` error path in `conductor_detect_test.go`.

---

*Auto-review generated by `/conductor:implement` on track completion*
