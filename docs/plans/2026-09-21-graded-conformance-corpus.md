# A graded conformance corpus for Perl 5 parsers

**Status:** design, not decomposed, not executed. Written 2026-09-21,
revised the same day after review by `crochet:assess`, a Ponytail pass, and
the B::SoN session. Every revision is marked.

**A SPIKE EXISTS AND IS NOT DELIVERED WORK.** Three files under
`conformance/01_literals/`, a glossary, and a runner under
`internal/conformance/` were written mid-design to test whether the proposed
format survives contact. They were produced OUTSIDE the assess/refine/execute
chain, so nothing in them is settled and nothing in this document should be
read as describing shipped behaviour.

What the spike is good for is evidence: the format section below proposes
specific semantics BECAUSE writing them exposed questions the sketch did not
raise, and those are marked as findings from the spike rather than as
decisions. Refinement decides them. The spike may be kept, rewritten, or
thrown away, and that is also refinement's call.

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

- **There is no file count and no schedule, and there should not be
  either.** Earlier revisions carried both: 300-400 files, 20-25 per day,
  2-3 weeks, 3-5 days of tooling. The count is removed because a total
  stated ahead of the work becomes a target, hit by writing files rather
  than by covering constructs. The durations are removed because they were
  human-development estimates and this corpus is written by agents, which
  makes a day-rate an estimate of the wrong quantity rather than an
  imprecise one. When a tier is done is answered by the dependency check
  and the `t/` sweep; what a file costs is stated below as the artifacts it
  needs, not as time.
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
needs nothing from `08_references`; `12_packages` needs nothing from
`10_io`. Three tiers can move without breaking the dependency check, which
by check 2 below would mean they are misplaced. The honest statement is that
the numbering is ONE TOPOLOGICAL SORT of a partial order -- which still gives
the filesystem an order, and still supports the check.

The oo/packages swap below is an instance of exactly this: two tiers whose
order the first draft got backwards because it sorted by how programs are
written rather than by what the grammar needs. Expect more of these from
the `t/` sweep.

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
    11_oo/             bless AND class/field/method/ADJUST, indirect new
    12_packages/       package, use, require, imports
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
- **REVISED: oo comes BEFORE packages, which reverses an earlier draft.**
  The old order had `11_packages` then `12_oo`, which made the OO tier
  depend on `package`/`use`. Perl says otherwise, measured under 5.42.0:

      class Foo { method m { 1 } }        syntax OK, no `package` anywhere
      bless {}, "Foo"                     needs a REF and a STRING, nothing else

  `class Foo {}` IS a package declaration -- a different spelling of tier
  12's own construct, not a consumer of it -- so the old order was
  circular. And `__PACKAGE__` stays `main` outside the class body, so
  `class` introduces its own scope rather than inheriting one.

  What `oo` actually needs is **08_references**: `bless` takes a reference,
  and every blessed-hash object is tier 08 plus a string. That dependency
  is real and is satisfied four tiers earlier.

  `use`/`require` move later because they are the heavier construct: they
  locate and load a FILE, then call `import`, which needs subroutines
  (tier 07) to mean anything. A tier list that puts them before the object
  system is ordering by how a program is usually WRITTEN rather than by
  what the grammar depends on.

- **oo covers both object systems.** Measured 2026-09-21: `class`,
  `field`, `method`, `ADJUST`, `:param` and `:isa` each parse clean IN
  ISOLATION. **CORRECTED: that is true per construct and false per file.**
  `class Foo { ADJUST { 1 } method m { 2 } }` is one Unknown swallowing
  both, and `t2_test.go:105-111` records 41 Unknowns across seven
  `class/*.t` files with "nothing in the chain owns ADJUST". The tier's
  construct files would all pass over a live bug; its adjacency file is
  what catches it. See "Three kinds of file" below.

**The twelve `hardMarkers` all place**, which is the check that the tier list
covers the known-hard surface: heredoc/format/qx/glob-angle to tier 13,
deref-brace and deref-at to tier 08, signature to tier 07, indirect-new to tier 11,
goto to tier 06, delete/exists/pkg-colon to tier 02. No residue, so no misc tier.

## Where the files come from

**REVISED: everything is authored. Extraction does not survive its own
check.**

Measured: of the 35 T1 files whose names match tier 04, 31 `use Test::More`
and 30 call `ok(` at top level. Every extracted tier 04 file would contain
tier-12 (`use`) and tier-07 (parenthesised call) constructs, failing
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

**REVISED, THEN REVISED AGAIN: the file counts are removed.** Earlier drafts
carried ~40 files for tiers 01-03, then ~85-110, then 300-400 total. Each
correction was real -- the first estimate was roughly half -- but a total
stated before the work is a target, and it would be hit by writing files
rather than by covering constructs. When a tier is done is answered by the
dependency check and the sweep.

**REVISED: the DURATIONS are removed too, and for a sharper reason than the
counts.** Earlier drafts carried "~40 validated files per day", then a
pessimistic "20-25 files/day", "2-3 weeks" for the smallest useful corpus,
and "3-5 days" of tooling. Every one of those is a human-development
estimate, inherited from reviewers reasoning about how long a person takes.
This corpus is written by agents, where the per-file cost is dominated by
things a day-rate does not model -- how many perl invocations a file needs,
whether the format is settled, whether the tooling exists yet. A
human-day-rate is not a slow estimate, it is an estimate of a different
quantity.

