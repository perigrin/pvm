<!-- ABOUTME: Chapter 7 of the Perl parser specification: how the implementer knows the parser is correct. -->
<!-- ABOUTME: Defines the parse oracle, the four metrics, the corpus, differential testing, and the fuzz and incremental invariants. -->

# 7. Conformance and Test Strategy

## 7.0 The problem this chapter solves

Every Perl parser project reports the same number: "we parse N% of files
without error." That number is nearly worthless. A file can parse cleanly
and still be parsed **wrong**:

```perl
sub f(\@) { }
my @a;
f(@a);          # perl passes a REFERENCE to @a
```

A parser that does not track prototypes builds a call node with a list
argument. No error node. No diagnostic. A confident, wrong tree, which every
downstream consumer — the type checker, go-to-definition, the semantic
highlighter — then acts on.

This is not hypothetical. It is the **current, measured blind spot in PSC**:

> nothing has ever checked whether tree-sitter's tree AGREES with perl's
> parse, only whether it contains ERROR nodes.
>
> — `docs/plans/2026-09-04-psc-parser-fix-options.md`

The same defect class was already found one level up, in type inference.
`internal/infer/precision_test.go` opens with it:

> "96.8% of value nodes typed" says how OFTEN inference answers. It says
> nothing about whether the answer is RIGHT, and the two move independently:
> a change that turns Unknowns into confident-but-wrong types raises
> coverage and makes the checker worse.

That chapter of work was fixed by making **perl itself the oracle**. This
chapter applies the identical move to parsing. Perl will tell you how it
parsed something. "Did we parse this the way perl did?" is therefore an
empirical question with a mechanical answer, not a matter of opinion.

Everything below follows from that one fact.

---

## 7.1 The parse oracle

### 7.1.1 The four channels, all verified

Perl exposes its own parse decisions through four independent channels. Each
was run on `perl v5.42.0 (x86_64-linux-thread-multi)` while writing this
chapter; the outputs shown are actual.

| # | Channel | Question it answers | Cost |
|---|---|---|---|
| 1 | `perl -c` | Does this compile at all? | ~8.8 ms |
| 2 | `prototype(\&f)` | What prototype is in scope for `f`? | ~9 ms |
| 3 | `perl -MO=Concise` | What optree did the parse produce? | ~25 ms |
| 4 | `perl -MO=Deparse` | Can perl re-emit the program it parsed? | ~25 ms |

#### Channel 1 — `perl -c`, the yes/no

```console
$ perl -c -e 'my $x = ;'
syntax error at -e line 1, near "= ;"
-e had compilation errors.
```

Exit status is the whole signal. This is the **weakest** oracle and the
cheapest. It answers metric (a) only — see §7.2.

#### Channel 2 — `prototype`, the direct read

```console
$ perl -e 'sub f(\@$;$){} print prototype(\&f), "\n"'
\@$;$
```

Perl hands back the prototype string verbatim. Any parser claiming to track
prototypes can be checked against this in one line, with no optree
interpretation at all.

#### Channel 3 — `B::Concise`, the parse-shape witness

This is the load-bearing one. The prototype example from §7.0, measured:

```console
$ perl -MO=Concise -e 'sub f(\@){} my @a; f(@a)' | grep -c srefgen
1
$ perl -MO=Concise -e 'sub f{}    my @a; f(@a)' | grep -c srefgen
0
```

Byte-for-byte identical call syntax. One `srefgen` op present, one absent.
**The prototype changed the parse, and the optree proves it.** No
hand-written expectation table is involved; perl reported it.

The `-exec` flag gives a flat, linear op sequence that is much easier to diff
than the nested default:

```console
$ perl -MO=Concise,-exec -e '$x = 1 + 2'
1  <0> enter v
2  <;> nextstate(main 1 -e:1) v:{
3  <$> const[IV 3] s/FOLD
4  <#> gvsv[*x] s
5  <2> sassign vKS/2
6  <@> leave[1 ref] vKP/REFC
```

Note line 3: `const[IV 3] s/FOLD`. Constant folding has already happened.
This is a real and important caveat — see §7.1.3.

**Concise is deterministic.** Two runs of the same input hash identically:

```console
$ for i in 1 2; do perl -MO=Concise -e 'my $x=1; sub f{$_[0]+1} f($x)' | md5sum; done
058ccb99c94ca8d50a80676fc7d9e6e9  -
058ccb99c94ca8d50a80676fc7d9e6e9  -
```

Determinism is what makes the optree usable as a golden value.

#### Channel 4 — `B::Deparse`, the round-trip

```console
$ perl -MO=Deparse -e 'my @a = map { $_*2 } grep /x/, @b; sub f(\@){} f(@a);'
my(@a) = map({$_ * 2;} grep(/x/, @b));
sub f (\@) {

}
f(@a);
```

