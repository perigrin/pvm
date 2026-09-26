---
title: Ungated defer is declined as a statement form the parser can already read
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0dde4-7e56-7338-81c2-f6670cccbb16
created: 2026-09-23T23:08:33.884597854Z
updated: 2026-09-26T15:47:37.830687536Z
sessions:
- start_sha: 5c84df85b7059abba4a396b055608d6900306f62
  end_sha: 3c0ac58c338f1b2269d264430e141346d1d45b02
  commits: 5
  started_at: 2026-09-26T12:53:13.691758691Z
  ended_at: 2026-09-26T15:47:37.830687536Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T12:53:13.691758691Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T15:47:37.830687536Z
observed_paths:
- conformance/mdtest/read-position.md
- conformance/mdtest/ungated.md
- internal/conformance/conformance_test.go
- internal/conformance/ratchet_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/lexer/brackets.go
- internal/lexer/expect.go
- internal/lexer/intuit.go
- internal/lexer/keyword.go
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/lexer/trivia_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/canon_test.go
- internal/parse/control.go
- internal/parse/fidelity.go
- internal/parse/parse.go
- internal/parse/shape.go
- internal/parse/t2_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use_test.go
- internal/parse/wordblock_test.go
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

---

**Measured 2026-09-26 at 5c84df85 by agent:m1e-1. This is TWO issues, not one,
and the canon defect is the larger of them. Nothing landed.**

## The cause is confirmed and the lever is in TWO places, not one

The mechanism the body names is right. Measured with a token dump:

    map { 1 } @a;    the `{` carries OpensBlock=true    (map is in blockTaking)
    zzz { 1 } @a;    the `{` carries OpensBlock=false   (zzz is not)

So the lexer half is `takesBlock`. But setting it for every word changes
NOTHING on its own -- the parser never reads the bit. `parseWordTerm`'s
`default:` branch (call.go:174) consumes no arguments for an unknown bareword,
and `parseLed`'s `case "(", "[", "{"` (expr.go:82) subscripts without ever
checking `OpensBlock`. Both halves are required.

## Perl gives THREE readings and our Index reading is not one of them

    perl -MO=Deparse -e 'my $v = zzz{a};'              my $v = (zzz { 'a' } );
    perl -MO=Deparse -e 'sub zzz {} my $v = zzz{a};'   my $v = zzz({'a'});
    perl -MO=Deparse -e 'sub zzz(&) {} zzz { 1 };'     &zzz(sub { 1; } );

Undeclared: an indirect method call on the block's value. Declared: a call with
an anonymous hash argument. Prototyped `(&)`: a call with a code reference.
Which one applies is a SYMBOL TABLE question, and all three agree on the
structure -- the brace group and what follows belong to one statement. What
perl NEVER produces is a call SUBSCRIPTED by a hash, and `zzz(){1}` is what we
produce. Confirmed the invocant is the block's VALUE, not an anon hash:

    perl -MO=Concise -e 'zzz { 1 } print "b";'
      <.> method_named[PV "zzz"]      invocant const[IV 1], the scope's value

which is why the runtime error names `package "1"` -- exactly the string the
corpus case records.

## The fix works on the target; the scope of its breakage is the problem

Three changes: `takesBlock` returns true for any word; a new
`case p.wordBlockFollows()` in `parseWordTerm` takes the block then the list;
`"defer"` out of `statementKeywords`. Measured together:

    the corpus case                      Unknown 1 -> 0, and it PASSES
    defer { 1 };                         1 -> 0
    defer { 1 } print "b";               1 -> 0
    zzz { 1 };                           canon zzz(){1}; -> zzz {1;};
    zzz { 1 } print "b";                 1 -> 0
    use feature 'defer'; sub f { ... }   1 -> 0

`go test ./internal/conformance/ -run '.../ungated'` reports the case PASSES and
asks only for the stale `refuses:` line to go: 212 -> 213 clean, 1 -> 0
refusing. Both refusal codes and the token fact are satisfied.

## THE CANON DEFECT IS REAL, SEPARATE, AND PRE-EXISTING

The body asked whether it still reproduces after 01a0dd71. It does, it is
BIGGER than one file, and it is NOT caused by `defer`. Proven at CLEAN HEAD
with no change of mine:

    my @b = map { $_ + 1 } @a;    Unknown=0  canon "my @b = map({$_ + 1;}@a);"
    my @b = map({$_ + 1;}@a);     Unknown=1  canon "my @b = map({$_ + 1});}@a);"