**What survives is the per-file WORK, stated as work rather than time.** An
authored file needs five things: source, a behavioural probe built so a
mis-grouped tree produces different output, a perl run pinning the expected
output, a construct declaration, and for lexical tiers a token assertion.
The original estimate assumed one thing, extraction, and the T1 evidence
says these are translations rather than lifts -- 889 of 986 files `use
Test::More`, and only 2 of 986 contain none of `use/require/sub/map/qw/=~`.
That ratio is the useful fact: it says the cost is five artifacts per file
and almost no file can be lifted, which is true whoever writes it.

**The negatives are not authored at all.** `t/lib/croak/`'s cases are
extracted by a runner -- see "The manifest alternative" below -- so they
carry none of that per-file cost. That is a structural difference in the
work, not a faster rate.

**The tooling is the real prerequisite, and it is a LIST rather than a
duration.** None of it is per-file and all of it blocks the first file. Its
usefulness is that it is enumerable and each item is separately checkable as
done -- which is what a bounded estimate was reaching for and failing to be.

    conformance/ and the file format   spiked, see the note below
    the runner                         spiked
    the glossary                       spiked, 5 entries
    tier READMEs
    the dependency lint                ops-subset, per "The checks"
    refusal codes on the ten Unknown sites
    ratchet integration                file-name-keyed, not path-keyed
    the croak extraction runner        with its `perl -c` split

**Nothing on that list is done.** The first three have a spike against them
-- code written mid-design to test the format, outside the chain -- which is
evidence about the shape of the work and not a delivery. The first three are
also the ones every file depends on, which is why they were the ones worth
spiking.

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

**MEASURED: could behaviour plus ops plus round-trip replace the token
layer?** The obvious objection to asserting on tokens is that the corpus
already has three checks, so the question is whether any of them reaches a
mis-lex. Tested against the two tier-01 cases:

    check                .5 vs 0.5           5e-1 lexed as `5e` `-` `1`
    behaviour            no, same to 17 sf   no, both print 0.5
    B::Concise ops       no, same optree     no, see below
    round-trip           NO -- round-trips   no, round-trips AND faithful
    Faithful             yes, caught it      no
    token assertion      yes                 yes

`my $x = .5` and `my $x = 0.5` produce byte-identical optrees:
`const[NV 0.5] s`. And the ops check cannot see the second case at all,
because **perl never builds the wrong tree** -- `5e - 1` is a syntax error,
so there is no optree to compare against. `B::Concise` tells us whether our
tree matches perl's for programs perl accepts; it cannot tell us we built a
DIFFERENT VALID tree from the same bytes.

Round-trip is blind for a structural reason rather than an accidental one:
it reassembles source bytes from the tree, and the bytes are all present
however they were grouped. `Number(5e) Operator(-) Number(1)` concatenates
back to `5e-1` exactly. It caught `.5` only because that mis-group happened
to strand a token; it is not a check on tokenisation and cannot be made
into one.

So the token layer is not redundant with the other three. It is the only
check that sees grouping, and grouping is what a lexer decides.

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

**Adjacency files, one closing each tier.** Two constructs side by side.

**REVISED: these were "tier N with tier N-1", and the renaming matters.**
Framed by tier, the file asserts something about OUR ordering, which a
review rightly objected is our implementation strategy rather than a fact
about Perl. But look at what the case actually is:

    class Foo { ADJUST { 1 } }                   parses
    class Foo { ADJUST { 1 } method m { 2 } }    does not

That is a fact about PERL. Every parser has an adjacency surface; ours is
where this bug lives. The file contents are identical either way -- only the
justification changes, and "two constructs adjacent" is portable where "tier
N with N-1" is not.

**REVISED AGAIN: "pair with the previous tier" does not reach the bug that
motivates this, and measurement proves it.** Probed against `parse.Parse`:

    class Foo { ADJUST { 1 } }                      0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }       1
    class Foo { ADJUST { 1 } field $x; }            1
    class Foo { ADJUST { 1 } ADJUST { 2 } }         1
    class Foo { method m { 2 } ADJUST { 1 } }       0
    class Foo { field $x; ADJUST { 1 } }            0
    class Foo { method m { 1 } method n { 2 } }     0
    package Bar; class Foo { ADJUST { 1 } }         0

The bug is precisely **`ADJUST` followed by any sibling**, including another
`ADJUST`. Every other pairing is clean.

`class`, `field`, `method` and `ADJUST` are ALL tier 11, so the failing
adjacency is within one tier. And the last line is what a previous-tier
pairing would have produced after the oo/packages swap -- `package` beside
`class` -- which parses at 0 and goes GREEN OVER THE LIVE BUG. The rule as
first written would have been satisfied by a file that catches nothing.

**So the rule is: every construct the tier introduces, in one body, each
adjacent to another.** For tier 11 that is one file:

    class Foo { field $x; ADJUST { 1 } method m { 2 } }

