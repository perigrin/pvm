---
stability: 1
covers: []
---

## Setup

Install Go 1.24+, git, and clone the repository. Run `go build -o git-zhi ./cmd/git-zhi/`
to verify the build.

## Branching

Create feature branches from `pu`. Merge back to `pu` via pull request.

## Review

All changes require a pull request. Tests must pass. No `--no-verify`.

## Deploy

Build the binaries with `go build` and place them on `$PATH`.
