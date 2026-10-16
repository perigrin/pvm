---
title: print FH &{...}(...) is refused
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T11:36:09.160250143Z
updated: 2026-10-16T11:36:09.160250143Z
---

## Context

- found by: the PR #486 review's I28 sweep (2026-10-16): psc check over perl 5.42's lib found parser refusals on code perl compiles
- commands: `go test ./internal/parse/`; perl 5.42 `-c` and `-MO=Deparse`

diagnostics.pm: `print THITHER &{...}(...)` -- a print to a bareword handle whose first argument is a `&{BLOCK}(ARGS)` call. perl 5.42 compiles it; internal/parse refuses it.

## Steps

- [ ] Failing test first; parse as perl does; the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] print FH &{...}(...) parses (`go test ./internal/parse/ -run '^TestPrintHandleAmpBlockCall$' -count=1 -v | grep -q '^--- PASS: TestPrintHandleAmpBlockCall'`)
