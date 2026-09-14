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

Definition order matters, and perl commits at the call site for a not-yet-seen
sub: `f(@a)` is an ordinary unprototyped call, while the paren-less forms are
an indirect method call or a compile error (chapter 3 §3.5.4). A single-pass
Go parser therefore matches perl on parenthesised forward calls and must
diagnose the rest, which is a smaller problem than it first appears.

## 0.4 The current parser cannot see this at all

`internal/parseoracle/fidelity_test.go` records it: the two programs above
produce **identical, error-free trees** in our tree-sitter parser. Coverage
scores both as a success. One of them is parsed wrong.

This is the blind spot the specification exists to close, and it is not
specific to tree-sitter — any parser without a prototype table has it.

### 0.4.1 The grammar also accepts source perl rejects

§0.4 is the coverage-vs-fidelity gap in its familiar direction: we chose a
different parse. There is a second direction, and it is worse — we accept
something that is not Perl at all.

The grammar reports **no error node** for a family of malformed assignments
that `perl -c` rejects outright, and it silently drops the right-hand side:

| Source | `perl -c` | `HasError()` | Our tree |
|---|---|---|---|
| `my $x = ;` | syntax error near `= ;` | `false` | `(_term (variable_declaration ...))` — RHS gone |
| `my @a = ;` | syntax error near `= ;` | `false` | `(_term (variable_declaration ...))` |
| `my %h = ;` | syntax error near `= ;` | `false` | `(_term (variable_declaration ...))` |
| `$x = ;` | syntax error near `= ;` | `false` | `(_term (_variables ...))` |
| `my ($a) = ;` | syntax error near `= ;` | `false` | `(_term (variable_declaration ...))` |
| `1 +;` | syntax error near `+;` | `false` | `(_term (primitive (number)))` — operator gone |
| `my $y = 1 +;` | syntax error near `+;` | `false` | **two sibling `_term`s** — one expression became two statements |
| `f( ;` | syntax error near `( ;` | `true` | error node, correctly declined |
| `return ;` | **syntax OK** | `false` | `(return_expression)` — valid, correctly accepted |

So this is a **family, not one construct**: every binding form and bare binary
operators share it. `f( ;` is the near miss that shows the grammar *can* report
these — it just does not for the assignment family. `return ;` is the trap:
it looks like the family and is legal Perl.

`my $y = 1 +;` is the most damaging row. A consumer reading that tree sees two
unrelated statements where the source wrote one expression, with the `+`
nowhere in the tree and nothing marking the loss.

**Why the harness escaped this only by luck.** `Compare` declined on
`HasError()`, and perl refuses the whole file, so `Facts.OK` was false and
comparison never reached a verdict. Put the construct inside a file that
otherwise compiles and it scored **`exact`** while being parsed wrong —
verified: all seven family members returned `BucketExact` with the detail
"perl took 0 reference(s), all explicit in the source".

**What was done, and what was given up.** The grammar lives in an external
module (`gotreesitter/grammars`, a compiled table), so fixing it at the source
is out of reach from this repo, and reimplementing perl's expression grammar to
second-guess the parser would be writing the parser twice. Neither was
attempted. Instead the limitation is made **detectable**:
`parser.Tree.IsDegenerate()` / `DegenerateKinds()`, and `Compare` now routes a
degenerate tree to `no-answer`.

The signal is structural, not semantic. A node kind beginning with `_` is a
*hidden* tree-sitter rule — `_term` is tree-sitter-perl's expression supertype.
Hidden rules are inlined into their parents in a successful parse and are never
supposed to surface as nodes. When one does, recovery bailed out mid-rule and
kept the fragment. The underscore is therefore an artifact of the *failure*,
not a property of the source, which is what separates these rows from
`return ;`. It is a signal about the *tree*, not about perl's verdict — see the
measured cost below, where two files that compile fine still leak.

What this buys: such a file can no longer score `exact`, and it lands in the
coverage number (`no-answer`) rather than inflating the fidelity one, with the
leaked rule named in the verdict detail for triage.

**Measured cost.** Over a perl5 checkout, 3585 files under 8KB, 3306 of which
parse cleanly: **2 files (0.06%) flag degenerate, and both compile under
`perl -c`** — `cpan/Digest/lib/Digest/base.pm` and
`dist/Tie-File/t/29a_upcopy.t`. The trigger is consecutive `sub NAME;` forward
declarations.

Those two are not noise, and they sharpen what the flag actually means. The
grammar genuinely degrades there too: `sub new;` comes back as a bare
`(bareword)` with the `sub` keyword gone, so the tree is wrong in the same way
the malformed family is wrong. What the flag does *not* mean is "perl rejects
this". It means "this tree is not a faithful parse" — the weaker claim, and the
only one the signal supports.

