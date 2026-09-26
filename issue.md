---
title: Ungated defer is declined as a statement form the parser can already read
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-23T23:08:33.884597854Z
updated: 2026-09-26T12:52:56.492773027Z
---

**Re-measured 2026-09-26, and the diagnosis below has changed. This was filed
as needing indirect object notation. It does not. It is the same gap
`01a0dd71` fixed for `ADJUST`: a brace group after a word is not read as a
block.**

`defer` sits in `statementKeywords` because the GATED form -- a block that
runs at scope exit -- is unimplemented. That entry also declines the UNGATED
reading, which is one corpus case: `ungated.md/`defer` without its feature
runs the block FIRST`, the last refusal in the corpus.

## Measured, 2026-09-26

    the corpus case                      Unknown=1
    defer { 1 };                         Unknown=1
    defer { 1 } print "b";               Unknown=1
    my $x = defer { 1 } print "b";       Unknown=0   <- expression path is fine
    use feature 'defer'; sub f { ... }   Unknown=1   <- gated refuses too

And the shape that names the cause -- an ARBITRARY word, not `defer`:

    zzz { 1 };              Unknown=0   canon "zzz(){1};"
    zzz { 1 } print "b";    Unknown=1   canon "zzz { 1 } print \"b\";"

`WORD BLOCK` alone parses as an INDEX CALL: `zzz(){1}`, a call subscripted by
an anonymous hash. That is exactly the mechanism `01a0dd71` found for
`ADJUST` (which was in neither phaser table, so `ADJUST { 1 }` was an index
call and everything after it failed) and `01a0ddac` records for a bare block
in a class body.

Three spellings of one gap. `ADJUST` was fixed by adding it to the phaser
tables; `defer` cannot be, because `defer BLOCK` is not a phaser -- but the
reading it needs is the same one.

## What perl does, and it is not indirect object notation

    perl -MO=Deparse -e 'sub f { defer { print "D\n" } print "body\n" }'
      sub f {
          defer {
              print "D\n"
          } print("body\n");
      }

NO ARROW. Deparse keeps the block form. Perl is reading `defer BLOCK ARG`,
not `BLOCK->defer(ARG)`.

The original filing claimed this is the same rule as `new Foo;`. Measured, it
is not: `new Foo` is `WORD BAREWORD`, which perl reads as
`'Foo'->new` -- a genuine method call -- and which we get WRONG in a separate
way (zero Unknowns, canon `new(Foo())`). That is now its own issue,
`01a0ddc5`, and it does not block this one.

## Why the ungated reading matters, from the corpus case

    $ perl -e 'sub f { defer { print "D\n" } print "body\n" } f();'
    D
    body
    Can't locate object method "defer" via package "1"

    $ perl -e 'use feature "defer"; sub f { defer { print "D\n" }
               print "body\n" } f();'
    body
    D

Ungated, the brace group is an anonymous hash constructor, so the block is
evaluated EAGERLY to build the hash -- printing `D` -- and only then does
perl look for a `defer` method and fail. Gated, the block runs at scope exit,
after `body`. Same two lines, opposite order, and the failure arrives after
the damage. Nothing at compile time distinguishes them.

The corpus case pins the ORDER: `D body end`, verified on 5.42.0.

## And a round-trip defect behind it, which is still true

With the entry removed, two T1 files improve -- `defer.t` 17 refusals to 11,
`defer_implicit_return_cleanup.t` to zero -- and one breaks:

    canon_test.go: defer.t: token 399:
      first emission has "CloseBracket }", second has "Semicolon ;"

`defer.t` stops re-emitting identically. Parsing further into the file
exposes a canonicalisation bug the refusal was hiding. TWO ratchets move, not
one: `TestCanonRatchet` also reports `defer.t: 0 -> 1`.

That measurement was taken before `01a0dd71` landed. RE-MEASURE IT -- the
ADJUST fix changed how a brace group after a word is read, and it may have
already fixed or already changed this.

## Scope

1. Re-measure the canon round-trip defect on `defer.t` at current HEAD. It
   may no longer reproduce.
2. Read `WORD BLOCK ARG` at statement level, following `01a0dd71`'s approach
   for `ADJUST`. Measure whether the fix is one rule covering `defer`, `try`,
   `catch`, `finally` and an arbitrary word, or whether `defer` needs its own
   entry -- and prefer the general rule if the measurement allows it.
3. The GATED form stays unimplemented; this issue is only about not declining
   the ungated reading. Keep `defer` refusing under `use feature 'defer'` if
   that is what the measurement says, and record it.

## Context

- paths: `internal/parse/parse.go` (`statementKeywords`, ~line 780),
  `internal/parse/use.go` and `internal/lexer/keyword.go` (where `01a0dd71`
  put ADJUST), `internal/parse/canon.go`
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- precedent: commit `a8094607`, issue `01a0dd71`
- siblings: `01a0ddac` (bare block in a class body), `01a0ddc5`
  (`WORD BAREWORD`)

## Acceptance criteria

- [ ] The ungated corpus case parses (`go test ./internal/conformance/ -run 'TestMdtestTopicsMakeTheSameClaims/ungated' -count=1 | grep -q '^ok'`)
- [ ] `defer BLOCK ARG` at statement level is a statement, not an index call (`go test ./internal/parse/ -run '^TestWordBlockArgIsAStatement$' -count=1 -v | grep -q '^--- PASS: TestWordBlockArgIsAStatement'`)
- [ ] The canon round-trip on `defer.t` holds (`go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the canon ratchet holds (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
