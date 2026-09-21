# A graded conformance corpus for Perl 5 parsers

**Status:** design, not yet decomposed. Written 2026-09-21.

## The problem

Neither existing corpus tells us where the parser is.

    T2 core      30.4%  (17 of 56)    perl's own t/, target 100%
    T1-easy      53.2%  (351 of 660)  PerlOnJava unit/, target 70%

Both are someone else's test suite, ordered alphabetically, measuring "how
much of this happens to parse" rather than "how far up the language have we
got". Neither number moves in a way that says what to build next. M1 reached
32/36 issues with both of those targets missed and nothing on the chain
owning either -- the count was hiding the gap.

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

The reference point is Raku's ROAST -- a graded, executable, construct-indexed
suite -- with one important difference. ROAST is normative: Raku is defined by
it. Perl 5 has no specification but its implementation, so this corpus is
DERIVED from perl rather than defining it. Every file is validated against
`perl` before it counts, which is the same rule PerlOnJava's own AGENTS.md
states:

> ALWAYS validate new unit tests with standard Perl before relying on them.
> Unit tests must encode standard Perl behavior, not PerlOnJava-specific
> behavior.

## The ordering criterion

**A tier sits where it does because the next tier cannot proceed without it.**

Not "how hard is it" and not "how central is it conceptually". Literals and
variables are self-contained -- `my $x = 1` needs no notion of context. The
moment operators arrive you cannot proceed without context: `@a + 1` is
arithmetic on a count and `"$x" . @a` is not, and the operator decides which.
So context sits immediately before operators, and no earlier.

That is the same rule as the dependency check, applied to concepts rather
than syntax, which means the ordering has one principle rather than three
competing ones.

It also resolves a tension worth recording. chromatic opens *Modern Perl*
with context, before any syntax, because a READER needs the frame
immediately. A PARSER needs it exactly when operators arrive. Same fact,
different consumers.

## The tiers

    0100_literals/       numbers, strings, quoting, qw
    0200_variables/      sigils, scalars, arrays, hashes, element access
    0300_context/        scalar vs list, the discriminating pairs
    0400_operators/      arithmetic, string, comparison, logical,
                         precedence, associativity, the and/&& cliff
    0500_scoping/        my, our, state, local, blocks
    0600_control/        if/unless, while/until, for/foreach, postfix
    0700_subroutines/    declaration, call forms, @_, return
    0800_references/     \, deref, arrow, anonymous constructors
    0900_regex/          match, substitution, binding, delimiters
    1000_io/             filehandles, heredocs, readline, print
    1100_packages/       package, use, require, imports
    1200_oo/             bless AND class/field/method/ADJUST
    1300_base/           what perl's own t/base assumes

Numbering is regenerable -- the number is derived from the classification, so
reordering means regenerating, not hand-editing. No sparse gaps needed.

Notes on placements that are not obvious:

- **0300_context is where the inference gap becomes measurable.**
  `types.NarrowByContext` exists and is called (`infer.go:524`) with NO source
  of context: the CST has no `Wantarray` node and no call-site propagation
  (§4.14.2). Three active defects are context questions wearing different
  hats -- rv2hv's remaining 10 sites, the `%` sigil-versus-modulus case, and
  `NarrowByContext` itself.
- **0800_references after 0700_subroutines**, following *Modern Perl*.
  Dereference syntax (`@$r`, `$h->{k}`, `@{$h{k}}`) is genuinely harder than a
  sub declaration, even though `\@a` alone looks simpler.
- **1200_oo covers both object systems.** `class`/`field`/`method` is
  different PARSING, not a variant of `bless`: `field $x :param = 1` puts
  attributes between the name and the initialiser, which `decl.go` handles
  specially because an expression parser reads the `:` as a ternary colon.
  Measured 2026-09-21: `class`, `field`, `method`, `ADJUST`, `:param` and
  `:isa` all parse clean today, so this tier's value is inference and edge
  cases rather than "does it parse".
- **1300_base is the terminus** -- the join where T2's first directory
  becomes reachable.

## Where the files come from, and the honest split

Extraction does not survive contact with the bottom tiers, and this is the
measured reason.