For the harness that is a small, known coverage loss rather than a wrong
verdict: those files land in `no-answer` instead of being scored. Erring toward
declining is the conservative direction, which is what makes the gate worth
having at this precision. `TestForwardDeclarationsAlsoLeak` pins it so the cost
cannot be quietly forgotten.

What this gives up, stated plainly:

- **We still do not reject the source.** `IsDegenerate()` says "this tree is
  not trustworthy", not "this is not Perl". A linter wanting a syntax error
  still has nothing to report.
- **It is a proxy, and not a precise one.** It detects the grammar giving up
  mid-rule, not malformed Perl. It over-fires on valid forward declarations
  (measured above) and stays silent on any malformed construct the grammar
  recovers from *without* leaking a hidden rule. This closes the measured
  family and makes no claim beyond it.
- **It is coupled to a grammar-internal naming convention.** If a future
  version renames `_term` or stops surfacing hidden rules on recovery, the
  signal degrades silently. `TestDroppedRHSFamilyHasNoErrorNode` pins the
  status quo so a grammar that starts emitting real error nodes fails loudly
  and points at the compensation to delete.
- **It does not survive the rewrite this specification proposes.** A hidden
  rule leaking is a property of *tree-sitter's* recovery, and a hand-written
  Go parser has no hidden rules to leak. If the parser described in chapters
  2-6 is built, `IsDegenerate()` becomes meaningless and must be replaced by
  that parser's own account of where it gave up.

  This is worth stating plainly because the harness is meant to be the
  rewrite's acceptance test. Most of it transfers: the oracle, the four
  buckets, the corpus, the pin, the ratchet all measure perl against *any*
  parser. This one signal does not — it is scaffolding against the parser we
  ship today, and it should be budgeted as such rather than mistaken for part
  of the permanent apparatus.

**The honest summary**: the grammar has two separate defects here — it accepts
the malformed assignment family, and it mis-parses forward declarations. One
detector catches both because both are the same underlying event: recovery
bailing out and leaking an internal rule name into the tree. Neither defect is
fixed; both are now visible.

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
| **Genuine syntax error** | **≥ 2** — see below |
| Other (exit status, timeouts) | 16 |

**Roughly two files in 620 fail for a reason about Perl** — and that is a lower
bound, for a reason worth stating.

`perl -c` reports **one** error and stops. A classifier reading its stderr
therefore sees only the *first* failure in each file, and a cheap early failure
masks whatever is behind it. `t/op/signatures.t` fails at line 32 with
`Unknown warnings category 'experimental::signature_named_parameters'` — feature
skew. Delete that pragma and it fails again at line 927:

```perl
sub tnamed01 (:$alpha, :$beta) { ... }   # named parameters in signatures
```
```
A signature parameter must start with '$', '@' or '%' ... near "(:"
```

Blead-only syntax, and a genuine parse failure on 5.42. It was filed as feature
skew only because the pragma error came first. `op/for-many.t` (§0.7) is the
other.

**The method matters more than the count.** One-error-per-file makes this a
lower bound on syntax errors and an upper bound on environmental ones; seeing
past the first failure means fixing it and re-running. The conclusion — that
these failures are overwhelmingly environmental rather than linguistic —
survives. The exact figure does not.

Two consequences for the harness. It must classify stderr: an exit status
cannot distinguish "your parser is wrong" from "this machine lacks
`Config.pm`", and a ratchet built on exit status silently encodes the second as
the first. And a low pass rate is a claim about the harness until proven
otherwise — the corpus is almost entirely compilable, so a number well under
94% means the setup is broken, not that Perl is hard.

## 0.6 The starting line, measured

The 44 files in `t/base`, `t/comp`, `t/cmd` and `t/opbasic` all compile under
perl, so any failure on them is ours. Against the parser shipping today:

**29 clean, 15 with error nodes — 65.9%.**

```
base/lex.t              comp/hints.t          comp/require.t
base/num.t              comp/package.t        comp/uproto.t
comp/decl.t             comp/parser.t         cmd/switch.t
comp/final_line_num.t   comp/parser_run.t     opbasic/arith.t
comp/form_scope.t       comp/proto.t          opbasic/qq.t
```

This is the number the conformance plan's M1 milestone
(`docs/plans/2026-09-05-parser-conformance-plan.md` §5) has to move, and it is worth reading the list rather than
the percentage. `comp/proto.t` and `comp/uproto.t` are the prototype files;
`base/lex.t` and `comp/parser.t` exist specifically to abuse the lexer;
`cmd/switch.t` is the heredoc-as-call-argument case already known from the
tree-sitter work. The failures cluster exactly where Chapters 2 and 3 say the
difficulty is, which is weak evidence that those chapters describe the real
problem rather than an imagined one.