which covers "X followed by anything" for every X in the tier, and is
cheaper than enumerating pairs. Where a tier genuinely depends on an
earlier one, the adjacency file pairs with the DECLARED PREREQUISITE rather
than with N-1 -- which matters for `09_regex`, `12_packages` and the other
tiers this document already says have no N-1 dependency. Pairing those with
N-1 would assert nothing.

The claim is about the constructs, not the
numbering.

This exists because of a hole the first draft had no way to see. Measured
2026-09-21:

    class Foo { ADJUST { 1 } }                   0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

`ADJUST` alone parses. `ADJUST` followed by anything does not. It is an
ADJACENCY bug, and **no one-construct-per-file corpus can ever reach it**,
because every tier file is one construct by definition. A corpus of this
shape would go green over a known open bug -- and does: the T2 shortfall map
records "nothing in the chain owns ADJUST" at `t2_test.go:105-111`.

An adjacency file per tier also gives the tier-necessity check teeth it
currently lacks. Nothing today verifies that a tier's PREREQUISITE is
actually used; if tier N cannot combine with N-1, the ordering claim is
false.

It does not cover every adjacency -- N with N-3 is unreached -- which is
what the `t/` sweep is for. But it covers the adjacent pair the ordering
actually claims.

**Negative files, in-tier where they have a home.** `t/lib/croak/` ships
cases in a flat source + `EXPECT` format, already pinned by this repo's own
corpus checkout, already parser-agnostic. The first draft had NO negative
coverage at all, and a suite that only tests what should parse cannot
distinguish a real parser from one that accepts everything.

**REVISED: "327 must-fail cases" was wrong twice, and the second error would
have produced wrong tests.**

327 is the count of `########` SEPARATORS, not of cases; the first case in
each file has no leading separator. Cases are `EXPECT` blocks: **338**, of
which 251 carry a `# NAME`.

The worse error: **croak does not mean must-not-parse.** Measured over 340
extracted cases under perl 5.42.0, running `perl -c` and then `perl`:

    compile-time failure  201   59%   genuinely `parsent`
    runtime-only failure   64   19%   perl COMPILES these
    exits 0                75   22%

Whole files are runtime croaks: `pp_ctl` 0 of 20 compile failures, `pp_hot`
0 of 11, `pp_sys` 0 of 13, `pp` 0 of 4. "goto into foreach" and "pipe()
croaks on bad left side" are diagnostics from running, not from parsing. **A
parser that emitted `parsent` for those would be wrong**, and a runner that
took the must-fail bit from the file would assert exactly that on 19% of
cases.

The 75 that exit 0 are mostly `signatures` (52 of 62), which is the
blead-5.45 against interpreter-5.42.0 skew `corpus.pin` already documents --
a version fact, not a parser fact.

**So the runner splits by `perl -c` rather than trusting the file:**

- compile-time failure  -> a `parsent` case
- runtime-only failure  -> a `parses` POSITIVE, and the croak is not our
  concern
- exits 0               -> excluded, with the pin's version skew named

Case identity is (file, ordinal), not `# NAME`, since 87 cases have no name.
One case needs a `-switch` or `--FILE--` and is skipped explicitly.

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

### The file format

**REVISED: this was a sketch across four revisions, and a spike turned it
into a concrete PROPOSAL.** Writing three files and a runner surfaced
questions the sketch did not raise -- what a repeated section means, whether
output is byte-exact, what a refusal looks like in a suite that must stay
pristine. What follows answers those, and every answer is a proposal for
refinement rather than a decision already taken. Where the spike contradicted
the sketch, the change is marked.

    conformance/01_literals/03_leading_decimal.t

    #!perl
    # A leading decimal point is part of the numeric literal: `.5` is one
    # token, not the concatenation operator `.` followed by `5`.
    #
    # TIER 01 literals
    # INTRODUCES numeric literal
    # USES nothing from a later tier
    #
    # MEASURED perl 5.42.0:
    #
    #   $ perl -e 'my $x = .5; print "$x\n"'
    #   0.5
    #
    #   $ perl -e 'printf "%.17g %.17g %.17g\n", .5, 0.5, 5e-1'
    #   0.5 0.5 0.5
    #
    # The three spellings are the SAME VALUE to 17 significant digits, so no
    # behavioural probe can see which was written. See ../GLOSSARY.md.
    #
    # STATUS refuses as of 38c95d23. Issue 01a0c13f-97f5.

    --- source
    my $x = .5;
    print "$x\n";

    --- expect parses

    --- expect output
    0.5

    --- expect tokens
    one numeric literal whose text is ".5"
    no operator whose text is "."

**Sections.** A line beginning `--- ` opens a section and runs to the next
one. Everything before the first section is the comment block. Sections may
appear in any order; a repeated section is an error rather than a
last-one-wins.

    --- source          required. The Perl the case runs.
    --- expect parses   the must-parse bit
    --- expect parsent  the must-NOT-parse bit
    --- expect output   what perl prints, byte for byte
    --- expect tokens   declared lexical facts, one per line

Exactly one of `parses`/`parsent` is required: a file asserting neither says
nothing, and the loader rejects it rather than passing vacuously.

**CONTACT CHANGED ONE THING: the trailing newline.** `--- expect output`
must match perl's bytes exactly, but a blank line before the next section is
formatting rather than content. Exactly one trailing newline is stripped, so
a file wanting a trailing blank line writes two. This is the only place the
format is not literal, and it is the only such rule.

