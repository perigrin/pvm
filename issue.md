---
title: "A call paren after a plain term parses: $x() and my $x() are refused by perl"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T23:49:35.11954988Z
updated: 2026-10-10T05:34:18.310522652Z
sessions:
- start_sha: 419a99b64470f2c2a547bcd922fd613573eb6dbf
  end_sha: baf52e02b9bfe6cccdacae5aafb9bd5e8d725c68
  commits: 5
  started_at: 2026-10-09T23:54:43.301236642Z
  ended_at: 2026-10-10T03:36:15.291909611Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T23:54:43.301236642Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-10T03:36:15.291909611Z
observed_paths:
- internal/lexer/fhslot_test.go
- internal/lexer/lexer.go
- internal/parse/anonhead_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/chain.go
- internal/parse/chain_test.go
- internal/parse/commafat_test.go
- internal/parse/decl.go
- internal/parse/expr.go
- internal/parse/parse.go
- internal/parse/postfixcallparen_test.go
- internal/parse/refusal.go
- internal/parse/scalarhandle_test.go
- internal/parse/term.go
- internal/parse/testdata/fuzz/FuzzParser/0c172203842b1f8c
- internal/parse/testdata/fuzz/FuzzParser/7bd29f70ee42a374
- internal/parse/testdata/fuzz/FuzzParser/a447ccd41522a9fd
- internal/parse/testdata/fuzz/FuzzParser/b45427e536ce28a7
- internal/parse/testdata/fuzz/FuzzParser/daec3eb606eaafb5
- internal/parse/testdata/fuzz/FuzzParser/f0168060c2eb4a04
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/trailingcomma_test.go
- internal/parse/unaryplus_test.go
- internal/parse/unclosed_test.go
- internal/parseoracle/subject_test.go
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

### The fuzz criterion replays the seeds (2026-10-10)

`make fuzz-parser` explores at random, so as a criterion it can fail on a
new find unrelated to this issue -- and a failing run writes its seed into
testdata, where every later `go test ./internal/parse/` replays it. This
issue's claim is its own inputs, which are committed as fuzz seeds; the
criterion replays them deterministically. The open-ended fuzz claim
belongs to 01a0a70f-0ec3 and 01a0de03, where it stays.

## Acceptance criteria

- [ ] a call paren after a plain term is refused, after a subscript or arrow accepted (`go test ./internal/parse/ -run '^TestPostfixCallParenFollowsPerl$' -count=1 -v | grep -q '^--- PASS: TestPostfixCallParenFollowsPerl'`)
- [ ] every committed fuzz seed replays clean (`go test ./internal/parse/ -run '^FuzzParser$' -count=1 -v | grep -q '^--- PASS: FuzzParser'`)
