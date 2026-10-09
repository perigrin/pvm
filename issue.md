---
title: Decide what a parenless call to an unknown callee does with its arguments
state: done
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0c13f-6816-79d2-945a-b84a12d39e6c
created: 2026-09-20T23:55:47.309990102Z
updated: 2026-10-09T17:21:19.842162981Z
sessions:
- start_sha: 6ae2242dc5addd89075b7c29d0a0777e69d757ef
  end_sha: 3ef0ae208b19053e94b91fbac2d6ddf19b4f4af4
  commits: 4
  started_at: 2026-09-26T09:53:36.480830299Z
  ended_at: 2026-09-26T11:06:23.766682673Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T09:53:36.480830299Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T11:06:23.766682673Z
observed_paths:
- conformance/mdtest/adjacency-06_control.md
- conformance/mdtest/argument-extent.md
- conformance/mdtest/unary.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- internal/conformance/testdata/corpus.ratchet
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/lexer/scanrows_test.go
- internal/lexer/testdata/corpus.ratchet
- internal/lexer/testdata/lex.t.tokens
- internal/parse/call.go
- internal/parse/decl.go
- internal/parse/expr.go
- internal/parse/fidelity.go
- internal/parse/infile_test.go
- internal/parse/keyword.go
- internal/parse/parse.go
- internal/parse/shape.go
- internal/parse/t2_test.go
- internal/parse/term.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parseoracle/parenless_test.go
---

**Rewritten 2026-09-21. This was filed as a three-way design choice for
perigrin. It is not a choice — perl answers it, and the answer rules out two
of the three routes.**

**Measured 2026-09-26 (criterion 4). The measurement CONTRADICTS this body's
central claim about where T2's declarations come from. See "What the
measurement says" below — the import half is worth ZERO on T2, and the
in-file half is worth 316 of 354.**

## What perl does, measured on 5.42.0

Re-verified 2026-09-26 on `/home/perigrin/.local/bin/perl`, v5.42.0:

    $ perl -MO=Deparse -e 'ok 8;'
    Number found where operator expected (Do you need to predeclare "ok"?)
    syntax error near "ok 8"

    $ perl -MO=Deparse -e 'ok 8, 9;'
    syntax error, same shape

    $ perl -MO=Deparse -e 'ok $x, 9;'
    $x->ok, '???'                    <-- INDIRECT OBJECT, not a call at all

    $ perl -MO=Deparse -e 'sub ok {} ok 8, 9;'
    sub ok { } ok 8, 9;              <-- declared first: a call

    $ perl -MO=Deparse -e 'sub f {} f 13 % 4;'
    sub f { } f 1;                   <-- arguments consumed, 13 % 4 folded

So: **perl refuses to compile a parenless call to an undeclared sub with a
numeric argument.** It is not a call perl parses and we fail to; it is a
syntax error in perl too.

Three consequences, each ruling something out:

**Route A (consume greedily) is WRONG, not merely risky.** Building a tree
for `ok 8` where perl builds none is a tree for a program perl rejects --
WRONG in the gate's sense, and M1's gate is Oracle WRONG = 0.

**The status quo is closer to correct than this issue assumed.** For a
genuinely undeclared callee, Unknown IS the honest answer. §4.8.3's
`Call{Resolved:false}` and an Unknown statement around it both say "I cannot
settle this", and perl cannot either.

**Route C is not a heuristic. It is perl's rule.** "Declared before the call
site" is exactly what `knowsShape()` models (call.go:152), what
`comp/proto.t:566` documents in the corpus -- "prototype checking done if
sub pre-declared" -- and what 01a0ad3f already built for imports.

## What the measurement says (2026-09-26, criterion 4)

Probe: parse each of T2's 56 files with `parse.Parse`, walk every Unknown
node, take its source span, and keep the spans whose text OPENS with a
bareword followed by whitespace and a non-delimiter -- the parenless-call
shape -- excluding statement keywords and word operators. For each such
callee, ask two separate questions: is there a `sub NAME` for it anywhere
in the same file's own tree, and would `parse.ParseWithLoader` with a
`DirLoader` on the file's own directory resolve it as an import with
`PrototypeKnown`.

    T2: 550 Unknown nodes total
    T2: 354 open with a bareword + argument (parenless-call shape)
      callee declared by `sub NAME` in the SAME FILE: 316   (89.3%)
      callee declared by a resolvable import:           0   (0.0%)
      callee genuinely undeclared:                      38  (10.7%)

Each number is its own count, not a share inferred from a total.