**The expectation is a DECLARED FACT, not a token dump.** "One numeric
literal whose text is `.5`" is answerable by any lexer. A dump of OUR token
kinds is answerable only by us, which is the coupling that makes a corpus
unusable by outsiders -- the same defect as asserting on a CST, one layer
down.

The grammar is deliberately two forms and no more:

    one <category> whose text is "<text>"
    no  <category> whose text is "<text>"

`<category>` is a GLOSSARY.md name. A fact naming an undefined category is
an error that says so, which is what stops the vocabulary growing silently
past its definitions. Counts other than one and zero are the extension
point, and there is no third form until a file needs one.

**The coupling should be one file.** A map from glossary names to this
lexer's kinds, and nothing else reading our token enum, so a project whose
lexer differs adopts the corpus by rewriting that map alone. The spike's
`internal/conformance/categories.go` is five entries and shows the shape is
achievable; whether it stays one file under a fuller corpus is what would
make the portability claim checkable rather than aspirational.

**A negative file is the same format with `--- expect parsent`** and no
output or token section. That is test262's `negative: {phase: parse}` and
the most portable thing here.

**An adjacency file is the same format** with the tier's constructs adjacent.

**The measured behaviour is in the comment**, per B::SoN: two of their three
burns were not loose assertions but tests passing where nobody could tell
whether that was still meaningful.

### Perl adjudicates before we do

The runner should check the FILE against perl before it checks the PARSER
against the file, and report a disagreement as `CORPUS BUG` rather than as a
refusal:

    file says `expect parses`, perl -c refuses it
    file says `expect parsent`, perl -c accepts it
    pinned output "0.5\n", perl prints "0.5000\n"

Without that order a wrong expectation reads as a parser failure, and the
corpus would accumulate cases that fail for reasons nobody checked. It is
the same compile-versus-run split the croak measurement forced: `perl -c`
answers the parse question, running answers the output question, and
conflating them is exactly the error that made 19% of croak cases look like
must-not-parse.

### A known refusal skips; a stale marker fails

**REVISED: the spike raised this and the sketch had no answer.** A corpus
whose purpose is to name what does not work yet cannot also be all-green,
and cannot be allowed to make `make test` noisy. The proposal below is what
the spike does; it is the part of the format most worth arguing with,
because it decides what the suite's green means.

    # STATUS refuses as of 38c95d23. Issue 01a0c13f-97f5.

A file carrying that line SKIPS, reporting its citation and what actually
happened -- the Unknown count, the token stream.

**The issue id is optional, and its absence is not a gap.** A refusal found
by writing the file cites `this file`, because the file already holds the
source, the measured perl behaviour and the tokens we produce instead. An
issue would be a second copy of that, free to go stale. `.5` cites an issue
only because it predates the corpus.

The reverse is an ERROR, loudly:

    file is marked `STATUS refuses` (01a0c13f-97f5) but now PASSES.
    Remove the STATUS line -- a stale marker hides a regression.

A marker that outlives its bug is worse than no marker, because it silences
a file that has started failing again for an unrelated reason. This is the
ratchet's rule at file granularity, and it is why a refusal is a skip rather
than a `t.Log`.

**One passing file per tier should not be optional.** In the spike that is
`02_decimal.t`, asserting `0.5`, which already works. A tier whose every
file refuses cannot demonstrate that passing is reachable, and cannot
distinguish "not implemented" from "the runner is broken". It is also the
baseline its neighbours deviate from --
`.5` and `5e-1` are the same construct with the integer part removed and an
exponent sign added -- so a regression there explains both rather than being
diagnosed twice.

### A refusal is the corpus working, not a bug report

Writing `03_leading_decimal.t` meant probing the lexer, which turned up a
second construct it does not handle:

    5e-1    ->  Number(5e) Operator(-) Number(1)
    5e+1    ->  Number(5e) Operator(+) Number(1)
    1.5e-3  ->  Number(1.5e) Operator(-) Number(3)
    5e1     ->  Number(5e1)

Signed exponents split; unsigned ones do not, which is why nothing had
noticed. `perl -e 'print 5e'` is a syntax error, so `Number("5e")` is a
token perl would reject.

**This is the corpus's normal mode and needs no ticket.** A construct the
parser does not reach yet is a file that refuses, and the file IS the
record: its source, the measured perl behaviour, and the exact token stream
we produce instead. Filing an issue as well would put the same information
in two places and make one of them stale. `.5` cites an issue only because
it was found before the corpus existed; `5e-1` was found by writing the
file, so it cites `this file`, which is a complete answer rather than a
gap.

The distinction the corpus DOES draw is between a refusal and a regression.
A refusal is marked and skips. A file that starts refusing without a marker
fails, and a marked file that starts passing fails. Both of those need
attention; neither of these two does.

**What the case does demonstrate is the token layer.** `5e` `-` `1` is a
well-formed subtraction: zero Unknowns, and the round trip agrees with
itself, so the parser cannot see it and neither can the oracle. Only the
token assertion catches it. The argument for lexical facts as a category was
previously made from the `printf "%.17g"` collapse alone -- a value
argument. This is the same claim against a construct where the TREE is also
blind, which is a stronger version of it.

