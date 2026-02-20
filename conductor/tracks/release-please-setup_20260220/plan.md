# Plan: CI/CD and Release-Please Setup

## Phase 1: CI Workflow

- [ ] Task: Create `.github/workflows/ci.yml` with triggers for `pull_request` and `push` to `main`
- [ ] Task: Configure OS matrix (ubuntu-latest, macos-latest, windows-latest) with Go 1.25
- [ ] Task: Add steps: checkout, setup-go, go test, go vet, go build
- [ ] Task: Add golangci-lint step using `golangci/golangci-lint-action`
- [ ] Task: Validate CI workflow runs successfully on the current codebase

## Phase 2: Release-Please Workflow

- [ ] Task: Create `.github/workflows/release-please.yml` triggered on push to `main`
- [ ] Task: Configure `googleapis/release-please-action` with `release-type: go`
- [ ] Task: Verify release-please configuration detects conventional commits

## Phase 3: README Badge and Finalization

- [ ] Task: Add release version badge to `README.md` header
- [ ] Task: Add CI status badge to `README.md` header
- [ ] Task: Verify badges render correctly
- [ ] Task: Conductor - User Manual Verification 'README Badge and Finalization' (Protocol in workflow.md)
