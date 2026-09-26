---
title: An ADJUST block followed by a method loses the method, and perl accepts both orders
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T11:20:16.240938445Z
updated: 2026-09-26T12:19:34.066552349Z
sessions:
- start_sha: f178e76df654244563b62e3b3f68f49cdfef252b
  end_sha: 7ccdfe71951a54c8814b1a62c2c9e769ca19a472
  commits: 4
  started_at: 2026-09-26T11:20:37.469293462Z
  ended_at: 2026-09-26T12:19:34.066552349Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T11:20:37.469293462Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T12:19:34.066552349Z
observed_paths:
- conformance/mdtest/adjacency-04_operators.md
- conformance/mdtest/adjacency-07_subroutines.md
- conformance/mdtest/adjacency-11_oo.md
- conformance/mdtest/named-operators.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- internal/conformance/conformance_test.go
- internal/conformance/file.go
- internal/conformance/testdata/corpus.ratchet
- internal/lexer/keyword.go
- internal/parse/keyword.go
- internal/parse/keywordgap_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
- internal/parse/use_test.go
---

An `ADJUST` block followed by a `method` declaration loses the method: the
refusal depends on the ORDER, and perl accepts both. One corpus case,
`adjacency-11_oo`, whose `refuses:` cites a CLOSED issue.

## Measured

`adjacency-11_oo.md`'s case has ONE Unknown. Bisected by progressive prefix it
appears at line 6, `method m { $x }` -- and that line alone is clean. Isolated:

    class body                                  Unknown
    field $x = 1;                                  0
    method m { 1 }                                 0
    field $x = 1; method m { $x }                  0
    field $x = 1; ADJUST { $x = 2 }                0
    field $x = 1; ADJUST { $x = 2 } method m { $x} 1   <- ADJUST then method
    field $x = 1; method m { $x } ADJUST { $x = 2} 0   <- REVERSED, clean
    field $x = 1; method m {...} method n {...}     0   <- two methods, clean
    ADJUST { 1 } method m { 1 }                     1   <- no field needed

So it is exactly `ADJUST` immediately followed by `method`. Neither construct
alone refuses, two methods in a row are fine, and reversing the order clears
it. The `field` is irrelevant.

The shape of the failure says `ADJUST`'s block is not being closed, so the
`method` that follows is read as part of it or as a term where a declaration
belongs.

## Perl accepts both orders

Measured on 5.42.0 at `/home/perigrin/.local/bin/perl`, both print `2`:

    use feature "class"; no warnings "experimental::class";
    class Foo { field $x = 1; ADJUST { $x = 2 } method m { $x } }
    print Foo->new->m;                                            -> 2

    class Foo { field $x = 1; method m { $x } ADJUST { $x = 2 } }
    print Foo->new->m;                                            -> 2

Order is not significant to perl. It is to us, which makes this a divergence
on valid input rather than a gap in coverage.

## Why a one-construct-per-case corpus would have missed it

Both constructs pass alone. This is precisely the adjacency property the
`adjacency-*` topics exist to catch, and it is the third such finding in this
milestone: the others were `no feature "signatures"` plus a prototype
(`01a0dd6f`) and a heredoc body plus a following statement.

## The stale citation, which is the second half

The case carries `refuses: 01a0c35f-a487-7aa8-934b-4e68dd0fbaa1`. That issue
is state DONE, milestone m3-conformance-corpus, titled "Tier READMEs, because
the dependency lint needs an INTRODUCES set that cannot drift" -- a
corpus-construction issue that never owned a parser gap. Repoint it here or
drop both lines when the fix lands.

This is the THIRD corpus case found citing a closed m3 issue. The others:
`named-operators`'s `undef` case cited `01a0c730` (now `01a0dd43`), and
`adjacency-07_subroutines` cites `01a0c432` (now `01a0dd6f`). All three
resolve as issues, which is why `TestRefusalCitationMustResolve` passed them;
they are simply the wrong issues. Worth a sweep of every `refuses:` in the
corpus against issue STATE, not just existence.

## Scope

1. Find where a class body's members are parsed and why `ADJUST`'s block does
   not end cleanly before a `method`. Measure whether `BUILD`, a nested
   `class`, or a plain `sub` after `ADJUST` fail the same way -- if they do,
   the fix is one rule about ADJUST's terminator, not one about `method`.
2. Check the reverse-order case still parses after the fix; it passes today
   and is the regression risk.
3. Repoint or remove the case's `refuses:`/`refusal:` lines.

## Context

- paths: `internal/parse/decl.go` (declaration forms), whatever parses a
  `class` body and `ADJUST`
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- entrypoint: the `ADJUST` handling in the class-body parse

## Acceptance criteria

- [ ] An ADJUST block followed by a method parses (`go test ./internal/parse/ -run '^TestAdjustThenMethod$' -count=1 -v | grep -q '^--- PASS: TestAdjustThenMethod'`)
- [ ] The adjacency-11 case parses (`go test ./internal/conformance/ -run 'TestMdtestTopicsMakeTheSameClaims/adjacency-11' -count=1 | grep -q '^ok'`)
- [ ] The case no longer cites a closed issue (`grep -q '01a0c35f' conformance/mdtest/adjacency-11_oo.md; test $? -ne 0`)
- [ ] REGRESSION GUARD, passes today: method-then-ADJUST still parses (`go test ./internal/parse/ -run '^TestAdjustThenMethod$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
