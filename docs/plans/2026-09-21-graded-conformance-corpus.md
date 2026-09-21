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

The reference point is Raku's ROAST -- a graded, executable,
construct-indexed suite -- with one difference. ROAST is normative: Raku is
defined by it. Perl 5 has no specification but its implementation, so this
corpus is DERIVED from perl. Every file is validated against `perl` before it
counts, which is the rule PerlOnJava's own AGENTS.md states:

> ALWAYS validate new unit tests with standard Perl before relying on them.

## The ordering criterion

**A tier sits where it does because the next tier cannot proceed without it.**

Not difficulty, not conceptual centrality. Literals and variables are
self-contained -- `my $x = 1` needs no notion of context. The moment
operators arrive you cannot proceed without context: `@a + 1` is arithmetic
on a count and `"$x" . @a` is not, and the operator decides which. So context
sits immediately before operators, and no earlier.

**REVISED: the ordering is a partial order, not a chain.** `my $x = 1` at
tier 0100 introduces `my` four tiers before `0500_scoping`; `0900_regex`
needs nothing from `0800_references`; `1100_packages` needs nothing from
`1000_io`. Three tiers can move without breaking the dependency check, which
by check 2 below would mean they are misplaced. The honest statement is that
the numbering is ONE TOPOLOGICAL SORT of a partial order -- which still gives
the filesystem an order, and still supports the check.

It also resolves a tension worth recording. chromatic opens *Modern Perl*
with context, before any syntax, because a READER needs the frame
immediately. A PARSER needs it when operators arrive. Same fact, different
consumers.

## The tiers

    0100_literals/       numbers, strings, quoting, qw
    0200_variables/      sigils, scalars, arrays, hashes, element access,
                         delete/exists, $::
    0300_context/        scalar vs list, the discriminating pairs
    0400_operators/      arithmetic, string, comparison, logical,
                         precedence, associativity, the and/&& cliff
    0500_scoping/        my, our, state, local, blocks
    0600_control/        if/unless, while/until, for/foreach, postfix, goto
    0700_subroutines/    declaration, call forms, @_, return, signatures
    0800_references/     \, deref, arrow, anonymous constructors, ${ }, @{ }
    0900_regex/          match, substitution, binding, delimiters
    1000_io/             filehandles, readline, print
    1100_packages/       package, use, require, imports
    1200_oo/             bless AND class/field/method/ADJUST, indirect new
    1300_opaque/         format, qx, <*>, heredocs, POD, __END__/__DATA__
    1400_recursive/      s///e, (?{ }), qr// with embedded code

**REVISED: `1300_base` is deleted.** It introduced no constructs -- it was
"the earlier tiers at real-world density", which is an exam, not a rung. It
also failed check 2 by construction: a tier that introduces nothing can move
anywhere.

**REVISED: two lexing tiers at the end.** 1300 is where the lexer must
DELIMIT without lexing the contents; 1400 is where it must RE-ENTER Perl
inside a delimiter, which is harder and structurally different. The plan's M0
text already names this set: "heredocs, POD, `__END__`, `q{}` with nested
delimiters, and `s///e` all to be *recognised* -- not understood, but
delimited correctly."

Numbering is regenerable -- the number is derived from the classification, so
reordering means regenerating, not hand-editing. No sparse gaps needed.

Notes on placements that are not obvious:

- **0300_context is where the inference gap becomes measurable.**
  `types.NarrowByContext` is called (`infer.go:524`) with NO source of
  context: the CST has no `Wantarray` node and no call-site propagation.
- **0800_references after 0700_subroutines**, following *Modern Perl*.
  Dereference syntax is harder than a sub declaration even though `\@a`
  alone looks simpler.
- **1200_oo covers both object systems.** Measured 2026-09-21: `class`,
  `field`, `method`, `ADJUST`, `:param` and `:isa` all parse clean today, so
  this tier's value is inference and edge cases rather than coverage.

**The twelve `hardMarkers` all place**, which is the check that the tier list
covers the known-hard surface: heredoc/format/qx/glob-angle to 1300,
deref-brace and deref-at to 0800, signature to 0700, indirect-new to 1200,
goto to 0600, delete/exists/pkg-colon to 0200. No residue, so no misc tier.

## Where the files come from

**REVISED: everything is authored. Extraction does not survive its own
check.**

