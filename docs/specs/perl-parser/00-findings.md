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
  family and makes no claim beyond it. §0.4.2 is that caveat cashing out: a
  construct the grammar mis-parses confidently, leaking nothing, found only
  when it produced a false `WRONG` in the sweep.
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

### 0.4.2 A deref block can swallow the reference inside it

§0.4.1 closes with a caveat: the hidden-rule signal "stays silent on any
malformed construct the grammar recovers from *without* leaking a hidden rule".
This is that case, found in the field rather than in theory. It arrived as a
**false `WRONG`** from the fidelity harness, which is the expensive kind of
defect — `WRONG` is the only bucket that fails a build, so a false positive
there makes the gate untrustworthy and an untrustworthy gate gets turned off.

`my @a = (1,2); my @b = @{ \@a };` compiles under `perl -c` and its optree
holds one `srefgen`. Our tree contains no reference at all, so the comparison
concluded we had committed to a different parse. We had not — we had failed to
parse it, and said so nowhere.

It is a third silent-loss family, and it is *not* a dropped token. The source
is all still there; the structure is wrong:

```
@{ \@a }  ->  (array (varname))              varname spans the text "\@a"
@{ $r }   ->  (array (varname (block ...)))  the block a correct parse builds
@a        ->  (array (varname))              varname spans the text "a"
```

The braces became *anonymous* tokens of the `array` node and the `block` was
never built, taking the refgen inside it with it. `varname` is defined to hold
an identifier, so a `varname` reading `\@a` is a node contradicting its own
rule — the same class of claim `_term` makes, reached by a different route.

**The family, measured.** Every row verified against perl 5.42.0 with
`-MO=Concise,-exec`:

Verdicts are the harness's, measured on the construct in isolation with perl's
real `srefgen` count — see "Measured cost" below for what happens when it sits
inside a corpus file that is already failing for another reason.

| Source | `perl -c` | `srefgen` | Our tree | Verdict before |
|---|---|---|---|---|
| `@{ \@a }` | syntax OK | 1 | `varname` = `\@a`, block lost | **WRONG** |
| `%{ \%h }` | syntax OK | 1 | `varname` = `\%h`, block lost | **WRONG** |
| `${ \$x }` | syntax OK | 1 | `varname` = `\$x`, block lost | **WRONG** |
| `&{ \&f }` | syntax OK | 1 | `varname` = `\&f`, block lost | **WRONG** |
| `@{ $r }` | syntax OK | 0 | `varname` → `block` — correct | exact |
| `@$r` | syntax OK | 0 | `array` → `varname` → `scalar` — correct | exact |
| `$r->@*` | syntax OK | 0 | `array_deref_expression` — correct | exact |
| `$$r[0]` | syntax OK | 0 | `array_element_expression` — correct | exact |

Neither existing signal sees the broken rows: `HasError()` is `false` and
`DegenerateKinds()` was empty. That is what separates this from §0.4.1 — there
the grammar gave up and left a fingerprint, here it confidently built the wrong
node.