Deparse output is **idempotent at the second pass** — deparsing the deparse
returns the identical text (verified above). This gives a fixpoint, and the
fixpoint is the comparison target: your parser's re-emitted source, fed back
through Deparse, must reach the same fixpoint text. That normalises away
whitespace, optional parens, and statement-terminator style, which are
exactly the differences you do not want to fail on.

### 7.1.2 Which channel answers which question

| Parse question | Channel | Signal |
|---|---|---|
| Is this even a program? | `-c` | exit status |
| Did `f @a` pass a list or a ref? | Concise | `srefgen` present |
| What prototype applies? | `prototype()` | the string |
| Is `%h` a hash or a modulus here? | Concise | `rv2hv` vs `modulo` |
| Is `/.../` a match or division? | Concise | `match` vs `divide` |
| Did `<FH>` read or glob? | Concise | `readline` vs `glob` |
| Is `{}` a block or a hashref? | Concise | `leaveloop`/`scope` vs `anonhash` |
| Which sub did `->m` resolve to? | Concise | `method_named` vs `method` |
| Did BEGIN change the parse? | Concise | run it and look |
| Is the whole structure right? | Deparse | fixpoint text equality |

The middle rows are the ones this chapter exists for. Each is a documented
Perl parsing ambiguity — Chapter 3 covers why they are ambiguous — and each
is **decidable by observation**.

### 7.1.3 What the oracle cannot tell you, stated plainly

The oracle is not free of caveats. Recording them here so nobody rediscovers
them as bugs.

**The optree is post-optimisation.** `1 + 2` arrives as `const[IV 3]`. Perl's
peephole optimiser runs between parse and the tree you read. Constant
folding, `sassign`-to-`padsv` collapsing, and `nextstate` elision all happen
before `B::Concise` sees anything. Consequence: **the optree is a lossy
witness of the parse.** It over-approximates agreement. Two different parses
can fold to the same optree.

Mitigation: compare on **op presence and shape**, never on exact op counts,
and confine folding-sensitive comparisons to the specific ops that mark the
ambiguity (`srefgen`, `rv2hv`, `match`, `readline`, `anonhash`). A
disagreement is real evidence; an agreement is weaker evidence than it looks.
This is the parse-level analogue of the "wider" bucket in the type oracle,
and it deserves the same suspicion the precision test gives it.

**BEGIN blocks execute.** `perl -c` and `B::*` both run `BEGIN`, `use`, and
`BEGIN`-time `eval`. That is precisely what makes them able to answer the
hard questions — and precisely what makes them unsafe on untrusted input.
The oracle harness is a **development and CI tool only.** It must never run
inside the LSP server; it runs in CI against a pinned perl (plan §4.4).

**The oracle needs a complete, valid program.** From the fix-options plan:

> `perl -c`, B::Concise and PPI all want a complete, valid program: perl's
> own parser is not incremental, not error-recovering, and not reusable
> across edits.

So the oracle validates **batch parses of files that compile**. It says
nothing about error recovery, partial input, or the half-typed buffer an
editor holds. Those need the property tests in §7.5 and §7.6 instead. Keep
the two categories separate in your head: the oracle proves correctness on
valid input; fuzzing proves robustness on invalid input. Neither substitutes
for the other.

**The oracle is version-pinned.** See §0.7 — this bit users already.

### 7.1.4 The Go harness

The harness — `oracle.Compiles`, `oracle.Prototype`, `oracle.Concise`,
`oracle.Deparse` and the `-exec` line parser — is scaffolding, and its code
sketch lives in the plan (`docs/plans/2026-09-05-parser-conformance-plan.md`
§1). What this chapter requires of it: stdlib plus `os/exec`; a compile
failure returned as an *answer*, not a harness error; Deparse driven to its
second-pass fixpoint (§7.1.1); and comparison on op presence, never on
position or count (§7.1.3).

### 7.1.5 Bucketing: exact / wider / WRONG / no-answer

Reuse the type oracle's four buckets verbatim. They already survived
adversarial review in this repo, and the semantics carry over exactly.

| Bucket | Meaning at the parse level | Action |
|---|---|---|
| **exact** | Our tree implies the same optree shape perl produced | none |
| **wider** | We produced a *less committed* tree that is not wrong — e.g. we emit a generic `CallExpression` where perl resolved a specific prototype-driven form, and we mark it unresolved | track; must not grow silently |
| **WRONG** | We committed to a parse perl did not make | **fail the build** |
| **no-answer** | We produced an ERROR node, or the oracle declined | track; this is the coverage metric |

The critical property, inherited from the precision test's comment:

> A wrong answer is the failure that matters: it is worse than saying
> nothing, because a consumer acts on it.

So: `assert.Zero(t, wrong)`. Coverage (`no-answer`) is a ratchet, not a gate —
it may only improve. **WRONG is a hard gate at zero from day one**, even
when coverage is 5%. A parser that handles ten constructs correctly is more
useful than one that handles a hundred with three lies in it, because the
first can be trusted within its stated scope and the second cannot be trusted
anywhere.

