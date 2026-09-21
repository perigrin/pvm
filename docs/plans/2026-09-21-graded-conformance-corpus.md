# A graded conformance corpus for Perl 5 parsers

**Status:** design, not yet decomposed. Written 2026-09-21, revised the same
day after review by `crochet:assess`, a Ponytail pass, and the B::SoN
session. Every revision is marked.

## The problem

Neither existing corpus tells us where the parser is.

    T2 core      30.4%  (17 of 56)    perl's own t/, target 100%
    T1-easy      53.2%  (351 of 660)  PerlOnJava unit/, target 70%

Both are someone else's test suite, ordered alphabetically, measuring "how
much of this happens to parse" rather than "how far up the language have we
got". Neither number moves in a way that says what to build next. M1 reached
32/42 issues with both of those targets missed.

Three specific defects in T1-easy as a metric, all measured:

- **Alphabetical.** `anonymous_*`, `app_*`, `archive_zip_*`. No progression.
- **Module-contaminated.** 271 of 660 files are module integration tests
  (42.8% clean) against 389 language tests (60.4% clean). `Archive::Zip`'s
  API is not a fact about Perl syntax.
- **A moving denominator.** "Easy" means "contains none of twelve source
  markers". Adding a marker shrinks the population and raises the rate --
  the marker list's own comment warns that ">= 70% of T1-easy would be
  satisfiable by shrinking the denominator".

## The proposal

A corpus ordered by what a parser must understand before it can understand
the next thing. One construct per file, descriptively named, numerically
prefixed so the filesystem carries the order, every file validated against
perl.

**REVISED: this is an EVOLVING CORPUS, not a settled artifact.** Nothing
here is a specification of Perl 5 and nothing in it is final. The tier
ordering is one topological sort of a partial order and will be reordered;
the numbering is regenerable precisely so it can be. Files will move between
tiers as measurement contradicts placement -- the `t/` sweep exists to
produce exactly those contradictions, and a construct that turns out to
belong two tiers earlier is the sweep working rather than a mistake.

Three things follow, and they should be read into every commitment below:

- **No number in this document is a target.** 300-400 files, 14 tiers, the
  cadence estimates: all are current best measurements, and the ones that
  have already moved are marked.
- **The corpus grows for measured reasons.** A file is added because
  something failed and named a gap, not to reach a count.
- **Adopters should expect churn.** If this ever becomes a standalone
  artifact for other parser authors, it will be versioned and pinned like
  any other moving dependency, not published as a fixed conformance suite.

The reference point is Raku's ROAST, with two differences rather than one.

ROAST is normative: Raku is defined by it. Perl 5 has no specification but
its implementation, so this corpus is DERIVED from perl. Every file is
validated against `perl` before it counts, which is the rule PerlOnJava's
own AGENTS.md states:

> ALWAYS validate new unit tests with standard Perl before relying on them.

**CORRECTED: ROAST is not graded, and an earlier draft claimed it was.** Its
96 directories are `S<NN>-<topic>`, a Synopsis index rather than a dependency
order -- S01 is `perl-5-integration`, which nobody bootstraps first. It
prescribes no order, tracks progress through per-implementation skip lists,
and is a semantics suite of ~155k tests that nobody bootstraps an
implementation off incrementally. "Perl 5 lacks a ROAST" is true; "Raku has
a graded one" is not. The grading is the part that is genuinely new, and
therefore the part with no prior art to lean on.

## The ordering criterion

**A tier sits where it does because the next tier cannot proceed without it.**

Not difficulty, not conceptual centrality. Literals and variables are
self-contained -- `my $x = 1` needs no notion of context. The moment
operators arrive you cannot proceed without context: `@a + 1` is arithmetic
on a count and `"$x" . @a` is not, and the operator decides which. So context
sits immediately before operators, and no earlier.

**REVISED: the ordering is a partial order, not a chain.** `my $x = 1` at
tier 01 introduces `my` four tiers before `05_scoping`; `09_regex`
needs nothing from `08_references`; `11_packages` needs nothing from
`10_io`. Three tiers can move without breaking the dependency check, which
by check 2 below would mean they are misplaced. The honest statement is that
the numbering is ONE TOPOLOGICAL SORT of a partial order -- which still gives
the filesystem an order, and still supports the check.

