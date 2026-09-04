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

## 0.5 Perl's test suite needs a correctly built shim

`t/test.pl:119` does `@INC = ()` and then unshifts `../lib`. `PERL5LIB` cannot
help, because `@INC` is cleared *after* it is read. 498 of 620 files use this
convention, so the corpus only compiles inside a tree where `../lib` is
populated.

**Populate it from both library roots.** The pure-perl library alone is not
enough: `Config.pm` and every XS module's `.pm` half live in the
architecture-specific directory, and omitting it costs ~170 files.

```sh
mkdir -p shim/lib && cp -r perl5/t shim/t

perl -e 'for (@INC) { print "$_\n" if -d && !/site_perl|vendor_perl/ }' \
  | while read d; do cp -rn "$d"/* shim/lib/ 2>/dev/null; done

cd shim/t && perl -c op/sub.t          # syntax OK
```

Measured on perl 5.42.0 with that shim:

| Directory | Compiles | Rate |
|---|---:|---:|
| base | 9/9 | 100% |
| comp | 25/25 | 100% |
| cmd | 5/5 | 100% |
| opbasic | 5/5 | 100% |
| lib | 10/10 | 100% |
| mro | 73/73 | 100% |
| uni | 30/30 | 100% |
| io | 44/44 | 100% |
| re | 79/80 | 98.8% |
| op | 223/228 | 97.8% |
| run | 27/28 | 96.4% |
| class | 10/12 | 83.3% |
| perf | 3/5 | 60% |
| porting | 16/37 | 43.2% |
| **Total** | **584/620** | **94.2%** |

### The first measurement was wrong, and the way it was wrong is instructive

An earlier shim copied only the pure-perl library and reported **411/620
(66.3%)**, with `class/` at 8.3% and `io/` at 50%. Those numbers were read as
facts about the corpus — "5.38 syntax this perl rejects", "XS-dependent" — and
they were nothing of the kind. Adding the architecture-specific directory moved
`class/` to 83%, `io/` to 100%, `mro/` to 100%, and the total to 94.2%.

Classifying every failure by its stderr rather than its exit status shows how
little of it was ever about the language:

| Cause | Files |
|---|---:|
| Missing module / `@INC` | 167 |
| Version or feature skew | 1 |
| **Genuine syntax error** | **1** — `op/for-many.t`, §0.6 |
| Other (exit status, timeouts) | 16 |

**One file in 620 fails for a reason about Perl.**

Two consequences for the harness. It must classify stderr: an exit status
cannot distinguish "your parser is wrong" from "this machine lacks
`Config.pm`", and a ratchet built on exit status silently encodes the second as
the first. And a low pass rate is a claim about the harness until proven
otherwise — the corpus is almost entirely compilable, so a number well under
94% means the setup is broken, not that Perl is hard.

## 0.6 The corpus and the oracle must be version-pinned together

The `perl5` checkout here is **blead 5.45** (`patchlevel.h`: `PERL_VERSION 45`);
the installed interpreter is **5.42.0**. Testing a parser against the newer
suite while asking the older interpreter for ground truth produces failures
that belong to neither.

`t/op/for-many.t:474` is a genuine syntax error on 5.42:

```perl
foreach my ( \@array ) ( ["A"], ["B"], ["C"] ) {   # refaliasing in multi-var foreach
```

Multi-var `foreach` alone is fine on 5.42 — `for my ($k,$v) (%h)` runs — so
this is specifically the newer refaliasing form. A parser measured against it
would be marked wrong for agreeing with the interpreter it was checked against.

**Consequence for the harness:** record the interpreter version *and* the
corpus commit in the ratchet header, and treat a mismatch as a reason to
re-baseline rather than as a regression.

## 0.7 The parser, not the type checker, is the latency problem

The intuitive assumption — a type checker walking every node must cost more
than a parse — is wrong here by one to two orders of magnitude. Measured with
`go test -bench` against perl5/lib on 2026-09-04:

| File | Lines | Parse | PSC `Analyze` | Ratio |
|---|---:|---:|---:|---:|
| `charnames.pm` | 484 | 6.5 ms | 0.48 ms | 14× |
| `FileHandle.pm` | 262 | 69.7 ms | 2.8 ms | 25× |
| `overload.pm` | 1701 | 99.6 ms | 2.2 ms | 45× |
| `sigtrap.pm` | 327 | 350.0 ms | 7.0 ms | 50× |
| `_charnames.pm` | 858 | 521.7 ms | 19.3 ms | 27× |

Three things follow.

**The parse cost is not a function of file size.** `sigtrap.pm` at 327 lines
costs 5× what `overload.pm` costs at 1701. Cost tracks grammar pathology, which
means it cannot be predicted, cannot be capped by refusing large files, and
cannot be debounced away — the user is typing in the file that is slow.

**A 350 ms parse is not 3× over an interactive budget, it is 35×.** No amount
of incremental machinery layered on top recovers that; the constant factor is
in the wrong place.

**PSC is already fast enough.** At 0.5-2 ms it fits comfortably. Optimising it
first — the obvious move, and the one an earlier draft of Chapter 6
recommended — would be optimising the cheap half by a factor of 40.

This is the strongest argument in this specification for replacing the parser,
and it is a stopwatch reading rather than a design preference.

`Analyze` does have an incremental defect, just not a latency one: its
annotation map is `map[uint32]types.Type` keyed by `StartByte`
(`infer.go:57-60`), so every key after an edit shifts and nothing survives a
re-parse. That costs cache correctness, and only becomes worth fixing once
re-parsing is cheap enough for reuse to matter.

## 0.8 A citation corrected

`toke.c:10568` is `call_sv` inside `S_new_constant`, which implements
overloaded constants via `$^H`. It is **not** the source-filter mechanism.
Source filters go through `Perl_filter_add` (`toke.c:5164`) and `FILTER_READ`
(`toke.c:5270`). Both are arbitrary compile-time code execution, but they are
different features, and the distinction matters when deciding what to detect
and bail out on.
