<!-- ABOUTME: Index for the Perl language specification targeting a Go stdlib-only parser. -->
<!-- ABOUTME: Start here; chapters are independent but chapter 0 constrains everything else. -->

# A Perl Language Specification for a Go Implementation

A specification sufficient to implement a Perl parser in Go using only the
standard library. The parser is a component of PSC, the static-analysis
compiler, not a project of its own: it feeds PSC's type inference and drives
PSC's incremental LSP, and chapter 6 §6.1.6 specifies what sits between the
tree and inference.

Derived by triangulating three sources. Where they disagree, **perl wins** — it
is the definition, the other two are approximations of it.

| Source | Language | Size | Role |
|---|---|---:|---|
| `perl5/toke.c` + `perly.y` | C | 16,605 | Ground truth |
| PerlOnJava | Java | 24,003 (parser) | A from-scratch parser that runs real Perl |
| perl-lsp | Rust | 36,499 (lexer+parser) | A from-scratch Perl LSP |

## Read in this order

| # | Chapter | Size | What it settles |
|---|---|---:|---|
| 0 | [Measured findings](00-findings.md) | ~300 | The numbers. Read first — several contradict the folklore. |
| 1 | [Scope and goals](01-scope.md) | ~130 | Constraints, and the three metrics kept distinct throughout |
| 2 | [Lexical structure](02-lexical-structure.md) | ~2.5k | Tokens, literals, quote-like operators, heredocs |
| 3 | [The lexer feedback problem](03-lexer-feedback.md) | ~1k | `PL_expect`, prototypes, what is undecidable |
| 4 | [Expression grammar](04-expressions.md) | ~1.7k | 32 precedence levels, the Pratt tables |
| 5 | [Statements and declarations](05-statements.md) | ~3.2k | Program structure, recovery, re-parse anchors |
| 6 | [Incremental parsing and LSP](06-incremental-lsp.md) | ~2.2k | The architecture, the latency budget; what inference walks and what the prior art says (§6.1.5–6.1.8) |
| 7 | [Conformance](07-conformance.md) | ~1k | How "correct" is measured: the oracle, the four metrics, differential testing |

Eight documents, roughly 12,600 lines. Chapters 2-5 specify the language; 6
specifies the machine; 7 specifies how you know it works.

## Moved out of the specification

- **The milestone ladder (M0–M6), the ratchet, CI shape, corpus vendoring and
  licensing, and the harness code sketches** are in
  [`docs/plans/2026-09-05-parser-conformance-plan.md`](../../plans/2026-09-05-parser-conformance-plan.md).
  They are a project schedule and scaffolding, which change as work
  progresses; chapter 7 keeps the definitions they measure against.
- **Prior art — PerlOnJava and perl-lsp assessed** (formerly appendix A1) is
  in [`docs/plans/2026-09-05-parser-prior-art.md`](../../plans/2026-09-05-parser-prior-art.md).
  Research notes on two other projects, not a specification of Perl. The
  prior art on type checkers that serve an LSP — TypeScript, rust-analyzer,
  Roslyn, Sorbet, Pyright — stays in the specification, at chapter 6 §6.1.5,
  because it decides the machine's shape rather than describing a project.

## The four findings that shape everything

**1. The undecidable part is small, but only if you measure it correctly.**
487 of perl's 620 test files contain `BEGIN` — 78%, which sounds fatal. But 466
of those are `chdir 't'; require './test.pl'` boilerplate that cannot change a
parse. Parse-relevant `BEGIN` is ~21 files (3.4%); source filters are 2 (0.3%).
The difference between those two readings is the difference between an
impossible project and a tractable one.

The same trap caught this document. An incomplete test shim reported that perl
could compile only 66% of its own suite, and those failures were written up as
properties of the corpus. Completing the shim took it to **94.2%**, and
classification showed exactly **one** file in 620 failing for a reason about
the language. Measure the harness before you believe the measurement.

**2. Fidelity is measurable, and nobody has measured it.**

```
$ perl -MO=Concise,-exec -e 'sub f(\@){} my @a; f(@a)'   # srefgen PRESENT
$ perl -MO=Concise,-exec -e 'sub f{}    my @a; f(@a)'    # srefgen ABSENT
```

Perl reports how it parsed something. So "did we parse this the way perl did?"
is a measurement, not an opinion. Our current parser produces **identical,
error-free trees** for those two programs — one of them is wrong, and coverage
scores both as a success (`internal/parseoracle/fidelity_test.go`).

**3. The parser is the latency problem, not the type checker.** Measured on
perl5/lib, parse costs 20-50× `Analyze`:

| File | Parse | `Analyze` |
|---|---:|---:|
| `overload.pm` (1701 lines) | 99.6 ms | 2.2 ms |
| `sigtrap.pm` (327 lines) | 350.0 ms | 7.0 ms |

Cost tracks grammar pathology, not file size, so it cannot be predicted or
debounced away. A 350 ms parse is 35× over an interactive budget.

The parse column is the tree-sitter parser's (2026-09-04), which is what
PSC still consumes. The hand-written `internal/parse/` has not been
benchmarked on these files; the `Analyze` column is unaffected by which
parser produced the tree.

**4. Modern Perl deletes an ambiguity class.** `S_intuit_method` returns 0 when
the `indirect` feature is off (`toke.c:5071`). `my $o = new Foo;` is a **syntax
error** under `use v5.36`. The construct that forces symbol-table consultation
to resolve a bareword is opt-out, and most new code has opted out.

## Status

The specification is complete. A parser exists and is measured against it.
At commit 3b9422eb: 84.2% of the corpus parses, 341 of 986 T1 files (34.6%)
are clean, 5,289 `Unknown` nodes remain, 620/620 corpus files round-trip,
and 1,000,000 fuzz executions pass. The M1 gate (issue 01a0a70f) is in
progress; against the plan's M1 targets (100% of T2, ≥70% of T1-easy, oracle
WRONG = 0) it is materially incomplete. What exists in code:

- `internal/lexer/` and `internal/parse/` — the hand-written lexer and
  parser this specification describes. 23 node kinds
  (`internal/parse/parse.go:146-201`); `Unknown` for what it declines to
  parse; `Call{Resolved:false}` for a call to an unseen sub.
- `internal/parseoracle/` — extracts perl's parse decisions as JSON, plus the
  test that records the fidelity gap.
- `internal/infer/` — PSC's type inference. It still consumes the tree-sitter
  tree through `internal/parser/`; its conversion to `internal/parse/`, and
  the lowering pass that conversion goes through, is chapter 6 §6.1.6.

Two measured defects in `internal/parse/` are recorded rather than hidden:
`($x)` loses its `Paren` node (`term.go:284-293`, against chapter 4 §4.10)
and `${$h->{k}}` lexes to one token and parses to one childless leaf (issue
01a0ad52). Both pass round-trip, which proves no byte was lost and nothing
about structure (chapter 7 §7.2(c)). Chapter 6 §6.1.8 lists them with the
rest of what is open.

## A caution about this document

Chapters were drafted by separate agents and then verified against running
perl. That verification changed real conclusions: five precedence bugs found in
the reference implementations, a latency claim inverted, a source-filter
citation corrected, a feature gate moved by two versions.

Claims here carry a citation (`toke.c:5071`) or a measurement. Where a claim is
an estimate, it says so. Treat anything without either as unverified — that is
how the errors above were caught, and it is unlikely they were the last.
