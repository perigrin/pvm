---
title: my Foo ($a, $b) = @_ does not lower as a list assignment
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T03:47:25.784950867Z
updated: 2026-10-16T03:47:25.784950867Z
---

## Context

- found by: the PR #486 review's parser fixes (2026-10-16)
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl`

`my Foo ($a, $b) = @_` -- a typed list declaration -- has its class-name child in front of the list, so internal/lower/decl.go does not reach the list-assignment branch, and inference uses scalar context. perl assigns the list.

## Steps

- [ ] Failing test first; fix against perl 5.42; the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] a typed list declaration lowers as a list assignment (`go test ./internal/lower/ -run '^TestTypedListDeclIsListAssign$' -count=1 -v | grep -q '^--- PASS: TestTypedListDeclIsListAssign'`)