### The glossary is a prerequisite

"One numeric literal" needs an outsider to agree what a numeric literal is,
and the boundaries are contestable: is `-1` one literal or a negation of
one? Is `.5e3` one token or two? Is `v5.42` a literal?

Moving assertions from the CST to the token stream shrank the contested
vocabulary from a whole tree grammar to a handful of lexical category names.
It did not eliminate it. **Without a glossary defining those names, every
adopter forks on the first ambiguous case**, and the corpus stops being
portable for the same reason a CST assertion would have.

**REVISED: the spike drafted one, and it is cheaper than it looked.** Five
categories, added when a file needed to assert about one rather than in
advance, which is the growth rule worth keeping whatever else changes. Every
boundary is DECIDED AGAINST PERL rather than asserted, which is this
project's standing rule applied to its own vocabulary:

    -1        TWO tokens. Negation, then a literal.
    5e-1      ONE token. The exponent sign is part of it.
    .5e3      ONE token. Leading point and exponent compose.
    1.        ONE token, and equals 1.
    v5.42     NOT a numeric literal -- a v-string, per perldata.

The `-1` entry is the one that earns the page. Measured:

    $ perl -MO=Concise -e 'my $x = -1;' | grep const
    const[IV -1] s/FOLD

The optree shows ONE folded constant, so it cannot distinguish `-1` from a
negation of `1` -- which is precisely why the token stream has to answer
this and why the asymmetry with `5e-1` needs writing down rather than
leaving to a reader's intuition.

**The glossary is not separable from the tier it serves.** A tier whose
assertions are token facts cannot be written before the vocabulary those
facts use is defined, or it asserts in terms nobody has agreed. So a tier
and the glossary entries its files need are one unit of work, and the
glossary grows an entry at a time rather than being written whole.

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

### The CST is NOT in the corpus, and the token layer is

Worth stating together, because they are usually asked as one question.

**No corpus file asserts CST shape, and none should.** That is the coupling
which makes a corpus unusable by anyone whose tree differs -- test262's
reason for asserting no ASTs, and the defect in both existing Perl attempts.
Serializing the CST is a check we run against OURSELVES
(`docs/plans/2026-09-21-cst-is-the-source-of-truth.md`), not a fact we ask
an adopter to reproduce. The round-trip oracle is the safety net used BEFORE
the CST is known to be right; a serialization test written today would pass
against a tree that had lost the source, because the same walk produces both
sides.

**The token layer is different in exactly the way that matters.** It is not
a tree, it is a grouping of bytes, and it is asserted as a declared fact in
the glossary's vocabulary rather than as our token kinds -- so an adopter
answers it with their own lexer. The measurement above shows the other three
checks cannot reach a mis-grouping, so removing it would leave a real gap;
removing CST assertions costs nothing because there are none.

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
classify a tier-09 file while it is at tier 02. The oracle already
collects `Ops`, the `B::Concise` op list, so the check is
`ops(file) ⊆ ∪ ops(tiers ≤ N)`, parser-independent and already this project's
method.

The caveat is sharper than "necessary but not sufficient": **the optimiser
can erase the construct a tier is about.** `my $x = 1+2` arrives as
`const[IV 3] s/FOLD` with no `add` op, so a literals-tier file and an
arithmetic-tier file can produce identical op lists. Ops alone cannot derive
a tier.

**RESOLVED: three sources, in layers, and the CST is not one of them yet.**
An earlier revision left this open between "a per-file declaration" and "the
CST once trusted", which is the chicken-and-egg the serialization note
resolves by ordering. The answer needs neither:

    the file DECLARES      the `# TIER` header in the sketch above is
                           the construct set. Authored files, so the
                           declaration is free.
    ops LINT it            `ops(file) subset-of union ops(tiers <= N)`
                           catches a file using something it did not
                           declare. Necessary, not sufficient, and
                           parser-independent.
    the token stream       classifies tiers 01-03, whose constructs are
                           lexical and where a declaration would be
                           checking itself.

The CST becomes a fourth source once it is trusted, and replaces the
declaration rather than the lint. Until then it is out.

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

## The manifest alternative, tested and rejected

**REVISED.** A review recommended a cheaper artifact: `conformance/tiers.txt`,
a MANIFEST ordering perl's existing `t/` by tier, zero files authored,
day-one value. It was the strongest objection to authoring the corpus at all
and deserved a measurement rather than an argument.

It was measured, over all 620 files of `t/` against 12 tier markers:

    3 tiers:  17 files      9 tiers:  93
    4 tiers:  33           10 tiers:  71
    5 tiers:  32           11 tiers:  85
    6 tiers:  51           12 tiers:  71
    7 tiers:  76
    8 tiers:  91

**Zero files touch fewer than three tiers; the median is nine.** And the 17
sparsest are degenerate rather than sparse -- `t/op/index_thr.t` is three
lines that `require './thread_it.pl'`, containing no Perl to grade.

The decisive case is smaller than the histogram. `t/base/if.t` is nine lines
and the most primitive file in perl's entire suite:

    print "1..2\n";
    $x = 'test';
    if ($x eq $x) { print "ok 1 - if eq\n"; } else { ... }

