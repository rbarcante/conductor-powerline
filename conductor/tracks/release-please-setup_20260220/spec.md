# Specification: CI/CD and Release-Please Setup

## Overview

Set up GitHub Actions CI workflows for PRs and merges, integrate release-please on the main branch for automated release management, and add a release version tag/badge to `README.md`.

## Functional Requirements

### 1. CI Workflow (PRs and Merges)

- Trigger on `pull_request` (all branches) and `push` to `main`
- Run on a matrix of macOS, Linux, and Windows
- Steps: `go test ./...`, `go vet ./...`, `golangci-lint run`, `go build .`
- Use the Go version from `go.mod` (1.25)

### 2. Release-Please Workflow (Main Branch)

- Trigger on `push` to `main`
- Use `release-please-action` with `release-type: go`
- Automatically creates release PRs with changelogs from conventional commits
- Creates GitHub Releases with tags on merge of release PR

### 3. README Version Badge

- Add a release version badge/tag near the top of `README.md`
- Add CI status badge as well

## Non-Functional Requirements

- Workflow files must be minimal and maintainable
- CI should complete in under 5 minutes for normal PRs
- No external dependencies beyond standard GitHub Actions

## Acceptance Criteria

- [ ] `.github/workflows/ci.yml` runs tests, lint, vet, and build on PR and push to main across macOS/Linux/Windows
- [ ] `.github/workflows/release-please.yml` runs on push to main and manages release PRs
- [ ] `README.md` displays release version and CI status badges
- [ ] CI passes on the current codebase

## Out of Scope

- goreleaser / prebuilt binary distribution (future track)
- Code coverage reporting in CI
- Dependabot / dependency update automation