Measured: of the 35 T1 files whose names match tier 0400, 31 `use Test::More`
and 30 call `ok(` at top level. Every extracted 0400 file would contain
tier-1100 (`use`) and tier-0700 (parenthesised call) constructs, failing
check 1 by construction. Either the harness idiom is exempted -- a hole in
exactly the check that makes the ordering a claim -- or everything is
authored.

The bottom tiers cannot be extracted at all. T1 files are
one-construct-SUBJECT, not one-construct-CONTENT:
`numeric_integral_nv_multiply.t` is about numeric literals and wraps every
assertion in `map`, `qw`, regex and `no warnings` -- six later tiers. There
is no tier-0100 fragment inside it; there is a numeric topic in tier-0900
syntax.

So **T1 is a source of CASES, cited by filename and pinned revision, never
copied.** That also collapses the licence question: authored files are T0 in
§7.3.6's terms, which that section already places in-repo, so §7.3.7's
"kept out of this repository" is not reversed and needs no amendment beyond
one sentence saying `conformance/` is project-authored.

**REVISED SIZE.** ~85-110 files for tiers 0100-0300 (the original estimate of
~40 was roughly half), and 300-400 authored files total -- twice §7.3.6's own
T0 estimate of ~200, which was the same idea. At a measured cadence of ~40
validated files per day that is 8-10 working days, plus tooling; about three
weeks before a full gate, or about five days if built in tier order.

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
  this: `canon.go` reads the parser's own `infix[]` and `prefix[]` at eight
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

`.5`, `0.5` and `5e-1` are the same value. A tier-0100 fixture asserting
which SPELLING appeared cannot be checked behaviourally, and forcing the
fixture to print moves the vacuity rather than removing it. So:

- **Lexical facts** (literal spelling, escapes, quoting style) are asserted
  on the CST. Behaviour is the wrong instrument.
- **Semantic claims** (precedence, associativity, binding, context) get a
  behavioural probe, built so a mis-grouped tree produces DIFFERENT output
  rather than merely output. **Effect ORDER is the discriminator** -- values
  collapse, effect sequences do not.

### Refusals carry a code, not a message

B::SoN measured their own codebase to answer this: **85 GAP raise sites, zero
structured codes, and 35 of 38 refusal assertions are bare `qr/GAP:/` with no
cause** -- the anti-pattern their own memory warns against, written before
the rule existed.

We are at zero raise sites, so we write codes rather than retrofit them.
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
converts to a round trip the moment the construct works. At ~4,900 Unknowns
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
classify a tier-0900 file while it is at tier 0200. The oracle already
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

**REVISED: the completion criterion is perl's own `t/`.** Not `t/base` --
that is nine files. All 620, which is the corpus the fidelity harness already
pins at `94e50866` and sweeps. It is external, unmovable, already
instrumented, and subsumes the tier list.

**REVISED: `t/` is a discovery mechanism, not a finish line.** The loop is:
build the corpus from T1's cases, implement against the corpus tier by tier,
then sweep `t/` and let whatever still fails name constructs the corpus does
not yet cover. Those become new corpus files. So `t/` failures are a gap in
the CORPUS first and the parser second, which keeps the corpus honest and
makes it grow for measured reasons.

Note this is a harder bar than the plan's M6, which asks for `t/op` >= 99%
and excludes `porting/` and `win32/`.

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

**The smallest useful corpus is tiers 0100-0400 plus a call-form slice from
0700: ~150-180 files, about five days.** 0100-0300 alone would be satisfied
on day one -- a probe measured 50 of 53 such cases already passing -- so a
gate on them says nothing. 0400 is the first tier where a parse can be WRONG
rather than Unknown, and the two largest T2 buckets are parenless call
argument position, which needs the 0700 slice.

## Open questions

- Does each tier get a README stating what it introduces and why it sits
  there? Recommended yes.
- `conformance/` at repo root, given the standalone intent.
- Does 0400_operators need internal grouping by precedence level? It is the
  largest tier by a distance.
- **REVISED:** does the construct set come from a per-file declaration or
  from the CST once trusted? Ops alone cannot supply it.
- Ordering sources: perl's own `perlintro`/`perlsyn` (installed, same
  licence), *Modern Perl* (CC BY-NC-SA) and *Programming Perl* (proprietary)
  as priors on teaching order. Reading them to decide what order to introduce
  concepts in is a fact about the language; lifting text is not.

## Next step

Prototype `0100_literals` -- the tier where classification is unambiguous, so
it shakes out the method before the cases that need judgement. Then
`0300_context`, where the format gets stressed.

Not decomposed into issues. This document is the input to
`crochet:refinement`.
