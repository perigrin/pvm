---
title: undef is missing from namedUnary, and its corpus case cites a closed issue
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T10:29:59.070448113Z
updated: 2026-09-26T11:06:33.23291779Z
sessions:
- start_sha: 3ef0ae208b19053e94b91fbac2d6ddf19b4f4af4
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T11:06:33.23291779Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T11:06:33.23291779Z
---

`undef` is missing from `namedUnary`, so the unary spelling refuses on every
container argument. One corpus case, and the case's `refuses:` cites a CLOSED
issue.

## Measured

Probed in `internal/parse`, counting Unknown nodes per spelling:

    undef @a;                    Unknown=1
    my @a=(1,2); undef @a;       Unknown=1
    undef $x;                    Unknown=1
    undef %h;                    Unknown=1
    undef &f;                    Unknown=0   <- the & sigil takes another path
    undef;                       Unknown=0   <- niladic
    my $y = undef;               Unknown=0   <- niladic
    my @b=(1,2); @b = undef;     Unknown=0   <- niladic

So it is the UNARY spelling with a container argument, and the niladic one is
already fine. `undef &f` parsing while `undef %h` does not is the detail that
names the cause: the `&` sigil begins a term the existing path accepts, and the
other three do not.

## Root cause

`internal/parse/keyword.go:31`'s `namedUnary` table has no `undef` entry. The
parser reads `undef` as a complete term and leaves `@a` stranded, which is the
`trailing_tokens` the corpus records.

## It IS a named unary by the table's own method

`keyword.go:6` documents how the set was derived -- deparse with `-p` and see
which side of the comma the argument lands on. Applied to `undef` on 5.42.0:

    perl -MO=Deparse,-p -e 'my $x; my $y; undef $x, $y;'
      -> ($x = undef, $y)          comma OUTSIDE  -> named unary

    perl -MO=Deparse,-p -e 'my $x; my $y; length $x, $y;'
      -> (length($x), $y)          the known-unary control

    perl -MO=Deparse,-p -e 'my $x; my $y; print $x, $y;'
      -> print($x, $y)             the known-list control

perl optimises `undef $x` into `$x = undef`, so the deparse does not show an
`undef(...)` call -- but the comma placement is the classifier the method
names, and it is unambiguous. The table's comment says the set is "80 of them",
derived from all 251 keywords; this is an omission from that sweep, not a
judgement call.

## The stale citation, which is the second half of this issue

`conformance/mdtest/named-operators.md`'s case `undef` is two operators
wearing one word` carries

    refuses: 01a0c730-b241-7765-aaa2-5260d050dce9
    refusal: trailing_tokens

`01a0c730` is state DONE, milestone m3-conformance-corpus, and is
"Thirty-five parsing-distinctive constructs T1 uses that the corpus never
names" -- it never owned this gap. A live refusal citing a closed issue is
exactly what `TestRefusalCitationMustResolve` exists to prevent going
unnoticed, and it slipped through because the citation resolves as an issue;
it is simply the wrong one.

Repoint it at this issue when the fix lands, or drop both lines if the fix
closes the case.

## Scope

1. Add `undef` to `namedUnary`. Measure whether that alone closes all four
   container spellings, or whether `%h` needs more than `@a` does.
2. Repoint the corpus case's `refuses:` at this issue, or remove it.
3. Check whether the case's two token facts (`one word whose text is "print"`,
   `no operator whose text is ")"`) pass once it parses -- the refusal code is
   `trailing_tokens` alone, so they may already hold.

## Context

- paths: `internal/parse/keyword.go` (`namedUnary`, line 31),
  `internal/parse/call.go:96` (the `namedUnary[text]` branch),
  `conformance/mdtest/named-operators.md`
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- entrypoint: `namedUnary` and its three readers in `call.go`

## Acceptance criteria

- [ ] The unary `undef` parses with a container argument (`go test ./internal/parse/ -run '^TestUndefIsANamedUnary$' -count=1 -v | grep -q '^--- PASS: TestUndefIsANamedUnary'`)
- [ ] The corpus case no longer cites a closed issue (`grep -q '01a0c730' conformance/mdtest/named-operators.md; test $? -ne 0`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the conformance suite holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
