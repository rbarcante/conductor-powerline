# Plan: Usage Not Showing — Only Fallback '--'

## Phase 1: Diagnosis & Debug Infrastructure

- [ ] Task: Add debug logging package (`internal/debug/`) with stderr output gated by `CONDUCTOR_DEBUG` env var
- [ ] Task: Instrument `oauth.GetToken()` with debug logs for each credential source attempt and result
- [ ] Task: Instrument `oauth.FetchUsage()` with debug logs for token retrieval, API call, and cache status
- [ ] Task: Instrument `main.run()` with debug logs for segment build and usage data status
- [ ] Task: Conductor - User Manual Verification 'Phase 1' (Protocol in workflow.md)

## Phase 2: Root Cause Investigation & Fix

- [ ] Task: Run the tool with `CONDUCTOR_DEBUG=1` to identify the exact failure point in the OAuth → API pipeline
- [ ] Task: Write failing test(s) that reproduce the identified root cause
- [ ] Task: Implement the fix — make the pipeline succeed with a valid token and real API
- [ ] Task: Verify fix by running the full statusline and confirming usage data appears (not `--`)
- [ ] Task: Conductor - User Manual Verification 'Phase 2' (Protocol in workflow.md)

## Phase 3: Test Coverage & Cleanup

- [ ] Task: Add unit tests for the debug logging package (enable/disable, stderr output)
- [ ] Task: Update existing OAuth and segment tests to cover the fixed behavior
- [ ] Task: Run full test suite and verify >80% coverage
- [ ] Task: Conductor - User Manual Verification 'Phase 3' (Protocol in workflow.md)
