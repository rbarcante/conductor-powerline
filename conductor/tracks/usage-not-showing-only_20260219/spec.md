# Spec: Usage Not Showing — Only Fallback '--'

## Overview

The block (5-hour) and weekly (7-day) usage segments always display the `--` placeholder instead of actual API usage data. The entire OAuth → API → segment pipeline needs debugging to identify and fix the failure point. Additionally, opt-in debug logging will be added so future failures in this chain are diagnosable without code changes.

## Functional Requirements

### FR-1: Diagnose and Fix OAuth → API Pipeline Failure

- Investigate the full chain: `GetToken()` → `FetchUsageData()` → `FetchUsage()` → segment rendering
- Identify which step fails on a real macOS system with a valid Claude Pro/Team subscription
- Fix the root cause so `block` and `weekly` segments display actual usage percentages
- Ensure fallback `--` still works when genuinely offline or unauthenticated

### FR-2: Add Opt-In Debug Logging

- Add a debug logging mechanism gated by an environment variable (e.g., `CONDUCTOR_DEBUG=1`)
- Log to stderr (statusline output on stdout must remain untouched)
- Log key pipeline steps: token retrieval result, API request/response status, cache hit/miss, segment values
- Silent by default — no stderr output unless debug mode is enabled

## Non-Functional Requirements

- No new external dependencies (use `log` stdlib package)
- Debug logging must not affect startup performance (<200ms target)
- All fixes must include corresponding test updates
- Maintain >80% code coverage

## Acceptance Criteria

- [ ] Running the statusline with a valid Claude OAuth token shows actual usage percentages (not `--`)
- [ ] Running without a token still shows `--` gracefully
- [ ] Setting `CONDUCTOR_DEBUG=1` produces diagnostic stderr output showing the pipeline steps
- [ ] All existing tests continue to pass
- [ ] New tests cover the debug logging paths

## Out of Scope

- Changing the API endpoint or OAuth token format
- Adding new segments or themes
- Modifying the config file schema
- Cross-platform testing (Windows/Linux) — this fix targets macOS first