Canon's own output for a form the parser reads CLEANLY today does not re-parse.
The fixpoint is broken for every `map`/`grep`/`sort` block at HEAD. The
mechanism: canon writes a `Block` child as `{` + statements + `}`, and a
Statement child gets a `;`, giving `{$_ + 1;}`. Re-lexed inside the parens the
`{` is classified from XTerm -- `map` is no longer adjacent to it -- so
`OpensBlock` is false, the `;` terminates a statement, and the `}` is orphaned.

This is why the fix breaks 46 files. It makes `print {$fh} ...` PARSE
(Unknown 1 -> 0, a real improvement), which puts a Block into a call's argument
list, which canon then cannot re-emit. Attribution, done both ways:

    45 of the 46 movers contain ZERO `defer`
    42 of the 46 contain `print {`
    3 contain neither and are a DIFFERENT regression (below)

## Two wrong trees the general rule introduces, which are its own defect

    my sub inner { 1 }              ->  my sub();inner({1;}inner())
    return run_callback { ... };    ->  return run_callback;   BYTES DROPPED

The first: `sub` sets `afterDeclName`, so the NAME's following `{` is
OpensBlock, and the new case grabs the name as a word-with-block. The second is
worse than an Unknown -- the block leaves the tree entirely, which the fidelity
harness scores WRONG. `caller_anonymous_callback.t` 1 -> 2 and
`subroutine_prototype.t` 5 -> 6 are these.

## Every ratchet delta, measured

    corpus            212 -> 213 clean, 1 -> 0 refusing        the target
    t1 parse          92 files BETTER, 2 WORSE
                      defer.t 17 -> 5            (43 defers; the body predicted 11)
                      defer_implicit_...t 1 -> 0 (its baseline is 1, not higher)
                      caller_anonymous_callback.t 1 -> 2   WRONG TREE
                      subroutine_prototype.t 5 -> 6        WRONG TREE
    t1canon           46 WORSE (0 -> 1), 1 BETTER      45 of 46 have no `defer`
    TestCanonTokenIdentity  46 files broke, all "Semicolon ;" vs "CloseBracket }"
    t1fidelity        12 WORSE (0 -> 1), 1 BETTER
    t2 shortfall      7 files better (base/num.t 6->3, base/rs.t 3->1,
                      class/method.t 6->3, comp/decl.t 3->2, comp/hints.t 34->31,
                      comp/proto.t 4->3, comp/require.t 7->6)
    TestClassCorpusRatchet  pin 24 -> 21

## Claims in the body measured FALSE

1. "defer.t 17 refusals to 11" -- it is 17 -> **5**.
2. "defer_implicit_return_cleanup.t to zero" -- true, but its baseline is **1**,
   so it is a one-node move, not the larger one the phrasing implies.
3. "TWO ratchets move, not one" -- **FOUR** move, and canon moves 46 files
   rather than `defer.t` alone.
4. "the expression path is fine" -- `my $x = defer { 1 } print "b";` is
   Unknown=0 but canons to `defer(){1}`, the index call perl never produces. It
   is not fine; it is silently WRONG, which is the more expensive failure.
5. Scope item 3, "keep `defer` refusing under `use feature 'defer'` if that is
   what the measurement says" -- the measurement says the OPPOSITE. Gated and
   ungated have the SAME tokens and the SAME statement boundary; they differ
   only in execution order, which is M2's. Verified on 5.42.0: gated prints
   `body D`, ungated prints `D body`. Refusing the gated form would refuse a
   shape we can read.

## What this should become

**This issue cannot land as filed.** The `WORD BLOCK` reading is correct and
worth having; it is blocked behind a canon defect that exists at HEAD, is
independent of `defer`, and is larger than this issue.

Split:

- A NEW issue owning the canon defect: a `Block` in an argument list must be
  re-emitted so it re-lexes as a block. The reproducer needs no `defer` and no
  change to the parser -- `map({$_ + 1;}@a)` is canon's own output and does not
  re-parse. Whoever fixes it should expect `t1canon` to IMPROVE.
- Then this issue, whose fix additionally needs a guard so the new case does not
  fire after `afterDeclName` (`my sub inner { 1 }`) and does not let `return`
  drop the block.

The RED test is written and reproduces both halves --
`TestWordBlockArgIsAStatement` fails at HEAD (exit 1, five shapes) and
`TestWordBlockRoundTrips` PASSES at HEAD and FAILS with the fix, which is the
canon defect isolated as a guard. It is not committed, because committing it RED
violates the 100%-pass rule and committing the fix would land 46 canon
regressions.
