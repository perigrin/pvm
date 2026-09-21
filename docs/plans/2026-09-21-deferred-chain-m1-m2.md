# Deferred chain: M1 and M2 as they stood before the corpus

**Status:** archived 2026-09-21, to be reconstituted after the
conformance corpus is written. This is a DEFERRAL, not a deletion.

The corpus rewrites what these milestones were measuring against, so
they are dropped now and rebuilt against the corpus later. This file
holds everything needed to recreate them: ids, dependency edges, and
full bodies with their measurements.

Completed work is NOT here -- M0's 8 and M1's 32 done issues stay in
the chain as history. Only open work was dropped.

**27 issues**, by milestone:

- `m1-parse-the-core` — 10 (3 in-progress, 7 pending)
- `m2-lower-the-tree` — 16 (16 pending)
- `psc-type-alignment` — 1 (1 pending)

## m1-parse-the-core

### 01a0a8cd — Remove the WRONG pin when the harness measures the new parser

- **id:** `01a0a8cd-8fcc-72a8-9ba1-c916eee9992d`
- **state:** in-progress
- **urgency:** normal
- **created:** 2026-09-16
- **blocked by:** `01a0a70f` — (outside this archive)
- **blocked by:** `01a0b835` — The parse oracle measures tree-sitter, not the parser M1 is building

Two files are pinned as WRONG in the corpus baseline, and the pin comes out
when they do.

    comp/our.t       metric 8
    op/universal.t   metric 16

Both are defects in the TREE-SITTER grammar, which is what the harness still
measures (`run.go:381` calls `parser.New()`). Findings §0.11 has the analysis
and `marker_gaps_test.go` has the regression tests.

## Why they are pinned rather than zero

Commit `a69d8f1c` introduced these and left `TestCorpusBaselineIsIntact`
tripped on purpose, as a signal. It stayed tripped, and a permanently-red
test is one nobody reads -- worse, a red suite cannot show the NEXT
regression. An absolute zero that has never held stops being a gate and
becomes scenery.

The pin names both PATHS rather than a count. A count of 2 would let one
defect be fixed and a different one appear with the gate silent; measured,
that swap passes a count and fails the pin.

## What removes it

Either:

1. **The M1 gate lands** and the harness measures `internal/parse` instead of
   the tree-sitter grammar. This is the expected route, and `comp/our.t`'s
   cause is already fixed there -- the M1 lexer reads `if (/TIE/)` as
   `Quote("/TIE/")`, not division. Verified.
2. **Someone fixes the tree-sitter grammar's two defects** directly. Real
   work on a parser M1-M6 exist to replace, so probably not worth it.

Either way the closing move is the same: delete the pin block in
`ratchet_test.go`, restore `!= 0`, re-baseline in the same commit, and update
findings §0.11.

If only ONE is fixed, replace the pin with the remaining path rather than
loosening it to a count.

## Acceptance

- [ ] The WRONG pin is gone and the gate asserts zero (`go test ./internal/parseoracle/ -run '^TestCorpusBaselineIsIntact$' -count=1 -v | grep -q '^--- PASS: TestCorpusBaselineIsIntact'`)
- [ ] The baseline holds no WRONG rows (`grep -c '^WRONG' internal/parseoracle/testdata/ratchet/baseline.txt | grep -q '^0$'`)
- [ ] Findings §0.11 records the change with the measurement that produced it

Findings: §0.11. Introduced by a69d8f1c; pinned by the commit that closes this issue's parent.

### 01a0b835 — The parse oracle measures tree-sitter, not the parser M1 is building

- **id:** `01a0b835-1270-7b32-9f0a-18fe7c555bae`
- **state:** in-progress
- **urgency:** normal
- **created:** 2026-09-19
- **blocked by:** `01a0ad3f` — (outside this archive)
- **blocks:** `01a0a8cd` — Remove the WRONG pin when the harness measures the new parser

The parse oracle grades the parser M1 exists to replace. `run.go:381` calls
`parser.New()` -- tree-sitter -- so every verdict in the 620-file corpus is a
statement about the OLD grammar, and M1's progress is invisible to it.

## Measured, 2026-09-19 at 528b9d83

    run.go:381                     parser.New()   (tree-sitter)
    grep -c '^WRONG' baseline.txt  2
    WRONG rows                     comp/our.t (8), op/universal.t (16)
    TestCorpusBaselineIsIntact     PASSES, with the pin

M1 built its own ratchet at `internal/parse/testdata/t1.ratchet` rather than
repointing this harness, so the two measure different parsers and neither
knows about the other.

## The seam already exists

`askSubject` (run.go:371-385) has two paths, and only the fallback is pinned
to tree-sitter:

    if opts.Subject != nil { ... }          // external subject binary
    tree, err := parser.New().Parse(src)    // <- line 381
    return TreeSitterSubject(tree), nil

The corpus, the verdict vocabulary, the buckets and the ratchet are all
parser-agnostic. What is missing is a `ParseSubject` alongside
`TreeSitterSubject` that answers `SubjectFacts` from an M1 tree.

## What SubjectFacts demands

The oracle does not ask "did it parse". It compares five markers per file
against perl's own answer, so the adapter has to answer in the same terms
(subject.go:31-64):

1. **`OK`** -- M1 reports this already.
2. **`Declined`** / **`DeclinedReason`** -- M1's `Unknown` node is the
   equivalent of tree-sitter's degenerate node. The field comment anticipates
   exactly this: "every parser can answer that question in its own terms."
3. **`CallSites`** -- each call and what it decided (`srefgen`, `rv2hv`,
   `match`, `readline`, `anonhash`), plus `SiteKindReference` for a backslash
   that is not a call. `sites_decided_test.go` is the start of this but is
   not wired to the oracle's vocabulary.
4. **`Prototypes`** -- needs cross-file resolution, which is 01a0ad3f. This
   is the blocker, not a nicety: without it every imported prototype reads as
   unresolved.
5. **`KnowsCallSites`** / **`KnowsPrototypes`** -- set honestly or an M1 that
   cannot answer is scored as AGREEING with perl. See CompareFacts. Getting
   this wrong flatters the numbers instead of failing.

## Why this is its own issue

01a0a8cd ("Remove the WRONG pin") assumed this repoint as route 1 and waits
on it. It cannot close first: both WRONG rows are still present because they
are tree-sitter defects and tree-sitter is still what gets measured. Deleting
the pin today turns a green gate red for two defects that are really there.

Bundling the repoint into a pin-removal is how that pin got stuck in the
first place -- a69d8f1c left the gate deliberately red, and a permanently-red
test is one nobody reads.

## Acceptance

- [ ] `askSubject`'s default subject is the M1 parser, not `parser.New()`
- [ ] A `ParseSubject` adapter answers all five `SubjectFacts` markers from an
      M1 tree, with `KnowsCallSites`/`KnowsPrototypes` set honestly
- [ ] The corpus is re-measured and the baseline regenerated in the same
      commit, with the bucket movement reported in BYTES as well as counts
- [ ] Whether `comp/our.t` and `op/universal.t` still score WRONG is measured
      and recorded -- findings says the M1 lexer already reads `if (/TIE/)` as
      `Quote("/TIE/")` rather than division, so comp/our.t's cause is fixed
- [ ] Findings §0.11 records which parser the harness measures

Findings: §0.11. Blocks 01a0a8cd. Blocked by 01a0ad3f (prototypes need
cross-file resolution before they can be answered at all).


## MEASURED 2026-09-19 at c3e65b4f: do not repoint yet

Both subjects run over the same 620-file shim, same corpus, same run:

                  tree-sitter    M1 subject
    exact              432          406
    wider               12          110
    WRONG                2           87
    no-answer          174           17

Repointing today takes WRONG from 2 to 87. WRONG is the gate bucket -- a file
the parser gets positively WRONG rather than declines -- so this would be the
largest regression in the project's history, not the win this issue assumed.

## Two things this issue got wrong

**1. The adapter already exists.** `internal/parse/cmd/subject/main.go` builds
SubjectFacts from an M1 tree, and `parse.Sites(root, src)` already returns
`[]parseoracle.SubjectCallSite` with all five markers. TestGoParserSubjectMeasures
has been measuring it over T2 since the M1 gate landed. The work I scoped as
"write a ParseSubject answering five markers" was done before this issue was
filed.

**2. Prototypes were never the blocker.** This issue was filed blocked on
01a0ad3f for cross-file resolution. That chain is now complete (f066f2ab,
7034e553, 9cf2ee6d) and the numbers above are measured WITH it. Prototypes are
not why the repoint fails.

## What actually blocks it

