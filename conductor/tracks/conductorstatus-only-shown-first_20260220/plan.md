# Plan: conductorStatus first-line visibility

## Phase 1: Modify Conductor Segment Visibility

- [ ] Task 1.1: Write failing tests — update `conductor_test.go` to assert `Enabled: false` for `ConductorActive` and `ConductorInstalled` states
- [ ] Task 1.2: Modify `Conductor()` in `conductor.go` — return `Enabled: false` for `ConductorActive` and `ConductorInstalled`
- [ ] Task 1.3: Update `buildRightSegments()` in `main.go` to check `seg.Enabled` before appending conductor segment (same pattern as context segment)
- [ ] Task 1.4: Run full test suite, verify all tests pass
- [ ] Task 1.5: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)
