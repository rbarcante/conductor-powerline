# Decisions: CI/CD and Release-Please Setup

## ADR-001: release-please without goreleaser

**Date:** 2026-02-20

**Decision:** Use release-please for release management without goreleaser for binary distribution.

**Rationale:** The project is primarily distributed via `go install`, so prebuilt binaries are not essential for the initial release workflow. goreleaser can be added in a future track if binary distribution becomes a priority.

## ADR-002: golangci-lint via GitHub Action

**Date:** 2026-02-20

**Decision:** Use the official `golangci/golangci-lint-action` in CI rather than installing golangci-lint manually.

**Rationale:** The official action handles caching, version pinning, and platform detection automatically. It's the recommended approach for GitHub Actions CI.