T1 files are one-construct-SUBJECT, not one-construct-CONTENT. Example:
`numeric_integral_nv_multiply.t` is about numeric literals, and every
assertion in it is wrapped in `map`, `qw`, regex, `no warnings` blocks and
string interpolation -- constructs from six later tiers. There is no
tier-0100 fragment to extract; there is a numeric topic expressed in
tier-0900 syntax. `string_interpolation.t` is the same shape: nominally
literals, actually `subtest` + `local` + `@{...}` + `$"`.

That follows from what T1 is. Tests for a RUNTIME, written to exercise
semantics, using whatever syntax is convenient. A parser curriculum needs the
opposite.

So:

- **Tiers 0100-0300: authored.** Minimal, no test harness -- `my $x = .5;`
  plus a parse assertion. Roughly 40 files. Cite a T1 file where one
  prompted a case, as a source of CASES rather than of text. No licence
  question: project-authored, T0 in §7.3.6's terms.
- **Tiers 0400 and up: extracted.** By then the harness idiom is introduced,
  so a `subtest` block comes across nearly verbatim with a real citation.

Citing copied text as copied and learned cases as learned is the point. A
uniform "extracted from" header across all tiers would overclaim for the
bottom three.

## Licensing

PerlOnJava is dual Artistic 1.0 / GPL 1.0+, copyright Flavio Glock -- the
same terms as perl's own tests. Both licences require the copyright notice
and licence terms to travel with redistributed copies, and Artistic 1.0 asks
that modified files be marked as changed. Renumbering and adding a header is
a modification.

Per-file header for extracted files:

    # Adapted from PerlOnJava src/test/resources/unit/<name>.t
    #   Copyright (c) Flavio Glock. Dual-licensed Artistic 1.0 / GPL 1.0+.
    #   Source revision: <sha>
    # Modified: extracted to tier NNNN and renumbered for the pvm
    #   conformance ordering; assertions otherwise unchanged.

Plus the licence texts in the corpus tree itself, so it stays
self-contained.

**This reverses a recorded decision.** §7.3.7 says the corpus is "kept out of
this repository", and the pins (`t1corpus.pin`, `corpus.pin`) exist so it is
not vendored. Copying a classified subset IN reverses that, defensibly --
this becomes our artefact with our ordering rather than a mirror -- but
§7.3.7 must be updated to say so rather than being silently contradicted.

Intent is that this may later become a standalone project for anyone
building a Perl 5 parser, so it should carry no dependency on pvm's
internals.

## The checks that make the ordering a claim rather than an opinion

1. **Dependency.** Each file's constructs were introduced at or before its
   tier. Mechanically verifiable: parse the file, extract its construct set,
   compare against the cumulative set of earlier tiers. Without this the
   numbering is editorial.
2. **Tier necessity.** A tier that could move earlier without breaking the
   dependency check is misplaced. This is the ordering criterion made
   testable.
3. **perl validation.** Every file compiles under perl before it counts.

## What this replaces

`TestT1EasyParseRate` and its 70% target, if the corpus proves out. "Parses
cleanly through tier N" says something; "70% of an alphabetical pile that
includes Archive::Zip integration tests" does not.

That is a plan change, not an addition. Filed here rather than done quietly:
a new corpus that lands ALONGSIDE the existing targets makes M1 further from
done, not closer.

## Open questions

- Does each tier get a README stating what it introduces and why it sits
  there? Recommended yes -- it is what makes the ordering arguable.
- Where does it live? `conformance/` at repo root, not
  `internal/parse/testdata`, given the standalone intent.
- Does 0400_operators need internal grouping by precedence level? It is the
  largest tier by a distance -- 32 levels, three associativities, the
  nonassoc rule and the and/&& cliff.
- Ordering sources: perl's own `perlintro`/`perlsyn` section order (installed,
  same licence), *Modern Perl* (CC BY-NC-SA) and *Programming Perl*
  (proprietary) as priors on teaching order. Reading them to decide what
  order to introduce concepts in is a fact about the language; lifting text
  is not.

## Next step

Prototype `0100_literals` -- the tier where classification is unambiguous, so
it shakes out the method before the cases that need judgement. Then
`0300_context`, which is where the format gets stressed.

Not decomposed into issues yet. This document is the input to
`crochet:refinement`.