The `wider` bucket carries the same trap the type work found. From
`precision_test.go`:

> bson hit this verifying its own oracle: it injected Str where Int was
> expected, and the check PASSED because the injection was not wrong, only
> wider.

The parse analogue: declaring every call "unresolved" scores 100% non-WRONG
and is useless. **Guard against it by asserting a floor on `exact`, not only
a ceiling on `wrong`.** A milestone that raises `wider` while `exact` stays
flat has bought nothing.

### 7.1.6 Making it fast enough for CI

Process spawn dominates: ~8.8 ms per `perl -c`, ~25 ms per `B::Concise`
(§7.1.1), so the 620-file corpus costs ~15 s single-threaded under Concise.
Caching on content hash, parallelism, one process per file rather than per
assertion, and a build-tag tier for the full sweep bring that to sub-second
on unchanged inputs. The mechanics are in the plan, §2.
---

## 7.2 The four metrics, defined precisely

Confusing these is the central failure mode. Name them, measure them
separately, and report all four.

### (a) parses-without-error — the weak metric

> Of N files, how many produce a tree containing no ERROR node?

**What everyone reports. Nearly worthless alone.** A parser that treats every
unrecognised construct as an opaque token scores 100% and understands
nothing. Both failure directions exist: PSC currently reports "495 confident
diagnostics" on a file it mis-read (`header_parser.t`), and the inverse — a
clean parse of a wrong tree — is undetectable by this metric by construction.

Useful only as a **coverage** number, and only when reported next to (b).

### (b) parses-correctly — the metric that matters

> Of N constructs whose parse perl will report, on how many does our tree
> agree with perl's?

This is the oracle metric of §7.1. **Nothing in PSC has ever measured it.**
It is the reason this chapter exists.

Report it bucketed: exact / wider / WRONG / no-answer. A single percentage
hides the distinction between "did not commit" and "committed wrongly", and
that distinction is the whole point.

### (c) round-trips — losslessness

> For source `S`, does `emit(parse(S)) == S`, byte for byte?

Two distinct properties, do not conflate them:

- **Lossless round-trip** (`emit(parse(S)) == S`) requires the CST to retain
  every byte: whitespace, comments, POD, `__END__` content. Chapter 6
  requires this for LSP formatting and rename, so it is a **hard invariant**,
  not a metric. Any file failing it is a bug.
- **Semantic round-trip** (`deparse(emit(parse(S))) == deparse(S)`) is the
  weaker, oracle-mediated form. It tolerates formatting difference and
  catches structural error. Use it where lossless round-trip is not yet
  achievable.

Round-trip is the **cheapest high-value test in the whole plan**: it needs no
perl, runs at memory speed, applies to every byte of every corpus file, and
catches a large class of position and span bugs that no other check sees.

### (d) stable-under-incremental-edit

> For any source `S` and edit `E`: `reparse(parse(S), E) ≡ parse(apply(S,E))`?

The incremental parser must be *indistinguishable from* the batch parser.
Not "close enough" — equal. Trees must compare equal on kind, span, and
child structure; only node identity and reuse bookkeeping may differ.

This is property-testable and belongs in `go test -fuzz`. §7.6.

### Summary table