Note also what this measures: **coverage**, the weak metric. All 29 "clean"
files are clean only in the sense of having no error node. How many are parsed
*correctly* is unmeasured, and §0.4 shows the parser cannot currently
distinguish cases perl distinguishes. The true starting number is therefore at
most 29, probably less.

§0.11 measures those 29 against perl for the first time. None of them is
WRONG on the one marker the harness currently tests — but only 7 of the 44
files exercise that marker at all, so 29 remains a ceiling rather than a
result.

## 0.7 The corpus and the oracle must be version-pinned together

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

## 0.8 The parser, not the type checker, is the latency problem

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

## 0.9 The lexer decides things the grammar cannot express

`toke.c`'s `S_lop` returns `FUNC` when the next character is `(` and `LSTOP`
otherwise. That single lookahead changes how much of the line a list operator
swallows, and no precedence table can represent it:

```
$ perl -e 'print (1+2)*3'     # prints 3
$ perl -e 'print 1+2*3'       # prints 7
```

Deparsed, the first is `(print(3) * 3)` — the parenthesis makes `(1+2)` the
*complete* argument list, and the multiplication applies to `print`'s return
value. The second is `print(7)`.

This is the general shape of the problem in Chapter 3: a Go implementation
needs its parser and lexer coupled, because the token a construct produces
depends on parse state and on lookahead simultaneously. A clean lexer/parser
split — the design every textbook recommends — cannot parse Perl, and both
reference implementations that tried it ended up re-scanning characters from
inside the parser to compensate.

## 0.10 A citation corrected

`toke.c:10568` is `call_sv` inside `S_new_constant`, which implements
overloaded constants via `$^H`. It is **not** the source-filter mechanism.
Source filters go through `Perl_filter_add` (`toke.c:5164`) and `FILTER_READ`
(the macro at `perl.h:4447`, dispatching to `Perl_filter_read` at
`toke.c:5268`). Both are arbitrary compile-time code execution, but they are
different features, and the distinction matters when deciding what to detect
and bail out on.

## 0.12 The sweep's cost is our parser, not perl

*2026-09-09.* The corpus sweep takes ~6 minutes, and the caching issue was
filed on the assumption that `B::Concise` dominated it — ~600ms/file,
extrapolated from `t/op/sub.t`. Building the cache disproved that.

| Phase | Cold | Warm |
|---|---:|---:|
| Oracle (perl) | ~33-39 s | **0.82 s** |
| Our tree-sitter parser | 5m28s | 5m28s |
| Whole sweep | ~6m50s | 6m18s |

The cache is a **40-48x** speedup on the oracle and an **8%** speedup on the
sweep, because roughly 87% of the sweep is our own parser. Measured directly
here at **412ms/file** over 40 `t/op` files with no perl in the loop at all.

**The cost that matters cannot be cached.** The oracle's answers are stable —
perl does not change between runs, which is what makes a content-hash cache
correct. Our parser's output is the thing under test; caching it would cache
the measurement.

Two consequences:

- Making the sweep fast enough to gate a commit is a **parser-performance**
  problem, not a harness problem. §0.8 already measured the same thing from
  another direction: a single parse costs 6-500 ms and tracks grammar
  pathology rather than file size.
- The caching issue's third acceptance criterion — "finishes fast enough to
  gate a commit" — is **not met and cannot be met by caching**. It was
  reported as failed rather than quietly redefined, which is the right
  outcome: 6m18s warm is still 6m18s.

The cache earns its place anyway. A re-baseline, a comparison change, or any
iteration on the *parser* side now costs 0.82s of perl instead of 35s, and the
oracle stops being a reason not to re-run.

## 0.11 Fidelity across the whole corpus, measured

*2026-09-13. Corpus `perl5` at `94e5086608`, interpreter 5.42.0 (the pin in
`internal/parseoracle/testdata/corpus.pin`).*

```
PARSEORACLE_SHIM=<shim> PERL5_CORPUS=~/dev/perl5 PARSEORACLE_TIMEOUT=8m \
  go test ./internal/parseoracle/ -run TestRatchetCorpus -parseoracle.corpus -parseoracle.update -v
```

All 620 `.t` files, each compiled by perl and by our parser and the two
bucketed against each other:

| Bucket | Files | Share of measured |
|---|---:|---:|
| exact | 405 | 68.3% |
| wider | 12 | 2.0% |
| **WRONG** | **2** | **0.3%** |
| no-answer | 174 | 29.3% |
| *measured (denominator)* | *593* | |
| excluded, environmental | 26 | — |
| runner error | 1 | — |

