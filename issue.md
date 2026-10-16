---
title: print STDOUT **2 produces a tree, though perl refuses it
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T03:47:25.870194718Z
updated: 2026-10-16T03:47:25.870194718Z
---

## Context

- found by: the PR #486 review's parser fixes (2026-10-16)
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl`

`print STDOUT **2;` is a perl 5.42 syntax error, but internal/parse produces a tree for it (its shape changed with review I22, which is otherwise correct). Refuse it as perl does.

## Steps

- [ ] Failing test first; fix against perl 5.42; the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] print FH **2 is refused (`go test ./internal/parse/ -run '^TestPrintHandleStarStarRefused$' -count=1 -v | grep -q '^--- PASS: TestPrintHandleStarStarRefused'`)
