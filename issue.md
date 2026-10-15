---
title: Stale tree-sitter references remain in CI config, hooks and e2e helpers
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-15T07:52:58.246587989Z
updated: 2026-10-15T07:52:58.246587989Z
---

## Context

- paths: `.github/WORKFLOW_TROUBLESHOOTING.md`, `.github/dependabot.yml`, `.github/hooks/run-go-vet.sh`, `.pre-commit-config.yaml`, `test/e2e/helpers/env.go`, `test/e2e/helpers/psc.go`, `internal/errors/rust_formatter.go`
- commands: `ag -i 'tree-sitter|treesitter|internal/parser'`

Found 2026-10-15 by the tree-sitter retirement (01a0c115-df67), which deleted internal/parser and gotreesitter but left these references, all stale since the CGO era: `make tree-sitter` in WORKFLOW_TROUBLESHOOTING.md; a dependabot entry for `/tree-sitter-typed-perl` (does not exist); `grep -v treesitter` in run-go-vet.sh; `internal/parser/` exclude patterns in .pre-commit-config.yaml; library paths including internal/parser in test/e2e/helpers/env.go; "run 'make tree-sitter'" skip messages in test/e2e/helpers/psc.go; `FormatTreeSitterError` (no caller) in rust_formatter.go.

## Steps

- [ ] Remove each; confirm the hooks and CI config still parse; `make test`; commit

## Acceptance criteria

- [ ] no tree-sitter reference remains outside dated plans and specs (`! ag -il 'tree-sitter|treesitter|internal/parser' .github .pre-commit-config.yaml test/e2e/helpers internal/errors`)