| Metric | Needs perl? | Cost | Gate |
|---|---|---|---|
| (a) no-error | no | fast | ratchet, may only improve |
| (b) agrees-with-perl | **yes** | 25 ms/file, cached | **WRONG == 0, hard fail** |
| (c) lossless round-trip | no | fast | **hard invariant, zero failures** |
| (c') semantic round-trip | yes | 50 ms/file | ratchet |
| (d) incremental == batch | no | fast | **hard invariant, zero failures** |

Note which gates are hard: (b)-WRONG, (c), and (d). All three are
**correctness invariants that hold at any coverage level**. Only the coverage
numbers are ratchets — committed per-file baselines that may only improve;
the mechanics are in the plan, §4. This ordering is deliberate — it lets you ship a parser
covering 30% of Perl and honestly claim it is correct on that 30%.

---

## 7.3 Corpus inventory

### 7.3.1 perl5's own test suite

Location: `/home/perigrin/dev/perl5/t/`. Measured 2026-09-04 against a blead
checkout at `94e5086608` (2026-07-16).

**620 `.t` files, 134,000 lines total.**

| Directory | Files | Lines | What it tests | Parse difficulty | Notes |
|---|---:|---:|---|---|---|
| `base/` | 9 | 1,379 | Bootstrap: does perl work at all | **Hard** | Tiny but `lex.t` is pathological. See §7.3.2 |
| `comp/` | 25 | 4,992 | **The compiler and parser itself** | **Hard** | `parser.t` (26K), `proto.t` (24K), `require.t` (21K). The single most valuable directory |
| `op/` | 228 | 71,454 | Every operator and builtin | Medium–Hard | The bulk. `signatures.t` 1,613 lines, `tr.t` 1,224, `lexsub.t` 1,091 |
| `re/` | 80 | 17,053 | Regex engine | **Very hard to lex** | `pat_advanced.t` 2,743 lines. Pattern *bodies* need not be parsed; pattern *delimiters* must be |
| `class/` | 12 | 1,268 | `use feature 'class'` | Medium | Small, modern, high value. `field.t` 8.4K |
| `io/` | 44 | 6,807 | Filehandles, layers | Medium | Filehandle syntax is a documented ambiguity |
| `run/` | 28 | 5,025 | Command-line switches | Easy–Medium | Much is `-e` strings inside strings |
| `mro/` | 73 | 6,971 | Method resolution order | **Easy** | 73 files, highly repetitive, `_utf8` variants doubling many. Good early volume |
| `uni/` | 30 | 4,847 | Unicode identifiers, casing | Medium | Tests the lexer's UTF-8 handling specifically |
| `opbasic/` | 5 | 1,751 | Ops without `Test::More` | **Easy** | Deliberately minimal-dependency. Excellent M1 target |
| `cmd/` | 5 | 466 | Control flow statements | **Easy** | Smallest useful set. Good M0 target |
| `perf/` | 5 | 1,986 | Optree shape and op counts | Medium | **`opcount.t` (42K) is a gift** — see §7.3.3 |
| `porting/` | 37 | 8,697 | *The perl source tree*, not the language | N/A | **Mostly exclude.** See below |
| `lib/` | 10 | 465 | Pragma behaviour helpers | Easy | Small |
| `test_pl/` | 6 | 328 | The test harness itself | Easy | Self-test of `t/test.pl` |
| `bigmem/` | 12 | 619 | Huge-string behaviour | Easy | Trivially parseable, skip at runtime |
| `win32/` | 9 | 1,145 | Windows-specific | Easy | Parse fine; exclude from runtime |
| `japh/` | 1 | 680 | Obfuscated "Just Another Perl Hacker" | **Adversarial** | See §7.3.2 |
| `benchmark/` | 1 | 84 | Benchmark harness | Easy | Negligible |

**`porting/` is about the perl source, not the language.** `podcheck.t`
(87K), `header_parser.t` (36K), `diag.t` (26K), `bench.t` (26K),
`libperl.t` (21K) test C headers, POD formatting, and the release process.
They are still *Perl programs* and so still parseable corpus, but they are
not language conformance and their failures carry no signal about the
grammar. **Exclude from milestone targets; keep in the fuzz/round-trip
corpus** where any valid Perl is useful.

**Harness note:** 526 of 620 files `require './test.pl'`; only 8 use
`Test::More`. So the corpus is nearly homogeneous, and one shared prelude
shape covers it.

### 7.3.2 The pathological files

Name these explicitly. They will break your parser, and they are supposed to.

**`t/base/lex.t` (16K, 129 tests) — deliberately abuses the lexer.** Opening
lines:

```perl
$x = $#[0];              # $# of an anonymous array? no: $#, then [0]
$x = '\\'; # ';          # a quoted backslash, then a comment containing a quote
eval '$foo{1} / 1;';     # is / division or a pattern? after a hash subscript
print <<'EOF';           # non-interpolating heredoc
print <<EOF;             # interpolating heredoc
eval <<\EOE, print $@;   # backslash-quoted heredoc terminator, mid-expression
```

Every one is a lexer-feedback case from Chapter 3. **Treat `lex.t` as a
milestone in its own right, not as one file among 620.** If it parses, the
lexer is real.

**`t/japh/abigail.t` (680 lines) — adversarial by construction.** Obfuscated
Perl written to be unreadable. It is valid Perl, so a correct parser handles
it, but it is a terrible early target and a superb late one. Note it does not
compile under `perl -c` in isolation (it depends on being run from `t/`), so
oracle comparison on it needs care.

**`t/comp/parser.t` (26K)** — perl's own tests for its parser, including
error recovery and pathological nesting. Directly relevant, and the best
single source of tricky-construct fixtures to lift into unit tests.

**`t/op/signatures.t` (1,613 lines)** and **`t/op/for-many.t`** — see §0.5 and §0.7,
these two contain post-5.42 syntax.

### 7.3.3 `t/perf/opcount.t` — a pre-built optree oracle

42K of assertions of the form "this program compiles to exactly these ops in
these counts." Perl's own maintainers wrote an optree-shape test suite, and
its assertions are a **ready-made expected-value table for the parse oracle**
— already reviewed, already maintained, already correct. Mine it early. It is
the highest-value 42K in the tree for this chapter's purposes.

### 7.3.4 Measured parseability, and the version trap

Compile rates per directory are in §0.5, measured with a correctly built
shim, together with the lesson an earlier run of this table taught: a bare
`perl -c` exit status conflates "does not parse" with "`BEGIN` blew up", and
only classifying stderr separates the two. The two genuine syntax errors on
5.42 — `op/signatures.t:927` and `op/for-many.t:474` — are blead-only
syntax, and §0.7 draws the consequence: **the corpus and the oracle must be
pinned to the same perl version, and the version recorded wherever a
baseline is kept.** How the harness classifies stderr, and the pin-file
format, are in the plan, §3.

### 7.3.5 PerlOnJava's corpus — a graded corpus someone already built

Location: `/home/perigrin/dev/PerlOnJava`. **2,055 `.t` files.**

| Location | Files | Character |
|---|---:|---|
| `src/test/resources/unit/` (top level) | 986 | **One construct per file**, descriptively named |
| `src/test/resources/unit/regex/` | 438 | Regex, one behaviour per file |
| `src/test/resources/unit/refcount/` | 59 | Reference counting |
| `src/test/resources/unit/overload/` | 18 | Operator overloading |
| `src/test/resources/unit/pack/`, `string/`, `runtime/` | 5 | Misc |
| `src/test/resources/module/*/t/` | ~350 | Vendored CPAN suites (Math-BigInt 49, Net-SSLeay 47, XML-Parser 45, Text-CSV 38, Unicode-Collate 37) |
| `dev/regex/tools/tests`, `dev/tools/tests` | 73 | Development scratch |

**Why this corpus is worth more than its size suggests.** The unit files are
one-construct-per-file with names that state the construct:

```
anonymous_code_attribute_order.t     attributes_prototype_warning_scope.t
b_deparse_prototype_block.t          array_autovivification.t
autoload_lvalue_assignment.t         boolean_bitwise_xor.t
regex/branch_reset_named_call.t      regex/accept_variable_lookbehind.t
```

Someone already did the decomposition work. When your parser fails
`anonymous_code_attribute_order.t`, the filename tells you the feature. That
is the difference between a 4,000-line failure and a bug report. **This is
the single best-shaped corpus of the three for a from-scratch parser**,
because a from-scratch parser fails in feature-shaped ways and this corpus is
indexed by feature.

Their discipline is worth copying verbatim (`AGENTS.md`):

> **ALWAYS validate new unit tests with standard Perl before relying on
> them.** Unit tests must encode standard Perl behavior, not PerlOnJava-specific
> behavior.
>
> **NEVER modify or delete existing tests.** Tests are the source of truth.
> If a test fails, fix the code, not the test.

They run them via `prove -r src/test/resources/unit` under real perl to
establish the expected result, then under their implementation. **That is the
oracle pattern, arrived at independently.** Three projects — PerlOnJava, the
PSC type work, and this chapter — converged on "run real perl, compare." That
convergence is itself evidence the approach is right.

License: dual Artistic/GPL, same as perl. See §7.3.7.

### 7.3.6 Recommended corpus tiers

| Tier | Contents | Files | Use |
|---|---:|---|---|
| **T0 Fixtures** | Hand-written, one construct each, in-repo | ~200 | Unit tests. Every bug gets one. Runs in `-short` |
| **T1 Graded** | PerlOnJava `unit/` (excl. `regex/`) | 986 | Feature-indexed milestone driver |
| **T2 Core** | perl5 `base cmd comp opbasic class` | 56 | Language core. Small, dense, high-signal |
| **T3 Volume** | perl5 `op mro uni lib test_pl` | 347 | Breadth |
| **T4 Regex** | perl5 `re/` + PerlOnJava `unit/regex/` | 518 | Lexer delimiter stress |
| **T5 Wild** | CPAN tarballs, PerlOnJava `module/*/t/` | 350+ | Real-world code |
| **T6 Adversarial** | `base/lex.t`, `japh/`, fuzz corpus | ~50 | Where it breaks |

T0–T2 is ~1,240 files and covers the plan's milestone ladder (§7.8) through
M4; T5 is not needed before M5.

### 7.3.7 Vendoring and licensing

perl5's tests are dual-licensed Artistic 1.0 / GPL 1.0+ (`perl5/Artistic`,
`perl5/Copying`), and PerlOnJava's carry the same dual licence. How the
corpus is fetched, pinned and kept out of this repository is a project
decision, not a fact about Perl: see the plan, §3.2. T0 fixtures are
project-authored and raise no licence question.