One run, 16m30s across 24 workers on a loaded machine. The previous
baseline's two identical runs took ~6 minutes each; the difference is load
and a second `perl -c` per file, not a change in what is measured.

**These numbers were 322 exact, 54.3%, 11 wider, 0 WRONG and 260 no-answer,
and before that 379 exact, 63.9%.** The history is: `run/switcht.t` moved
into `exact` when `wantsTaint` stopped freezing perl's refusal of a
`#!./perl -t` shebang as a parse result (379 → 380); 58 files then moved out
of `exact` when two verdict-correctness defects were fixed (→ 322); and 59
files moved again when the oracle's population was widened and the
comparison made per-statement (→ 373). Each step is a measurement getting
honest, in one direction or the other.

**The oracle was measuring a narrower population than the subject.**
`-MO=Concise,-exec` dumps `PL_main_root` only. A named sub, an anonymous sub
and a `BEGIN` block are each a separate CV and never appear in it, so a
prototype-forced reference inside any sub body — which is where most of the
corpus keeps its calls — was invisible, while the subject counted every
backslash it could see. `parse_facts.pl` now walks every CV with `B`
directly (`main_root`, the special-block arrays with `B::save_BEGINs` asked
for first, every named sub in every package, each anonymous sub through the
`anoncode` op that closes over it) and reports each reference with the line
of the statement it belongs to, taken from the nearest `nextstate`. Only
COPs naming this file count, so the subs a test pulls in from `t/test.pl`
are not held against it. The comparison then decides one statement at a
time: a reference perl took is explained only by a hedge in the same
statement, and a subject backslash perl folded away (`\1`, `\"x"`) can mask
nothing beyond its own statement, so it no longer makes a file unanswerable.

The 59 files that moved, with why:

- **51 files, `no-answer` → `exact`** — the surplus group, less one. Under
  whole-file totals a subject backslash that perl represented as something
  other than `srefgen` (a folded literal, a list-form `refgen`) made the
  totals incomparable and the file unanswerable. Per statement it masks
  nothing, and every one of these 51 agrees at every statement perl took a
  reference in.
- **1 file, `no-answer` → `wider`** — the 52nd surplus file, `op/concat2.t`:
  two references inside sub bodies the old oracle never saw, both hedged.
- **2 files, `exact` → `wider`** — `re/qr-72922.t` (`sub s1`, line 30,
  `Internals::SvREFCNT($$re_weak_copy)` under the prototype `\[$@%&*]`; the
  subject hedges the qualified call) and `op/or.t` (line 17, `return bless
  \$instance => $class` inside a sub; our grammar reads `\$instance =>` as
  an autoquoted bareword and drops the backslash, then hedges `bless`, so
  the verdict is a hedge over a tree that lost source). Both were `exact`
  only because the references sat inside a sub.
- **2 files, `wider` → `exact`** — `op/getpid.t` and `uni/goto.t`. Their
  `wider` came from `srefgen` ops that are not references at call sites:
  `my $pid2 : shared` is rewritten by op.c's `apply_attrs_my` into
  `attributes->import(PKG, \$pid2, 'shared')`, and `goto &NAME` is a `goto`
  whose operand perly.y parsed as an entersub term and `newLOOPEX` wrapped in
  a `REFGEN`. Neither has a backslash the source could have written, and the
  population now excludes both shapes by recognising the ops that make them.
- **3 files, `wider` → WRONG** — below.

**Two files the review measured did not move, and the reason is the same
exclusion.** `op/die_goto.t` (5 → 7 under `-stash=main`) and `mro/isarev.t`
(3 → 4) each carry their extra `srefgen` under a `goto &foo` inside a sub.
Those are `goto`'s rewrite, not references the parse took, and counting
them would score every `goto &sub` in the corpus WRONG for a parse both
sides agree on. They stay `exact`, with the two excluded ops named here
rather than counted there.

**WRONG is three, and all three are our grammar silently shredding a
statement.** Perl took six, five and four references in regions where the
subject's tree reports none, hedges none, and says nothing. That is a tree
that stopped being a parse of its source, exactly what the degenerate gate
in `adapter.go` exists to decline, and a shape it does not yet detect.
Whole-file hedging was hiding all three as `wider`; per-statement scoring is
what surfaced them. The bucket is doing its job.

Bisected, the cause is **two** grammar gaps, not one, and neither is the
`q~ ... ~` first suspected — that construct parses correctly:

- `mro/package_aliases.t:192` and its utf8 twin: `~ =~ s\__code__\$$_{code}\r,`
  — a substitution delimited by **backslash**. `toke.c`'s `Perl_scan_str`
  takes the next non-whitespace character as the delimiter with no
  allow-list, and guards the escape case with `close_delim_code != '\\'`;
  our scanner checked the escape first, so the closing delimiter was eaten
  and the scan ran to EOF. One line derails the remaining 200.