That is tiers 01, 02, 04, 06 and 10 in nine lines. **A manifest can ORDER
`t/`'s files but cannot ASSIGN any of them a tier, because `t/`'s floor is
already above tier 06.** Nothing in `t/` exercises tier 01 alone, so tier 01
cannot be failed through a manifest -- which was the whole of the day-one
value. Authoring wins.

**One part of the manifest case survives and should be taken.**
`t/lib/croak/` IS a manifest that already exists: 338 cases in a flat
`source` + `EXPECT` format, parser-agnostic, already pinned by this repo's
corpus checkout. A runner extracts them rather than a person writing them.

**But the extraction is a `perl -c` split, not a copy** -- see "Negative
files" above, where the measurement is. 201 of 340 are compile-time
failures and become `parsent`; 64 compile fine and become positives; 75
exit 0 under 5.42.0 and are excluded. Trusting the filename would assert
must-not-parse on cases perl parses.

**That still removes the negatives from the authoring cost entirely.** The
category that would otherwise be the most tedious to write is produced by a
runner, and the per-file cost above applies to positives only. The split is
work the runner does once, not work per case.

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
tier 07** -- stated as a construct set rather than a file count or a
duration, for the reason at the top of this document. 01-03 alone would be
satisfied
on day one -- an uncommitted probe measured 50 of 53 such cases passing and
an independent one 13 of 16; neither is reproducible and both should be
treated as indicative only. The three failures in each were the same:
leading decimal `.5` (01a0c13f-97f5), heredoc (01a0c13f-aaf8), and the
ADJUST adjacency. So tier 01 is one landed issue away rather than satisfied,
which means a gate on tiers 01-03 says almost nothing. Tier 04 is the first
tier where a parse can be WRONG
rather than Unknown, and the two largest T2 buckets are parenless call
argument position, which needs the 07 slice.

## Questions this document has closed

Both of the two that were open at decomposition are now answered, and the
answers live in the chain rather than here:

- **Does each tier get a README?** Yes, and for a reason stronger than
  documentation: the dependency lint needs a machine-readable INTRODUCES set
  per tier, and that set has to live somewhere it cannot drift from the
  files it describes. In a Go table it is a second list; in the tier
  directory the drift shows up in the same diff.
- **Does `04_operators` need internal grouping by precedence level?**
  Deferred, deliberately, to the issue that writes tier 04. A grouping
  invented before the files exist will be the wrong grouping; decide it when
  the directory is actually unreadable.

Answered within this document, and recorded here because earlier revisions
listed them as open:

- Does the construct set come from a per-file declaration or from the CST
  once trusted? Proposed: the `# TIER` header declares, the ops lint checks,
  the token stream classifies tiers 01-03, and the CST is out until trusted.
- `conformance/` at repo root? Proposed yes, given the standalone intent.
- What does a file look like? Proposed under "The file format".

## On sources

Perl's own `perlintro`/`perlsyn` (installed, same licence), *Modern Perl*
(CC BY-NC-SA) and *Programming Perl* (proprietary) are priors on teaching
order. Reading them to decide what order to introduce concepts in is a fact
about the language; lifting text is not.

## What is deferred, and where to

Decomposed into `m3-conformance-corpus`. Two commitments in this document
are deliberately NOT in that milestone, recorded here so the deferral is
named rather than silent:

- **The `t/` sweep.** The loop this document describes -- build the corpus,
  implement against it tier by tier, then sweep `t/` and let what still
  fails name constructs the corpus does not cover -- runs AFTER the corpus
  exists. It is the first work of the milestone that follows
  `m3-conformance-corpus`, and it needs the sweep's denominator ("files perl
  compiles") which the harness already computes. Nothing in m3 runs it.
- **Adjacency beyond the adjacent pair.** An adjacency file per tier covers
  "tier N with its declared prerequisite". N with N-3 is unreached by
  construction, and the sweep is what finds it. Same destination.

Everything else in this document is owned by an issue in
`m3-conformance-corpus`.

## The M1-M6 rewrite, measured 2026-09-26

"The corpus comes first, and M1-M6 are rewritten around it" was written
before the corpus existed. It has existed since `b4180102` and been
through a format port since. This section is what the rewrite turns out
to be once measured, and it is smaller than the sentence implies because
three of its five parts already happened.

### What already happened, and where

- **`T1-easy >= 70%` is gone.** `easy_test.go` deleted whole at
  `b4180102`; the twelve `hardMarkers` moved into the tier READMEs that
  claim them, leaving `hardMarkerCount = 12` as the only Go.
- **`T2 100%` never existed as a gate.** Measured at `b4180102`: a
  comment and a `t.Logf`, no `if` anywhere. Removing it removed a claim,
  not a check. `TestT2CoreParses`'s real gate -- the per-file
  `shortfall` map, which fails in BOTH directions -- is untouched and
  says so in its own log line: "Reported, not targeted -- the tier
  corpus gates what the parser reads."
- **The 01-04-plus-07 slice exists.** `TestSmallestUsefulCorpus` demands
  every file reach `passed` or `knownRefusal`.

So the rewrite's remaining work is not "replace the gates". It is the
question those three left open.