The hedge. An Unknown statement reports five Unresolved sites spanning it --
one per marker -- which `internal/parse/sites.go` documents as the whole
reason Unknown exists: "emit Unknown and stay quiet is not a refusal. It is
the claim that nothing is there."

    $ go run ./internal/parse/cmd/subject t/op/bop.t
    {"ok":true,...,"call_sites":[
      {"kind":"reference","line":22,...,"unresolved":true},
      {"kind":"hash","line":22,...,"unresolved":true},
      {"kind":"match","line":22,...,"unresolved":true},
      {"kind":"readline","line":22,...,"unresolved":true},
      {"kind":"anonhash","line":22,...,"unresolved":true}, ...

Correct, and expensive: every hedged site where perl found nothing scores
against the file. The tree-sitter path avoids this by declining the whole
FILE instead (174 no-answer against 17), so those files are never scored at
all.

So the two subjects are not answering the same question. One declines by the
file, the other by the statement, and the harness's buckets price those very
differently. That is the real work here, and it is a design question about
what a hedge costs -- not a repoint.

## Recommendation

Split. This issue should become:

1. **Price the hedge.** Decide whether five Unresolved sites per Unknown is
   the right report, or whether a statement-level decline should map to
   no-answer for that statement rather than WRONG for the file. That is a
   change to CompareFacts or to Sites, and it needs its own measurement.
2. **Then repoint**, when the M1 subject's WRONG count is at or below
   tree-sitter's 2.

01a0a8cd (remove the WRONG pin) stays blocked on the repoint either way, and
its own analysis is unaffected: both pinned files are tree-sitter defects.

Paused rather than closed: the issue is real and its goal is right, but the
route it describes is not available and the next step is a decision about
hedge pricing that belongs to perigrin.


## 2026-09-19: the earlier diagnosis was partly wrong

I paused this issue saying the blocker was hedge PRICING -- that the M1
subject reports five Unresolved sites per Unknown while tree-sitter declines
whole files, so the two answer different questions. That is still true as far
as it goes, and it is NOT the whole story.

Running the T2 gate (which needs PARSEORACLE_SHIM and had not been run in
this session) surfaced a plain missing-site bug:

    cmd/subval.t  rv2hv  perl read a hash at 2 site(s) the subject did not,
                  and at lines 182, 184 the subject committed to the
                  statement with no such site and no hedge

Both lines are `%$href`. `isHashTerm` tests a LEAF's text for a leading `%`,
so it catches `%hash` and misses every DEREFERENCE -- `%$href` is a Unary
`%` over a Term `$href`, and the Term case sees only `$href`. perl emits
rv2hv for both.

Fixed in 461ea788:

    T2 subject   2 WRONG -> 1, and the 1 is the pinned comp/proto.t M4 TODO
    T2 exact     41 -> 42 of 56 measured (73.2% -> 75.0%)

So some part of the 87 WRONG this issue measured was never about hedging at
all -- it was a site kind the subject did not know how to report. How much is
being measured now over the full corpus.

## What this means for the repoint

The 620-file numbers in this issue were taken before that fix AND before
today's parser work (01a0b983, 01a0b99b, which between them removed 1,025
false unresolved calls from T1). They are stale and should not be used to
decide anything.

The decision this issue was paused on -- how a hedge should be priced --
remains open, but it should be made against a re-measurement, and against the
per-statement findings 01a0ba79 adds. A ranked table of where WRONG comes
from turns "87 versus 2" into a list of things to fix.

### 01a0bae2 — A brace before -> is a hashref, and two anonhash cases the tree misreads

- **id:** `01a0bae2-e4fe-74e9-82cb-562c3fa8adf6`
- **state:** in-progress
- **urgency:** normal
- **created:** 2026-09-19
- **blocked by:** `01a0badc` — (outside this archive)

35 WRONG anonhash sites, and most are span attribution -- but three are real
and each is the "is `{` a block or a hashref" question §7.1.2 names.

## The real cases

**`grep {a=>$_}->{a}`** (op/grep.t:105). The `->` makes the brace a HASHREF,
not grep's block. Measured:

    $ perl -MO=Concise,-exec -e 'my @res = grep {a=>$_}->{a}, ("x");' | grep -c anonhash
    1

**`use overload '%{}' => sub { +{} }`** (op/coreamp.t:23) and
**`'${}' => sub { \my $x }`**. `+{}` is the disambiguating unary plus that
forces a hashref reading, and perl emits NO anonhash for it:

    $ perl -MO=Concise,-exec -e 'my $x = sub { +{} };' | grep -c anonhash
    0

So the subject reporting one there is over-reporting, the opposite direction
from the grep case. Both are the same ambiguity decided differently.

**`\state %y = {1,2}`** (op/lvref.t:370) and **`bless {`** (op/sort.t:944,
op/switch.t:1228) need measuring individually.

## The rest is span attribution

    op/sprintf2.t:347           x8   "my @tests = ("
    mro/package_aliases.t       x6   "for("
    mro/package_aliases_utf8.t  x6   "for("
    op/bop.t:693                x3   "my %op_pairs = ("
    op/pos.t:138                x2   "my %expected = ("

A hash literal assigned to a hash emits no anonhash at all:

    $ perl -MO=Concise,-exec -e 'my %op_pairs = ( a => 1 );' | grep -c anonhash
    0

so these are the first line of a multi-line statement whose ops belong
elsewhere -- 01a0badc's shape. Fix that first and re-measure before touching
this; the residue is what this issue is actually about.

## Acceptance

- [ ] `grep {a=>$_}->{a}` reports an anonhash site (`go test ./internal/parse/ -run '^TestBraceBeforeArrowIsAHashref$' -count=1 -v | grep -q '^--- PASS: TestBraceBeforeArrowIsAHashref'`)
- [ ] `+{}` does NOT report one, matching perl (`go test ./internal/parse/ -run '^TestBraceBeforeArrowIsAHashref$' -count=1 -v | grep -q '^--- PASS: TestBraceBeforeArrowIsAHashref'`)
- [ ] The remaining cases are measured and recorded, blocked on 01a0badc (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.1.2, §4.9.2.


## Re-measured after 01a0badc, 2026-09-19: 35 sites -> 2

This issue was filed at 35 anonhash sites and blocked on the span fix. That
fix landed (09e38fa9) and took it to 2. The span attribution WAS almost all
of it, as this issue predicted.

The two that remain:

    op/coreamp.t:23   use overload '%{}' => sub { +{} }
    op/undef.t:200    my $h= { baz => 1 }; my @k= keys %{($h=undef)||{}};

## What I established

**coreamp.t is a genuine over-report.** `+{}` is the disambiguating unary
plus, and perl emits NO anonhash for it:

    $ perl -MO=Concise,-exec -e 'my $x = sub { +{} };' | grep -c anonhash
    0

We report one. That is a claim perl contradicts, and it is the whole of
this site.

**undef.t is not what I assumed and I did not finish it.** I expected perl
to fold `{ baz => 1 }`, but measuring says otherwise:

    my $h = { baz => 1 };          1 anonhash
    my $h = {};                    0 anonhash, 1 emptyavhv/ANONHASH
    my $h = {}; my $j = {a=>1};    1 anonhash  (the empty one is emptyavhv)

and parse_facts.pl:142 folds emptyavhv-flagged-ANONHASH back into the
marker, so perl's total for that line should be 2 -- which is what we
report. The count agrees and the site is still WRONG, so the disagreement is
about WHICH STATEMENT owns one of them, not how many exist.

That needs the oracle's own line list for the file, which I did not run.
Paused there rather than guessing: two sites does not justify more
speculation, and a wrong diagnosis committed to an issue is worse than an
open question.

## Recommendation

Worth doing as a pair with whoever next touches anonhash, or folding into
the "hedge versus detect" decision 01a0bae1-b76f raises -- `+{}` may be
better answered by declining than by adding a special case for the unary
plus.

### 01a0c10b — T2 core is 30.4% against a 100% target: what the 39 files actually need

- **id:** `01a0c10b-7148-77e5-8923-16fbce93c64b`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c13f` — Measure the 147 unbucketed Unknowns in four comp files before fixing any of them
- **blocked by:** `01a0c13f` — T2's target excludes what M1 does not own, and says which milestone does

T2 core is M1's headline metric and it is not close. The plan's target is
**100% of 56 files**; the measurement is **30.4%, 17 files**, 39 short.

Nothing on the chain owns that gap. This issue records what the gap IS, so
the work can be decomposed from a measurement rather than from a guess.

## The measurement, 2026-09-20 at b6e7c351

649 Unknown nodes across 44 of the 56 files. Bucketed by construct, not by
file:

      179   opbasic/arith.t    `%` in a PARENLESS call's arguments
      ~151  many              a parenless call whose argument is a bare number
       65   comp/parser.t     (unbucketed)
       48   base/num.t        leading-decimal literals, among others
       41   comp/proto.t      M4 prototype work, not M1's
       37   base/lex.t        heredocs, leading decimals

## Four constructs isolated by minimal reproduction

Each measured against this parser; the first two are most of the corpus.

**1. `%` as modulus in a parenless call.** 179 nodes, all of arith.t.

    tryeq $T++, 13 % 4, 1, 'x';     1 Unknown
    f 13 % 4;                       1 Unknown
    f($a, 13 % 4);                  0   <-- inside parens it works
    my $x = 13 % 4;                 0

So the lexer's expect state is wrong for an argument position after a
parenless call name: it wants a term, reads `%` as a hash sigil, and the
statement falls to Unknown. §3.2 says a leading `%` in TERM position is a
sigil and nowhere else -- the bug is that this position is not a term.

**2. A parenless call with a bare number argument.**

    ok 8;                           1 Unknown
    ok 8 - 1;                       1 Unknown

`ok`, `print` and `is` account for ~151 rows between them, and this is the
shape. Likely the same root cause as (1): the argument position is not being
entered correctly.

**3. A leading-decimal literal.**

    my $x = .5;                     1 Unknown
    my $x = int($f * 100 + .5);     2 Unknown
    my $x = 0.5;                    0

**4. Heredocs in statement position.**

    print <<'EOF';\nbody\nEOF\n    1 Unknown

## What this issue is

A MEASUREMENT, filed so the T2 gap is visible on the chain. The fixes belong
in their own issues, decomposed from the buckets above -- (1) and (2) may
well be one issue and one root cause, which is worth establishing before
either is scoped.

comp/proto.t's 41 are M4's and should be excluded from any T2 target M1 is
held to.

## Acceptance

- [ ] Each bucket above has an issue or is explicitly out of M1's scope, recorded here
- [ ] The T2 shortfall map's per-file counts are regenerated after each lands (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

## Why this is not itself the fix

Filing 39 files' worth of work as one issue is the thing that produced a
32/35 milestone with two targets missed. The buckets are the units; this is
the survey that names them.

Measured 2026-09-20. Plan: M1. Spec: §3.2, §7.3.

### 01a0c13f — Decide what a parenless call to an unknown callee does with its arguments

- **id:** `01a0c13f-50ee-77c8-afe5-9c5e6ebb9638`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocks:** `01a0c13f` — A parenless call consumes its arguments, so its statement stops falling to Unknown

**Rewritten 2026-09-21. This was filed as a three-way design choice for
perigrin. It is not a choice — perl answers it, and the answer rules out two
of the three routes.**

## What perl does, measured on 5.42.0

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

## So what is the actual bug

Not that we decline `ok 8`. That `ok` in T2 **is** declared -- `Test::More`
exports it -- and we are not using that. The T1 measurement from 01a0ad3f
says the machinery works when the module is reachable: 460 -> 532 clean
files, 3,322 -> 2,531 Unknown nodes, when Test::More resolved.

The work is to make the same resolution reach T2, and to check the in-file
`sub NAME` case, which needs no module at all.

## Scope

1. Measure how many of T2's ~330 parenless-call Unknowns have a callee that
   IS declared -- by an import whose module is readable, or by a `sub NAME`
   earlier in the same file. That number is the size of the real fix.
2. The remainder are genuinely undeclared. Confirm perl also rejects them,
   file-by-file, and they stay Unknown as correct behaviour.

## The fragments become corpus fixtures

The probes above are a discriminating pair the corpus does not have:
`sub ok {} ok 8, 9` versus bare `ok 8, 9` is precisely what the fidelity
harness exists to score, and nothing in T2 tests the declared/undeclared
boundary directly. Land them as oracle fixtures rather than throwing them
away with the scratch file.

## Acceptance

- [ ] The declared/undeclared boundary is a fixture perl and the parser are both scored on (`go test ./internal/parseoracle/ -run '^TestParenlessCallBoundary$' -count=1 -v | grep -q '^--- PASS: TestParenlessCallBoundary'`)
- [ ] `sub NAME` earlier in the same file makes a parenless call to it parse (`go test ./internal/parse/ -run '^TestInFileSubDeclarationIsKnown$' -count=1 -v | grep -q '^--- PASS: TestInFileSubDeclarationIsKnown'`)
- [ ] A genuinely undeclared parenless call stays Unknown, because perl rejects it too (`go test ./internal/parse/ -run '^TestUndeclaredParenlessCallStaysUnknown$' -count=1 -v | grep -q '^--- PASS: TestUndeclaredParenlessCallStaysUnknown'`)
- [ ] The T2 share whose callee is declared is measured and recorded here, before any fix (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

Measured 2026-09-21. Plan: M1. Spec: §4.8.3, §3.2.

### 01a0c13f — A parenless call consumes its arguments, so its statement stops falling to Unknown

- **id:** `01a0c13f-6816-79d2-945a-b84a12d39e6c`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c13f` — Decide what a parenless call to an unknown callee does with its arguments
- **blocks:** `01a0c13f` — Measure the 147 unbucketed Unknowns in four comp files before fixing any of them
- **blocks:** `01a0c13f` — T2's target excludes what M1 does not own, and says which milestone does

The decision issue establishes WHICH route the `default:` arm of
`internal/parse/call.go` takes for a parenless call to a callee this parser
does not know. This issue takes it, and moves the two ratchets in the same
commit.

## Prerequisites

- [ ] Decide what a parenless call to an unknown callee does with its
      arguments (the blocker). Do not start before it names a route.

## Context

- paths: `internal/parse/call.go`, `internal/parse/shape.go`,
  `internal/parse/t2_test.go`, `internal/parse/testdata/t1.ratchet`
- commands: `go test ./internal/parse/ -count=1`
- entrypoints: `internal/parse/call.go:160-175` (the `default:` arm),
  `internal/parse/shape.go:149` (`knowsShape`)

## What this fixes

Every parenless call to a sub this parser has not seen declared, WITH ANY
ARGUMENT, currently loses its enclosing statement to Unknown -- not because
the Call node is wrong (§4.8.3 is satisfied: `Call{Resolved:false}` is
emitted) but because no arguments are consumed and the leftover tokens sink
the statement.

Measured at b6e7c351: ~330 of T2's 649 Unknown nodes, including all 179 in
`opbasic/arith.t` and the ~151 across `ok`/`is`/`print` sites. This is the
single largest item in the T2 gap.

## Steps

- [ ] Write a failing test asserting `ok 8;` parses with no Unknown node,
      alongside `ok $x`, `ok "s"`, `ok 8,'name'` and `f 13 % 4`, in a file
      that declares or imports the callee the way the chosen route requires
- [ ] Run it and watch it fail for the right reason -- the statement is
      Unknown, not the Call (`go test ./internal/parse/ -run '^TestParenlessCallTakesArguments$' -count=1 -v`)
- [ ] Implement the chosen route, minimally
- [ ] Run the test to green
- [ ] Run the guard the blocker's route-A risk names: a bareword hash key
      (`$h{foo}`), a class name (`Foo->new(1)`) and a fat comma LHS
      (`foo => 1`) each still parse as they do today, with no Call swallowing
      the rest of the statement
- [ ] Measure the T1 ratchet delta in NODES and in BYTES, and the T2 per-file
      movement, before regenerating anything
- [ ] Regenerate `internal/parse/testdata/t1.ratchet` and update the
      `shortfall` map in `internal/parse/t2_test.go` in the SAME commit, with
      the byte measurement in the commit message
- [ ] Run the whole package: `go test ./internal/parse/ -count=1`
- [ ] Commit

## Acceptance Criteria

### Positive Scenarios

- [ ] A parenless call to an unknown-or-declared callee consumes its
      arguments and the statement is not Unknown (`go test ./internal/parse/ -run '^TestParenlessCallTakesArguments$' -count=1 -v | grep -q '^--- PASS: TestParenlessCallTakesArguments'`)
- [ ] A bareword that is NOT a call -- hash key, class name, fat comma LHS --
      still parses unchanged (`go test ./internal/parse/ -run '^TestBarewordIsNotAlwaysACall$' -count=1 -v | grep -q '^--- PASS: TestBarewordIsNotAlwaysACall'`)
- [ ] The T2 shortfall map is regenerated in the same commit and the metric is
      green (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The T1 ratchet is regenerated in the same commit and is green (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] The whole parse package is green (`go test ./internal/parse/ -count=1`)

Spec: §4.8.3. Plan: M1. Survey: 01a0c10b.

### 01a0c13f — A leading-decimal literal is a number, not an Unknown

- **id:** `01a0c13f-97f5-7f98-b32d-07245ec6ddfe`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocks:** `01a0c13f` — Measure the 147 unbucketed Unknowns in four comp files before fixing any of them

A numeric literal with no digit before the point is a number in perl and an
Unknown here.

## Context

- paths: `internal/lexer/` (number scanning), `internal/parse/term.go`,
  `internal/parse/t2_test.go`, `internal/parse/testdata/t1.ratchet`
- commands: `go test ./internal/lexer/ ./internal/parse/ -count=1`
- entrypoints: wherever the lexer scans a numeric literal; the `.` is being
  read as an operator rather than as the start of a number.

## Measured, 2026-09-20 at b6e7c351

    my $x = .5;                    1 Unknown
    my $x = int($f * 100 + .5);    2 Unknown
    my $x = 0.5;                   0

`base/num.t` carries 48 Unknowns and `base/lex.t` 37; leading decimals are
named in the survey as a contributor to both. The exact share is what the
first step measures.

A leading `.` is unambiguous in TERM position: nothing else can start there.
In operator position it is concatenation, which must keep working. The rule
is the lexer's expect state, the same mechanism §3.2 uses for a leading `%`.

## Steps

- [ ] Measure how many of `base/num.t`'s 48 and `base/lex.t`'s 37 Unknowns
      are leading decimals, and record the number here
- [ ] Write a failing test: `my $x = .5;` has no Unknown node, and `.5`
      lexes as one numeric literal whose text is `.5`
- [ ] Run it and confirm it fails for that reason (`go test ./internal/lexer/ -run '^TestLeadingDecimalLiteral$' -count=1 -v`)
- [ ] Implement: in term-expecting position, a `.` followed by a digit starts
      a number
- [ ] Write the companion test that concatenation is unharmed: `$a . 5`,
      `$a .5` and `"x" . .5` each keep their operator reading
- [ ] Run both to green
- [ ] Measure the T1 delta in nodes AND bytes; regenerate
      `internal/parse/testdata/t1.ratchet` and the `shortfall` map in
      `internal/parse/t2_test.go` in the SAME commit
- [ ] Run `go test ./internal/lexer/ ./internal/parse/ -count=1`
- [ ] Commit

## Acceptance Criteria

### Positive Scenarios

- [ ] A leading-decimal literal is one number and the statement is clean (`go test ./internal/lexer/ -run '^TestLeadingDecimalLiteral$' -count=1 -v | grep -q '^--- PASS: TestLeadingDecimalLiteral'`)
- [ ] Concatenation still reads as concatenation (`go test ./internal/lexer/ -run '^TestDotIsStillConcatenation$' -count=1 -v | grep -q '^--- PASS: TestDotIsStillConcatenation'`)
- [ ] The T2 metric is green with the shortfall map regenerated in the same
      commit (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The T1 ratchet is green with its baseline regenerated in the same commit (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)

Spec: §3.2. Plan: M1. Survey: 01a0c10b bucket 3.

### 01a0c13f — A heredoc in statement position parses, body and all

- **id:** `01a0c13f-aaf8-7c2f-a037-754feea1cf77`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocks:** `01a0c13f` — Measure the 147 unbucketed Unknowns in four comp files before fixing any of them

A heredoc introducer in statement position sinks its statement.

## Context

- paths: `internal/lexer/` (heredoc scanning), `internal/parse/call.go`,
  `internal/parse/t2_test.go`, `internal/parse/testdata/t1.ratchet`
- commands: `go test ./internal/lexer/ ./internal/parse/ -count=1`
- entrypoints: the lexer's heredoc handling; then whatever consumes the
  introducer token in `internal/parse`.

## Measured, 2026-09-20 at b6e7c351

    print <<'EOF';
    body
    EOF
                                   1 Unknown

`comp/require.t`'s shortfall comment in `internal/parse/t2_test.go` already
records this shape by name: reaching inside a `BEGIN { ... }` block "exposes
a heredoc the parser does not yet read". `base/lex.t` (37) is named in the
survey as heredocs plus leading decimals.

The first step is a SPLIT question: does the lexer emit a heredoc token whose
body runs to the terminator and the parser drops it, or does the lexer not
recognise the introducer at all? The fix lands in a different file each way,
so answer it before changing anything.

## Steps

- [ ] Determine whether the lexer produces a heredoc token for the three-line
      form above, and record the answer here. Measure how many of
      `base/lex.t`'s 37 and `comp/require.t`'s 12 Unknowns are heredocs.
- [ ] Write a failing test: the heredoc statement above parses with no
      Unknown node and round-trips byte-identically
- [ ] Run it and confirm the failure is the one measured (`go test ./internal/parse/ -run '^TestHeredocInStatementPosition$' -count=1 -v`)
- [ ] Implement minimally, in whichever layer the first step named
- [ ] Add the quoting variants the corpus uses: `<<'EOF'`, `<<"EOF"`,
      `<<EOF` and the indented `<<~EOF`
- [ ] Run to green
- [ ] Measure the T1 delta in nodes AND bytes; regenerate
      `internal/parse/testdata/t1.ratchet` and the `shortfall` map in
      `internal/parse/t2_test.go` in the SAME commit
- [ ] Run `go test ./internal/lexer/ ./internal/parse/ -count=1`
- [ ] Commit

## Acceptance Criteria

### Positive Scenarios

- [ ] A heredoc in statement position parses with no Unknown node (`go test ./internal/parse/ -run '^TestHeredocInStatementPosition$' -count=1 -v | grep -q '^--- PASS: TestHeredocInStatementPosition'`)
- [ ] All four quoting forms parse and round-trip (`go test ./internal/parse/ -run '^TestHeredocQuotingForms$' -count=1 -v | grep -q '^--- PASS: TestHeredocQuotingForms'`)
- [ ] The T2 metric is green with the shortfall map regenerated in the same
      commit (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The T1 ratchet is green with its baseline regenerated in the same commit (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)

Plan: M1. Survey: 01a0c10b bucket 4.

### 01a0c13f — Measure the 147 unbucketed Unknowns in four comp files before fixing any of them

- **id:** `01a0c13f-c7d7-7d9c-9a85-60a59aa26bc0`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c13f` — A parenless call consumes its arguments, so its statement stops falling to Unknown
- **blocked by:** `01a0c13f` — A leading-decimal literal is a number, not an Unknown
- **blocked by:** `01a0c13f` — A heredoc in statement position parses, body and all
- **blocks:** `01a0c10b` — T2 core is 30.4% against a 100% target: what the 39 files actually need

Four files carry 147 of T2's 649 Unknown nodes and nobody knows what is in
them. This issue MEASURES first and files second. Do not guess a cause and do
not write a fix inside this issue.

## Context

- paths: `internal/parse/t2_test.go` (the shortfall map is the inventory)
- commands: `go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v`
- entrypoints: `internal/parse/t2_test.go`'s `shortfall` map; the corpus at
  `$PERL5_CORPUS/t/` (default `~/dev/perl5/t/`)

## What is unmeasured, at b6e7c351

    comp/parser.t   65 Unknowns
    comp/hints.t    36
    comp/colon.t    25
    comp/redef.t    21
    -------------------
                   147

The survey (01a0c10b) bucketed 502 of T2's 649 nodes by construct. These 147
it listed as "(unbucketed)". Two of them already have partial notes in the
`shortfall` map's comments -- `comp/parser.t` is recorded as holding a
symbol-table dereference `$::{"..."}` that `TestDerefBlockContentsAreDecided`
names as still opaque -- but no file has a construct breakdown.

Note the overlap risk: the parenless-call fix and the leading-decimal fix may
account for a share of these. Measure AFTER those land, or measure now and
re-measure after, but say which.

## Why measurement is the deliverable

The survey exists because filing 39 files of work as one issue is what
produced a milestone with two targets missed. Filing a FIX for 147 unbucketed
nodes would repeat that at smaller scale. The unit of work is a construct
with a minimal reproduction, and this issue produces those units.

## Steps

- [ ] For each of the four files, extract every Unknown node's source span
      and group the spans by construct
- [ ] Reduce each group to a minimal reproduction of two or three lines,
      measured -- the form `f 13 % 4;  1 Unknown` the survey uses
- [ ] Record the per-file breakdown in this issue: construct, node count,
      minimal reproduction
- [ ] For each construct that is M1's, file its own issue in
      `m1-parse-the-core` with the measurement and a failing-test step
- [ ] For each construct that is NOT M1's (prototypes are M4's; anything the
      m2 refinement already blocked, such as block-form filehandles), record
      the exclusion here with the milestone it belongs to
- [ ] Commit the breakdown as a note where it can be found -- the shortfall
      map's comments are where the other per-file findings live

## Acceptance Criteria

### Positive Scenarios

- [ ] Every one of the 147 nodes is attributed to a named construct or
      explicitly recorded as residual, with the arithmetic shown
- [ ] Each M1-owned construct has its own issue in `m1-parse-the-core` (`git zhi issue list --milestone m1-parse-the-core | grep -c .`)
- [ ] The findings are committed beside the counts they explain, so the next
      reader does not re-measure (`grep -q 'comp/parser.t' internal/parse/t2_test.go`)
- [ ] The T2 metric is still green -- this issue changes no parse behaviour (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)

Plan: M1. Survey: 01a0c10b, the "(unbucketed)" row.

### 01a0c13f — T2's target excludes what M1 does not own, and says which milestone does

- **id:** `01a0c13f-e0e2-7b4f-a9d2-06f7edad1e7f`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c13f` — A parenless call consumes its arguments, so its statement stops falling to Unknown
- **blocks:** `01a0c10b` — T2 core is 30.4% against a 100% target: what the 39 files actually need

M1's gate says "100% of T2 (56 files)". Some of those 56 files cannot be M1's
to fix, and a target that includes them is unreachable by construction. The
plan itself warns against exactly this: "Do not carry an unreachable 100% into
a gate: that is how a gate becomes advisory."

This issue makes T2's denominator honest, in the test, with each exclusion
named and attributed to the milestone that owns it.

## Context

- paths: `internal/parse/t2_test.go`,
  `docs/plans/2026-09-05-parser-conformance-plan.md` (§"M1 -- Parses the core
  without error")
- commands: `go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v`
- entrypoints: `internal/parse/t2_test.go`, `TestT2CoreParses` and its
  `shortfall` map

## What is out of M1's scope, measured at b6e7c351

    comp/proto.t          41 Unknowns   prototype recognition -- M4
    t/class/*.t (9 files) 52 Unknowns   class syntax; ADJUST and class
                                        phasers are recorded in the shortfall
                                        map as owned by NOTHING on the chain
    block-form filehandle  1 per site   `print {$fh} "x"` -- blocked per the
                                        m2 refinement

`comp/proto.t` is the unambiguous one: M4 is titled "Prototypes, context, and
the hard ambiguities", and 41 nodes of prototype work counted against M1's
100% is 41 nodes M1 cannot move without doing M4's job early.

The class files need a judgement, not an assumption -- some class syntax is
already landed (`class/class.t` left the map) and some is not. The step below
says to decide each with a measurement rather than by category.

## What this issue does NOT do

It does not lower the bar to whatever today's number happens to be. A file is
excluded only when the construct blocking it is owned by a named later
milestone or a named blocked issue. Everything else stays in, at 100%.

## Steps

- [ ] Write a failing test asserting T2's M1-scoped denominator: the file
      list minus the exclusions, each exclusion carrying a comment naming the
      milestone or issue that owns it
- [ ] Run it and see it fail on today's numbers (`go test ./internal/parse/ -run '^TestT2ScopeIsM1s$' -count=1 -v`)
- [ ] Implement the split in `TestT2CoreParses`: report BOTH numbers -- clean
      files over all 56, and clean files over the M1-scoped denominator -- so
      the plan's metric and the reachable one are both visible and neither is
      quietly replaced
- [ ] For each excluded file, confirm by measurement that its Unknowns are
      ACTUALLY the excluded construct, not a mixture. A file whose Unknowns
      are half prototypes and half something M1 owns is not excluded; it is
      partially fixable and stays in.
- [ ] Update the plan's M1 table to state the scoped target beside the raw
      one, with the exclusion list and its reason
- [ ] Run to green and commit

## Acceptance Criteria

### Positive Scenarios

- [ ] T2's M1-scoped denominator is asserted in the test, with each exclusion
      naming its owning milestone (`go test ./internal/parse/ -run '^TestT2ScopeIsM1s$' -count=1 -v | grep -q '^--- PASS: TestT2ScopeIsM1s'`)
- [ ] `TestT2CoreParses` reports both the raw 56-file rate and the M1-scoped
      rate, and is green (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] `comp/proto.t` is excluded with M4 named as its owner (`grep -q 'comp/proto.t' internal/parse/t2_test.go`)
- [ ] The plan records the scoped target beside the raw one (`grep -qi 'scoped' docs/plans/2026-09-05-parser-conformance-plan.md`)

Plan: M1, M4. Survey: 01a0c10b's closing note.

## m2-lower-the-tree

### 01a0c0e7 — The lowering pass: internal/infer must read internal/parse, not tree-sitter

- **id:** `01a0c0e7-8234-7f04-80e5-1fab8682f9bd`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c115` — Retire the tree-sitter vocabulary from internal/infer
- **blocks:** `019d1449` — Expand PSC inference to cover more Perl type semantics

`internal/infer` reads the tree-sitter tree through `internal/parser/`. That
tree is being burned down, so every inference rule written against it is work
that gets deleted. Nothing in `internal/infer` should be extended until it
reads `internal/parse/` instead.

## Why no partial work survives

§6.1.6 counted it, and the count is the argument:

> `internal/infer/*.go` switches on 34 distinct tree-sitter-perl grammar
> names (`subroutine_declaration_statement`,
> `ambiguous_function_call_expression`, `array_element_expression`,
> `hash_deref_expression`, `func0op_call_expression`, `refgen_expression`,
> ...), and `internal/parse/parse.go:146-201` emits 23 coarse kinds with none
> of those names (`statement`, `term`, `binary`, `index`, `declaration`,
> `call`, ...). Porting the facade hands PSC a `Kind()` that returns
> `"binary"` where it expects `"assignment_expression"`. **Ten methods port;
> the vocabulary does not.**

So this is not a migration that can be done incrementally behind a shim. The
vocabulary changes underneath every switch.

## And the old tree is re-derived from TEXT in places

§4.14.2 records where `internal/infer` reaches past the tree entirely:

    findOperatorText    infer.go:645-659
    callElementType     infer.go:1313, telling map from grep by
                        strings.HasPrefix on source

Those are the places new rules would have to reproduce, and they are exactly
what a lowered tree exists to remove.

## Scope

The lowering pass of §6.1.6: a pass over `internal/parse`'s tree producing
the vocabulary §4.14 specifies, which `internal/infer` then consumes instead
of `internal/parser`.

§6.1.6 is explicit about what NOT to build with it: "One lowering pass,
specified in §6.1.6, and no dataflow IR until a feature demands one." Neither
HIR nor PIR as perl-lsp built them.

Key facts by NODE IDENTITY, not byte offset. §6.1.6 item 4, and rust-analyzer
rejected the offset key by name: "offsets change after edits. So, as a rule
of thumb, we avoid using offsets, text ranges or syntax trees as keys and
values for queries". §6.4.3 observes the same defect in PSC's existing
`map[uint32]types.Type`.

## Size

5,012 non-test lines in `internal/infer` at 3b9422eb, 34 grammar names to
re-express. This is a multi-session milestone-sized piece of work and should
probably be decomposed before it is executed -- filed here so that the thing
everything else waits on is ON the chain rather than implied by a spec
section.

## What it unblocks

  019d1449  Expand PSC inference       -- blocked on this, now explicitly
  area 5 of that issue                 -- NarrowByContext has no source of
                                          context until the lowered tree
                                          carries Wantarray and call-site
                                          context

## Acceptance

Deliberately not written. The ACs for a pass this size should come from a
decomposition pass (crochet:refinement) against §6.1.6 and §4.14, not from
one sitting.

Plan: M2/M3 adjacent -- it is what makes §6.1.6's "PSC is being converted to
the new tree regardless" true. Spec: §6.1.6, §4.14, §6.4.3.

### 01a0c10e — Re-measure §4.14.2 against HEAD and rewrite the status table

- **id:** `01a0c10e-6a8d-721d-be45-a07ae1ee2033`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocks:** `01a0c10f` — Lower literals, names and Hole: NumLit, StrLit, Interp, QwList, Bareword
- **blocks:** `01a0c110` — Lower variables and subscripts: Var, Deref, Index, Key with Arrow
- **blocks:** `01a0c111` — Lower operators: Binary.Op from the field, Repeat and Assign from the Paren flag
- **blocks:** `01a0c112` — Lower declarations: Decl kinds, the list form by child count, sub declarations
- **blocks:** `01a0c112` — Lower calls: five sites to one Call, and the filehandle slot from the Handle flag
- **blocks:** `01a0c113` — Lower special forms: MapGrepSort.Op, the eval boundary, and the renames

## Context

- paths: `docs/specs/perl-parser/04-expressions.md` (§4.14.2, §4.14.3, §4.14.6), `internal/parse/parse.go`, `internal/parse/statustable_test.go` (new)
- commands: `go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2; `internal/parse/parse.go:236-299`

§4.14.2 is the decomposition axis for the whole lowering (seam D slices by its
SHIPPED / LOWERING / PARSER status column). Its baseline is commit 3c997a6b and
HEAD is 39+ commits past it. The table is measurably STALE, in the direction of
LESS work: several rows marked PARSER now parse clean.

Verified at HEAD (b6e7c351) by parsing each fragment and dumping the tree:

	map { $_ } @a;          0 Unknowns; call(map) resolved=true with a block child
	sort { $a <=> $b } @a;  0 Unknowns
	grep { $_ } @a;         0 Unknowns
	delete $h{k};           0 Unknowns; call(delete) with an index child
	exists $h{k};           0 Unknowns
	sub { 1 };              0 Unknowns; declaration(sub) with a block child
	my @r = (1..5);         0 Unknowns; binary ".." -- the "Range infers Str" row is GONE
	$h->{k};                index "{" with Arrow SET
	(a => 1, b => 2);       list with Fat SET on the two barewords
	($x); ("a") x 3;        term with Paren SET
	print $fh "x";          term with Handle SET
	print {$fh} "x";        1 Unknown  -- STILL BLOCKED
	new Foo(1);             1 Unknown  -- STILL BLOCKED
	qx/ls/;                 parses as term "qx/ls/" -- splitting still PARSER
	&$code(1);              1 Unknown  -- STILL BLOCKED

Bigger than the row edits: **§4.14.3 asks for three bits on `parse.Node` --
`Paren`, `Arrow`, and the filehandle slot -- and all three SHIPPED, plus `Fat`.**
`parse.go:236-299` documents `Arrow`, `Paren`, `Fat` and `Handle` with the same
"meaningful on one kind, false elsewhere" contract as `Resolved`. Every §4.14.3
row that reads "PARSER" for one of those four is wrong at HEAD.

The grammar-name count is also contested: the parent brief says 34, §4.14 says
57, §6.1.4/§6.4.1 say 34. The count that matters is stated here so it stops
drifting: **distinct string literals appearing in a `case` arm of a switch on
`node.Kind()` across the four non-test files of `internal/infer`.** Measure it,
state the method inline in the spec next to the number, and reconcile the three
sites to that one figure. (A raw `case "..."` grep over the same four files
yields 83, which is an upper bound -- it includes switches on operator text and
on sigils, not just `Kind()`.)

This issue produces a spec update plus a REGRESSION TEST that fails the next
time the parser makes a row stale, so the table cannot silently rot again.

## Steps

- [ ] Write `internal/parse/statustable_test.go` with `TestStatusTableAtHead`: a
      table of (source fragment, expected Unknown count, expected flags set)
      covering every §4.14.2 row whose status this issue changes, plus the four
      rows that stay PARSER. Assert against `parse.Parse`.
- [ ] Run to verify failure: `go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v`
- [ ] Fix the EXPECTATIONS to the measured truth, not the parser -- this test
      pins reality, it does not change it. Re-run to green.
- [ ] Commit the test.
- [ ] Count the grammar names by the stated method over `internal/infer/infer.go`,
      `symbols.go`, `project.go`, `diagnostics.go`.
- [ ] Rewrite §4.14.2 status column at HEAD. Change the preamble baseline from
      3c997a6b to the HEAD sha. Every changed row cites the fragment and the
      measured tree.
- [ ] Rewrite §4.14.3: `Paren`, `Arrow`, `Handle`, `Fat` are SHIPPED; strike the
      "Three of these ask for a bit on `parse.Node`" paragraph, which is done.
- [ ] Reconcile the grammar-name count across §4.14, §6.1.4, §6.4.1 and the
      parent brief 01a0c0e7 to one number with the method stated.
- [ ] Run the full parse suite: `go test ./internal/parse/ -count=1`
- [ ] Commit the spec update.

## Acceptance Criteria

### Positive Scenarios
- [ ] The status table is pinned by an executable test that measures the real parser (`go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v | grep -q "^--- PASS: TestStatusTableAtHead"`)
- [ ] The parse package is green with the new test in it (`go test ./internal/parse/ -count=1`)
- [ ] §4.14.2 no longer claims its 3c997a6b baseline (`test -z "$(grep -c "against 3c997a6b" docs/specs/perl-parser/04-expressions.md | grep -v "^0$")"`)

### 01a0c10e — Node identity: stable ids for internal/lower that survive an edit above them

- **id:** `01a0c10e-c81e-7565-88fb-d973f18ed667`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocks:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocks:** `01a0c113` — The position map: a side table keyed by id, with erased nodes mapping forward
- **blocks:** `01a0c115` — Swap the annotation key from StartByte to node identity

## Context

- paths: `internal/lower/id.go` (new), `internal/lower/id_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestIDStableUnderEditAbove$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/06-incremental-lsp.md` §6.1.6 item 4, §6.4.3

This is the highest-risk item in the milestone and it has zero dependencies, so
it starts first. §6.1.6 item 4 is binding: **facts are keyed by node identity,
never by byte offset.** rust-analyzer rejects offset keys by name, and §6.4.3
records the same defect in PSC today -- `infer.go:57-60` keys
`map[uint32]types.Type` on `StartByte`, so every edit invalidates the whole map,
and `infer.go:184` lets a parent written after its child at the same offset
overwrite the child.

**The acceptance test that separates a real scheme from a fake one**: insert one
statement at the TOP of a two-sub file, re-parse, re-lower, and assert that no
id inside either sub changed. Byte offset fails it. Preorder index fails it.
Both are the obvious implementations, so the test exists to reject them.

The shape §6.1.6 item 4 names is `crates/span/src/ast_id.rs`: an enumeration
index by kind, name and disambiguator -- ids are allocated per enclosing
declaration, so the counter inside `sub b` restarts and does not see what
happened above it. Rule 6 of §4.14.1 constrains the meaning: identity is per
SYNTAX SITE. One id for one `sub { $i }`, even though three loop iterations
create three cells; a table keyed on a captured value keys on
`(node id, capture instance)`, not on node id alone. That is a scope-resolution
concern and no cell node exists in §4.14, so nothing here needs to model it --
but the id must not be defined in a way that forecloses it.

Scope: `internal/lower/id.go` only. No node types, no lowering rules. It ships
the allocator and its invariant test.

## Steps

- [ ] Write `TestIDStableUnderEditAbove`: parse `sub a { my $x = 1; } sub b { my $y = 2; }`,
      allocate ids over the tree, record them; parse `my $z = 0;` + the same
      text, allocate again, and assert every id inside `sub a` and `sub b` is
      unchanged.
- [ ] Run to verify failure (the package does not exist yet):
      `go test ./internal/lower/ -run "^TestIDStableUnderEditAbove$" -count=1 -v`
- [ ] Implement the allocator in `internal/lower/id.go`. Per-declaration
      enumeration by kind + name + disambiguator, in the manner of
      `ast_id.rs`. Run to green. Commit.
- [ ] Write `TestIDDistinguishesSameSpanNodes`: the `infer.go:184` defect
      directly -- parse `my $x = 5; my @y = ($x); my $z = $y[0];`, assert that
      the nodes sharing a start byte get DIFFERENT ids, and that a map keyed by
      id holds one entry per node rather than collapsing them.
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Write `TestIDDistinguishesSiblingsWithSameShape`: two structurally
      identical statements in one sub (`my $a = 1; my $a = 1;`) get different
      ids -- the disambiguator does its job.
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] An edit above a sub leaves every id inside it unchanged (`go test ./internal/lower/ -run "^TestIDStableUnderEditAbove$" -count=1 -v | grep -q "^--- PASS: TestIDStableUnderEditAbove"`)
- [ ] Nodes that share a start byte get distinct ids (`go test ./internal/lower/ -run "^TestIDDistinguishesSameSpanNodes$" -count=1 -v | grep -q "^--- PASS: TestIDDistinguishesSameSpanNodes"`)
- [ ] Structurally identical siblings get distinct ids (`go test ./internal/lower/ -run "^TestIDDistinguishesSiblingsWithSameShape$" -count=1 -v | grep -q "^--- PASS: TestIDDistinguishesSiblingsWithSameShape"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c10f — The §4.14 vocabulary types and Hole in internal/lower/node.go

- **id:** `01a0c10f-37f8-7a43-b3ce-ea194f7defe0`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10e` — Node identity: stable ids for internal/lower that survive an edit above them
- **blocks:** `01a0c10f` — Lower literals, names and Hole: NumLit, StrLit, Interp, QwList, Bareword
- **blocks:** `01a0c110` — Lower variables and subscripts: Var, Deref, Index, Key with Arrow
- **blocks:** `01a0c111` — Lower Slice{Kind}: the five-way switch §4.14.1 rule 5 specifies
- **blocks:** `01a0c111` — Lower operators: Binary.Op from the field, Repeat and Assign from the Paren flag
- **blocks:** `01a0c112` — Lower declarations: Decl kinds, the list form by child count, sub declarations
- **blocks:** `01a0c112` — Lower calls: five sites to one Call, and the filehandle slot from the Handle flag
- **blocks:** `01a0c113` — Lower special forms: MapGrepSort.Op, the eval boundary, and the renames
- **blocks:** `01a0c113` — The position map: a side table keyed by id, with erased nodes mapping forward

## Context

- paths: `internal/lower/node.go` (new), `internal/lower/node_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14 (the Go listing), §4.14.6

Pure declaration: the node types §4.14 specifies, with `nodeBase` carrying the
identity from `internal/lower/id.go`. No lowering rules -- those are separate
issues and they each depend on this one.

Two binding constraints from §6.1.6 shape the declaration:

**Every lowered field is optional (§6.1.6 item 5).** The CST this pass reads is
routinely broken; a lowering that assumes well-formed input panics on the
half-typed code the LSP exists for. rust-analyzer makes every accessor optional
for exactly this reason.

**`Hole` may stand anywhere a `Node` may (§4.14.2, `Hole` row).** All 4,944 T1
`Unknown` nodes are childless leaves -- 0 with children, 0 nested. The census
shows `Unknown` as a `binary` operand (281), an `index` base or subscript (74),
and once each under `unary` and `declaration`. §4.14.6 leaves open whether
`Decl.Init`, `Index.Base` and `Match.Target` tolerate a `Hole`; this issue
SETTLES that as yes, by test, because the census already shows the CST producing
those positions.

§4.14.6 open questions this issue closes by choosing, with the choice recorded
in a comment at the declaration:

- **`Paren`: node or bit.** §4.14.1 rule 1 says the information survives and
  `parse.Node.Paren` now SHIPS. Whether the lowered tree carries a `Paren` node
  or a `Parenthesized` flag on `Repeat`/`Assign`/`Decl` is "settled by writing
  those two rules and the position-map entry". Only two rules read it. Make the
  call here and say why.
- **`MethodCall.Indirect` / `.Ambiguous`.** `new Foo(1)` is still `Unknown` at
  HEAD, so the honest-uncertainty channel carries no signal. §4.14.6: "If it is
  permanently out of M1 scope, drop both fields rather than ship fields that are
  always false." Decide; do not ship an always-false field.
- **`AnonList`/`AnonHash` vs `ArrayLiteral`/`HashLiteral`.** §4.14.6 marks this
  "perigrin's call". Ask rather than guess.

Out of scope, explicitly: no dataflow IR, no HIR, no PIR (§6.1.6). This file
declares one tree.

## Steps

- [ ] Write `TestHoleStandsAnywhereANodeMay`: construct a `Decl` with a `Hole`
      `Init`, an `Index` with a `Hole` `Base`, a `Match` with a `Hole` `Target`,
      and a `Binary` with `Hole` operands. Assert each constructs and that a
      generic child walk reaches the `Hole` without panicking.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v`
- [ ] Declare the §4.14 types in `internal/lower/node.go` with `nodeBase`
      embedding the identity from `id.go`. Run to green. Commit.
- [ ] Write `TestZeroValueNodeIsUsable`: every declared type's zero value can be
      walked and its optional fields read as nil without panicking (§6.1.6 item 5).
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Record the three §4.14.6 decisions in comments at the declarations, and
      update §4.14.6 to strike the ones now settled.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] A Hole is accepted in Decl.Init, Index.Base, Match.Target and as a Binary operand (`go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v | grep -q "^--- PASS: TestHoleStandsAnywhereANodeMay"`)
- [ ] Every node type is usable at its zero value (`go test ./internal/lower/ -run "^TestZeroValueNodeIsUsable$" -count=1 -v | grep -q "^--- PASS: TestZeroValueNodeIsUsable"`)
- [ ] The package compiles and is green (`go test ./internal/lower/ -count=1`)

### 01a0c10f — Lower literals, names and Hole: NumLit, StrLit, Interp, QwList, Bareword

- **id:** `01a0c10f-9671-733f-8b98-30a78f634b0f`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocked by:** `01a0c10e` — Re-measure §4.14.2 against HEAD and rewrite the status table
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/lower/literals.go` (new), `internal/lower/literals_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Literals and names"; §4.14.1 rule 3

First slice of the lowering rules, chosen first because it is the smallest and
establishes the walk shape the other five slices copy. Status per §4.14.2, as
re-measured by the status-table issue:

- `NumLit` SHIPPED -- `term "42"`, `term "0x1f"`, `term "1_000"`; `IsInt` from `Text`
- `StrLit` SHIPPED -- delimiter is `Text[0]`
- `StrLit.Backtick` LOWERING -- `Text[0] == backtick`. `qx//` stays PARSER
  (`internal/lexer/quote.go:22-31` has no `qx` row); at HEAD `qx/ls/` parses as
  one `term "qx/ls/"`, so the lowering may emit a `StrLit` with `Backtick` set
  and the body unsplit, or a `Hole`. Pick one and say why.
- `QwList` LOWERING -- `term "qw(a b)"` is opaque and **§4.14.2 assigns the body
  split to the lowering**. perl: `(qw(a b)) x 2` and `("a","b") x 2` both give
  `a b a b`; scalar `qw(a b c)` and `("a","b","c")` both give `c`.
- `Bareword{Quoted}` LOWERING -- positional over `Call{Resolved:false}`, which
  §4.14.1 rule 3 shows is overloaded four ways: an unresolved call; a hash key
  (`$h{k}` -> `[index "{" [term "$h"] [call "k"]]`); the left of a fat comma;
  a class name as invocant; and the stat-buffer bareword (`-r _`). Separate them
  by position -- second child of an `Index{"{"}`, left operand of `Binary "=>"`,
  left operand of `Binary "->"`, operand of a file test -- and produce
  `Bareword{Quoted:true}` for the first two.
  **At HEAD the fat-comma case is easier than §4.14.2 says**: `parse.Node.Fat`
  ships and is set on the bareword inside a `list`, so that arm reads a flag
  rather than a position. Verified: `(a => 1, b => 2)` gives a `list` whose two
  `call` children carry `Fat`.
- `Interp` PARSER -- `"a$b c"` is one `term`; the lexer defers interpolation.
  Emit a `StrLit{Interpolated:true}` with `Parts` empty rather than a `Hole`,
  since the text is intact.
- `Hole` LOWERING -- one `Hole` per `Unknown`, per constraint 3 of the milestone.
  Every T1 `Unknown` is a childless leaf (0 with children, 0 nested), so this is
  a leaf-to-leaf rename plus the id.

## Steps

- [ ] Write `TestLowerLiterals`: a table of (source, expected lowered shape)
      over each row above, including `qw(a b)` splitting to its words and a
      backtick string setting `Backtick`.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerBarewordByPosition`: `$h{k}`, `a => 1`, `Foo->new` and
      `-r _` each produce the node §4.14.1 rule 3 asks for. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLowerUnknownToHole`: a fragment that parses to `Unknown`
      (`new Foo(1);`) lowers to exactly one `Hole` with a span covering the same
      bytes. Run to verify failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Literals lower to their §4.14 kinds with IsInt, Backtick and the qw body split (`go test ./internal/lower/ -run "^TestLowerLiterals$" -count=1 -v | grep -q "^--- PASS: TestLowerLiterals"`)
- [ ] The four overloads of Call{Resolved:false} separate by position (`go test ./internal/lower/ -run "^TestLowerBarewordByPosition$" -count=1 -v | grep -q "^--- PASS: TestLowerBarewordByPosition"`)
- [ ] Each Unknown lowers to one Hole covering the same span (`go test ./internal/lower/ -run "^TestLowerUnknownToHole$" -count=1 -v | grep -q "^--- PASS: TestLowerUnknownToHole"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c110 — Lower variables and subscripts: Var, Deref, Index, Key with Arrow

- **id:** `01a0c110-839f-77fb-8a3d-30fa5a8d6e96`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocked by:** `01a0c10e` — Re-measure §4.14.2 against HEAD and rewrite the status table
- **blocks:** `01a0c111` — Lower Slice{Kind}: the five-way switch §4.14.1 rule 5 specifies
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/lower/vars.go` (new), `internal/lower/vars_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Variables and dereference"; §4.14.1 rules 2 and 7

Status per §4.14.2, corrected at HEAD:

- `Var{Sigil,Name}` SHIPPED — `term "$x"`, `"@a"`, `"%h"`, `"*foo"`, `"$#a"`,
  `"&foo"`, `"$Foo::x"`; the sigil is the leading bytes. `$::x` is still PARSER.
  §4.14.1 rule 7: `(Sigil, Name)` is the identity pair and the sigil is
  load-bearing — `$_`, `@_` and the file-test bareword `_` are three different
  things. This matters concretely: **three paths in `internal/infer` re-derive
  `@_` and `$_[0]` from source text** (`infer.go:3659`, `:3753`, `:4193`). The
  lowering produces `Var{"@","_"}` and `Index{Var{"$","_"}, 0}` so those rules
  compare fields instead.

  The consumer evidence that the sigil is a field and not three kinds: at 8 of 9
  switch sites in `internal/infer` the `scalar`/`array`/`hash` arms are identical
  modulo the sigil string (`infer.go:1669-1673`, `:3560-3564`, `:3851-3855`,
  `:4072-4077`, `:1942-1946`, `:3088-3090`, `:1402-1406`), and
  `lookupNarrowedType` (`:1890-1896`) and `sigildName` (`symbols.go:403`) already
  take the sigil as a parameter.

- `Index`, `Key` SHIPPED. **`Arrow` is no longer PARSER**: `parse.Node.Arrow`
  ships (`parse.go:270`) and is set in the arrow-before-subscript branch
  (`expr.go:83-91`). Verified at HEAD: `$h{k}` gives `index "{"` with no flag and
  `$h->{k}` gives `index "{" arrow`.

  §4.14.1 rule 2 is why it matters, and it is a semantic difference rather than a
  spelling: `$h{k}` reads `%h`, `$h->{k}` reads `$h`. Measured on perl 5.42.0,
  `my %h = (k => "hash"); my $h = { k => "ref" }; print $h{k}, $h->{k};` prints
  `hashref`.

  What inference actually branches on is one step removed — "is the base a
  reference being dereferenced". `inferElementType` (`infer.go:1270-1280`) reads
  the old grammar's `container_variable`-vs-`scalar` split for exactly this, and
  `$$x[0]` is a dereference with no arrow. **So the lowering derives deref-ness
  as `Arrow || base is a Deref`, and keeps `Arrow` itself for the LSP.** Ship
  both; they are not the same predicate.

  `{` selects `Key`, `[` selects `Index`; chains nest (`$x->{a}[0]{b}`).

- `Deref{Sigil,Expr}` PARSER — `${$x}`, `$$x`, `@{$x}`, `@$x`, `%$x`, `$#{$x}`,
  `$#$x` are each one childless `term` (issue 01a0ad52). Emit the best node the
  text supports and a `Hole` for the inner expression rather than guessing; do
  not re-lex the span.

- `Deref{Postfix}` LOWERING for `->@*` / `->%*`: `$x->@*` is
  `[binary "->" [term "$x"] [term "@*"]]`, verified at HEAD. `->$#*` and postfix
  slices stay PARSER.

Out of scope: `Slice{Kind}` is its own issue — it is a five-way switch and the
largest free block in §4.14.2.

## Steps

- [ ] Write `TestLowerVarAndSubscript`: `$x`, `@a`, `%h`, `*foo`, `$#a`, `&foo`,
      `$Foo::x` lower to `Var` with the right `(Sigil, Name)` pair; `$_` and `@_`
      are distinguishable; `$h{k}` and `$h->{k}` lower to `Key` differing in
      exactly `Arrow`; `$h->{a}[0]{b}` nests.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerDerefness`: `$h->{k}`, `$$x[0]` and `${$x}[0]` all report
      deref-ness true, and `$h{k}` reports false — the predicate
      `Arrow || base is Deref`, which is what `inferElementType` needs.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerPostfixDeref`: `$x->@*` and `$x->%*` lower to
      `Deref{Postfix:true}` with the right sigil. Run to verify failure,
      implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Var carries (Sigil, Name) and the two subscript spellings differ in exactly Arrow (`go test ./internal/lower/ -run "^TestLowerVarAndSubscript$" -count=1 -v | grep -q "^--- PASS: TestLowerVarAndSubscript"`)
- [ ] Deref-ness is Arrow-or-base-is-Deref, so `$$x[0]` counts and `$h{k}` does not (`go test ./internal/lower/ -run "^TestLowerDerefness$" -count=1 -v | grep -q "^--- PASS: TestLowerDerefness"`)
- [ ] Postfix deref lowers for `->@*` and `->%*` (`go test ./internal/lower/ -run "^TestLowerPostfixDeref$" -count=1 -v | grep -q "^--- PASS: TestLowerPostfixDeref"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c111 — Lower Slice{Kind}: the five-way switch §4.14.1 rule 5 specifies

- **id:** `01a0c111-72d3-70fa-88a8-a9a387f04261`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocked by:** `01a0c110` — Lower variables and subscripts: Var, Deref, Index, Key with Arrow
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/lower/slice.go` (new), `internal/lower/slice_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.1 rule 5

§4.14.2 calls `Slice{Kind}` "one five-way switch, and the largest single block
of this list the lowering gets for free". It is carved out of the variables
issue for that reason: it is self-contained, and it is the clearest measurable
win in the milestone.

**Why the five kinds must not collapse.** They produce four different result
types (§4.4.7). Re-measured on perl 5.42.0: `@a[0,2]` gives elements,
`@h{qw(x y)}` gives values, `%a[0,1]` gives index/element pairs, `%h{'x'}` gives
key/value pairs. Collapsing them to one "slice" loses exactly the information
PSC needs most.

**The CST already carries the discriminant.** Every slice is `Kind: Index` —
the sole `Kind: Index` construction site is `parseSubscript` in
`internal/parse/expr.go` — and the pair `(base child's first byte, Index.Text)`
selects the kind. Verified at HEAD:

	@a[0,1]     index "["  over term "@a"     -> SliceArray
	@h{'a','b'} index "{"  over term "@h"     -> SliceHash
	%a[0,1]     index "["  over term "%a"     -> SliceKVArray
	%h{'a'}     index "{"  over term "%h"     -> SliceKVHash
	(1,2)[0]    index "["  over a list        -> SliceList

Nothing has to be re-derived from source text; the discriminant is two field
reads.

**The defect this removes, measured.** The old grammar's `slice_expression`,
`slice_container_variable` and `anonymous_slice_expression` appear NOWHERE in
`internal/infer`, and the consequence is measurable: `my @a=(1,2); my @s=@a[0,1];`
gives `@a elem=Int` but `@s elem=Unknown`. Inference branches only two ways on
this today (KV or not, because only the element type is at stake) and does not
branch at all yet. The other three values serve the LSP and diagnostics.

**Do not follow B::SoN here.** It maps `aslice`, `kvaslice`, `hslice`, `kvhslice`
all to one `Slice` (`FromOptree/OpMap.pm:209-231`) and `aelem`, `helem`,
`multideref`, `gelem` all to one `Subscript` (`:205-208`), because the optree it
reads has already typed the result. §4.14.1 rule 5 is explicit: that is a phase
difference, not a disagreement — PSC is upstream of it and must not follow.

The `SliceList` case matters for a second reason: §4.14.1 rule 1 records that
`(1,2)[0] x 3` is an `index` over a `list`, not a parenthesised operand, and
perl agrees (`my @a = ((1,2)[0]) x 3` gives `1 1 1`). Getting `SliceList` right
here keeps the Paren rule honest in the operators issue.

## Steps

- [ ] Write `TestLowerSliceKind`: a five-row table, one per `SliceKind`, each
      asserting the lowered node's `Kind` field from the fragments above.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v`
- [ ] Implement the five-way switch on `(base sigil, Index.Text)`. Run to green.
      Commit.
- [ ] Write `TestLowerSliceIsNotSubscript`: `@a[0]` lowers to a `Slice`, and
      `$a[0]` lowers to an `Index` — the sigil on the base is what separates
      them, not the bracket. Run to verify failure, implement, run to green,
      commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] All five SliceKind values are produced from their fragments (`go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v | grep -q "^--- PASS: TestLowerSliceKind"`)
- [ ] A slice and a single-element subscript lower to different kinds (`go test ./internal/lower/ -run "^TestLowerSliceIsNotSubscript$" -count=1 -v | grep -q "^--- PASS: TestLowerSliceIsNotSubscript"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c111 — Lower operators: Binary.Op from the field, Repeat and Assign from the Paren flag

- **id:** `01a0c111-8c3e-7451-97d9-10307441ba2d`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocked by:** `01a0c10e` — Re-measure §4.14.2 against HEAD and rewrite the status table
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/lower/ops.go` (new), `internal/lower/ops_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerBinaryOpFromField$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Operators"; §4.14.1 rule 1

This slice kills one of the three measured consumer defects the lowering exists
for: **`findOperatorText` (`infer.go:645-659`) re-derives the operator by
scanning a node's unnamed children for text that matches an operator table.**
`parse.Node.Text` already holds the operator, so the lowering SETS `Binary.Op`
from a field read and that function has no successor.

Status per §4.14.2, corrected at HEAD:

- `Unary{Op,Postfix}` SHIPPED — `unary "!"`, `"neg"`, `"~"`, `"++"`, `"ref"`
  (for backslash), `"pos"` (for unary plus); `postfix "++"`. `Postfix` is a KIND
  split in the CST (`Kind == Postfix` vs `Kind == Unary`), so the flag is a
  rename. `extractNegatedGuard` (`infer.go:2285`) currently tests text `== "!"`;
  the lowered `Op` replaces that comparison.
- `Binary{Op}` SHIPPED; **the lowering sets `Op` from `Text`.** The bridge to
  B::SoN's split family is a total table from `Op` and belongs in the consumer,
  not here: `==`→NumEq, `eq`→StrEq, `<`→NumLt, `lt`→StrLt, `.`→Concat,
  `x`→Repeat, `//`→DefinedOr. Keep one `Binary{Op}`.
- `Ternary` SHIPPED — `ternary "?:"`.
- `CmpChain` SHIPPED — `cmp_chain "<"` with flat operands. **No consumer**:
  `internal/infer` has no chained-comparison rule at all. §4.14.6 leaves the
  typing rule open and says it is "settled by writing the rule and seeing
  whether it wants a flat operand list or nested Binarys". Lower it faithfully
  (flat, as the CST has it) and leave the typing to the inference issue; do not
  invent a rule here.
- `Range{Exclusive}` — `$a..$b` is `binary ".."`, `$a...$b` is `binary "..."`.
  **The measured defect in §4.14.2 is GONE at HEAD**: `my @r = (1..5);` now
  parses as `declaration "my"` over `term "@r"` and `binary ".." paren` with
  `term "1"` and `term "5"` children. It no longer reads as `Str`, and the
  lexer no longer takes `1..5` as one Number token. Re-verify this before
  implementing — if it holds, `Range` is LOWERING for both variable and literal
  endpoints, which §4.14.2 marked PARSER.
- `Repeat{ListRepeat}` LOWERING. **`parse.Node.Paren` ships**, so §4.14.1 rule
  1's fact is a field read rather than the unsound span heuristic the rule
  describes at length. Verified at HEAD: `('a') x 3` gives
  `binary "x"` over `term "'a'" paren` and `term "3"`. `qw(a b) x 2` is the
  other list-repeat spelling and keys on the `qw` prefix of the term's `Text`.
  perl: `('a') x 3` gives three elements, `'a' x 3` gives one.
  **Do not reimplement the span heuristic.** §4.14.1 rule 1 spends a paragraph
  showing it is unsound (721 of 6,574 T1 nodes with a paren-leading span are
  not parenthesised operands, and `PrototypeNode` leaves carry their parens in
  `Text`). The flag exists now; read it.
- `Assign{ListAssign}` LOWERING — `binary "="`, `"+="`, ...; `($a,$b) = (1,2)`
  has a `list` LHS, `@a = (1,2)` an aggregate-sigil `term`, and a single
  parenthesised scalar sets `Paren`. `my (...) =` is under `Decl` and belongs to
  the declarations issue.
- `List{Fat}` SHIPPED as `List`, and **`Fat` is no longer PARSER**:
  `parse.Node.Fat` ships and is set. Verified at HEAD: `(a => 1, b => 2)` gives
  a `list` whose two `call` children carry `Fat`. §4.5.4: the fat comma quotes
  the word to its left, so without this the lowering could not tell `(a, 1)`
  (a call) from `(a => 1)` (a string).
- `Paren` — §4.14.6 leaves node-vs-flag open; the vocabulary issue settles it.
  Whichever it chose, this issue is where the two rules that read it get
  written, which is what §4.14.6 says settles the question.

## Steps

- [ ] Write `TestLowerBinaryOpFromField`: a table over `+ - . x == eq < lt // ..
      ... =~ !~ =` asserting `Binary.Op` matches `parse.Node.Text` exactly, and
      that no source rescan is needed. Include a case with an operator that is
      also a substring of an operand, which is what `findOperatorText` gets
      wrong.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerBinaryOpFromField$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerListRepeatFromParenFlag`: `('a') x 3` lowers to
      `Repeat{ListRepeat:true}`, `'a' x 3` to `ListRepeat:false`, and
      `qw(a b) x 2` to `true` via the `qw` prefix. Run to verify failure,
      implement, run to green, commit.
- [ ] Write `TestLowerRangeEndpoints`: `$a..$b`, `1..5` and `$a...$b` each lower
      to a `Range` with both endpoints present and `Exclusive` set only for the
      three-dot form. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerListFat`: `(a => 1, b => 2)` lowers to a `List` whose
      `Fat` slice marks elements 0 and 2, and `(a, 1, b, 2)` marks none.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerAssignListness`: `($a,$b) = (1,2)` and `@a = (1,2)` set
      `ListAssign`, `$a = 1` does not. Run to verify failure, implement, run to
      green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Binary.Op comes from the node's field, retiring findOperatorText (`go test ./internal/lower/ -run "^TestLowerBinaryOpFromField$" -count=1 -v | grep -q "^--- PASS: TestLowerBinaryOpFromField"`)
- [ ] List repeat reads the Paren flag, not a span heuristic (`go test ./internal/lower/ -run "^TestLowerListRepeatFromParenFlag$" -count=1 -v | grep -q "^--- PASS: TestLowerListRepeatFromParenFlag"`)
- [ ] Ranges keep both endpoints, literal ones included (`go test ./internal/lower/ -run "^TestLowerRangeEndpoints$" -count=1 -v | grep -q "^--- PASS: TestLowerRangeEndpoints"`)
- [ ] Fat commas survive inside a list (`go test ./internal/lower/ -run "^TestLowerListFat$" -count=1 -v | grep -q "^--- PASS: TestLowerListFat"`)
- [ ] List assignment is distinguished from scalar assignment (`go test ./internal/lower/ -run "^TestLowerAssignListness$" -count=1 -v | grep -q "^--- PASS: TestLowerAssignListness"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c112 — Lower declarations: Decl kinds, the list form by child count, sub declarations

- **id:** `01a0c112-8a65-7d5c-94ec-69859b2928d3`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocked by:** `01a0c10e` — Re-measure §4.14.2 against HEAD and rewrite the status table
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/lower/decl.go` (new), `internal/lower/decl_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerDeclListForm$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Declarations"; §4.12.2 rule 2

Status per §4.14.2:

- `Decl{Kind}` SHIPPED — `declaration` with `Text` the declarator. T1 census of
  `Declaration.Text`: my 4,905, sub 1,775, package 569, local 398, our 318,
  state 9, field 7, class 5, method 4. `internal/infer` handles `my` only;
  `local`/`state`/`our`/`field` are unhandled. Lower all of them into
  `Decl.Kind` — the lowering's job is to carry the fact, not to type it.
  §4.14.6 leaves the *inference* treatment of `local`/`state`/`our`/`field`
  open, and `local`'s restore-on-scope-exit is a flow fact belonging to §6.1.7,
  which is unbuilt. Out of scope here.

- `Decl{Vars,Init}` LOWERING, and this is the semantically load-bearing part.
  The CST has TWO shapes, verified at HEAD:

	my $x = f()       -> [declaration "my" [term "$x"] [call "f"]]
	                     two children, no "="
	my ($x) = f()     -> [declaration "my" [binary "=" [term "$x" paren] [call "f"]]]
	                     one child, the "="
	my ($x,$y) = f()  -> the same with a list LHS

  **The list-vs-scalar fact §4.12.2 rule 2 needs is the child count.** Measured
  on perl: with `sub f {(1,2,3)}`, `my ($x)=f()` gives 1 and `my $y=f()` gives 3.
  The predicate §4.14.2 states is "one child and it is `binary "="`".

  §4.14.2 recommends normalising the parser to one shape before the lowering is
  written. **That recommendation is declined here and the reason is recorded**:
  it is a parser change with a corpus-wide blast radius, the predicate is two
  field reads, and `parse.Node.Paren` now also marks the parenthesised LHS
  independently — so the lowering has two agreeing signals and needs neither a
  parser change nor a source rescan. If the executor finds the predicate does
  not hold over the corpus, file the parser normalisation rather than widening
  this issue.

- `Decl{Attrs}` LOWERING — `my $x :shared` gives `term ":shared"` as a second
  child.
- `Decl{RefAlias}` LOWERING — `my \$x = $y` gives
  `[declaration "my" [binary "=" [unary "ref" [term "$x"]] [term "$y"]]]`.
  No consumer in `internal/infer`; §4.14.2 marks the consumer SPECULATIVE. Lower
  it because it is nearly free and the shape is unambiguous, but write no
  inference rule for it.
- sub declarations SHIPPED — `sub f { 1 }` is
  `[declaration "sub" [term "f"] [block ...]]`; `sub f;` likewise without a
  block. `internal/infer` uses the old `subroutine_declaration_statement` at 8
  sites; all eight lower to this one shape. Attributes
  (`sub f :prototype($) { 1 }`) are still PARSER — the declaration spans `sub f`
  and the rest is `Unknown`, so emit a `Hole`.
- signatures PARSER — without the feature, `sub f ($x, $y) { 1 }` gives
  `[declaration "sub" [term "f"] [prototype "($x, $y)"] [block ...]]`, which is
  what perl does too. With `use v5.36;` the same text mis-parses (the body is
  read as a hash slice of the signature) because `internal/lexer/proto.go`
  returns false when `l.signatures` is set and nothing downstream picks the
  signature up. `internal/infer` reads `signature` (2) and
  `mandatory_parameter` (2), both unwrapped to their `scalar`. Lower the
  prototype form; emit a `Hole` for the 5.36 mis-parse rather than lowering a
  wrong tree.

- `AnonSub` — **SHIPPED in expression position at HEAD, and the statement-position
  defect is FIXED.** §4.14.2 records `sub { 1 };` producing two statements, a
  wrong tree that round-trips. Verified at HEAD it is now one
  `declaration "sub"` with a `block` child. Re-verify before implementing. In
  expression position `my $f = sub { 1 }` gives
  `[declaration "my" [term "$f"] [declaration "sub" [block ...]]]`. `Sig` stays
  PARSER per the signatures row.

## Steps

- [ ] Write `TestLowerDeclListForm`: `my $x = f()` lowers to a `Decl` with
      `ListAssign` false, `my ($x) = f()` and `my ($x,$y) = f()` to true — by the
      child-count predicate. Assert against the `Paren` flag too, so the two
      signals are shown to agree.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerDeclListForm$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerDeclKinds`: one case per declarator in the T1 census
      (my, sub, package, local, our, state, field, class, method) lowering to
      the right `Decl.Kind`. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerSubDeclaration`: `sub f { 1 }`, `sub f;`, `sub { 1 };` and
      `my $f = sub { 1 };` each lower to the shape §4.14 specifies, with
      `sub f :prototype($) { 1 }` producing a `Hole` for the attribute region
      rather than a wrong node. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerDeclAttrsAndRefAlias`: `my $x :shared` fills `Attrs`,
      `my \$x = $y` sets `RefAlias`. Run to verify failure, implement, run to
      green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The list-vs-scalar declaration form is derived from child count (`go test ./internal/lower/ -run "^TestLowerDeclListForm$" -count=1 -v | grep -q "^--- PASS: TestLowerDeclListForm"`)
- [ ] Every declarator in the T1 census lowers to a Decl.Kind (`go test ./internal/lower/ -run "^TestLowerDeclKinds$" -count=1 -v | grep -q "^--- PASS: TestLowerDeclKinds"`)
- [ ] Named, forward, anonymous and expression-position subs all lower (`go test ./internal/lower/ -run "^TestLowerSubDeclaration$" -count=1 -v | grep -q "^--- PASS: TestLowerSubDeclaration"`)
- [ ] Attributes and ref-alias declarations lower (`go test ./internal/lower/ -run "^TestLowerDeclAttrsAndRefAlias$" -count=1 -v | grep -q "^--- PASS: TestLowerDeclAttrsAndRefAlias"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c112 — Lower calls: five sites to one Call, and the filehandle slot from the Handle flag

- **id:** `01a0c112-a8bb-7d7f-b156-d4ad99474e87`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocked by:** `01a0c10e` — Re-measure §4.14.2 against HEAD and rewrite the status table
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/lower/call.go` (new), `internal/lower/call_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Calls"; §4.8.1, §4.8.3

This slice retires a measured consumer defect and collapses five duplicated
consumer sites into one.

**The five sites.** §4.14.2 records that five places in `internal/infer` do the
same `GetBuiltin` + arity + argument-type check and differ only in where the
name hides: `infer.go:693` (`function_call_expression`,
`ambiguous_function_call_expression`), `:881` (`func1op_call_expression`),
`:1009` (`func0op_call_expression`), `:4004` and `:4048` (`function` or
`bareword` child). All five lower to one `Call{Name, Args, Resolved}`. That
duplication is the proof the erasure is VARIATION and not structure (§4.14.5).

**`Call.Handle` is no longer fragile.** §4.14.2 marks it "LOWERING by child
shape, fragile; a bit at `call.go:167` is the right fix". **That bit SHIPPED**:
`parse.Node.Handle` exists (`parse.go:299`) and is set at `call.go:143`.
Verified at HEAD: `print $fh "x"` gives `call "print"` over `term "$fh" handle`
and `term "\"x\""`. So read the flag; do not infer the slot from child count.

The defect this removes is measured and documented in the consumer itself:
`infer.go:1155-1165` records that counting the handle as argument 1 made every
typed-handle `print` a false `Str` mismatch, and it works around that by
skipping `indirect_object`. With `Handle` on the lowered node, `Call.Handle` is
a separate field from `Call.Args` and the workaround has no successor.

`print {$fh} "x"` is STILL `Unknown` at HEAD — `call.go:177` has the branch but
the lexer is not marking that `{` as `OpensBlock`. Emit a `Hole`; do not guess.

Remaining rows:

- `Call.Parenthesized` LOWERING — `f(1,2)` gives a `call "f"` whose span
  includes the `(`; `map $_+1, @a` does not. §4.8.1: this affects the paren
  cliff and prototypes.
- `Call.Ampersand` / `.BareAmpersand` LOWERING — `&foo;` gives `term "&foo"`;
  `&foo();` gives `[index "(" [term "&foo"]]`. Bare iff there is no enclosing
  `index "("`. §4.8.3: `&foo` passes the caller's `@_`, which is why the bare
  form is STRUCTURE and not a spelling.
- `Call.Fn` (call through a reference) LOWERING for two shapes, verified at
  HEAD: `$code->(1)` gives `[binary "->" [term "$code"] [term "1" paren]]` and
  `&{$code}(1)` gives `[index "(" [index "{" [term "&"] [term "$code"]] ...]`.
  `&$code(1);` is still `Unknown` — emit a `Hole`.
- `MethodCall` LOWERING, three shapes verified at HEAD: `$obj->method(1)` gives
  `[binary "->" [term "$obj"] [call "method" [term "1"]]]`; `Foo->new` gives
  `[binary "->" [call "Foo"] [call "new"]]`; `$obj->$name(1)` gives
  `[binary "->" [term "$obj"] [index "(" [term "$name"] [term "1"]]]`.
  `HasParens` comes from the `call` span. `Indirect`/`Ambiguous` stay PARSER —
  `new Foo(1)` is still `Unknown` at HEAD. Whether those two fields exist at
  all was settled by the vocabulary issue per §4.14.6; honour that decision.

`Call{Resolved}` is SHIPPED and is §4.14.1 rule 3's honest-uncertainty channel:
`Resolved:false` means "a call to something I have not seen", and PSC widens to
`Any` at exactly those points rather than guessing. The four non-call overloads
of that flag are handled by the literals/barewords issue.

## Steps

- [ ] Write `TestLowerCallHandleFromFlag`: `print $fh "x"`, `print STDERR "x"`
      and `say $fh "x"` lower to a `Call` with `Handle` set and `Args` holding
      only the real arguments; `print $fh, "x"` lowers with no `Handle` and two
      args; `print {$fh} "x"` lowers to a `Hole`.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerCallFlags`: `f(1,2)` sets `Parenthesized`, `map $_+1, @a`
      does not; `&foo;` sets both `Ampersand` and `BareAmpersand`; `&foo();`
      sets `Ampersand` only. Run to verify failure, implement, run to green,
      commit.
- [ ] Write `TestLowerCallThroughReference`: `$code->(1)` and `&{$code}(1)`
      lower to `Call` with `Fn` set and `Name` empty; `&$code(1)` lowers to a
      `Hole`. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerMethodCall`: the three shapes above lower to `MethodCall`
      with the right `Invocant`, `Name`/`Dynamic` and `HasParens`; `new Foo(1)`
      lowers to a `Hole`. Run to verify failure, implement, run to green,
      commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The filehandle slot comes from the Handle flag and is not an argument (`go test ./internal/lower/ -run "^TestLowerCallHandleFromFlag$" -count=1 -v | grep -q "^--- PASS: TestLowerCallHandleFromFlag"`)
- [ ] Parenthesized, Ampersand and BareAmpersand are derived correctly (`go test ./internal/lower/ -run "^TestLowerCallFlags$" -count=1 -v | grep -q "^--- PASS: TestLowerCallFlags"`)
- [ ] Calls through a code reference lower with Fn set (`go test ./internal/lower/ -run "^TestLowerCallThroughReference$" -count=1 -v | grep -q "^--- PASS: TestLowerCallThroughReference"`)
- [ ] All three method-call shapes lower to one MethodCall kind (`go test ./internal/lower/ -run "^TestLowerMethodCall$" -count=1 -v | grep -q "^--- PASS: TestLowerMethodCall"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c113 — Lower special forms: MapGrepSort.Op, the eval boundary, and the renames

- **id:** `01a0c113-c3b2-75d8-ae0e-4f146d610c49`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocked by:** `01a0c10e` — Re-measure §4.14.2 against HEAD and rewrite the status table
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/lower/special.go` (new), `internal/lower/special_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Special forms"; §4.14.1 rule 4

This slice retires the third and worst of the measured consumer defects.

**`callElementType` (`infer.go:1313`) tells map from grep by
`strings.HasPrefix(strings.TrimSpace(call.Text(source)), "grep")`** — two
operations with opposite element-type rules, separated by a string match on
source. `MapGrepSort.Op` fixes exactly that: one field, three values, no text.

**§4.14.2's block-form row is STALE and this is the largest correction in the
milestone.** It says "block forms all `Unknown`", marks them PARSER and blocks
them on issue 01a0ac52 (the §4.9.2 brace classification). Verified at HEAD, all
three block forms parse clean:

	map { $_ } @a;          call "map"  resolved=true, block child, term "@a"
	grep { $_ } @a;         call "grep" resolved=true, block child, term "@a"
	sort { $a <=> $b } @a;  call "sort" resolved=true, block child, term "@a"

So `MapGrepSort` is LOWERING for BOTH forms. Re-verify before implementing; if
it holds, §4.14.1 rule 4's `BlockGuessed` also needs revisiting — the flag exists
to record when the brace heuristic fired, and if the lexer's brace stack is now
making the call deterministically there may be nothing to record. Settle that
here by looking at how the block child is produced, and update rule 4.

Remaining rows, all verified at HEAD:

- `DoBlock` LOWERING (rename) — `do { 1 }` gives `[call "do" [block ...]]`.
  `blockResultType` (`infer.go:1417`) already computes a block's value for
  `map`, so the consumer exists.
- `DoFile` PARSER — `do 'file.pl';` is `Unknown`. Consumer SPECULATIVE. Hole.
- `EvalBlock`, `EvalStr` LOWERING — `eval { 1 }` gives `[call "eval" [block ...]]`;
  `eval "1"` and `eval $s` give `[call "eval" [term ...]]`. No consumer today,
  and §4.14.2 says these are needed anyway: **`EvalStr` is the soundness
  boundary where the result is `Any`.** Lowering them is what lets inference
  stop being wrong there rather than silently right by accident.
- `AnonList`, `AnonHash` SHIPPED — `anon_array`, `anon_hash`; `+{a=>1}` gives
  `[unary "pos" [anon_hash ...]]`. Statement-start `{a=>1};` is still a
  wrong-tree PARSER case (it parses as a `block`; perl's `B::Deparse` gives
  `+{'a', 1}`). Do not lower the wrong tree — emit a `Hole` for that case or
  lower the `block` faithfully as a block, whichever the executor can justify.
  The naming (`AnonList`/`AnonHash` vs `ArrayLiteral`/`HashLiteral`) was settled
  by the vocabulary issue per §4.14.6.
- `Ref` LOWERING — `\@a` gives `[unary "ref" [term "@a"]]`; `\($a,$b)` gives
  `[unary "ref" [list ...]]`.
- `FileTest` LOWERING (rename) — `-e $f` gives `[call "-e" resolved=true [term "$f"]]`.
  §4.14.2 marks the distinct kind SPECULATIVE: the old grammar folds `-e` into
  `func1op_call_expression`, the builtin table returns `Bool`, and that works.
  Lower it as a rename because it is free; write no file-test-specific inference
  rule. `-r _` takes a `Bareword` operand per §4.14.1 rule 3, owned by the
  literals issue. A nil operand means `$_`.
- `Readline{Magic}` LOWERING for `<STDIN>`, `<$fh>`, `<>` — each one `term`.
  `<<>>` is PARSER (the lexer does not know the 5.22 form). Note this kind's
  consumer value: `infer.go:78-84` uses `my $line = <$T>` as the canonical
  irreducible `Unknown`, and the kind is what lets it NAME that rather than
  merely produce it.
- `Glob` PARSER — `<*.c>;` is `Unknown`; `glob("*.c")` is a `call "glob"`.
  Consumer SPECULATIVE.
- `LoopEx` SHIPPED — `last FOO` gives `[loop_control "last" [label "FOO"]]`;
  `return` is also `loop_control` (`return (1,2)` gives a `list` child).
  `goto &foo;` is PARSER. `internal/infer` reads `return_expression` (3), and
  `blockAlwaysExits` (`infer.go:3312-3342`) detects `return`/`die`/`exit` by
  TEXT and will need `last`/`next` for correct branch joins — `LoopEx.Op` is
  what makes that a field read.
- `Wantarray` LOWERING (rename) — `wantarray;` gives
  `call "wantarray" resolved=true`. This is the one exact name agreement with
  B::SoN. `NarrowByContext` (`infer.go:1291`) exists and has no context source;
  this kind is the first half of giving it one.
- `Match{Op}` SHIPPED as an opaque leaf — `s/a/b/g`, `m/x/`, `qr/x/`, `/a/`,
  `tr/a/b/`, `y/a/b/` each one `term` with the raw operator as `Text`.
  `Match{Target,Negated}` LOWERING — `$x =~ /a/` is `binary "=~"`, `$x !~ /a/`
  is `binary "!~"`; a nil `Target` means `$_`.
  `Match{Pattern,Replace,Flags}` PARSER — inside the one `Quote` token by
  design (`internal/lexer/quote.go`). Do not split them here. B::SoN's
  `Transliterate.pm` records why the `tr` case is not a regex: reading `tr`'s
  operands as a pattern compiles something the source never wrote.
  `nearestPatternBefore` (`infer.go:1724-1751`) resolves `$1` by "the last match
  before this point", so **the side table must preserve total source ORDER** —
  that is the position-map issue's problem, noted here because this is the kind
  that depends on it.

## Steps

- [ ] Write `TestLowerMapGrepSortOp`: block and expression forms of all three
      lower to `MapGrepSort` with the right `Op`, `Block`/`FirstExpr` and
      `List` — with no source-text inspection anywhere in the path.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerEvalBoundary`: `eval { 1 }` lowers to `EvalBlock` and
      `eval "1"` / `eval $s` to `EvalStr` — the soundness boundary is a distinct
      kind. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerSpecialRenames`: `do { 1 }`, `wantarray`, `-e $f`, `\@a`,
      `<STDIN>`, `<>`, `[1,2]`, `{a=>1}` in expression position, `last FOO`,
      `next`, `return (1,2)` each lower to their §4.14 kind. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLowerMatchTargetAndNegated`: `$x =~ /a/` sets `Target`,
      `$x !~ /a/` sets `Negated` too, and a bare `/a/` leaves `Target` nil
      (meaning `$_`). Run to verify failure, implement, run to green, commit.
- [ ] Update §4.14.2's MapGrepSort row and §4.14.1 rule 4 to what HEAD measures.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] map, grep and sort are separated by a field, retiring the HasPrefix test (`go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v | grep -q "^--- PASS: TestLowerMapGrepSortOp"`)
- [ ] String eval is a distinct kind from block eval (`go test ./internal/lower/ -run "^TestLowerEvalBoundary$" -count=1 -v | grep -q "^--- PASS: TestLowerEvalBoundary"`)
- [ ] The special-form renames all produce their §4.14 kind (`go test ./internal/lower/ -run "^TestLowerSpecialRenames$" -count=1 -v | grep -q "^--- PASS: TestLowerSpecialRenames"`)
- [ ] Match carries Target and Negated, with nil meaning the topic variable (`go test ./internal/lower/ -run "^TestLowerMatchTargetAndNegated$" -count=1 -v | grep -q "^--- PASS: TestLowerMatchTargetAndNegated"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c113 — The position map: a side table keyed by id, with erased nodes mapping forward

- **id:** `01a0c113-dfd3-76a4-8b61-ae61b324bcb4`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10e` — Node identity: stable ids for internal/lower that survive an edit above them
- **blocked by:** `01a0c10f` — The §4.14 vocabulary types and Hole in internal/lower/node.go
- **blocks:** `01a0c115` — Cutover part 2: flip the LSP, routing position queries through the position map

## Context

- paths: `internal/lower/positions.go` (new), `internal/lower/positions_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestPositionMapErasedNodeMapsForward$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/06-incremental-lsp.md` §6.1.6 items 1 and 3

§6.1.6 item 1 is the boundary this issue defends: **`internal/parse/` stays the
only tree the parser produces and the only tree the LSP's structural features
read.** Positions for every byte, trees for broken code, round-trip, folding,
rename, format and any future extract-variable refactor read the CST. Sorbet
marks exactly that boundary — its `preserveConcreteSyntax` mode
(`ast/desugar/Desugar.cc:32`) exists for one caller, extract-variable
(`main/lsp/ExtractVariable.cc:314`) — because source-rewriting features must
see what the user typed. Type queries do not.

§6.1.6 item 3 specifies the connection: **a side table in both directions, and
an erased CST node maps FORWARD to its survivor.** rust-analyzer's
`expr_map`/`expr_map_back` (`expr_store.rs:168-186`) and Roslyn's
`BoundNode.Syntax` (`BoundNode.cs:24`) are the two shapes; the side table is
chosen here so **the lowered tree carries no syntax pointers**. The map is
many-CST-to-one-lowered, which is what trivia needs too (43,250 T1 nodes,
15.9%, all erased).

The canonical case is `lower.rs:1630` verbatim: the paren's CST node maps to the
inner expression's lowered node, so **hover on the paren resolves a type**. Note
this is independent of how §4.14.6's Paren node-or-flag question was settled by
the vocabulary issue: either way, the paren's BYTES must resolve.

**Total source ORDER is a hard requirement, not a nicety.** §4.14's preamble
states it and names the consumer: `nearestPatternBefore` (`infer.go:1724-1751`)
resolves `$1` by "the last match before this point". A table that can answer
"which node is at this offset" but not "which match precedes this offset" does
not meet the contract. Spans live in this table, keyed by node id — NOT on the
lowered nodes (§6.1.6 item 4: "never stored on CST nodes", and the lowered tree
is a value type for the same reason).

**Do not build incrementality here.** §6.4.3 is explicit: re-run `Analyze` on
the whole tree after each parse. Measured, `Analyze` is 0.5-2 ms against a parse
of 6-500 ms on the same files, so it is nowhere near the tightest constraint. If
it stops fitting the fix is to make `Analyze` incremental per-sub — a much later
project.

## Steps

- [ ] Write `TestPositionMapErasedNodeMapsForward`: lower `($x);` and assert
      that an offset inside the parens resolves to the lowered node for `$x` —
      the `lower.rs:1630` case. Add a trivia case: an offset inside a comment
      between two statements resolves to the enclosing lowered node rather than
      nothing.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestPositionMapErasedNodeMapsForward$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestPositionMapPreservesSourceOrder`: lower a file with two matches
      and a `$1` between them, and assert the table answers "the last match
      before this offset" with the first match, not the second — the
      `nearestPatternBefore` contract.
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Write `TestPositionMapInsideHole`: an offset inside a `Hole`'s span
      resolves to the `Hole`, so the LSP can report "no information" rather than
      a type (§4.14.2, Hole row). Run to verify failure, implement, run to
      green, commit.
- [ ] Write `TestLoweredTreeHasNoSyntaxPointers`: a structural assertion that no
      lowered node type holds a `*parse.Node` field (§6.1.6 item 3). Reflection
      over the declared types is enough. Run to verify failure, implement, run
      to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] An erased paren's bytes resolve forward to the surviving lowered node (`go test ./internal/lower/ -run "^TestPositionMapErasedNodeMapsForward$" -count=1 -v | grep -q "^--- PASS: TestPositionMapErasedNodeMapsForward"`)
- [ ] The table answers last-match-before-offset in source order (`go test ./internal/lower/ -run "^TestPositionMapPreservesSourceOrder$" -count=1 -v | grep -q "^--- PASS: TestPositionMapPreservesSourceOrder"`)
- [ ] An offset inside a Hole resolves to the Hole (`go test ./internal/lower/ -run "^TestPositionMapInsideHole$" -count=1 -v | grep -q "^--- PASS: TestPositionMapInsideHole"`)
- [ ] No lowered node type carries a syntax pointer (`go test ./internal/lower/ -run "^TestLoweredTreeHasNoSyntaxPointers$" -count=1 -v | grep -q "^--- PASS: TestLoweredTreeHasNoSyntaxPointers"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### 01a0c115 — Swap the annotation key from StartByte to node identity

- **id:** `01a0c115-88d1-73f0-8d38-a4182644108b`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10e` — Node identity: stable ids for internal/lower that survive an edit above them
- **blocks:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

## Context

- paths: `internal/infer/infer.go` (lines 57-60, 184, and the annotation reads), `internal/infer/project.go:30`, `internal/psc/lsp.go`, `internal/infer/annotations_test.go` (new)
- commands: `go test ./internal/infer/ -run "^TestAnnotationsKeyedByIdentity$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/06-incremental-lsp.md` §6.4.3; §6.1.6 item 4

§6.4.3 names this defect precisely: PSC's annotation map is keyed by `StartByte`
(`infer.go:57-60`, `map[uint32]types.Type`). **Byte-keyed annotations are
invalidated wholesale by any edit, because every key after the edit shifts.**
rust-analyzer's contributor guide rejects this key by name and states the
invariant: "syntax tree is a value type ... Using the tree as a store for
semantic info is convenient in traditional compilers, but doesn't work nicely in
the IDE."

There is a second, sharper defect at `infer.go:184`, and it is a correctness bug
rather than a performance one. The write is:

	if typ != types.Unknown { annotations[node.StartByte()] = typ }

so a PARENT written after its CHILD at the same offset overwrites the child.
Measured on `my $x = 5; my @y = ($x); my $z = $y[0];`, byte 0 carries four
nodes and the `number` inside `(1..5)` reads back as its parent's type.

This issue swaps the key to the node identity from `internal/lower/id.go`. It is
parallel to the lowering-rule slices and joins them at the cutover.

**The second site is `internal/infer/project.go:30`** — the same byte key in the
cross-file cache. §6.4.3's exception applies there and must be preserved:
`ProjectIndex` results are cached per file path and **must survive a keystroke
in an unrelated buffer.** Invalidate a `ProjectIndex` entry only when THAT
file's content changes. Do not let the key swap quietly widen that invalidation.

**Explicitly out of scope, per §6.4.3: do NOT try to incrementally update the
annotation map.** Re-run `Analyze` on the whole tree after each parse.
Justification, measured: `Analyze` is a bottom-up walk over an in-memory tree
with no I/O, 0.5-2 ms for 40,000 nodes, against a parse of 6-500 ms on the same
files. It fits the 10 ms p95 budget with room. If it ever stops fitting, the fix
is to make `Analyze` incremental per-sub — a much later project — not to patch a
byte-keyed map. An implementation here that adds a `DamageSet` or any
invalidation machinery is out of scope.

The public signature of `Analyze`/`AnalyzeWithOptions` changes with the key
type, which is why the two `internal/psc` callers (`lsp.go:53`,
`check_command.go:94`) are named above: `lsp.go` also reads the map by byte for
`DefinitionAtByte` and hover, and that read goes through the position map from
the position-map issue rather than through the annotation key.

## Steps

- [ ] Write `TestAnnotationsKeyedByIdentity`: analyse
      `my $x = 5; my @y = ($x); my $z = $y[0];` and assert the map holds a
      distinct entry per node rather than collapsing the four nodes at byte 0 —
      the `infer.go:184` defect, asserted directly.
- [ ] Run to verify failure: `go test ./internal/infer/ -run "^TestAnnotationsKeyedByIdentity$" -count=1 -v`
- [ ] Swap the key type. Run to green. Commit.
- [ ] Write `TestAnnotationsSurviveEditAbove`: analyse a two-sub file, insert a
      statement at the top, re-analyse, and assert the annotations inside both
      subs are keyed the same as before — the edit did not shift them.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestProjectIndexInvalidatesPerFile`: a `ProjectIndex` entry for
      file A survives a change to file B and is dropped by a change to A —
      §6.4.3's exception, preserved across the key swap.
      Run to verify failure, implement, run to green, commit.
- [ ] Run the infer suite: `go test ./internal/infer/ -count=1`
- [ ] Run the psc suite, since the callers see the signature change:
      `go test ./internal/psc/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Nodes sharing a start byte get distinct annotations (`go test ./internal/infer/ -run "^TestAnnotationsKeyedByIdentity$" -count=1 -v | grep -q "^--- PASS: TestAnnotationsKeyedByIdentity"`)
- [ ] Annotation keys survive an edit above the annotated nodes (`go test ./internal/infer/ -run "^TestAnnotationsSurviveEditAbove$" -count=1 -v | grep -q "^--- PASS: TestAnnotationsSurviveEditAbove"`)
- [ ] ProjectIndex invalidation stays per-file (`go test ./internal/infer/ -run "^TestProjectIndexInvalidatesPerFile$" -count=1 -v | grep -q "^--- PASS: TestProjectIndexInvalidatesPerFile"`)
- [ ] The infer package is green (`go test ./internal/infer/ -count=1`)
- [ ] The psc package still builds and passes against the new signature (`go test ./internal/psc/ -count=1`)

### 01a0c115 — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker

- **id:** `01a0c115-a44b-7393-a828-ad29844b20b4`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c10f` — Lower literals, names and Hole: NumLit, StrLit, Interp, QwList, Bareword
- **blocked by:** `01a0c110` — Lower variables and subscripts: Var, Deref, Index, Key with Arrow
- **blocked by:** `01a0c111` — Lower Slice{Kind}: the five-way switch §4.14.1 rule 5 specifies
- **blocked by:** `01a0c111` — Lower operators: Binary.Op from the field, Repeat and Assign from the Paren flag
- **blocked by:** `01a0c112` — Lower declarations: Decl kinds, the list form by child count, sub declarations
- **blocked by:** `01a0c112` — Lower calls: five sites to one Call, and the filehandle slot from the Handle flag
- **blocked by:** `01a0c113` — Lower special forms: MapGrepSort.Op, the eval boundary, and the renames
- **blocked by:** `01a0c115` — Swap the annotation key from StartByte to node identity
- **blocks:** `01a0c115` — Cutover part 2: flip the LSP, routing position queries through the position map

## Context

- paths: `internal/infer/lowered.go` (new), `internal/infer/lowered_test.go` (new), `internal/psc/check_command.go:94`
- commands: `go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v`
- entrypoints: `internal/infer/infer.go:64` (`Analyze`), `internal/psc/check_command.go:83-97`

The strangler-fig step. `internal/infer` has three entry points — `Analyze`
(`infer.go:64`), `AnalyzeWithOptions` (`infer.go:121`) and `NewProjectIndex`
(`project.go:38`) — and exactly TWO non-test callers:

	internal/psc/lsp.go:53            annotations, diagnostics, symbols := infer.Analyze(tree, source, nil)
	internal/psc/check_command.go:94  _, diags, _ := infer.AnalyzeWithOptions(tree, source, nil, opts)

That is the whole boundary, which is what makes a staged cutover possible
instead of a big bang. This issue adds `infer.AnalyzeLowered` ALONGSIDE the
existing entry points and flips the BATCH caller only — `check_command.go`, the
smaller surface, with no incremental path, no hover, and no position queries to
get right at the same time. The LSP flip is a separate issue and comes second.

Nothing is deleted here. The old walkers keep working against
`internal/parser/` until the LSP flip lands and the retirement issue removes
them; a half-migrated `internal/infer` with both paths live is the intended
state for exactly one issue's duration.

**What `AnalyzeLowered` must reproduce.** The diagnostics `check_command`
already emits. The cutover is not the place to change what PSC reports — a
diagnostic that changes here is either a fix that belongs in its own issue or a
regression. The test for that is differential: run both paths over the same
corpus and compare.

**Where the vocabulary bites.** `Kind()` on the old tree returns tree-sitter
grammar names; the lowered tree has §4.14's kinds and none of those names. Every
rule ported here is REWRITTEN against the lowered vocabulary, not shimmed. §6.4.1
is explicit that an earlier draft concluded from the ten-method count that PSC
would not change — "not one of its 4,265 lines" — and that this was the wrong
count. Ten methods port; the vocabulary does not.

**What gets measurably simpler, and these are the things to check in review:**
- the five `GetBuiltin` sites (`infer.go:693`, `:881`, `:1009`, `:4004`,
  `:4048`) collapse to one `Call` arm
- `findOperatorText` (`infer.go:645-659`) has no successor — `Binary.Op` is a
  field
- `callElementType`'s `strings.HasPrefix(..., "grep")` (`infer.go:1313`) becomes
  a switch on `MapGrepSort.Op`
- the `infer.go:1155-1165` filehandle workaround has no successor —
  `Call.Handle` is a separate field from `Call.Args`
- the three `@_`/`$_[0]` text re-derivations (`infer.go:3659`, `:3753`, `:4193`)
  compare `Var{Sigil,Name}` fields
- the 8-of-9 `scalar`/`array`/`hash` switch sites that differ only by sigil
  string become one arm reading `Var.Sigil`

Scope discipline: **no dataflow IR** (§6.1.6). §6.1.7 owns flow and is unbuilt.
If a rule wants flow, it keeps whatever approximation it has today.

## Steps

- [ ] Write `TestCheckCommandOnLoweredTree`: run `checkFile` through the lowered
      path over a fixture exercising each lowered kind, and assert the
      diagnostics match what the current path emits for the same fixture.
- [ ] Run to verify failure: `go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v`
- [ ] Add `infer.AnalyzeLowered` in `internal/infer/lowered.go`, porting rules
      slice by slice against the lowered vocabulary. Run to green. Commit at
      each slice — literals, variables, operators, declarations, calls, special
      forms — rather than once at the end.
- [ ] Write `TestLoweredAndLegacyAgreeOnCorpus`: a differential test running
      both paths over the parse package's existing corpus fixtures and asserting
      the diagnostic sets are equal. Any divergence is triaged as fix-or-
      regression before the flip, not after.
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Flip `check_command.go:94` to `AnalyzeLowered`. Run to green. Commit.
- [ ] Run the full suite: `make test`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The batch checker produces its diagnostics from the lowered tree (`go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v | grep -q "^--- PASS: TestCheckCommandOnLoweredTree"`)
- [ ] Both paths agree over the corpus (`go test ./internal/psc/ -run "^TestLoweredAndLegacyAgreeOnCorpus$" -count=1 -v | grep -q "^--- PASS: TestLoweredAndLegacyAgreeOnCorpus"`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)
- [ ] The whole repository is green (`make test`)

### 01a0c115 — Cutover part 2: flip the LSP, routing position queries through the position map

- **id:** `01a0c115-c0db-7923-b276-984794affd39`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c115` — Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker
- **blocked by:** `01a0c113` — The position map: a side table keyed by id, with erased nodes mapping forward
- **blocks:** `01a0c115` — Retire the tree-sitter vocabulary from internal/infer

## Context

- paths: `internal/psc/lsp.go` (lines 22-23, 53, 109, 122, 196), `internal/psc/lsp_test.go`
- commands: `go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v`
- entrypoints: `internal/psc/lsp.go:53`

The second and last caller. `check_command.go` flipped first because it is a
batch path with no position queries; `lsp.go` is harder for exactly the reason
that made it second — it does not only call `Analyze`, it READS the annotation
map by byte offset:

	lsp.go:22-23   diagnostics []infer.Diagnostic; symbols *infer.SymbolTable
	lsp.go:53      annotations, diagnostics, symbols := infer.Analyze(tree, source, nil)
	lsp.go:109     Diagnostics(uri string) []infer.Diagnostic
	lsp.go:122     SymbolTable(uri string) *infer.SymbolTable
	lsp.go:196     DefinitionAtByte(uri string, offset uint32) (*infer.Symbol, bool)

`DefinitionAtByte` is the query that has to change shape: with annotations keyed
by node identity, a byte offset no longer indexes the map directly. **It goes
offset -> position map -> node id -> annotation**, which is what the position
map issue built, and it is why that issue's source-order and Hole requirements
are load-bearing here rather than theoretical.

Two behaviours must survive the flip and both have tests to write:

**Hover on an erased node still resolves.** §6.1.6 item 3's canonical case:
hover inside `($x)` resolves the type of `$x`, because the paren's CST node maps
forward to the inner lowered node (`lower.rs:1630`). Today it resolves because
the paren left no node behind and the byte key happened to land on `$x`. After
the flip it must resolve deliberately.

**Hover inside a Hole reports no information, not a wrong type.** §4.14.2's Hole
row: a position query inside a Hole's span returns the Hole, so the LSP reports
"no information" rather than a type. T1 has 4,944 Unknowns covering 29.1% of
corpus bytes, so this is the common case, not an edge case.

§6.1.6 item 1 stays true after this flip and is worth re-checking in review:
**the LSP's STRUCTURAL features still read the CST.** Folding, rename, format,
round-trip and any future extract-variable read `internal/parse`. Only type
queries go through the lowered tree. A flip that routes a structural feature
through the lowered tree has broken the boundary Sorbet's `preserveConcreteSyntax`
marks, and the fix is to route it back.

Do not build incrementality (§6.4.3): re-run analysis on the whole tree after
each parse. Measured, that is 0.5-2 ms against a 6-500 ms parse.

## Steps

- [ ] Write `TestLSPHoverThroughPositionMap`: a hover at an offset inside `($x)`
      resolves the type of `$x` via offset -> position map -> id -> annotation.
- [ ] Run to verify failure: `go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v`
- [ ] Route `DefinitionAtByte` and the hover path through the position map.
      Run to green. Commit.
- [ ] Write `TestLSPHoverInsideHoleReportsNothing`: a hover inside an `Unknown`
      span returns no type rather than the enclosing node's type.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLSPStructuralFeaturesStillReadCST`: the structural queries the
      server exposes resolve against `internal/parse` and not the lowered tree —
      §6.1.6 item 1's boundary, asserted rather than assumed.
      Run to verify failure, implement, run to green, commit.
- [ ] Flip `lsp.go:53` to `AnalyzeLowered`. Run to green. Commit.
- [ ] Run the full suite: `make test`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Hover on a parenthesised expression resolves through the position map (`go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v | grep -q "^--- PASS: TestLSPHoverThroughPositionMap"`)
- [ ] Hover inside a Hole reports no information rather than a wrong type (`go test ./internal/psc/ -run "^TestLSPHoverInsideHoleReportsNothing$" -count=1 -v | grep -q "^--- PASS: TestLSPHoverInsideHoleReportsNothing"`)
- [ ] Structural features still read the CST (`go test ./internal/psc/ -run "^TestLSPStructuralFeaturesStillReadCST$" -count=1 -v | grep -q "^--- PASS: TestLSPStructuralFeaturesStillReadCST"`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)
- [ ] The whole repository is green (`make test`)

### 01a0c115 — Retire the tree-sitter vocabulary from internal/infer

- **id:** `01a0c115-df67-7d69-8883-73d40b0ae15d`
- **state:** pending
- **urgency:** normal
- **created:** 2026-09-20
- **blocked by:** `01a0c115` — Cutover part 2: flip the LSP, routing position queries through the position map
- **blocks:** `01a0c0e7` — The lowering pass: internal/infer must read internal/parse, not tree-sitter

## Context

- paths: `internal/infer/infer.go`, `internal/infer/lowered.go`, `internal/parser/`, `internal/infer/retired_test.go` (new)
- commands: `go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v`
- entrypoints: `internal/infer/infer.go`; `internal/parser/parser.go`

This is the last 10-20% of the migration, filed as its own issue because that is
where the discipline usually fails: the old path stays live "just in case",
`internal/parser` keeps a consumer, and the milestone gets called done at 85%.

Both callers are flipped by the time this starts. What remains:

- the old walkers in `internal/infer/infer.go` that switch on tree-sitter
  grammar names, now with no caller
- the three text-re-derivation sites the lowering replaced:
  `findOperatorText` (`infer.go:645-659`), `callElementType`'s
  `strings.HasPrefix(..., "grep")` (`infer.go:1313`), and the `@_`/`$_[0]`
  re-derivations (`infer.go:3659`, `:3753`, `:4193`)
- the `infer.go:1155-1165` filehandle workaround, whose comment documents the
  false `Str` mismatch it exists to avoid — `Call.Handle` removes the cause
- the five duplicated `GetBuiltin` sites (`infer.go:693`, `:881`, `:1009`,
  `:4004`, `:4048`)
- whether `internal/parser/` still has a non-test consumer at all

**The deletion IS the acceptance criterion**, and it is asserted by test rather
than by inspection: a test that greps the non-test files of `internal/infer` for
tree-sitter grammar names and fails if any remain. That is the executable form
of "the vocabulary migrated", and it is the only one that cannot be satisfied by
a comment.

Use the number the status-table issue established by its stated method (grammar
names in a `case` arm of a `Kind()` switch) as the starting inventory, so this
test and that count measure the same thing.

**`internal/parser/` may still have legitimate consumers** — it is a thin façade
over `gotreesitter` and the LSP's structural features were never its only
possible user. Do not delete the package on the assumption that it is dead;
measure. If it has no non-test consumer, propose removing it and let perigrin
decide, since deleting a package is not a refactor.

If anything cannot be deleted, **do not leave it silently**: file a follow-up
issue naming what remains and why, and link it here. CLAUDE.md: "If you find
yourself saying 'let''s defer this to a future project', write down where it is
deferred to. Unlabeled deferrals become drift."

## Steps

- [ ] Write `TestNoTreeSitterVocabularyRemains`: read the non-test `.go` files of
      `internal/infer` and assert none contains a tree-sitter grammar name in a
      `case` arm. Seed the forbidden list from the status-table issue's count.
- [ ] Run to verify failure: `go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v`
- [ ] Delete the dead walkers. Run to green. Commit.
- [ ] Write `TestNoSourceTextRederivation`: assert `findOperatorText` and the
      `strings.HasPrefix(..., "grep")` test are gone, and that no remaining
      inference rule reaches for `source` to decide which operator or which
      operation it has. Run to verify failure, delete, run to green, commit.
- [ ] Measure whether `internal/parser/` has a non-test consumer. Record the
      answer in the issue; if none, propose removal and ask rather than delete.
- [ ] Fold `internal/infer/lowered.go` back into the package's normal layout if
      the split was only a migration scaffold, so no file is named for a
      transition that is over.
- [ ] Run the full suite: `make test`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] No tree-sitter grammar name survives in inference (`go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v | grep -q "^--- PASS: TestNoTreeSitterVocabularyRemains"`)
- [ ] No inference rule re-derives an operator or an operation from source text (`go test ./internal/infer/ -run "^TestNoSourceTextRederivation$" -count=1 -v | grep -q "^--- PASS: TestNoSourceTextRederivation"`)
- [ ] The infer package is green (`go test ./internal/infer/ -count=1`)
- [ ] The whole repository is green (`make test`)

## psc-type-alignment

### 019d1449 — Expand PSC inference to cover more Perl type semantics

- **id:** `019d1449-1880-729b-860b-ac29464f6cd7`
- **state:** pending
- **urgency:** normal
- **created:** 2026-03-22
- **blocked by:** `01a0c0e7` — The lowering pass: internal/infer must read internal/parse, not tree-sitter

**2026-09-20: blocked on 01a0c0e7, and the split recommendation below is
superseded.**

That recommendation was "undef propagation now against the tree-sitter tree,
parameter inference later against the lowered one", on the argument that
undef propagation is "a lattice question more than a tree question" and so
mostly survives the conversion.

perigrin's direction: **do no work against tree-sitter, it is being burned
down.** That is a stronger constraint than "most of it survives" -- surviving
partially is not a reason to write code against a tree that is being deleted,
and §6.1.6 counts why nothing survives cleanly anyway:

> Ten methods port; the vocabulary does not.

`internal/infer` switches on 34 tree-sitter grammar names that
`internal/parse` does not emit. So BOTH halves of this issue wait on the
lowering pass, not just the parameter half.

Also stale below: area 3's note to "check for double-ownership before
starting" against 01a0ad3f. That issue is DONE (f066f2ab, 2026-09-19).
Cross-file inference remains out of scope here regardless.

---

Expand PSC's type inference beyond arity and container-level mismatches to value-level checking, as the formal papers describe.

Filed 2026-03-22. Re-measured 2026-09-17 at 3c997a6b by building `./cmd/psc`
and running it: two of the six expansion areas have since shipped, one
acceptance criterion already passes, and two do not. What follows is the
current state, not the original list.

## Measured state of the six areas

**1. String-to-number coercion — SHIPPED.** The first acceptance criterion
below already passes:

    my $x = "hello";
    my $y = $x + 1;
    -> ac1.pl:2:9: warning: left operand of "+": expected Num, got Str
       [coercion-mismatch]

**2. Undef propagation — NOT IMPLEMENTED.** `my $x; my $y = $x + 1;` produces
no diagnostic at all. This is the real remaining work in this issue.

**3. Cross-file return type inference — UNMEASURED here.** Untested in this
pass because it needs a multi-file fixture, and because it now overlaps
01a0ad3f (resolve imports by parsing the module's own source), which owns
how a second file is reached at all. Check for double-ownership before
starting.

**4. User-defined function signatures from usage — NOT IMPLEMENTED.**
`sub add { my ($a,$b) = @_; return $a+$b } add("x", 2);` produces no
diagnostic; nothing infers that `$a` is used numerically.

**5. Context-sensitive inference — PARTLY SHIPPED.** `types.NarrowByContext`
exists and is called at `internal/infer/infer.go:524`. §4.14.2 records the
gap that remains: `NarrowByContext` has no *source* of context, because the
CST has no `Wantarray` node and no call-site context propagation. That half
is blocked on the lowering, not on this issue.

**6. Pattern-matching and type guards — LARGELY SHIPPED.** `defined`, `ref`
and `isa` guards are implemented (`infer.go:2065-2082`, `:2536-2538`).
Guard-based flow narrowing shipped in v1.0.0-rc52. What remains is CST
extraction for `ref($x) eq 'TYPE'` and `$x isa Foo` — the type-system half is
done and the walker's `extractGuardPattern` needs the two node-kind branches.
That is GitHub #384, effort:medium, and it should be its own issue rather
than a line here.

## What this issue is now

Areas 2 and 4: undef propagation, and parameter types inferred from usage.
Everything else has shipped, moved, or belongs elsewhere.

## The framing that changed underneath this issue

Filed before the lowering decision. Chapter 6 §6.1.6 (committed ef839f93)
establishes that inference does not walk the CST directly: a lowering pass
produces the tree it reads, with the vocabulary of §4.14. `internal/infer`
still consumes the OLD tree-sitter tree through `internal/parser/`, switching
on 57 grammar names.

This matters for scope. Work done here against the tree-sitter tree is work
that the conversion will have to redo. Two options, and the issue should pick
one rather than drift:

- **Do it now against the current tree.** Undef propagation is a lattice
  question more than a tree question, so most of it survives the conversion.
  Cheapest if the conversion is far off.
- **Wait for the lowered tree.** §4.14.2 records that `internal/infer`
  re-derives structure from source TEXT in several places -- `findOperatorText`
  (`infer.go:645-659`), and `callElementType` (`:1313`) telling map from grep
  by `strings.HasPrefix` on source. Those are exactly the places new inference
  rules would have to reproduce.

Recommend the first for undef propagation and the second for parameter
inference, which needs call-site argument shapes the lowered `Call` node
provides and the grammar names do not.

## Acceptance

- [ ] A non-numeric string literal in numeric context diagnoses (`go test ./internal/infer/ -run '^TestCoercionMismatchOnStringInNumericContext$' -count=1 -v | grep -q '^--- PASS: TestCoercionMismatchOnStringInNumericContext'`)
- [ ] `my $x; my $y = $x + 1;` diagnoses undef in a numeric context (`go test ./internal/infer/ -run '^TestUndefInNumericContext$' -count=1 -v | grep -q '^--- PASS: TestUndefInNumericContext'`)
- [ ] `my $x; if (defined $x) { my $y = $x + 1 }` does NOT diagnose — the guard narrows it (`go test ./internal/infer/ -run '^TestDefinedGuardSuppressesUndefDiagnostic$' -count=1 -v | grep -q '^--- PASS: TestDefinedGuardSuppressesUndefDiagnostic'`)
- [ ] Undef propagates through assignment: `my $x; my $y = $x; my $z = $y + 1;` diagnoses at the addition (`go test ./internal/infer/ -run '^TestUndefPropagatesThroughAssignment$' -count=1 -v | grep -q '^--- PASS: TestUndefPropagatesThroughAssignment'`)
- [ ] A parameter used numerically in its body is inferred Num, and a string argument at a call site diagnoses (`go test ./internal/infer/ -run '^TestParameterTypeInferredFromBody$' -count=1 -v | grep -q '^--- PASS: TestParameterTypeInferredFromBody'`)
- [ ] A parameter used in no typed position stays Unknown rather than being guessed (`go test ./internal/infer/ -run '^TestUnconstrainedParameterStaysUnknown$' -count=1 -v | grep -q '^--- PASS: TestUnconstrainedParameterStaysUnknown'`)
- [ ] No new diagnostic fires on the existing corpus fixtures (`go test ./internal/infer/ -count=1 2>&1 | grep -qv FAIL`)
- [ ] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The third and sixth criteria are the negative scenarios: an undef check with
no guard awareness would fire on every `my $x;` in the corpus, and parameter
inference that guesses is worse than one that declines. §4.8.3's rule applies
to types as much as to parses -- an honest Unknown is never a wrong answer.

Not in scope: cross-file inference (area 3, overlaps 01a0ad3f), context
sources for `NarrowByContext` (area 5, needs the lowered tree), and
`ref($x) eq 'TYPE'` / `$x isa Foo` extraction (area 6, GitHub #384, its own
issue).

Key files: internal/infer/infer.go, internal/types/signatures.go
See also: GitHub #409, GitHub #384