- `op/filetest.t:96`: `is(-s -f $ro_empty_file, 0, ...)` — **stacked file
  tests inside a call's argument list**. `-s -f $file` parses correctly on
  its own; inside an argument list the statement is shredded into loose
  fragments under `source_file` with `HasError()` false. The root cause is
  in the GLR runtime rather than the Perl grammar: a zero-width external
  token marking the call rule was discarded in favour of a same-byte DFA
  token, so `-s` was lexed as unary minus and `s` as the substitution
  operator.

Both are fixed in a fork of the grammar
(`perigrin/gotreesitter`, pinned by commit), and wiring it in moved the
measurement again: **373 → 411 exact (62.9% → 69.3%)**, no-answer 208 → 170,
and all three of those WRONG verdicts cleared. `package_aliases.t` now parses
with no error at all; the other two recovered their references and are left
reporting honest errors for two *separate* pre-existing gaps the shred had
been hiding — the grammar does not accept Unicode identifiers
(`\%Ｏｒｇａｎ::`, `sub Hyᚹ::ｳ {}`), and `op/filetest.t:215`'s `split //`
errors on the pristine runtime too. Both were proven independent of these
fixes.

**The fork briefly created a fourth WRONG, and finding it was the point of
checking.** `op/splice.t` moved `wider → WRONG` on the first post-fork sweep,
because the GLR fix makes every parenthesised call a
`function_call_expression` where a bare `f(...)` used to be
`ambiguous_function_call_expression`. Line 104's
`Internals::SvREADONLY(@readonly_array, 1)` is a prototyped builtin whose
declaration a static parser has never seen: perl takes a reference, we report
a plain call, and reading the new node kind as a *commitment* blamed the
parser for a prototype it cannot know. That is the `wider` bucket's entire
reason for existing.

The predicate now asks the source rather than the node kind — only `&f(@a)`,
which perl documents as bypassing the prototype (perlsub, "Prototypes"),
counts as settled. `isHedgedCall` and `sourceSettledTheCall` in `compare.go`.
Re-swept, that moved exactly one file, `op/splice.t` back to `wider`, and
left every other verdict untouched, taking WRONG to 0 for the `srefgen`-only
metric.

This is the shape of hazard a grammar change carries. The fix was correct and
the measurement improved, but a node kind the harness read as "the parser is
sure" silently changed meaning underneath it, and nothing but a re-sweep and
a deliberate check would have caught it. A verdict of WRONG is the only one
that fails a build, so a false one is the most expensive defect this harness
can have.

**Adding the four markers raised WRONG to 2, and both are real.** Neither is a
marker defect; both are grammar defects that a reference-only metric could not
see, which is the case for the other four markers made concrete:

- `op/universal.t` — perl builds an anonymous hash at lines 16, 30, 104, 151,
  184, 195 and 229. The subject reports six of seven, missing line 16,
  `$a = {};`. That statement is not unusual and the adapter is not at fault:
  `$a = {};` alone, after a `plan` call, after a `BEGIN` block, and in a copy
  of this file's own first twenty lines all yield an anonhash site. Only the
  full-file parse loses it, and there no site of any kind is reported for
  lines 14–18 — the statement is gone from the tree with `HasError()` false.
  That is the silent-shred class the stacked-filetest bug belonged to.
- `comp/our.t` — perl matches at lines 32 and 33, the bare `/TIE/` and
  `/calls/` inside `for ($AUTOLOAD =~ /TieAll::(.*)/)`. The subject does
  report match sites covering both lines, but they belong to the enclosing
  `for` and `if` statements and are spent against perl's match for the `for`
  list itself; the innermost statement owning line 32 has none of its own.
  Whether WRONG is the right verdict here is genuinely open — the subject did
  see matches at those lines, so "committed with no site" overstates the
  error. Recorded rather than settled, in `marker_gaps_test.go`.

**Twelve `wider` files are every prototype-driven reference the subject
hedged**, and WRONG being small is worth less than it sounds: the comparison
tests one marker, `srefgen`, and an honest refusal buckets `wider`. Two
things changed here. A hedge now has to sit in the statement it explains, so
`srefgen=5, hedged=1, committed=4` no longer scores `wider` on the strength
of one hedge somewhere in the file. And a hedge is now recognised from the
source rather than the node kind, because the grammar fork stopped spelling
"unresolved" as a distinct kind.

**`exact` is weaker than "parses like perl", and the gap is now much
smaller.** The headline rate is the wrong number to read: what matters is how
many `exact` verdicts had any parse decision in play at all, because a file
where perl made none is a file the two sides agree about nothing on.