It also resolves a tension worth recording. chromatic opens *Modern Perl*
with context, before any syntax, because a READER needs the frame
immediately. A PARSER needs it when operators arrive. Same fact, different
consumers.

## The tiers

    01_literals/       numbers, strings, quoting, qw
    02_variables/      sigils, scalars, arrays, hashes, element access,
                       delete/exists, $::
    03_context/        scalar vs list, the discriminating pairs
    04_operators/      arithmetic, string, comparison, logical,
                       precedence, associativity, the and/&& cliff
    05_scoping/        my, our, state, local, blocks
    06_control/        if/unless, while/until, for/foreach, postfix, goto
    07_subroutines/    declaration, call forms, @_, return, signatures
    08_references/     \, deref, arrow, anonymous constructors, ${ }, @{ }
    09_regex/          match, substitution, binding, delimiters
    10_io/             filehandles, readline, print
    11_packages/       package, use, require, imports
    12_oo/             bless AND class/field/method/ADJUST, indirect new
    13_opaque/         format, qx, <*>, heredocs, POD, __END__/__DATA__
    14_recursive/      s///e, (?{ }), qr// with embedded code

**REVISED: `13_base` is deleted.** It introduced no constructs -- it was
"the earlier tiers at real-world density", which is an exam, not a rung. It
also failed check 2 by construction: a tier that introduces nothing can move
anywhere.

**REVISED: two lexing tiers at the end.** 13 is where the lexer must
DELIMIT without lexing the contents; 14 is where it must RE-ENTER Perl
inside a delimiter, which is harder and structurally different. The plan's M0
text already names this set: "heredocs, POD, `__END__`, `q{}` with nested
delimiters, and `s///e` all to be *recognised* -- not understood, but
delimited correctly."

**REVISED: two digits, not four, and no gaps.** An earlier draft used
`0100_`/`0200_` to reserve 100 slots per tier for insertion. That is a hedge
against expensive renumbering, and renumbering is not expensive: the number
is DERIVED from the classification, so a reorder is a regeneration rather
than a hand-edit. Sparse numbering would also fossilise the first ordering,
which this document says explicitly will change.

Files inside a tier are numbered the same way -- `01_literals/03_leading_decimal.t`
-- and renumber with the tier. A file's identity is its NAME; the number is
its current position, and positions move.

Notes on placements that are not obvious:

- **03_context is where the inference gap becomes measurable.**
  `types.NarrowByContext` is called (`infer.go:524`) with NO source of
  context: the CST has no `Wantarray` node and no call-site propagation.
- **08_references after 07_subroutines**, following *Modern Perl*.
  Dereference syntax is harder than a sub declaration even though `\@a`
  alone looks simpler.
- **12_oo covers both object systems.** Measured 2026-09-21: `class`,
  `field`, `method`, `ADJUST`, `:param` and `:isa` each parse clean IN
  ISOLATION. **CORRECTED: that is true per construct and false per file.**
  `class Foo { ADJUST { 1 } method m { 2 } }` is one Unknown swallowing
  both, and `t2_test.go:105-111` records 41 Unknowns across seven
  `class/*.t` files with "nothing in the chain owns ADJUST". The tier's
  construct files would all pass over a live bug; its combination file is
  what catches it. See "Three kinds of file" below.

**The twelve `hardMarkers` all place**, which is the check that the tier list
covers the known-hard surface: heredoc/format/qx/glob-angle to 1300,
deref-brace and deref-at to 0800, signature to 0700, indirect-new to 1200,
goto to 0600, delete/exists/pkg-colon to 0200. No residue, so no misc tier.

## Where the files come from

**REVISED: everything is authored. Extraction does not survive its own
check.**

Measured: of the 35 T1 files whose names match tier 04, 31 `use Test::More`
and 30 call `ok(` at top level. Every extracted 0400 file would contain
tier-11 (`use`) and tier-07 (parenthesised call) constructs, failing
check 1 by construction. Either the harness idiom is exempted -- a hole in
exactly the check that makes the ordering a claim -- or everything is
authored.

