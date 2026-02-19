# Specification: Context Window Percentage Segment

## Overview

Add a context window percentage segment that displays the current Claude Code context window utilization as a percentage. The segment renders on the **right side** of the powerline (with left-pointing separators). It reads `context_window` data from the stdin hook JSON.

## Functional Requirements

### 1. Data Source

Parse `context_window.current_usage` and `context_window.context_window_size` from the hook JSON stdin data. Calculate:

```
percent = round((input_tokens + cache_creation_input_tokens + cache_read_input_tokens) / context_window_size * 100)
```

### 2. Dynamic Icons (Nerd Fonts mode)

| Range       | Icon | Description   |
|-------------|------|---------------|
| < 50%       | `○`  | Empty circle  |
| 50% – 80%   | `◐`  | Half circle   |
| > 80%       | `●`  | Full circle   |

Text fallback (no Nerd Fonts): `CTX`

### 3. Color Thresholds

Colors change dynamically based on percentage:

| Range       | Theme Key           | Color Family |
|-------------|---------------------|--------------|
| < 50%       | `context`           | Green/cool   |
| 50% – 80%   | `context-warning`   | Yellow/amber |
| > 80%       | `context-critical`  | Red/warm     |

### 4. Right-Side Rendering

The renderer must support a right-side section using left-pointing arrow separators (Nerd Font left arrow glyph). The context segment is always rendered on the right, regardless of `segmentOrder` position.

### 5. Configuration

Enabled/disabled via `segments.context.enabled` in config JSON. Added to default segment order.

### 6. Graceful Degradation

If `context_window` data is missing from hook JSON, the segment is not shown. Follows the "silent by default" principle.

## Non-Functional Requirements

- No new dependencies (pure stdlib)
- >80% test coverage for new code
- Sub-200ms total startup not impacted

## Acceptance Criteria

- [ ] Context % displays correctly when hook JSON includes context_window data
- [ ] Icon changes based on percentage thresholds (empty/half/full circle)
- [ ] Colors change based on percentage thresholds (green/yellow/red)
- [ ] Segment renders on right side with left-pointing arrows
- [ ] Segment hidden when context_window data is absent
- [ ] All 6 themes include context color definitions (normal/warning/critical)
- [ ] Text fallback mode works without Nerd Fonts
- [ ] >80% test coverage

## Out of Scope

- Historical context tracking
- Context prediction/estimation
- Right-side rendering of other segments (only context goes right for now)