| | one marker (`srefgen`) | five markers |
|---|---:|---:|
| exact | 411 | 405 |
| …with a marker in play | **79** | **231** |
| …with nothing measured | 332 | 174 |
| **verified surface** | **13%** | **39%** |

The verified surface nearly tripled. A parser reading `%h` as modulus, `/x/`
as division, `<FH>` as a glob and `{}` as a block used to score the same as
one that got them right; now those four decisions are checked wherever perl
makes them, one statement at a time, by the same rule `srefgen` always used.

On the 44 files of §0.6 the same shift shows in miniature: 30 exact, of which
**22 now have a marker in play** where 5 did before.

The remaining 174 markerless exacts are not a defect — they are files where
perl genuinely made none of the five decisions. Closing that gap further means
more markers (`method_named` for `->m` resolution, `leaveloop`/`scope` for
block-vs-hashref beyond the empty case), not a different rule.

**`no-answer` is 174 files, and they are not all the same thing.** 153 carry
a category, which means our parser produced an error node and the taxonomy
named the construct — that is our coverage gap, the population the
conformance plan's M1 has to move, and the honest reading of "how much of
Perl we cannot parse".

The remaining 21 carry no category, and they are a mixture rather than one
kind: 6 are files perl never finished parsing (`skip_all` inside `BEGIN`,
`ok:1 op_count:0` — `lib/cygwin.t`, `op/refstack.t`, `uni/greek.t`,
`uni/latin2.t`, `win32/signal.t`, `win32/system.t`); 1 is a file perl itself
declined (`class/inherit.t`); and 3 declined for a surplus our side could not
make comparable (`io/open.t` at 11 references, `re/reg_mesg.t` at 9,
`op/stat.t` at 3).

**Of the last 7, five carry an error node the taxonomy did not name** —
`base/lex.t`, `japh/abigail.t`, `op/for-many.t`, `op/mkdir.t` and
`re/bigfuzzy_not_utf8.t` all parse with `HasError()` true; only `run/dtrace.t`
and `win32/popen.t` parse cleanly. So an empty category did not mean an
unproblematic file: those five belong with the 153 as coverage gaps and were
missing from that count because `CategoriseSource` returned nothing for them.
Measured directly rather than inferred from the column, after an earlier
draft of this paragraph assumed the opposite. §0.13 measured the shape — the
trees have `HasError()` with no `ERROR` node because recovery halted — and
closed the rule gap: every file with an error now gets a site, and the
column at the next sweep will name them. The effect on the headline is nil,
since all 21 are already outside `exact`.

The coverage gap fell by 38 files when the grammar fork landed — files our
parser could not read before and can now — which is the one number here that
is about the parser rather than about the harness.

**What the walk still does not see, each with its mechanism.** Three files
retag their own COPs with `#line` (`base/lex.t`, `op/warn.t`,
`test_pl/examples.t`): ops after the directive name another file and are
dropped, which is why their metric fell to 0. `ADJUST` blocks and `field`
initialisers under `feature 'class'` are CVs held outside the stash, and
`format` bodies are FORM slots; none is walked. A string `eval` is not
compiled at `-c` time. And the statement is the finest attribution perl
offers, so a folded `\1` in the *same* statement as a prototype-driven
reference the subject missed spends itself on that reference. Measured
with the rule's own grouping over the whole corpus, **12 statements in 7
files** (`class/destruct.t`, `op/crypt.t`, `op/leaky-magic.t`, `op/not.t`,
`op/stat_errors.t`, `op/universal.t`, `test_pl/can_isa_ok.t`) hold an
unspent backslash beside a call, out of 9,386 statements carrying any site.
Those 12 are the whole surface on which a prototype-driven reference could
still hide; closing it means the adapter not reporting a reference for a
literal operand, and it is not closed here.

**Two corpus facts stay separate from the parser's score.** 26 files fail
because this machine's environment cannot satisfy them — a missing module is
not a parse error — and they leave the denominator by way of `Classify`. One
file, `t/re/pat_psycho.t`, is a runner error: it calls `test.pl`'s
`watchdog(5 * 60)` from a `BEGIN` block, so `perl -c` forks a monitor that
inherits the oracle's pipe. It costs its 60-second timeout and reaches no
bucket.

**Cost.** The plan's §2 estimated ~25 ms per file for `B::Concise`; measured
on real corpus files it is **~600 ms** (`op/sub.t`, 1055 ops), and ten files
cost 9.3 s sequentially of which the prototype probe is only ~10%.

**That is not where six minutes goes**, as this paragraph previously
concluded. §0.12 measured the sweep directly: ~87% of wall time is our own
parser, not `B::Concise`. The oracle is the cheaper half. The content-hash
cache §2 describes therefore buys less than the estimate implied — it is
still worth having (`01a095cb`), but it is a speedup on the smaller term.