The bottom tiers cannot be extracted at all. T1 files are
one-construct-SUBJECT, not one-construct-CONTENT:
`numeric_integral_nv_multiply.t` is about numeric literals and wraps every
assertion in `map`, `qw`, regex and `no warnings` -- six later tiers. There
is no tier-01 fragment inside it; there is a numeric topic in tier-09
syntax.

So **T1 is a source of CASES, cited by filename and pinned revision, never
copied.** That also collapses the licence question: authored files are T0 in
§7.3.6's terms, which that section already places in-repo, so §7.3.7's
"kept out of this repository" is not reversed and needs no amendment beyond
one sentence saying `conformance/` is project-authored.

**REVISED SIZE.** ~85-110 files for tiers 01-03 (the original estimate of
~40 was roughly half), and 300-400 authored files total -- twice §7.3.6's own
T0 estimate of ~200, which was the same idea.

**REVISED TWICE: the cadence was optimistic and the tooling was unsized.**
The "~40 validated files per day" figure had no citation and came from a
cadence measured on extraction-shaped work. An authored file under this
revision needs source, a behavioural probe built so a mis-grouped tree
produces different output, a perl run to pin the expected output, a
construct declaration, and for lexical tiers a token assertion. The T1
evidence says these are translations rather than lifts: 889 of 986 files
`use Test::More`, and only 2 of 986 contain none of
`use/require/sub/map/qw/=~`. Pessimistic cadence is 20-25 files/day.

The tooling does not exist at all: `conformance/`, the runner, tier
READMEs, the dependency lint, refusal codes on the ten Unknown sites,
ratchet integration. 3-5 days.

So the smallest useful corpus is **2-3 weeks**, not five days, and the full
300-400 is longer. That is the number to plan against.

## What a corpus file contains

**REVISED after the B::SoN session.** This is the part the first draft had
most wrong.

### Structural first, behavioural as the backstop

The first draft proposed validating a CST by emitting Perl from it, running
that, and comparing output to running the original -- modelled on B::SoN's
lowering gate. That characterisation was wrong, and their measurement says
so: **350 test files, 290 structural and 60 behavioural.** The round trip is
the backstop, not the primary check.

The history is the argument. They ran structural-only for months and shipped
six live wrong answers -- `sort bylen` sorting 4 items where perl sorts 3,
`@$r` losing mutations -- every one a structurally plausible graph, none
caught by a green suite, all six found by a human reading graphs and one by
accident. Their conclusion, which transfers:

> Behavioural comparison is what you add because CST inspection doesn't
> scale to a reviewer's attention. Keep both.

### The failure mode to design against

> **A shared assumption cancels out.** If the emitter reproduces a fold the
> walker made, the round-trip agrees with itself and the miscompile stays
> invisible.

Two rules from them, both of which apply to `canon`:

- The emitter is **deliberately dumb** -- it emits what the node says, never
  what the source probably meant.
- The emitter **must not import from the producer**. We currently violate
  this: `canon.go` reads the parser's own `infix[]` and `prefix[]` at ten
  sites, documented as a virtue. Their diagnosis is better than the one I
  reached: the coupling is CORRECT for round-tripping -- the emitter really
  must ask what the parser asked -- and the defect is that one wrong answer
  serves both askers.

Their recommended check needs no refactor and does not involve the emitter:
**31 adjacent precedence pairs, each a fixture whose two groupings produce
different observable output, run under perl.** That challenges the table from
outside. Shape verified:

    print 1 + 2 * 3;      7
    print((1 + 2) * 3);   9
    sub f{print "f";1} sub g{print "g";2} sub h{print "h";3}
    my $r = f() + g() * h(); print "=$r";      fgh=7

They also make separation structural rather than prose -- separate modules,
no common imports, and a test that fails if that changes -- on the grounds
that their own prose rule did not hold.

### Lexical facts are structural claims; no behavioural probe can see them

    perl -e 'printf "%.17g %.17g %.17g\n", .5, 0.5, 5e-1'
    0.5 0.5 0.5