---

## 7.4 The ratchet

A committed per-file baseline that may only improve: any file regressing
fails the build, any file improving requires a baseline update in the same
commit. The format, the Go implementation and the CI shape are in the plan,
§4.

---

## 7.5 Differential testing

### 7.5.1 Three implementations, one arbiter

Chapter 1 already states the rule:

> Three independent implementations are triangulated throughout. Where they
> disagree, **perl wins** — it is the definition.

| Implementation | Access from Go | Speed | Role |
|---|---|---|---|
| **perl** | `os/exec` | 25 ms | **Arbiter. Always right by definition.** |
| **tree-sitter-perl** | via gotreesitter, in-process | µs | Cheap second opinion; the current PSC parser |
| **PerlOnJava** | `os/exec` on `jperl` | JVM startup, seconds | Expensive; use on disagreements only |

### 7.5.2 The triage rule

```
Go and perl agree                     -> pass
Go and perl disagree                  -> OUR BUG. Always. No exceptions.
Go and tree-sitter disagree,
    perl unavailable                  -> INVESTIGATE, do not fail
Go and tree-sitter disagree,
    Go matches perl                   -> tree-sitter bug. File upstream.
Go and PerlOnJava disagree,
    both match perl                   -> both fine (tree shape differs legitimately)
perl fails to compile the input       -> NO ANSWER. Exclude from scoring.
```

