<!-- ABOUTME: Measured findings that shape the Perl parser specification. -->
<!-- ABOUTME: Every number here was produced by running perl, not estimated. -->

# 0. Measured Findings

Numbers used elsewhere in this specification, with the command that produced
them. Where a figure would ordinarily be quoted from folklore, it was measured
instead — several times the folklore turned out to be badly wrong.

## 0.1 The undecidability surface is far smaller than it looks

Perl's compile-time execution makes correct static parsing undecidable *in
general*. The question that matters is how often it bites. Measured across
perl5's 620 `.t` files:

| Construct | Files | Share | Assessment |
|---|---:|---:|---|
| Contains `BEGIN` | 487 | 78.5% | **Misleading — see below** |
| `BEGIN` that is `@INC`/`chdir` boilerplate | 466 | 75.2% | Parse-irrelevant |
| `BEGIN` that could affect the parse | 21 | 3.4% | The real figure |
| `use constant` | 27 | 4.4% | Creates `()`-prototyped terms |
| Source filters | 2 | 0.3% | The only hard stop |

The headline number is a trap. "78% of files use `BEGIN`" is true and useless:
almost all of it is the `chdir 't' if -d 't'; require './test.pl'` preamble,
which cannot change how anything parses. The parse-relevant figure is roughly
**1-3%**, depending on how strictly "relevant" is drawn.

This is the finding that decides whether the project is viable. A parser that
is exactly right on 97% of files and *knows which 3% it is unsure about* is a
useful parser. One that is silently wrong on an unknown 78% is not.

## 0.2 Modern Perl deletes an entire class of ambiguity

`S_intuit_method` (`toke.c:5071`) returns 0 immediately when the `indirect`
feature is off:

```c
    if (!FEATURE_INDIRECT_IS_ENABLED)
        return 0;
```

Measured consequence:

```perl
package Foo; sub new { bless {}, shift }
package main;
my $o = new Foo;          # default perl: works, prints "Foo"
```
```perl
use v5.36;
my $o = new Foo;          # syntax error at line 4, near "new Foo"
```

Indirect object syntax — the construct that forces a parser to consult the
symbol table to decide whether a bareword is a method call — is a **syntax
error** under `use v5.36`. In code that opts into a modern feature bundle,
barewords become statically resolvable. The hardest legacy ambiguity is
opt-out, and most new code has already opted out.

## 0.3 Prototypes change the parse, and perl will say so

```
$ perl -MO=Concise,-exec -e 'sub f(\@){} my @a; f(@a)'   # srefgen PRESENT
$ perl -MO=Concise,-exec -e 'sub f{}    my @a; f(@a)'    # srefgen ABSENT
```

The first passes a reference, the second a flattened list, from identical
call-site text. Confirmed to still hold when the prototype is installed by a
string `eval` inside `BEGIN` — the undecidable case — and when installed by a
glob assignment (`BEGIN { *g = sub(\@){} }`).

Definition order matters, and perl's own behaviour on a not-yet-seen sub is the
conservative default. A single-pass Go parser is therefore already
bug-compatible with perl on forward calls, which is a smaller problem than it
first appears.

## 0.4 The current parser cannot see this at all

`internal/parseoracle/fidelity_test.go` records it: the two programs above
produce **identical, error-free trees** in our tree-sitter parser. Coverage
scores both as a success. One of them is parsed wrong.

This is the blind spot the specification exists to close, and it is not
specific to tree-sitter — any parser without a prototype table has it.

## 0.5 Perl's test suite needs a built perl

`t/test.pl:119` does `@INC = ()` and then unshifts `../lib`. `PERL5LIB` cannot
help, because `@INC` is cleared *after* it is read. 498 of 620 files use this
convention, so the corpus only compiles inside a tree where `../lib` is
populated.

With that shim in place, measured on perl 5.42.0:

| Directory | Compiles | Rate | Note |
|---|---:|---:|---|
| base | 9/9 | 100% | |
| comp | 25/25 | 100% | |
| cmd | 5/5 | 100% | |
| opbasic | 5/5 | 100% | |
| lib | 9/10 | 90% | |
| mro | 62/73 | 84.9% | |
| op | 160/228 | 70.2% | |
| uni | 21/30 | 70% | |
| re | 55/80 | 68.8% | |
| run | 18/28 | 64.3% | |
| io | 22/44 | 50% | XS-dependent |
| porting | 10/37 | 27% | Tests perl's source, not the language |
| class | 1/12 | 8.3% | 5.38 syntax this perl rejects |
| **Total** | **411/620** | **66.3%** | |

**This is the ceiling.** A conformance target of "parse all 620 files" would be
measured against files perl itself cannot compile here. Targets must be stated
against the 411, or against a named subset.

`base`, `comp`, `cmd`, and `opbasic` — 44 files at 100% — are the natural first
milestone: perl compiles every one, so any failure is ours.

## 0.6 A citation corrected

`toke.c:10568` is `call_sv` inside `S_new_constant`, which implements
overloaded constants via `$^H`. It is **not** the source-filter mechanism.
Source filters go through `Perl_filter_add` (`toke.c:5164`) and `FILTER_READ`
(`toke.c:5270`). Both are arbitrary compile-time code execution, but they are
different features, and the distinction matters when deciding what to detect
and bail out on.
