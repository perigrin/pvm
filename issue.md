---
title: A compound operator after declaration attributes is accepted, though perl refuses it
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T03:47:25.709100469Z
updated: 2026-10-16T03:47:25.709100469Z
---

## Context

- found by: the PR #486 review's parser fixes (2026-10-16)
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl`

`my $x :shared -= 1;` -- perl 5.42 -c: "Invalid separator character" -- but internal/parse accepts it (the compound operator is recorded on the declaration since review I10). Refuse it as perl does.

## Steps

- [ ] Failing test first; fix against perl 5.42; the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] a compound operator after attributes is refused (`go test ./internal/parse/ -run '^TestCompoundOpAfterAttrsRefused$' -count=1 -v | grep -q '^--- PASS: TestCompoundOpAfterAttrsRefused'`)