The rule is one sentence: **perl always wins.** Write it in the harness as a
comment, because the temptation to "fix" a corpus file arrives every time.
PerlOnJava's `AGENTS.md` states it as policy — "Tests are the source of
truth. If a test fails, fix the code, not the test" — and they are right.

### 7.5.3 Differential testing has already paid out here

This is not a speculative benefit. Running tree-sitter against the same
corpus is precisely what produced the finding in the fix-options plan:

> shipped blob      89 of 227 t/op files with parse errors
> with this patch   64 of 227 (28%)
> **upstream floor  31-35 of 227 — measured by running tree-sitter's OWN CLI
> over the same corpus**
>
> The target of 0 in the brief was wrong and unreachable.

A differential run against another implementation **corrected a milestone
target that was set by guesswork.** That is the argument for building the
comparison harness early: it tells you not only where you are, but whether
where you are aiming exists.

### 7.5.4 Harness sketch

```go
// ABOUTME: Differential harness — runs several Perl parsers on one input and reports disagreement.
// ABOUTME: Perl is the arbiter; every other implementation is a second opinion.

type Verdict int

const (
    Agree Verdict = iota
    OurBug          // we disagree with perl. always our fault.
    TheirBug        // we match perl, another impl does not
    NoAnswer        // perl could not compile it
    Investigate     // impls disagree, perl silent
)

func Differential(ctx context.Context, src []byte) (Verdict, string) {
    ours, ourErr := goparser.Parse(src)

    perlOps, perlErr := oracle.Concise(ctx, perlBin, string(src))
    if _, bad := perlErr.(*oracle.CompileError); bad {
        // perl will not compile it, so there is no ground truth here.
        return NoAnswer, "perl declined: " + perlErr.Error()
    }

    switch {
    case ourErr != nil:
        return OurBug, "we failed where perl succeeded: " + ourErr.Error()
    case !impliesSameOptree(ours, perlOps):
        return OurBug, diffOptree(ours, perlOps)
    }

    // We match perl. Now a cheap second opinion, for information only.
    if ts, err := treesitter.Parse(src); err == nil && !sameShape(ours, ts) {
        return TheirBug, "tree-sitter disagrees with perl-confirmed parse"
    }
    return Agree, ""
}
```

`impliesSameOptree` is the interesting function and the one to write
carefully. It must **not** try to reproduce perl's optree — that would be
reimplementing the compiler. It checks the *marker ops* from §7.1.2:

```go
// markers maps a parse decision we make to the op perl emits when it makes
// the same decision. Presence, not position or count — the peephole
// optimiser reorders and folds, so counts are not stable.
var markers = []struct {
    ourPredicate func(*ast.Node) bool
    perlOp       string
}{
    {isPrototypeRefArg, "srefgen"},
    {isHashDeref, "rv2hv"},
    {isMatchNotDivide, "match"},
    {isReadlineNotGlob, "readline"},
    {isAnonHashNotBlock, "anonhash"},
}
```

Start with five markers. Add one every time the oracle catches a disagreement
class the markers missed. This is deliberately incremental: a complete optree
model is not required to get value, and attempting one is how this turns into
a second compiler.

---

## 7.6 Fuzzing

Go's native fuzzer is standard library — `testing.F`, `go test -fuzz` — so
this costs no dependency.

### 7.6.1 What to fuzz

perl-lsp's targets (`perl-lsp/fuzz/fuzz_targets/`, 22 of them) are a
well-chosen list, arrived at by finding real bugs. The subset that maps onto
this project:

| Target | Why it earns its slot |
|---|---|
| `lexer_tokenization` | The lexer is where Perl parsers actually break |
| `heredoc_parsing` | Heredocs suspend the line; nesting and `<<~` are notorious |
| `substitution_parsing` | `s///e` re-enters the parser on the replacement |
| `quote_operators` | Arbitrary delimiters, bracket nesting, `q{a{b}c}` |
| `declaration_parsing` | `my`/`our`/`state`/`field` and their attribute tails |
| `incremental_edit_sequences` | The §7.2(d) invariant. See §7.6.4 |
| `unicode_positions` / `utf16_roundtrip` | LSP positions are UTF-16 offsets |
| `structured_perl_programs` | Grammar-guided generation, not just bytes |

