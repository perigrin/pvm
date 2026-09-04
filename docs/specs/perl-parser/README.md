<!-- ABOUTME: Index for the Perl language specification targeting a Go stdlib-only parser. -->
<!-- ABOUTME: Start here; chapters are independent but chapter 0 constrains everything else. -->

# A Perl Language Specification for a Go Implementation

A specification sufficient to implement a Perl parser in Go using only the
standard library, to drive an incremental LSP and feed the PSC type checker.

Derived by triangulating three sources. Where they disagree, **perl wins** — it
is the definition, the other two are approximations of it.

| Source | Language | Size | Role |
|---|---|---:|---|
| `perl5/toke.c` + `perly.y` | C | 16,605 | Ground truth |
| PerlOnJava | Java | 24,003 (parser) | A from-scratch parser that runs real Perl |
| perl-lsp | Rust | 36,499 (lexer+parser) | A from-scratch Perl LSP |

## Read in this order

| # | Chapter | Lines | What it settles |
|---|---|---:|---|
| 0 | [Measured findings](00-findings.md) | 127 | The numbers. Read first — several contradict the folklore. |
| 1 | [Scope and goals](01-scope.md) | 133 | Constraints, and the three metrics kept distinct throughout |
| 2 | [Lexical structure](02-lexical-structure.md) | 2,499 | Tokens, literals, quote-like operators, heredocs |
| 3 | [The lexer feedback problem](03-lexer-feedback.md) | 987 | `PL_expect`, prototypes, what is undecidable |
| 4 | [Expression grammar](04-expressions.md) | 1,652 | 33 precedence levels, the Pratt tables |
| 5 | [Statements and declarations](05-statements.md) | 2,595 | Program structure, recovery, re-parse anchors |
| 6 | [Incremental parsing and LSP](06-incremental-lsp.md) | 1,817 | The architecture, the latency budget |
| 7 | [Conformance](07-conformance.md) | 1,647 | The milestone ladder and how "done" is measured |
| A1 | [Prior art](A1-prior-art.md) | 1,215 | What the other two got right and wrong |

Roughly 12,700 lines. Chapters 2-5 specify the language; 6 specifies the
machine; 7 specifies how you know it works.

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

**4. Modern Perl deletes an ambiguity class.** `S_intuit_method` returns 0 when
the `indirect` feature is off (`toke.c:5071`). `my $o = new Foo;` is a **syntax
error** under `use v5.36`. The construct that forces symbol-table consultation
to resolve a bareword is opt-out, and most new code has opted out.

## Status

The specification is complete. **No parser has been written** — that decision
is open. What exists in code is the measuring instrument:

- `internal/parseoracle/` — extracts perl's parse decisions as JSON, plus the
  test that records the fidelity gap.

## A caution about this document

Chapters were drafted by separate agents and then verified against running
perl. That verification changed real conclusions: five precedence bugs found in
the reference implementations, a latency claim inverted, a source-filter
citation corrected, a feature gate moved by two versions.

Claims here carry a citation (`toke.c:5071`) or a measurement. Where a claim is
an estimate, it says so. Treat anything without either as unverified — that is
how the errors above were caught, and it is unlikely they were the last.
