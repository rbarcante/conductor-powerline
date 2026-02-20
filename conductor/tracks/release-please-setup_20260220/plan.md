# Plan: CI/CD and Release-Please Setup

## Phase 1: CI Workflow

- [x] Task: Create `.github/workflows/ci.yml` with triggers for `pull_request` and `push` to `main`
- [x] Task: Configure OS matrix (ubuntu-latest, macos-latest, windows-latest) with Go 1.25
- [x] Task: Add steps: checkout, setup-go, go test, go vet, go build
- [x] Task: Add golangci-lint step using `golangci/golangci-lint-action`
- [x] Task: Validate CI workflow runs successfully on the current codebase

## Phase 2: Release-Please Workflow

- [x] Task: Create `.github/workflows/release-please.yml` triggered on push to `main`
- [x] Task: Configure `googleapis/release-please-action` with `release-type: go`
- [x] Task: Verify release-please configuration detects conventional commits

## Phase 3: README Badge and Finalization

- [x] Task: Add release version badge to `README.md` header
- [x] Task: Add CI status badge to `README.md` header
- [x] Task: Verify badges render correctly
- [x] Task: Conductor - User Manual Verification 'README Badge and Finalization' (Protocol in workflow.md)