Order for a from-scratch parser: **lexer, heredoc, quote-operator, then
incremental.** The first three are where a hand-written Perl lexer dies.

### 7.6.2 The invariants

Four, and only the first is about crashing.

```go
// ABOUTME: Fuzz invariants for the lexer — properties that hold on ANY input, valid or not.
// ABOUTME: The parser must never panic, hang, or produce non-monotonic positions.

func FuzzLexer(f *testing.F) {
    // Seed from the corpus: real Perl mutates into interesting near-Perl.
    for _, s := range []string{
        "my $x = 42;",
        "print <<'EOF';\nbody\nEOF\n",
        "s{a}{b}ge;",
        "q{nested {braces} here}",
        "$x = $#[0];",              // from t/base/lex.t
        "my @a = map { $_*2 } @b;",
        "sub f(\\@) {}",
        "<<~EOT;\n  indented\n  EOT\n",
    } {
        f.Add(s)
    }

    f.Fuzz(func(t *testing.T, src string) {
        // INVARIANT 1: never panic. The LSP must not die on a half-typed buffer.
        toks := lexer.Tokenize([]byte(src))   // panics fail the test automatically

        // INVARIANT 2: positions are monotonic and in bounds.
        prev := 0
        for i, tok := range toks {
            if tok.Start < prev {
                t.Fatalf("token %d start %d < previous end %d", i, tok.Start, prev)
            }
            if tok.End < tok.Start {
                t.Fatalf("token %d has End %d < Start %d", i, tok.End, tok.Start)
            }
            if tok.End > len(src) {
                t.Fatalf("token %d End %d exceeds input length %d", i, tok.End, len(src))
            }
            prev = tok.End
        }

        // INVARIANT 3: lossless. Concatenating every token's text — trivia
        // included — reproduces the input exactly. This is metric (c), and
        // it catches more real bugs than the other three combined.
        var sb strings.Builder
        for _, tok := range toks {
            sb.WriteString(src[tok.Start:tok.End])
        }
        if sb.String() != src {
            t.Fatalf("lossy tokenization:\n got %q\nwant %q", sb.String(), src)
        }
    })
}
```

**Invariant 4 — termination — needs care**, because `go test -fuzz` does not
detect infinite loops; it hangs. A Perl lexer *will* infinite-loop during
development: an unterminated heredoc or quote-like operator that fails to
advance the cursor is the classic case. Guard structurally rather than with a
timeout:

```go
// INVARIANT 4: every lexer step consumes at least one byte.
//
// A timeout would catch this too, but only after the fuzzer has stalled for
// minutes. Asserting forward progress at the loop head fails in microseconds
// and points at the offending state directly.
func (l *Lexer) next() Token {
    before := l.pos
    tok := l.scan()
    if l.pos == before && tok.Kind != EOF {
        panic(fmt.Sprintf("lexer made no progress at %d in state %v", l.pos, l.state))
    }
    return tok
}
```

Ship that check in the production lexer behind a build tag, or leave it in
unconditionally — one integer comparison per token is not a cost worth
optimising, and the failure it prevents is a hung editor.

### 7.6.3 Structure-aware fuzzing

Random bytes rarely reach the parser; they die in the lexer. To fuzz the
*parser*, generate Perl-shaped input. `testing.F` gives you a byte string, so
use it as a **seed for a generator** rather than as source:

```go
func FuzzParser(f *testing.F) {
    f.Add([]byte{1, 2, 3})
    f.Fuzz(func(t *testing.T, seed []byte) {
        src := generatePerl(seed) // deterministic: same seed, same program
        tree, err := parser.Parse([]byte(src))
        if err != nil {
            return // generator may emit invalid Perl; that is fine
        }
        // Round-trip must hold for anything that parses.
        if got := tree.Source(); got != src {
            t.Fatalf("round-trip failed:\n got %q\nwant %q", got, src)
        }
    })
}
```

`generatePerl` walks the seed bytes as choices in a small grammar — statement,
then expression, then term — with a depth cap. 200 lines, no dependencies,
and it reaches parser states that byte mutation never will.

### 7.6.4 Fuzzing corpus seeds

Seed the fuzzer from the real corpus. `go test -fuzz` stores its corpus under
`testdata/fuzz/<FuzzTestName>/`, and seeding it from perl5's `t/` gets the
fuzzer past the "learning what Perl looks like" phase entirely:

```console
$ go run ./cmd/seedcorpus -from $PERL5_CORPUS -to testdata/fuzz/FuzzLexer
```

