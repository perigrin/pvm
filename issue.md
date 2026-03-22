---
title: Fix psc check CLI to produce diagnostics
state: pending
urgency: normal
milestone: psc-type-alignment
created: 2026-03-22T02:43:24.620376-04:00
updated: 2026-03-22T02:43:24.620376-04:00
---

The `psc check` CLI exits 0 with no output on files containing known type errors.
The Go-level E2E tests (`TestPSCCheckDiagnostics`) pass — they invoke `psc.NewCommand()`
directly and get arity-mismatch and type-mismatch diagnostics on stderr.

The CLI help shows flags (--recursive, --format, --strict, --dump-ast, --show-inferred)
that do not exist in `internal/psc/check_command.go`. The router or help system is
intercepting the check subcommand and routing it through a different code path.

## Acceptance

- [ ] `./build/psc check test/e2e/testdata/check/diagnostics.pl` produces diagnostic output
- [ ] Exit code is non-zero when diagnostics are found
- [ ] CLI output matches Go API test expectations
- [ ] Root cause of CLI/API disconnect is documented

Key files: internal/psc/check_command.go, internal/cli/router.go, internal/cli/help.go
See also: GitHub #406