### The measurement that decides the rest

This parser over all 209 positive corpus cases, at `b0f1a9e9`:

    192 clean, 17 with Unknown        91.9%

    04_operators    29/33      11_oo         16/18
    07_subroutines  10/13      13_opaque      7/15
    all other ten tiers       100%

Against T2 core in the same tree: **19 of 56 files clean, 33.9%**, from
`TestT2CoreParses`'s own log line. (An earlier revision of this section
said 17 of 56, 30.4% -- that is the figure at the top of this document,
measured 2026-09-21, restated here as current without re-running it. The
number this section exists to correct was itself stale.)

Two instruments ~58 points apart. The tempting conclusion -- "the corpus
does not contain what T2 fails on" -- is FALSE, and checking it is what
this section is for.

`conformance/mdtest/argument-extent.md` holds five tier-07 cases on
exactly the largest T2 bucket: a parenless call's argument extent, greedy
and prototype-cut, plus the undeclared callee as `parses: no`. The
construct was never missing.

What is wrong is the CITATION. Both refusing cases record
`refuses: 01a0c432-fbd5`, an issue whose state is `done`. A reader
following it lands on finished work and concludes the corpus says nothing
about the construct.

### Why the gap is the citations and not the coverage

A corpus green at 92% while the parser fails two thirds of real perl
files looks like a corpus that has stopped making a claim. Measured, it
is making the claim and pointing it at closed work.

Naming all seventeen refusals rather than counting them by tier:

    sixteen   `refuses: unfiled`        BY DESIGN, see below
    five      cite a CLOSED issue       01a0c432 (x3), 01a0c35f, 01a0c730
    three     cite a live issue         01a0ce57, 01a0cf64, 01a0d087
    --------
    twenty-four cases record a refusal

FIVE point at nothing actionable, and `01a0d0c7` -- whose title already
says five -- owns them. Two earlier revisions of this section got the
count wrong in opposite directions: fourteen (counting `unfiled` as a
defect) and then four (dropping one of the three `01a0c432` sites, and
listing `01a0c35f` as both closed and live while omitting `01a0cf64`).
The measurement is `grep -rh 'refuses: 01a0' conformance/mdtest/ | sort |
uniq -c`, which nothing before this revision ran.

**And "of seventeen" was the wrong denominator.** A citation belongs to a
CASE; seventeen is the number of cases where the PARSER refuses. Measured,
the two sets overlap without either containing the other:

    24   record a refusal
    17   the parser refuses
    12   a token fact fails
     5   both

So sixteen `unfiled` and seventeen parse-refusals are different sets whose
overlap is ten, and an earlier revision's "these ARE the 17" was false. Its
supporting sentence -- that the unfiled tiers match the dirty count -- was
false too: three `unfiled` cases are in tier 01, which the same section's
dirty count reports as 100% clean, because a case with a RECORDED refusal
skips and never appears dirty.

The repair is a CITATION repair rather than new cases: the cases exist.

Counting by tier is what hid it. "07_subroutines 10/13 clean" is three
refusals unnamed, and two of the three were the parenless-call cases.

### The construct, isolated

Declaration form and prototype are IRRELEVANT. Argument count is
everything:

    sub answer { 42 }   print answer, "\n";   clean    <- corpus has this
    sub answer { 42 }   answer;               clean
    sub f { }           f(8);                 clean
    sub f { }           f 8;                  UNKNOWN
    sub f;              f 8;                  UNKNOWN
    sub f ($) { }       f 8;                  UNKNOWN
    sub tryeq ($$$$){}  tryeq 1, 13 % 4, 1;   UNKNOWN  <- arith.t, 179 nodes

A parenless call to a DECLARED sub parses with zero arguments and fails
with any. `subroutines.md`'s "Four spellings of one call" passes because
all four spellings call with NO arguments; `argument-extent.md` covers
the argument-bearing forms and REFUSES, which is why the construct is
both covered and invisible in a clean/dirty count.

The arith.t shape adds nothing to it. `f 1, 2`, `f 13 % 4, 2` and
`sub tryeq ($$$$) {} tryeq 1, 13 % 4, 1, 'x'` all refuse identically,
one `trailing_tokens` each -- so an EXPRESSION in argument position is
not a distinct behaviour from a literal, and no new case would measure
anything the corpus does not already measure.

`%` is not implicated. `sub f {} f 13 % 4` and `sub f {} f 8` fail
identically, so the "`%` read as a hash sigil" reading in `01a0c10b` is
a symptom of argument position, not a modulus bug.

### What perl says, which bounds what the corpus may assert

Measured on 5.42.0, and it rules out the obvious fix:

    ok 8;                     SYNTAX ERROR -- "Number found where
                              operator expected (Do you need to
                              predeclare ok?)"
    ok $x, 9;                 syntax OK -- $x->ok, '???'
                              INDIRECT OBJECT, not a call
    sub ok {} ok 8, 9;        syntax OK -- a call
    sub f {} f 13 % 4;        syntax OK -- f 1, constant folded

So a parser that greedily consumes arguments after an UNDECLARED callee
builds a tree for a program perl rejects, which is WRONG in the gate's
sense and M1's gate is Oracle WRONG = 0. The corpus must assert the
declared case parses and the undeclared case does not -- two cases, not
one, and the second is `parses: no`.

