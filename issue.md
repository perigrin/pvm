---
title: Remove phantom PSC commands and flags from help system
state: pending
urgency: normal
milestone: psc-type-alignment
created: 2026-03-22T02:43:32.068658-04:00
updated: 2026-03-22T02:43:32.068658-04:00
---

The PSC CLI help displays commands and flags that do not exist in the codebase:

Phantom commands: compile, def, format, generate-type, import-type
Phantom flags on check: --recursive, --format, --strict, --dump-ast, --show-inferred

The actual PSC subcommands (from `internal/psc/command.go`) are: parse, analyze, check, lsp.

## Acceptance

- [ ] `psc --help` shows only: parse, analyze, check, lsp, version
- [ ] `psc check --help` shows only flags registered by `newCheckCommand()`
- [ ] Help system reflects actual cobra command tree, not aspirational content

Key files: internal/cli/help.go, internal/psc/command.go
See also: GitHub #407
