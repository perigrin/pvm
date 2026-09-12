<!-- ABOUTME: The decision memo for whether to write a Perl parser in Go. -->
<!-- ABOUTME: Summarises the specification's evidence; the decision itself is perigrin's. -->

# Should we write a Perl parser in Go?

The specification in `docs/specs/perl-parser/` (~12,700 lines) exists to make
this decidable. It does not make the decision. This memo states the evidence
and the options.

## What is actually broken

Two problems, both measured, both in the parser we ship today.

**1. Latency, by one to two orders of magnitude.**

| File | Lines | Parse | PSC `Analyze` |
|---|---:|---:|---:|
| `overload.pm` | 1701 | 99.6 ms | 2.2 ms |
| `sigtrap.pm` | 327 | **350.0 ms** | 7.0 ms |
| `_charnames.pm` | 858 | **521.7 ms** | 19.3 ms |

An interactive budget is ~10 ms. Cost tracks grammar pathology, not file size,
so it cannot be predicted, capped, or debounced — the user is typing in the
file that is slow. No incremental layer recovers a constant factor this large.

**2. Fidelity is unmeasured, and known to be wrong.**

```perl
sub f(\@){} my @a; f(@a);     # perl passes a REFERENCE
sub f{}    my @a; f(@a);      # perl passes a LIST
```

Our parser produces **identical, error-free trees** for both. Coverage scores
it a success. `internal/parseoracle/fidelity_test.go` records this.

**Coverage on the easiest corpus is 29/44 (65.9%)** — files perl compiles
without complaint, that we cannot parse.

## What makes it tractable

The three reasons this is not a doomed project, all measured:

- **The undecidable part is ~1-3%**, not the 78% a naive `BEGIN` grep suggests.
  466 of 487 `BEGIN` blocks in perl's suite are `@INC` boilerplate. Source
  filters are 2 files in 620.
- **`use v5.36` deletes the bareword ambiguity class.** Indirect object syntax
  is a syntax error under modern feature bundles (`toke.c:5071`).
- **Fidelity is checkable.** Perl reports its own parse decisions, so every
  approximation gets a number instead of an argument.

## The cost

~35,000-40,000 lines of Go, plus ~30,000 of test. That estimate comes from two
independent implementations converging within 15%: PerlOnJava's parser is
24,003 lines of Java, perl-lsp's lexer+parser ~36,500 lines of Rust. Parsing
Perl costs roughly 25k lines in any language.

This is a months-long project, not a sprint.

## What nobody has done

**Incremental Perl parsing does not exist.** perl-lsp has ~7,000 lines of it
behind a feature flag that production never calls; its own source says so:

> After applying the edits the *entire* document is reparsed — incremental
> *parsing* is future work.
> — `text_sync.rs:1-9`

Its published "70-90% reuse" comes from a path that does a full parse first.
This is genuinely unsolved territory, which is either the interesting part or
the risk, depending on appetite.

## Options

**A. Do nothing.** The parser is slow and sometimes wrong, but PSC works well
enough on small files. Costs nothing. The 350 ms parse stays.

**B. Fix the current parser incrementally.** Continue the gotreesitter scanner
work. Known remaining: the heredoc-as-call-argument GLR bug (~32 files),
`_regexp_open_*`, five `_RECOVER_*` tokens. Cheap per step, and each step has
been landing. But it cannot fix the 350 ms — that is the runtime, not the
grammar — and it cannot fix fidelity, since tree-sitter has no prototype table.

**C. Write the Go parser.** Fixes both problems. ~40k lines. The spec exists
and the measuring instrument is built, so it would not start cold.

**D. Build the fidelity harness first, decide after.** ~1-2 days. Runs the
oracle across the 584-file corpus and produces exact/wider/WRONG buckets for
the current parser. Converts "sometimes wrong" into a number.

## Recommendation

**D, then decide between B and C.**

The fidelity number does not exist yet. Everything above says the current
parser is wrong in a specific way on a specific construct; what nobody knows is
whether that is 2% of real files or 20%. That single number changes which of B
and C is correct, it costs a day or two against a months-long commitment, and
the harness is needed for C anyway — it is the acceptance test.

If fidelity comes back fine, B plus a performance investigation may be enough.
If it comes back bad, C is justified on correctness and the latency argument
becomes a bonus rather than the case.

One caveat worth stating: the specification was written by agents and corrected
against running perl at every step, and that correction process caught five
precedence bugs in the reference implementations, an inverted latency claim, a
misattributed `toke.c` citation, and a ceiling that was wrong by 28 points
because of my own broken test harness. The numbers here have been checked. They
have not been checked twice.
