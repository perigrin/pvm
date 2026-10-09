---
title: "A #line inside code interpolated in a string is not honoured"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T07:54:13.120721591Z
updated: 2026-10-09T07:54:13.120721591Z
---

## Context

- paths: `internal/parse/sites.go` (`PerlLines`, `lineDirective`), the lexer's string tokens
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by 01a11f08-7251 (the subject honours `#line`), 2026-10-09: perl honours a `#line N` comment inside code interpolated in a string -- comp/parser.t's `"${\n#line 53\n_}"` -- because toke.c lexes the interpolated block as code. `PerlLines` does not, since the lexer reads the string as one token. No corpus verdict depends on it today (the oracle drops comp/parser.t after its `#line 23 KASHPRITZA`); it belongs with parsing inside interpolation (the deferred real fix for the I group of the WRONG triage).

## Steps

- [ ] Once interpolated code is lexed as code, measure perl on `#line` inside `${ ... }` and `@{[ ... ]}`; failing test first; implement

## Acceptance criteria

- [ ] a #line inside interpolated code renumbers as perl does (`go test ./internal/parse/ -run '^TestPerlLinesInsideInterpolation$' -count=1 -v | grep -q '^--- PASS: TestPerlLinesInsideInterpolation'`)
