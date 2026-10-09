---
title: "A call paren after a plain term parses: $x() and my $x() are refused by perl"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T23:49:35.11954988Z
updated: 2026-10-09T23:54:43.301236642Z
sessions:
- start_sha: 419a99b64470f2c2a547bcd922fd613573eb6dbf
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T23:54:43.301236642Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T23:54:43.301236642Z
---

## Context

- paths: `internal/parse/expr.go` (the infix `(` path: a postfix call paren after a term), `internal/parse/canon.go`, `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `make fuzz-parser`; `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-09 by `make fuzz-parser` (after 01a0ed89): `my$0()` -- canon is not a fixpoint (`my$0();` then `;;`). Seed parked at scratchpad/pending-seeds/a447ccd41522a9fd. The parser reads a postfix `(` as a call subscript after ANY term. Measured on 5.42:

- refused: `$x()`, `my $x()`, `$$x()`, `@x()`, `f()()`, `(f)()`, `1()`, `$x++()`
- accepted: `$x[0]()`, `$x->[0]()`, `$x->()()` (a call paren after a subscript), `&$x()`, `$x->$m()`

`$x->$m()` and the invalid `$x()` reach the same infix `(` path with the same left node (Term `$m`), so the rule needs to know whether an arrow (or a subscript, or `&`) came before.

## Steps

- [ ] Failing test from the measurements; refuse a call paren perl refuses (existing refusal site); keep every accepted form; commit the seed; `make fuzz-parser` until a full run passes; `make test`; commit

## Acceptance criteria

- [ ] a call paren after a plain term is refused, after a subscript or arrow accepted (`go test ./internal/parse/ -run '^TestPostfixCallParenFollowsPerl$' -count=1 -v | grep -q '^--- PASS: TestPostfixCallParenFollowsPerl'`)
- [ ] the parser fuzzer finds nothing (`make fuzz-parser`)