**THE IMPORT HALF IS WORTH ZERO ON T2, AND THIS BODY'S REASON FOR EXPECTING
OTHERWISE IS FALSE.** This body said "`ok` in T2 **is** declared, because
`Test::More` exports it". Measured: **0 of T2's 56 files contain
`use Test::More`.** T2 is perl's own core suite; it gets its `ok` from
`require './test.pl'` (14 files) or from a `sub ok` written in the file
itself (8 files: `base/num.t`, `comp/colon.t`, `comp/fold.t`,
`comp/line_debug.t`, `comp/redef.t`, `comp/use.t`, `opbasic/concat.t`,
`opbasic/magic_phase.t`). `require` is deliberately not resolved --
`use.go:118` says so, and correctly: `require` is a runtime load with no
import at all. So 01a0ad3f's import machinery is not the thing that reaches
T2. The in-file half is.

The 316 in-file-declared sites, by callee, with the prototype perl reads:

    tryeq=147        sub tryeq ($$$$)        opbasic/arith.t
    ok=67            sub ok ($$) / ($;$) / no proto
    is=29            in-file sub is
    like=13  testing=13  try=11  tryeq_sloppy=11 ($$$$)
    star2=5  sreftest=4  star=4  do_require=3  a_sub=2
    DESTROY=1 a_hash_ref=1 array_ref_plus=1 mymkdir=1
    sub_aref=1 sub_array=1 unop=1

By file:

    opbasic/arith.t=169  comp/parser.t=33  comp/proto.t=32
    comp/colon.t=25      comp/redef.t=20   comp/use.t=10
    comp/fold.t=8        opbasic/magic_phase.t=7  comp/require.t=3
    comp/retainedlines.t=3  comp/line_debug.t=2  comp/uproto.t=2
    comp/hints.t=1       opbasic/concat.t=1

`opbasic/arith.t` alone is 169 of the 316, and its shortfall is 179 -- the
single densest file in the T2 shortfall map, and `tryeq ($$$$)` is why.

The 38 genuinely-undeclared callees are not a gap: `ADJUST` (5, a class
phaser in no table), `O`, `undef`, `is` and `like` where the file uses the
test.pl versions, `continue`, `plan`, `shift`, `split`, `refcount_is`,
`P`, `T`. perl would reject the bareword-call reading of each, so Unknown
stays correct for them.

## A SECOND, STRUCTURAL FINDING the body did not anticipate

`parseRoot` (parse.go:404-420) ALREADY reads the file's own `sub NAME`
declarations into `root.imports` with `PrototypeKnown: true`. Two things
stop that reaching a call site:

1. It runs AFTER the whole file is parsed, so the calls were already
   decided. This is deliberate for the `Imports()` accessor -- a sub may be
   declared below the call -- but it means `knowsShape()` never saw it
   during the parse.
2. It is inside `if res != nil`, and T2 measures with `parse.Parse`, whose
   resolver is nil by contract. So T2 gets no in-file declarations at all,
   with or without a module on disk.

Fixing the in-file half means declaring a `sub NAME` into the shape table
AS IT IS PARSED, which is also what perl does -- top to bottom -- and is
what `conformance/mdtest/argument-extent.md`'s third case pins: `f 1, 2;`
followed by `sub f {}` is a syntax error BECAUSE the declaration is below.
A pre-pass over the whole file would get that case WRONG.

## So what is the actual bug

Not that we decline `ok 8`. That a `sub NAME` declared EARLIER IN THE SAME
FILE is a declaration `knowsShape()` should see, and does not, because the
only path that records one runs after the parse and only under a resolver.

## Scope

1. MEASURED, above. 316 of 354, all in-file; 0 by import.
2. The remaining 38 are genuinely undeclared and stay Unknown as correct
   behaviour.

## The fragments become corpus fixtures

The probes above are a discriminating pair the corpus does not have:
`sub ok {} ok 8, 9` versus bare `ok 8, 9` is precisely what the fidelity
harness exists to score, and nothing in T2 tests the declared/undeclared
boundary directly. Land them as oracle fixtures rather than throwing them
away with the scratch file.

## What the fix cost, and what moved (2026-09-26)

`parseSubDecl` records each `sub NAME` and its prototype into the shape table
AS IT IS READ, before its own body and before anything below it. That is the
whole of the in-file half -- `knowsShape` needed no change, and neither did
`ShapeOf`. The issue's claim that this is "the SAME resolution reaching T2"
is TRUE of the machinery and FALSE of the route: imports were never going to
reach T2, and the in-file case needed a new recording site.