**The defect is far narrower than "deref".** The grammar handles a `\` inside a
deref block correctly in every neighbouring shape, and the near misses differ
by a single character:

| Source | Our tree |
|---|---|
| `@{ \@a }` | collapses |
| `@{ \@a, }` | correct block, refgen intact |
| `@{ +\@a }` | correct block, refgen intact |
| `@{ \ @a }` | correct block, refgen intact |
| `@{ \\$r }` | correct block |
| `"@{[ \@a ]}"` | correct |

So the trigger is precisely: a lone backslash tight against a sigilled
variable, alone in the braces. Anything else in the block and the correct rule
wins.

**What was done, and what was given up.** Three options were on the table.

1. *Count it structurally in the comparison.* Viable — unlike the report's
   assumption, the tree does retain a distinguishing signal, and a precise one.
   **Rejected anyway.** It would score these files `exact`, asserting a parse we
   did not make. The tree says `@a` is an array *named* `\@a`; treating that as
   a reference we understood would launder a parse failure into the fidelity
   number, and every downstream consumer of that tree is still being lied to.
2. *Extend degeneracy detection.* **Chosen.** `varname` gains the same
   treatment `_term` has: a leaf `varname` whose text opens with `\` marks the
   tree untrustworthy, and the verdict becomes `no-answer`. A measuring
   instrument that declines to answer is behaving correctly; one that answers
   wrongly is not.
3. *Fix the grammar.* Out of reach for the same reason as §0.4.1 — a compiled
   table in `gotreesitter/grammars`, external to this repo. It remains the only
   real fix; this is compensation.

**The first attempt over-fired, and the corpus caught it.** "A leaf `varname`
whose text opens with `\`" is the obvious rule and it is wrong: `$\` is the
output record separator, and the grammar parses it *correctly* into a `varname`
whose text is a lone backslash. The sweep moved two files —
`t/op/tiehandle.t` and `t/uni/lex_utf8.t` — from `exact` to `no-answer` for no
reason at all. Requiring a **sigil after the backslash** separates the
punctuation variable from the collapse, and those rows are now pinned in
`TestCollapsedDerefDetectorIsNarrow`.

That is worth recording for what it says about the method rather than the bug.
The unit tests were green; only running the thing against 620 real files
surfaced it. A narrowness test is only as good as the shapes somebody thought
to put in it, and the corpus thinks of more.

**Measured cost: zero files, and the zero is worth reading carefully.** With
the corrected detector the 620-file corpus ratchet passes unchanged — no file
moves bucket, and the committed baseline needed no re-basing. That is not
because the defect is theoretical. Nine corpus files contain the construct, and
every one is accounted for:

- **Seven were already `no-answer`** for an unrelated, larger failure in the
  same file (`base/lex.t`, `io/open.t`, `op/ref.t`, `op/split.t`,
  `op/localref.t`, `op/sub_lval.t`, `op/magic.t`). The false `WRONG` was
  *masked*, not absent — remove the larger failure and it surfaces.
- **`perf/optree.t`** has `'@{\@_}'` inside a string literal, correctly not
  parsed as code.
- **`op/tie.t`** has `&{\&$$elem}`, which is `\&$…` rather than `\` against a
  plain name. The grammar builds a correct block for it; the collapse needs the
  backslash against a bare sigilled name.

So the fix is a **latent-fault fix**: it removes a false `WRONG` that this
corpus happens to hide behind other failures, and that a different corpus — or
this one after the other defects are fixed — would expose. The 19 files
matching across the whole perl5 checkout (`${\$_}`, `@{\@pkg}`,
`&{\&utf8::is_utf8}` and relatives) are the scale of the construct in the wild.

Declining on every dereference would instead score 100% non-`WRONG` and measure
nothing — the failure mode the `Bucket` doc comment warns about — which is what
`TestCollapsedDerefDetectorIsNarrow` exists to prevent.

The caveats of §0.4.1 carry over unchanged: this says nothing about whether
perl accepts the source (every case it flags *compiles*), it is a proxy rather
than a syntax check, and it does not survive the rewrite this specification
proposes. `TestDerefOfExplicitRefCollapses` pins the status quo, so a grammar
that starts parsing these correctly fails loudly and points at the
compensation to delete.

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
(`toke.c:5270`). Both are arbitrary compile-time code execution, but they are
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

*2026-09-08. Corpus `perl5` at `94e5086608`, interpreter 5.42.0 (the pin in
`internal/parseoracle/testdata/corpus.pin`).*

```
PARSEORACLE_SHIM=/tmp/oracletree PERL5_CORPUS=~/dev/perl5 \
  go test ./internal/parseoracle/ -run TestCorpusSweep -parseoracle.corpus -v
```

All 620 `.t` files, each compiled by perl and by our parser and the two
bucketed against each other:

| Bucket | Files | Share of measured |
|---|---:|---:|
| exact | 380 | 64.1% |
| wider | 11 | 1.9% |
| **WRONG** | **0** | **0.0%** |
| no-answer | 202 | 34.1% |
| *measured (denominator)* | *593* | |
| excluded, environmental | 26 | — |
| runner error | 1 | — |

Two independent runs produced identical counts. Wall time was 5m52s and
6m10s across 24 workers.

**WRONG is zero, and that is worth less than it sounds.** The comparison
tests one marker — `srefgen`, perl taking a reference at a call site — and
our grammar emits `ambiguous_function_call_expression` where it cannot settle
a prototype. An honest refusal buckets `wider`, never WRONG, so the eleven
`wider` files are every prototype-driven call in the corpus and the zero is
mostly a statement about what the harness currently asks.

**`exact` is weaker than "parses like perl", and by a measurable amount.** On
the 44 baseline files of §0.6 the sweep reports 29 exact and 15 no-answer —
reproducing that section's coverage split exactly. But only **7 of those 44
files take a reference at all**, and **25 of the 29 exact verdicts had no
marker in play**. For those 25, `exact` means only "perl took no reference
our source did not write", which is true of any file that takes no
references. The 64.1% is therefore a ceiling on agreement, not a measurement
of it; adding markers (`rv2hv`, `match`, `readline`, `anonhash`) is what
converts it into one.

**One file moved after this was first measured, and not because the parser
changed.** `run/switcht.t` was `no-answer` because the harness compiled it
without the taint flag its `#!./perl -t` shebang asks for — perl refused with
`"-t" is on the #! line, it must also be used on the command line`, an
environmental refusal frozen as a parse result. Fixing `wantsTaint` moved it to
`exact`: 379 → 380 and 203 → 202. The parser parsed it correctly all along.

**The `no-answer` third is our coverage gap, not perl's.** 202 files carry an
error node from our parser. Those 203 are the population the conformance
plan's M1 has to move, and they are the honest reading of "how much of Perl
we cannot parse".

**Two corpus facts stay separate from the parser's score.** 26 files fail
because this machine's environment cannot satisfy them — a missing module is
not a parse error — and they leave the denominator by way of `Classify`. One
file, `t/re/pat_psycho.t`, is a runner error: it calls `test.pl`'s
`watchdog(5 * 60)` from a `BEGIN` block, so `perl -c` forks a monitor that
inherits the oracle's pipe. It costs its 60-second timeout and reaches no
bucket.

**Cost.** The plan's §2 estimated ~25 ms per file for `B::Concise`; measured
on real corpus files it is **~600 ms** (`op/sub.t`, 1055 ops), and ten files
cost 9.3 s sequentially of which the prototype probe is only ~10%. That is
where six minutes goes, and it is a real argument for the content-hash cache
§2 describes — deferred to its own issue rather than built speculatively
here.