`.5`, `0.5` and `5e-1` are the same value. A tier-01 fixture asserting
which SPELLING appeared cannot be checked behaviourally, and forcing the
fixture to print moves the vacuity rather than removing it. So:

- **REVISED: lexical facts are asserted on the TOKEN STREAM, not the CST.**
  (literal spelling, escapes, quoting style.) Three reasons, and the third
  is the one that decides it for a public artifact:

  test262 asserts no ASTs at all -- positive tests assert behaviour, and
  parse expectations are only `negative: {phase: parse, type: SyntaxError}`.
  That is *why* a dozen independent engines could adopt it: it never
  mandates node names. A CST assertion is the coupling that makes a corpus
  unusable by anyone whose tree differs from ours, and both existing Perl
  attempts (tree-sitter-perl's corpus, PPI's `.dump`) demonstrate it.

  The lexer is the better instrument anyway. It is M0-complete, round-trips
  620/620, imports nothing from tree-sitter, and NEVER FOLDS -- where the
  optree arrives post-peephole with `1+2` already `const[IV 3]`. Its token
  kinds are exactly the lexical surface the lower tiers are distinguished
  by: `Number` for `.5`, the heredoc introducer, `qw`, the sigils.

  And it decouples tiers 01-03 from the CST redesign
  (`2026-09-21-cst-is-the-source-of-truth.md`) entirely. A CST assertion
  written today would be rewritten if operators become leaves; a token fact
  is M0-stable.

  The expectation is still written as a DECLARED FACT rather than a token
  dump -- "this file contains exactly one numeric literal whose text is
  `.5`" is a claim any lexer can answer, where a dump of our token kinds is
  not. Token kinds are our vocabulary; less coupled than a CST, not
  uncoupled.

- **Semantic claims** (precedence, associativity, binding, context) get a
  behavioural probe, built so a mis-grouped tree produces DIFFERENT output
  rather than merely output. **Effect ORDER is the discriminator** -- values
  collapse, effect sequences do not.

### Three kinds of file, and two of them are new

**REVISED.** The tier list does not grow; these are file KINDS within a
tier, distinguished by a naming convention.

**Construct files.** One construct, the thing the tier introduces. What the
first draft described.

**Combination files, one closing each tier.** A file using tier N together
with tier N-1.

This exists because of a hole the first draft had no way to see. Measured
2026-09-21:

    class Foo { ADJUST { 1 } }                   0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

`ADJUST` alone parses. `ADJUST` followed by anything does not. It is an
ADJACENCY bug, and **no one-construct-per-file corpus can ever reach it**,
because every tier file is one construct by definition. A corpus of this
shape would go green over a known open bug -- and does: the T2 shortfall map
records "nothing in the chain owns ADJUST" at `t2_test.go:105-111`.

A combination file per tier also gives the tier-necessity check teeth it
currently lacks. Nothing today verifies that a tier's PREREQUISITE is
actually used; if tier N cannot combine with N-1, the ordering claim is
false.

It does not cover every adjacency -- N with N-3 is unreached -- which is
what the `t/` sweep is for. But it covers the adjacent pair the ordering
actually claims.

**Negative files, in-tier where they have a home.** `t/lib/croak/` ships 327
must-fail cases in a flat source + `EXPECT` format, already pinned by this
repo's own corpus checkout, already parser-agnostic. The first draft had NO
negative coverage at all, and a suite that only tests what should parse
cannot distinguish a real parser from one that accepts everything.

They place by construct, and they are better in-tier than gathered:

    Unterminated here-doc, qw//, q//      13_opaque
    Missing operator before "@foo"        04_operators
    foo found where operator expected     07_subroutines

A negative case is a BOUNDARY on the construct its tier introduces --
"here is `q//`, here is where it stops being `q//`". Gathered into a tier of
their own they would be disembodied error messages.

Two kinds have no home and go last, with the discovery-sweep findings:
cases where two failures interact (`... after strict error, w/fatal
warnings`) and `[perl #NNNNNN]` regression cases, which are historical
rather than structural.

