---
title: "A fully-qualified call with a parenless argument list is refused: Carp::croak sprintf(...)"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T11:36:08.803943228Z
updated: 2026-10-16T11:36:08.803943228Z
---

## Context

- found by: the PR #486 review's I28 sweep (2026-10-16): psc check over perl 5.42's lib found parser refusals on code perl compiles
- commands: `go test ./internal/parse/`; perl 5.42 `-c` and `-MO=Deparse`

autodie.pm: `Carp::croak sprintf(...)` -- a fully-qualified sub call with no parens around its arguments. perl 5.42 compiles it (syntax OK); internal/parse refuses it. A `Package::name LIST` call of a sub perl knows takes its list as perl does.

## Steps

- [ ] Failing test first; parse as perl does; the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] a parenless fully-qualified call parses (`go test ./internal/parse/ -run '^TestQualifiedParenlessCall$' -count=1 -v | grep -q '^--- PASS: TestQualifiedParenlessCall'`)
