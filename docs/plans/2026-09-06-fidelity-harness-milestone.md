<!-- ABOUTME: Architect context for the parser-fidelity-harness milestone. -->
<!-- ABOUTME: git-zhi milestones carry no body, so the context lives here and issues cite it. -->

# Milestone: parser-fidelity-harness

## Context

The parser we ship is measured on **coverage** — does a file parse without an
ERROR node — and never on **fidelity**: does it parse the way perl parses.
Those are different questions, and a file can score perfectly on the first
while being parsed wrong.

Demonstrated in `internal/parseoracle/fidelity_test.go`:

    sub f(\@){} my @a; f(@a);     perl passes a REFERENCE (srefgen in the optree)
    sub f{}    my @a; f(@a);      perl passes a LIST

Our parser produces **identical, error-free trees** for both. One of them is
wrong and nothing catches it.

This milestone builds the instrument that turns that from an anecdote into a
number. It does not write a parser. Its output is the measurement that decides
whether the 35-40k-line Go frontend in `docs/specs/perl-parser/` is justified —
and if that rewrite happens, this harness is its acceptance test either way.

**What exists.** `internal/parseoracle/` is 254 lines: a Perl script
(`testdata/parse_facts.pl`) that compiles a file and emits perl's own parse
facts as JSON — the linear op sequence, srefgen/entersub counts, and every
prototype in scope — plus two Go tests. Everything is unexported and lives in
`_test.go`, so nothing outside the package can call it.

**The corpus.** perl's own `t/` is 620 files, of which perl 5.42.0 compiles
584 (94.2%) given a correctly built shim; see `00-findings.md` §0.5. The shim
matters: `t/test.pl:119` clears `@INC`, so these tests only compile inside a
tree where `../lib` is populated from BOTH the pure-perl and
architecture-specific library roots. Omitting the second costs ~170 files.

**Known traps, all measured.** `perl -c` reports one error and stops, so a
stderr classifier sees only each file's first failure — the syntax-error count
in §0.5 is a lower bound for that reason. The corpus here is blead 5.45 while
the interpreter is 5.42.0, so corpus and oracle must be version-pinned
together. And the optree is post-peephole (`1+2` arrives as `const[IV 3]`), so
comparisons must test for the presence of a marker op, never op counts or an
exact sequence.

## File Structure

Existing, to be extended:

    internal/parseoracle/testdata/parse_facts.pl   107 lines, emits perl's parse facts
    internal/parseoracle/fidelity_test.go          147 lines, two characterisation tests
    internal/parseoracle/README.md                 the shim recipe and the ceiling table

New:

    internal/parseoracle/oracle.go        exported API: run perl, decode facts, cache
    internal/parseoracle/compare.go       exact / wider / WRONG / no-answer bucketing
    internal/parseoracle/corpus.go        shim construction, stderr classification
    internal/parseoracle/testdata/corpus.pin    perl + perl5 revision pinning
    internal/parseoracle/testdata/ratchet/baseline.txt   committed per-file verdicts

Specification and plan:

    docs/specs/perl-parser/07-conformance.md      what "correct" means; the metrics
    docs/plans/2026-09-05-parser-conformance-plan.md   §1 harness, §2 CI speed,
                                                       §3 corpus, §4 ratchet

## Design Rationale

Four stages, and the ordering is forced by what each one needs from the last.

**The oracle must be callable before anything can call it.** Today it is
test-only. Lifting it to an exported API with a decoded result type is the one
piece everything else depends on, so it is the head of the chain.

**The corpus must be reproducible before a measurement means anything.** The
shim, the version pin, and the stderr classifier are independent of the
comparison logic and can proceed in parallel with it — but a number produced
against an unpinned corpus is not a number, it is an anecdote.

**The comparison is where the judgment lives.** Bucketing into exact / wider /
WRONG is not mechanical: a static parser that answers "I don't know" is
behaving correctly, and must not score as WRONG. This is the piece most likely
to need revision after seeing real output, so it is deliberately separated from
the plumbing that feeds it.

**The ratchet comes last**, because a baseline of per-file verdicts is only
meaningful once the verdicts are trustworthy. Committing it earlier freezes
noise.

Critical chain: exported API → comparison → corpus run → baseline. Corpus
logistics parallelise against the comparison work.