TWO ADJACENT DEFECTS the fix exposed, both pre-existing and both only
reachable once a local sub resolved:

1. A WORD AFTER `->` WAS ROUTED THROUGH `parseWordTerm`. So in a file
   declaring `sub new`, `Foo->new->bar` read `new` as a list operator and
   swallowed `->bar`. 19 T1 files regressed before this was fixed --
   `autoload.t`, `tied_hash_autoviv_refloop_cleanup.t` and 17 more, each
   stranding the second arrow as `not_a_term`. perl reads `C->new->foo` as two
   method calls with `sub new` right there, measured. Fixed in `expr.go`'s
   arrow branch: a method name is a name, and it takes its own parens when the
   source wrote them.
2. `endsArgumentList` DID NOT LIST THE CONCATENATION OR COMPARISON OPERATORS.
   Measured on 5.42.0 with `sub f { 1 }` in scope: `f . 2` is `f() . '2'`,
   `f == 2` is `f() == 2`, `f eq 2` is `f() eq 2` -- all infix, all cutting the
   list -- while `f + 2` is `f(2)`, `f - 2` is `f(-2)`, `f * 2` is `f(*2)` and
   `f <=> 2` pulls in `File::Glob`, all of which START a term. So `.`, `==`,
   `!=`, `=~`, `!~`, `eq`, `ne`, `cmp` were added and `+`, `-`, `*`, `<` were
   deliberately left out.

RATCHET DELTAS, every mover attributed to a file containing a parenless call
to a declared callee:

    T1 parse       493 -> 505 clean, 2280 -> 2248 nodes; 23 better, 0 worse
    T1 fidelity    b_gv_io.t, xml_parser_current_element.t 1 -> 0
    T1 canon       unmoved
    T2 shortfall   comp/colon.t and opbasic/magic_phase.t left the map;
                   11 more improved. opbasic/arith.t 179 -> 21,
                   comp/proto.t 41 -> 4, comp/redef.t 21 -> 1,
                   comp/use.t 11 -> 1, comp/parser.t 55 -> 31
    corpus         204 -> 206 clean, 8 -> 6 refusing

TWO CORPUS CASES CLOSED, and they are the pair this topic exists for.
`conformance/mdtest/argument-extent.md`'s "A parenless call is greedy" and
"A `($)` prototype cuts the extent to one" differ only in the declaration
above a byte-identical call site. perl prints `f[1 2]` and `g[1]2`, measured,
which is what the cases pin. Their `refuses:` markers are removed; their token
facts are untouched and still hold. The third case, "An undeclared callee is a
syntax error", still asserts `parses: no` and still does -- correctly.

ONE RATCHET MOVED THE WRONG WAY WITH A CORRECT TREE. `subroutine.t` 18 -> 19,
on `(blessed $_[0] && 1)`. perl deparses that as `blessed($_[0] && 1)`, which
is exactly what we now build; the refusal is canon dropping the source's
REDUNDANT grouping paren once the call absorbs the list. Untracked: nothing in
the chain owns paren provenance on an absorbed argument list.

ONE DELIBERATE DEVIATION, named in the code. The declaration is recorded
before its own body, which is one shape more permissive than perl: perl needs
a `sub f;` above `sub f { f 1 }` for the inner parenless call, measured. The
shape does not occur -- ZERO parenless calls to the enclosing sub's own name
across all 1,042 T1 and T2 files -- and accepting one is `wider`, not WRONG.

## Acceptance criteria

- [x] The declared/undeclared boundary is a fixture perl and the parser are both scored on (`go test ./internal/parseoracle/ -run '^TestParenlessCallBoundary$' -count=1 -v | grep -q '^--- PASS: TestParenlessCallBoundary'`)
- [x] `sub NAME` earlier in the same file makes a parenless call to it parse (`go test ./internal/parse/ -run '^TestInFileSubDeclarationIsKnown$' -count=1 -v | grep -q '^--- PASS: TestInFileSubDeclarationIsKnown'`)
- [x] A genuinely undeclared parenless call stays Unknown, because perl rejects it too (`go test ./internal/parse/ -run '^TestUndeclaredParenlessCallStaysUnknown$' -count=1 -v | grep -q '^--- PASS: TestUndeclaredParenlessCallStaysUnknown'`)
- [x] The T2 share whose callee is declared is measured and recorded here, before any fix (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

Measured 2026-09-21, re-measured 2026-09-26. Plan: M1. Spec: §4.8.3, §3.2.