Diagnostic text is NOT asserted -- only the must-fail bit. Perl's messages
are not a stable contract and asserting them would make the corpus track
perl's wording rather than its grammar. The vocabulary is Guacamole's
(`lib/Guacamole/Test.pm`): `parses` / `parsent`, which is the Perl-native
spelling of test262's model.

### Refusals carry a code, not a message

B::SoN measured their own codebase to answer this: **85 GAP raise sites, zero
structured codes, and 35 of 38 refusal assertions are bare `qr/GAP:/` with no
cause** -- the anti-pattern their own memory warns against, written before
the rule existed.

**CORRECTED: we are at TEN raise sites, not zero.** `Kind: Unknown` is
constructed in `parse.go` (4), `expr.go` (3), `term.go` (2) and `chain.go`
(1), and `Node` has no reason field. So this is a ten-site retrofit with a
"which cause is this" judgement at each, not greenfield -- smaller than
B::SoN's 85, but not free, and the earlier draft's "so we write codes rather
than retrofit" was wrong about the cost.
`Unknown` carries a reason code alongside its human message; tests assert on
the code, and the message is free to be rewritten. The deciding argument is
not brittleness but the census: their 5-facts/11-artifacts classification was
assembled by hand and two classifications contradicted the message's own
wording. With codes it is a group-by.

Their caution, recorded rather than discarded: codes drift too, they are a
second thing to sync, and their six substring assertions survived 30-odd
commits. And at our volume: **pick codes for the CAUSE, not the construct**,
or the census is the raise-site list with extra steps. The grouping is the
product.

**A refusal test is scaffolding.** It encodes a LIMITATION, not a fact, and
converts to a round trip the moment the construct works. At ~3,966 Unknowns
(3,317 over T1 per `t1.ratchet:6`, 649 over T2 per 01a0c10b; an earlier
draft said 4,900 with no source)
the age distribution of unconverted refusals is itself a measurement of where
the parser stopped growing.

**The measured behaviour goes in the comment** -- "perl prints 1 2 3 here, we
print nothing". Two of their three burns were not loose patterns; they were
tests passing where nobody could tell whether that was still meaningful.

### Serializing the CST is the destination

See `docs/plans/2026-09-21-cst-is-the-source-of-truth.md`. The round-trip
oracle is the safety net used BEFORE the CST is known to be right; a
serialization test written today would pass against a tree that had lost the
source, because the same walk produces both sides.

### The SoN IR is NOT a third layer, yet

Asked directly. Source→CST is near-lossless; source→SoN is deliberately
LOSSY, and the wire has no `Maybe[T]` -- `Scalar` currently spells
`Maybe[Str]`. A type-checker validated against that would bake in the
approximation. The wire also moves: `Loop.bound` was added the day before
this was written.

> Layers 1 and 2 are a conformance corpus. Layer 3 would be a regression
> corpus for our lowering decisions -- a different and more fragile artifact.

## The checks that make the ordering a claim rather than an opinion

1. **Dependency.** Each file's constructs were introduced at or before its
   tier.
2. **Tier necessity.** A tier that could move earlier without breaking check
   1 is misplaced. **REVISED:** this is a partial order, so the check is that
   a tier's PREREQUISITE is among the tiers before it, not that it could not
   move at all.
3. **perl validation.** Every file compiles under perl before it counts.

**REVISED: what supplies the construct set.** The parser cannot -- it cannot
classify a tier-09 file while it is at tier 0200. The oracle already
collects `Ops`, the `B::Concise` op list, so the check is
`ops(file) ⊆ ∪ ops(tiers ≤ N)`, parser-independent and already this project's
method.

The caveat is sharper than "necessary but not sufficient": **the optimiser
can erase the construct a tier is about.** `my $x = 1+2` arrives as
`const[IV 3] s/FOLD` with no `add` op, so a literals-tier file and an
arithmetic-tier file can produce identical op lists. Ops alone cannot derive
a tier. Either each file declares its construct set and the check verifies
nothing ELSE appears, or the CST supplies it once the CST is trusted -- which
is the same chicken-and-egg the serialization note resolves by ordering.

## What this replaces