Commit crashers, never delete them. Every crash becomes a permanent
regression test — which is exactly the PerlOnJava rule ("EVERY externally
observed failure requires permanent tracked regression coverage") applied to
fuzzing.

---

## 7.7 Incremental-parse testing

### 7.7.1 The property

> For any source `S` and any edit `E`:
> `reparse(parse(S), E)` ≡ `parse(apply(S, E))`

Structural equality on kind, span, and children. Node identity and reuse
bookkeeping may differ — that is the *point* of incrementality — but nothing
observable to a consumer may.

This is the single most valuable property test in the project, because
incremental parsing is where subtle, data-dependent, impossible-to-reproduce
bugs live. A user reports "sometimes go-to-definition breaks after I type a
quote", and without this test you will never find it.

### 7.7.2 Property test with `go test -fuzz`

The fuzzer generates the edits. This is exactly what fuzzing is for: the
space of (source, edit) pairs is far too large to enumerate and far too
structured to sample by hand.

```go
// ABOUTME: Property test: incremental re-parse must equal full re-parse, for any edit.
// ABOUTME: The fuzzer generates edit sequences; the invariant is tree equality.

func FuzzIncremental(f *testing.F) {
    f.Add("my $x = 1;\nprint $x;\n", 5, 1, "42")
    f.Add("print <<'EOF';\nbody\nEOF\n", 6, 3, "\"\"")  // break the heredoc quote
    f.Add("s{a}{b}g;", 2, 1, "{")                        // unbalance a delimiter

    f.Fuzz(func(t *testing.T, src string, offset, delLen int, insert string) {
        // Clamp the fuzzer's integers into a valid edit. Rejecting instead
        // would throw away most inputs.
        if len(src) > 64*1024 || len(insert) > 1024 {
            t.Skip()
        }
        offset = clamp(offset, 0, len(src))
        delLen = clamp(delLen, 0, len(src)-offset)
        edited := src[:offset] + insert + src[offset+delLen:]

        old := parser.Parse([]byte(src))
        inc := parser.ReparseIncremental(old, parser.Edit{
            Start: offset, OldEnd: offset + delLen, NewText: insert,
        }, []byte(edited))
        full := parser.Parse([]byte(edited))

        if diff := treeDiff(inc, full); diff != "" {
            t.Fatalf("incremental != full after edit at %d (-%d, +%q):\n%s",
                offset, delLen, insert, diff)
        }
    })
}
```

### 7.7.3 The edits that actually break parsers

Seed these specifically. Each breaks the assumption that an edit's effect is
*local*, which is the assumption incremental parsing rests on:

| Edit | Why it is nasty |
|---|---|
| Insert `"` mid-file | Every subsequent token re-lexes. Maximum blast radius |
| Insert `<<EOF` | A heredoc suspends the rest of the line; the *next* line's meaning changes |
| Delete a heredoc terminator | The heredoc swallows the remainder of the file |
| Insert `=pod` at column 0 | POD mode until `=cut`. Everything after becomes trivia |
| Insert `__END__` | Everything after becomes data |
| Change `sub f()` to `sub f(\@)` | **A prototype edit changes how calls parse — possibly earlier in the file.** The one edit whose effect propagates *backwards* |
| Insert `{` | Every brace after it re-associates |
| Split a token (`$foo` -> `$f|oo`) | Edit lands inside a token, not between two |

The prototype row deserves emphasis. Perl allows a prototype to affect calls
*textually before* the declaration in some arrangements, and via `BEGIN` it
can affect anything. **Any incremental parser must have a defined
invalidation story for prototype edits, and this test is what proves the
story is true.** If your answer is "invalidate the whole file on a prototype
change" — that is a fine answer, and this test confirms you actually do it.

### 7.7.4 Multi-edit sequences

Single edits are the easy case. Real editors send bursts. Extend the property
to sequences, because state corruption often needs several edits to surface:

```go
func FuzzIncrementalSequence(f *testing.F) {
    f.Fuzz(func(t *testing.T, src string, script []byte) {
        tree := parser.Parse([]byte(src))
        cur := src
        for _, e := range decodeEdits(script, len(cur)) {  // seed -> edit list
            cur = applyEdit(cur, e)
            tree = parser.ReparseIncremental(tree, e, []byte(cur))
            // Check after EVERY edit. Corruption compounds, and checking only
            // at the end tells you a bug exists without telling you which edit
            // caused it.
            if diff := treeDiff(tree, parser.Parse([]byte(cur))); diff != "" {
                t.Fatalf("diverged after edit %+v:\n%s", e, diff)
            }
        }
    })
}
```

---

## 7.8 Milestones, ratchet, and the first test

The milestone ladder M0–M6 with its per-milestone targets, the ratchet — a
committed per-file baseline that may only improve, with its Go
implementation and CI shape — the first test to write, and the checklist are
a project plan, not a specification of Perl. They live in
`docs/plans/2026-09-05-parser-conformance-plan.md` (§4–§7).

Three invariants from that ladder are spec, not plan, and hold at every
milestone: **WRONG = 0** on the oracle (§7.1.5), **lossless round-trip** on
everything that parses (§7.2c), and **incremental == batch** (§7.2d).