## 0.13 What our parser cannot parse, by construct

*2026-09-14. Measured by parsing every `no-answer` file of the §0.11 baseline
directly through `internal/parser` and reading the line where recovery gave
up; each construct below was then reduced to a minimal source and that source
verified to fail on its own. This is the implementation order for a
hand-written parser: the lexer, not the grammar, owns the top of the list.*

**The taxonomy column could not say what was in 86 of its 153 categorised
files, and said nothing at all about 8 more that carry an error.** The 86
were `General`; the 8 had `HasError()` true and no `ERROR` node, so
`CategoriseSource` returned nothing. After this measurement the split of the
174 `no-answer` files is:

```
Identifier 37   Regex 27   QuoteLike 21   Subroutine 17   Operator 17
ControlFlow 17  General 19  ModernFeature 5  Dereference 1  none 13
```

Measured with the rules in `taxonomy.go` at this commit against the §0.11
baseline's rows; the baseline's own category column is regenerated only by a
corpus sweep, so it lags this table until the next `-parseoracle.update`.

The 13 with no category are the files perl never finished, plus
`class/inherit.t`, `run/dtrace.t`, `win32/popen.t` and the four `run/switch*.t`:
none of them carries an error node, verified directly. An empty category now
means what it says.

### The ranked list

Counts are files in the corpus whose first error is this construct. Where a
construct was already partly claimed by another category, the total is given
and the previous filing noted.

1. **Non-ASCII identifiers under `use utf8` — 34 files.** Every `t/mro/*_utf8.t`
   and most of `t/uni/`. `use utf8; package Føø::Bær;` fails; so do
   `@ᕘ::ISA = 'x'`, `*ᕘ::ᕘ_Ƒ운ℭ = sub {}` and `$Àlìcè::VERSION = 1`. The
   declaration `sub ᕘ { 1 }` alone parses, but `(shift)->SUPER::ᕘ` does not,
   so the failure is in the identifier character class at every position, not
   only after `package`. Six of the 34 were previously filed as QuoteLike or
   Subroutine because the same line also held a `qw//` or a `::`.

2. **The repetition operator `x` glued to its left operand — 13 files.**
   `my @a = ((1)x3, 2);` fails; `(1) x 3` parses. The lexer reads `x3` as an
   identifier. The quoted forms `'x'x8` and `"\x{ffff}"x3` and the call form
   `chr (0xdf)x4` fail the same way. In its simplest shape,
   `my @a = (1)x3;`, the tree is *degenerate with no error node*: the
   grammar silently drops the repetition.

3. **`format` bodies — 8 files.** The empty format `format STDERR =\n.\n`
   fails outright; a picture line `@ @<<` followed by its argument line
   fails; a format inside a sub body takes the whole sub with it. Two of the
   eight (`op/write.t`, `comp/form_scope.t`) parse to a degenerate tree with
   no error node at all. The body is line-oriented and ends at a lone `.`,
   which is the same shape as a heredoc and wants the same lexer mode.

4. **Bodiless forward declarations `sub NAME;` — 7 files.** `sub bar;` alone
   at the top of a file parses. After a statement it does not: `f(1);\nsub
   bar;` produces an error node on the declaration, and `1;\nsub bar;` is
   silently degenerate with the `sub` keyword dropped (§0.4.1 measured that
   leak on cleanly-parsing files too). The prototype form
   `sub Hash::Util::bucket_ratio (\%);` inside a sub body breaks the
   enclosing sub (`op/hash.t`).

5. **Regex modifiers on the closing-delimiter line — 6 files.** A brace- or
   bang-delimited body whose last line is only `}ge;`, `}x;` or `!x;`:
   `$x =~ s{\n a\n}{\n b\n}ge;` fails. The bang form is worse:
   `$x =~ s!a!b!x;` is degenerate with no error node, while
   `$x =~ s!a!b!g;` parses — the `x` is being read as the repetition
   operator, which is construct 2 again from the other side.

6. **Post-5.36 keywords used as legacy identifiers — 5 files.** A sub named
   `try` and then called (`sub try { 1 }\ntry(1, 2);` — `run/runenv.t`),
   a sub named `defer` (`op/multideref.t`), `true()`/`false()` called as
   functions (`perf/opcount.t`), `method Pack (...)` as an indirect-object
   call (`op/method.t`; `new Pack (...)` parses), and a forward-declared
   `method forwarded;` inside a `class` (`class/method.t`). The grammar
   reserves these words unconditionally; perl reserves them per feature.