It already does both. `argument-extent.md`'s "An undeclared callee is a
syntax error" is that `parses: no` case, and it was written in
`b4180102`, whose own commit message records reaching this conclusion by
measurement: "AC1 asks that a parenless call to an unknown callee have
its argument extent MEASURED rather than assumed. Measured, there is no
extent, because there is no call." This section re-derived it five days
later without checking, which is the same failure in miniature.

### Five refusals citing closed issues, and the two denominators

Found while measuring: sixteen cases carry `refuses: unfiled`.

    heredocs.md            3      comparison.md       3
    formats.md             2      pod-and-data.md     2
    vstrings.md            2      arithmetic.md       1
    numeric-point.md       1      adjacency-04/-13    2

They are NOT the seventeen parse-refusals, and an earlier revision said
they were. Measured: sixteen unfiled, seventeen parse-refusals, overlap
ten. By tier the unfiled are 04 five, 13 eight, and **01 three** -- which
the dirty count above reports as 100% clean, because a recorded refusal
SKIPS and so never counts as dirty. Two different questions, two different
denominators.

**They are recorded BY DESIGN and are not the defect.** `unfiled` is a
deliberate sentinel: `internal/conformance/file.go:275`, the topic-format
spelling of `selfRecorded` ("this file"), meaning THE CASE IS ITS OWN
RECORD. The case body already states the measured behaviour and the tokens
we produce instead, so an issue would be a second copy free to go stale.
`TestRefusalCitationMustResolve/unfiled_is_not_looked_up` asserts it and
passes, and its comment rejects the contrary reading in advance: treating
the sixteen as citations pointing at nothing is "a claim about the corpus
that is not true".

**Measured: ZERO are stale.** All sixteen still refuse. Two probes were
wrong before one was right, and both errors are this chain's recurring
defect in new clothes:

  - The first counted Unknown nodes and reported six stale. `verdict()`
    collects TWO kinds of evidence -- refusal codes AND token facts -- and
    six cases refuse on a fact with a clean parse, so counting nodes alone
    read them as passing.
  - The second called all sixteen defects without reading the reader.
    `grep -rn unfiledRefusal internal/` finds the const, its rationale and
    the test asserting it, in one command.

Both times the corpus was right and the instrument measuring it was not.

Split by what actually refuses:

    ten     the parser refuses -- heredocs (3), formats (2),
            pod/data (2), the two adjacency bodies, the logical-op case
    six     the parse is CLEAN and a TOKEN FACT fails
              three   a word-shaped operator (`x`, `cmp`, `and`)
              three   two `scanNumber` gaps (`5e-1`, two vstring forms)

The logical-operators case is in the TEN, not the six. It carries
`refusal: not_a_term` (comparison.md:112) as well as a failing `xor`
fact -- the only one of the four word-shaped-operator cases that does.
An earlier revision counted it in both halves and called all four
"clean", which is false for it and is why the cleanup issues overlapped.

### Three of those facts contradict the glossary

`x`, `cmp` and `and` are each asserted as `one operator whose text is
"..."`. The lexer emits `Word` and the corpus scores that as a refusal.
(`xor` makes the same assertion, but its case also carries a parser
refusal, so it belongs to the ten above and is repaired with them.) The glossary is the authority the token layer names, and it
says:

> **operator** -- Punctuation denoting an operation: `+`, `.`, `=~`,
> `->`, `?`, `:`.

> **word** -- A bare identifier: `foo`, `Foo::Bar`, `my`, `print`.
> Keywords are not distinguished from other identifiers at this layer.

`x` is not punctuation. By the glossary's own definition the lexer is
RIGHT and these corpus facts are wrong. `01a0ce57` independently reached
the same conclusion in prose -- "the binary `x` is lexed correctly as a
Word" -- while the corpus scored the opposite.

The conflation is a layer confusion, and both layers are individually
right. Perl emits `repeat` for `$a x 3`, an operator AT THE OP LAYER;
`categories.go` maps the word "operator" straight to `lexer.Operator`, a
TOKEN kind. A fact wanting to say "this spells perl's repeat op" has no
vocabulary for it and reached for the one word that appears in both
layers.

So this cluster is not parser work at all. Either the four facts are
rewritten in token vocabulary, or the glossary grows a category for
word-shaped operators and the lexer follows -- and that is a decision
about the glossary, which is the corpus's contract with B::SoN and Chalk,
not a local fix.

The remaining two are real lexer gaps: `5e-1` lexes as
`Number(5e) Operator(-) Number(1)`, and neither `65.66.67` nor
`v65.66.67` forms a vstring.

### What this does NOT change

- The `t/` sweep stays deferred to its own milestone, as recorded above.
  Extending the corpus is not the sweep; the sweep is what finds the
  constructs the corpus still lacks AFTER this.
- The `shortfall` ratchet stays. Two instruments measuring different
  things is correct; one of them pretending to measure the other is not.
- M2-M6's gates are untouched by this section. M2 lowers the tree this
  parser builds, so it inherits whatever M1's coverage becomes; nothing
  in M2's sixteen issues reads the corpus and nothing needs to.