`TestT1EasyParseRate` and its 70% target, and the T2 100% target. "Parses
cleanly through tier N" says something; "70% of an alphabetical pile that
includes Archive::Zip integration tests" does not.

**REVISED TWICE: `t/` is a discovery mechanism, not a finish line.** An
earlier revision made "all 620 files of perl's `t/` parse clean" the
completion criterion and then, two paragraphs later, called `t/` a discovery
mechanism. Both cannot hold, and the first is unreachable by construction:
`internal/parseoracle/testdata/corpus.pin` records the corpus at blead 5.45
against interpreter 5.42.0, and its own comment names `t/op/for-many.t:474`
as a genuine syntax error under 5.42 rather than a parser bug. The ratchet
baseline already carries `no-answer` and `excluded` statuses for exactly
this.

Carrying an unreachable 100% into a gate is the trap `01a0c13f-e0e2` was
filed to fix for T2 -- reproduced one paragraph after this document killed
the T2 gate. So the completion criterion is deleted.

**The corpus is the gate. `t/` is the sweep.** The loop is: build the corpus
from T1's cases, implement against the corpus tier by tier, then sweep `t/`
and let whatever still fails name constructs the corpus does not yet cover.
Those become new corpus files. A `t/` failure is a gap in the CORPUS first
and the parser second, which keeps the corpus honest and makes it grow for
measured reasons. The sweep's denominator is "files perl compiles", which
the harness already computes.

## Sequencing

**REVISED: the corpus comes first, and M1-M6 are rewritten around it.**

What survives that rewrite, measured rather than assumed:

- **`internal/lexer` and `internal/parse` survive.** They import tree-sitter
  NOWHERE -- verified, zero import lines, three comments only.
- **The oracle's perl side survives** and is now tree-sitter-free as of
  `e9ab4733`: adapter.go, compare.go and taxonomy.go deleted, the built-in
  subject fallback removed, `go list -deps | grep internal/parser` empty.
- **Corpus-independent metrics survive**: lossless round-trip, no-panic
  fuzz, Oracle WRONG=0, Oracle exact, canonical re-emission.
- **Corpus-tied gates die**: T1-easy >= 70%, T2 100%, M3's >= 90% of 584,
  M4's T1 >= 90%, M6's `t/op` >= 99%.
- **`hardMarkers` is scaffolding.** Mine it for tier placements, then delete
  `easy_test.go` whole. It encodes "constructs this parser failed on at
  4bc971ec", which in a graded corpus is a tier placement rather than a
  filter.

**The smallest useful corpus is tiers 01-04 plus a call-form slice from
0700: ~150-180 files, about five days.** 01-03 alone would be satisfied
on day one -- an uncommitted probe measured 50 of 53 such cases passing and
an independent one 13 of 16; neither is reproducible and both should be
treated as indicative only. The three failures in each were the same:
leading decimal `.5` (01a0c13f-97f5), heredoc (01a0c13f-aaf8), and the
ADJUST adjacency. So 0100 is one landed issue away rather than satisfied --
which is
gate on them says nothing. 0400 is the first tier where a parse can be WRONG
rather than Unknown, and the two largest T2 buckets are parenless call
argument position, which needs the 07 slice.

## Open questions

- Does each tier get a README stating what it introduces and why it sits
  there? Recommended yes.
- `conformance/` at repo root, given the standalone intent.
- Does 04_operators need internal grouping by precedence level? It is the
  largest tier by a distance.
- **REVISED:** does the construct set come from a per-file declaration or
  from the CST once trusted? Ops alone cannot supply it.
- Ordering sources: perl's own `perlintro`/`perlsyn` (installed, same
  licence), *Modern Perl* (CC BY-NC-SA) and *Programming Perl* (proprietary)
  as priors on teaching order. Reading them to decide what order to introduce
  concepts in is a fact about the language; lifting text is not.

## Next step

Prototype `01_literals` -- the tier where classification is unambiguous, so
it shakes out the method before the cases that need judgement. Then
`03_context`, where the format gets stressed.

Not decomposed into issues. This document is the input to
`crochet:refinement`.
