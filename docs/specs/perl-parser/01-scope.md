<!-- ABOUTME: Chapter 1 of the Perl parser specification: scope, goals, and constraints. -->
<!-- ABOUTME: Defines what a Go-stdlib-only Perl parser must do and, more importantly, must not promise. -->

# 1. Scope and Goals

## 1.1 What is being specified

A parser for the Perl programming language, implemented in Go using only the
Go standard library, sufficient to power:

1. **A Language Server (LSP)** that re-parses incrementally as the user types.
2. **The PSC type checker**, which consumes the resulting syntax tree.

This document is a specification for an implementer. It is not a tutorial on
Perl and not a survey of parsing techniques. Where it states a rule, that rule
was read out of Perl's own source, or measured by running perl, and it is
cited.

## 1.2 Hard constraints

| Constraint | Consequence |
|---|---|
| Go standard library only | No tree-sitter, no ANTLR, no goyacc-generated tables from a third-party grammar. Hand-written lexer and recursive-descent/Pratt parser. |
| `CGO_ENABLED=0` | Cross-compilation stays trivial. This is PVM's stated premise and the reason the current tree-sitter binding exists at all. |
| Incremental re-parse | The architecture is constrained from the start; retrofitting incrementality onto a batch parser is a rewrite. |
| Feeds PSC | The tree is not an end in itself. Node kinds and context propagation must serve type inference. |

`goyacc` ships with the Go toolchain but is not in the standard library, and
generating an LALR parser from a transcription of `perly.y` would inherit
Perl's grammar conflicts without inheriting `toke.c`'s lexer feedback — the
part that actually resolves them. Recursive descent is the recommendation, and
Chapter 4 gives the precedence table in the form a Pratt parser needs.

## 1.3 The correctness bar, stated honestly

Perl cannot be parsed correctly by a static parser. This is not a limitation of
effort; it follows from the language definition:

```perl
BEGIN { eval "sub f(\\@) {}" }   # a prototype installed at compile time
my @a;
f(@a);                            # parses as f(\@a) — or as f(@a) if the BEGIN failed
```

The parse of line 3 depends on running line 1. A source filter goes further and
rewrites the file arbitrarily before perl sees it (`toke.c` calls into Perl code
for this). Any static parser therefore **approximates**.

The useful question is not "is it correct" but "how wrong, how often, and on
what". That is measurable, and Chapter 7 specifies the measurement. Three
metrics are kept distinct throughout this document:

| Metric | Meaning | Notes |
|---|---|---|
| **Coverage** | Parses without producing an error node | The weak metric. Everyone reports it. |
| **Fidelity** | Parses *the way perl parses* | The metric that matters. Verifiable — see below. |
| **Stability** | Incremental re-parse equals full re-parse | Property-testable. Non-negotiable for an LSP. |

A file can score perfectly on coverage and still be parsed wrong. `f @a` read
as a list where perl took a reference produces a clean tree and a false one.

## 1.4 Fidelity is measurable: perl reports its own parse

Perl will tell you how it parsed something. This converts every approximation
in this specification from a matter of opinion into a measurable quantity.

```
$ perl -MO=Concise -e 'sub f(\@){} my @a; f(@a)'   # srefgen present
$ perl -MO=Concise -e 'sub f{}    my @a; f(@a)'    # srefgen absent
```

Identical source shape; different parse; the optree says which. The same holds
for a prototype installed at compile time by a string `eval` inside `BEGIN` —
measured, `srefgen` still appears — which is the undecidability argument of
1.3 confirmed rather than asserted.

Three other probes, with their limits:

| Probe | Reports | Limit |
|---|---|---|
| `perl -MO=Concise` | The optree — the parse as perl resolved it | Verbose; needs op-level interpretation |
| `prototype(\&f)` | A sub's prototype string directly | Only answers the prototype question |
| `perl -MO=Deparse` | Source regenerated from the optree | **Weaker than it looks.** It emits source that *re-parses* to the same tree, not source that shows the parse. `sub f(\@){} my @a; f(@a)` deparses to `f(@a)`, not `f(\@a)` — the `srefgen` is invisible. Use it for round-trip checks, never for disambiguation questions. |
| `perl -c` | Compile-only pass/fail | Binary; says nothing about *how* it parsed |

Every "the static parser must guess here" in this document is accompanied by
the probe that checks the guess. An unmeasured approximation is a defect
waiting to be discovered by a user; a measured one is a known number on a
dashboard.

## 1.5 Non-goals

- **Executing Perl.** No runtime, no optree, no bytecode. PerlOnJava does this;
  this project does not.
- **Regex internals.** A regex literal is lexed to its delimiters and modifiers
  and kept as a unit. The pattern's own grammar is out of scope, with the
  exception of interpolated variables and capture-group counting, which PSC
  needs.
- **Source filters.** Detected and declared unanalyzable. See Chapter 3.
- **Perfect fidelity on pathological input.** `t/op/lex.t` exists to abuse the
  lexer. Failing some of it is acceptable and expected; failing silently is not.
- **`format`/`write`** beyond skipping the picture-line block intact.

## 1.6 Structure of this specification

| Chapter | Contents |
|---|---|
| 1 | Scope and goals (this chapter) |
| 2 | Lexical structure — tokens, literals, quote-like operators, heredocs |
| 3 | The lexer feedback problem — `PL_expect`, prototypes, undecidability |
| 4 | Expression grammar — precedence, associativity, term and call forms |
| 5 | Statements, declarations, program structure |
| 6 | Incremental parsing and LSP architecture |
| 7 | Conformance and test strategy |
| A1 | Prior art — PerlOnJava and perl-lsp assessed |

Chapters 2-5 specify the language. Chapter 6 specifies the machine that parses
it incrementally. Chapter 7 specifies how the implementer knows it works.

## 1.7 Sources

Three independent implementations are triangulated throughout. Where they
disagree, **perl wins** — it is the definition.

| Source | Language | Role |
|---|---|---|
| `perl5/toke.c` (14,708 lines), `perly.y` (1,897) | C | Ground truth. The lexer and grammar as perl defines them. |
| PerlOnJava (~24,000 lines of parser) | Java | A from-scratch recursive-descent parser that runs real Perl. Proves the approach works. Ships ~1,935 `.t` files. |
| perl-lsp (~36,500 lines of lexer+parser) | Rust | A from-scratch Perl LSP with incremental parsing. The closest analogue to this project's target. |

The last two are approximations of the first. Their divergences are noted where
found — they mark either a deliberate trade-off worth copying or a bug worth
avoiding.