7. **A bareword call followed by `&&` — 3 files.** `my $x = foo && 1;` fails;
   `foo || 1` and `foo and 1` parse. The `&` after a bareword is read as the
   code sigil. `if (is_miniperl && !eval ...)` in `io/open.t` and
   `op/mkdir.t` is this, and both were among the 8 uncategorised: their
   trees have no `ERROR` node because recovery *halted* — the `source_file`
   node ends before the source does.

8. **`given`/`when` — 3 files.** `given ($x) { when (1) { } }` and the
   `CORE::given(1) {` form. Deprecated, but the corpus tests it.

9. **Numeric literal forms — 3 files.** The hexadecimal float `0x0p0`
   (`op/hexfp.t`, an error node) and `0x0.b17217f7d1cf78p0` (`op/sprintf2.t`);
   the underscore directly after the radix prefix, `0x_1234` (`op/oct.t`),
   which parses to a degenerate tree.

10. **The `'` package separator — 2 files.** `$main'a = 1;` and
    `sub CORE'print'foo { 43 }`.

11. **Declared references — 2 files.** `our \$T = \$::T;` and `my \$x = \$y;`
    (`re/opt.t`); `foreach my (\@a) ([1]) { }` (`op/for-many.t`, filed under
    ControlFlow by its `foreach`).

Single files, each verified: the string bitwise operator `22 &. 66`
(`op/bop.t`; `^.` fails the same way); a lexical `our sub foo { 42 }` inside
a block (`op/lexsub.t`); a heredoc whose tag contains spaces, `<<"        --"`,
in list context (`re/pat_advanced.t`); a filetest followed by a pre-increment,
`-f ++$d` (`japh/abigail.t`; `-f $d` parses); the label form `sub f { foo: }`,
which yields a MISSING node rather than an error (`op/attrs.t`); an
argument-less `eval` in `is(eval, 33, ...)` (`op/eval.t`) and `require;` after
a statement (`op/override.t`, degenerate); `print $a < $b` (`base/num.t`,
where `<` after `print $scalar` opens a glob — `print $a > $b` parses); a
subscript of the punctuation array, `$x = $#[0];` (`base/lex.t`, a halted
parse); the bare word `END` as a statement inside a sub (`uni/class.t`);
`print $? & 0xFF ? ... : ...` (`op/magic.t`, degenerate, and `print $? & 1
? ...` parses — not reduced further).

### What could not be minimised, and why

Of the 19 files still `General`, thirteen are constructs from the list above
that have no rule, because a rule for one or three files would describe
those files rather than a construct: the numeric literals, the declared
reference, the label, argument-less `eval` and `require`, `$#[0]`, `END`,
`print $a <`, `print $? & 0xFF`, the orphaned `{` before a class block, and
`op/hash.t`, whose prototype forward declaration sits inside a sub and puts
the site on the enclosing `sub validate_hash {`. The other six are below.

Five files stay `General` because the construct is not at the site.
`io/pipe.t`, `op/numconvert.t` and `op/split.t` all put the first `ERROR`
node on a `split //` line, and every prefix of each file that ends at or
after that line parses cleanly: the trigger is later in the file, recovery
placed the node early, and no window under 140 lines reproduces it.
`op/stat.t` halts at line 492 and the smallest balanced failing window is
lines 336–500. `comp/hints.t` fails at a `BEGIN {` on line 264 that no
window of its neighbours reproduces. `re/bigfuzzy_not_utf8.t` fails on one
line of fuzzer bytes inside a string literal; a string of raw `\xff` bytes
parses, so the byte that breaks it is not known, and the Identifier rule
currently claims the line on the non-ASCII runes around it — a known false
positive, one file. `comp/final_line_num.t` ends in `print 1+` on purpose;
perl rejects it too, and it belongs outside the denominator rather than in
any category.

### What the measurement says about the parser, beyond the list

**The first error node is not always where the parse broke.** Three files
above carry their `ERROR` node eighty lines before the construct that
caused it. Any tool that reads the site from the tree, including the
taxonomy, inherits that placement.

**A tree can carry an error and no error node, three ways.** A MISSING
token (`sub f { foo: }`); a halted parse, where `source_file` ends early and
nothing after it is in the tree at all (`$x = $#[0];` alone does this); and
a degenerate tree, where a hidden rule leaks and no error is recorded. The
taxonomy now reads a site from each. The degenerate site is one statement
early: `print 1;\nsub bar;` leaks its hidden node onto `print 1;`, so the
construct is on the line after the site.

**Silent acceptance is the larger half of several rows.** `(1)x3`,
`0x_1234`, `s!a!b!x`, `true()`, `require;` after a statement and
`1;\nsub bar;` all come back with `HasError()` false. A parser that scored
these as clean would be wrong five ways before it reached the error nodes.
The degenerate detector (§0.4.1) is what makes them visible, and it is why
they are in `no-answer` rather than `exact`.
