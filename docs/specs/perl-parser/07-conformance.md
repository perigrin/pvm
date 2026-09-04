<!-- ABOUTME: Chapter 7 of the Perl parser specification: how the implementer knows the parser is correct. -->
<!-- ABOUTME: Defines the parse oracle, the four metrics, the ratchet, the corpus inventory, and a graded milestone ladder. -->

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
inside the LSP server. §7.7 covers sandboxing.

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

**The oracle is version-pinned.** See §7.3.4 — this bit users already.

### 7.1.4 The Go harness

Stdlib plus `os/exec`. No dependencies. The shape follows
`internal/infer/precision_test.go`, which already works.

```go
// ABOUTME: Runs real perl to observe how perl itself parsed a program.
// ABOUTME: Perl is the oracle, so no hand-written expectation tables are needed.

package oracle

import (
    "bufio"
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"
)

// Op is one line of B::Concise -exec output, parsed into fields.
type Op struct {
    Seq   int    // "3"
    Class string // "<$>"
    Name  string // "const"
    Arg   string // "[IV 3]"
    Flags string // "s/FOLD"
}

// Concise returns perl's optree for src, as a flat -exec sequence.
//
// -exec is used rather than the nested default because a linear sequence
// diffs cleanly and does not require reconstructing perl's tree shape.
func Concise(ctx context.Context, perl, src string) ([]Op, error) {
    cmd := exec.CommandContext(ctx, perl, "-MO=Concise,-exec", "-e", src)
    out, err := cmd.Output()
    if err != nil {
        // A compile failure is an ANSWER, not a harness failure.
        if ee, ok := err.(*exec.ExitError); ok {
            return nil, &CompileError{Stderr: string(ee.Stderr)}
        }
        return nil, err
    }
    return parseConcise(string(out))
}

type CompileError struct{ Stderr string }

func (e *CompileError) Error() string { return "perl did not compile it: " + e.Stderr }

func parseConcise(s string) ([]Op, error) {
    var ops []Op
    sc := bufio.NewScanner(strings.NewReader(s))
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        // Skip "-e syntax OK" and B::Concise commentary lines.
        if line == "" || strings.HasPrefix(line, "#") || strings.HasSuffix(line, "syntax OK") {
            continue
        }
        var op Op
        // "3  <$> const[IV 3] s/FOLD"
        if _, err := fmt.Sscanf(line, "%d %s", &op.Seq, &op.Class); err != nil {
            continue
        }
        rest := strings.TrimSpace(line[strings.Index(line, op.Class)+len(op.Class):])
        name, arg, flags := splitOpBody(rest)
        op.Name, op.Arg, op.Flags = name, arg, flags
        ops = append(ops, op)
    }
    return ops, sc.Err()
}

// Has reports whether the optree contains an op with the given name.
// This is the primary comparison primitive: presence, not position.
func Has(ops []Op, name string) bool {
    for _, o := range ops {
        if o.Name == name {
            return true
        }
    }
    return false
}

// Names returns the op-name sequence, for whole-shape comparison.
func Names(ops []Op) []string {
    out := make([]string, len(ops))
    for i, o := range ops {
        out[i] = o.Name
    }
    return out
}

// Prototype asks perl directly what prototype is in scope for a named sub.
func Prototype(ctx context.Context, perl, src, sub string) (string, bool, error) {
    prog := src + fmt.Sprintf(`
; my $p = prototype(\&%s);
print defined($p) ? "P:$p\n" : "U\n";`, sub)
    out, err := exec.CommandContext(ctx, perl, "-e", prog).Output()
    if err != nil {
        return "", false, err
    }
    line := strings.TrimSpace(string(out))
    if line == "U" {
        return "", false, nil
    }
    return strings.TrimPrefix(line, "P:"), true, nil
}

// Compiles is the cheap yes/no gate.
func Compiles(ctx context.Context, perl, path string) bool {
    return exec.CommandContext(ctx, perl, "-c", path).Run() == nil
}

// Deparse returns perl's re-emission of the program, driven to its fixpoint.
//
// Deparse output is idempotent at the second pass (measured), so two rounds
// normalise away whitespace, optional parens and terminator style.
func Deparse(ctx context.Context, perl, src string) (string, error) {
    once, err := deparseOnce(ctx, perl, src)
    if err != nil {
        return "", err
    }
    return deparseOnce(ctx, perl, once)
}
```

Splitting the op body is fiddly but mechanical:

```go
func splitOpBody(s string) (name, arg, flags string) {
    if i := strings.IndexAny(s, "[ "); i >= 0 {
        name = s[:i]
        rest := s[i:]
        if strings.HasPrefix(rest, "[") {
            if j := strings.Index(rest, "]"); j >= 0 {
                arg, rest = rest[:j+1], rest[j+1:]
            }
        }
        flags = strings.TrimSpace(rest)
        return name, arg, flags
    }
    return s, "", ""
}
```

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

Measured: 100 sequential `perl -c` invocations take **0.88 s wall** — ~8.8 ms
each, dominated by process spawn. `B::Concise` runs ~25 ms. Naively, the 620
perl5 `.t` files under Concise cost ~15 s single-threaded. That is already
tolerable, but the corpus grows and the fixture set grows faster.

Four techniques, in order of payoff:

1. **Cache on content hash.** Perl's answer for a given (perl version, source
   bytes) never changes. Key a `testdata/oracle_cache/<sha256>.json` file on
   `sha256(perlVersion + "\x00" + src)`. Commit the cache. CI then runs zero
   perl processes on unchanged inputs, and the cache doubles as a review
   artifact — a diff in the cache is a diff in what perl said, which is
   exactly what a reviewer wants to see.

2. **Parallelise with `t.Parallel()` + a semaphore.** Process spawn is the
   cost, and it parallelises linearly to core count.

3. **Batch per file, not per construct.** One `perl -MO=Concise` run per
   source file, then extract every claim from the one optree. Do not spawn
   per assertion.

4. **Tier by build.** `go test -short` runs the fixture set only (fast, ~200
   cases, sub-second from cache). The full corpus sweep runs on a `corpus`
   build tag, nightly and pre-merge.

```go
func TestOracleCorpus(t *testing.T) {
    if testing.Short() {
        t.Skip("corpus sweep: run without -short")
    }
    perl, err := exec.LookPath("perl")
    if err != nil {
        t.Skip("perl not found in PATH — the parse oracle runs a real interpreter")
    }

    sem := make(chan struct{}, runtime.NumCPU())
    for _, path := range corpusFiles(t) {
        t.Run(path, func(t *testing.T) {
            t.Parallel()
            sem <- struct{}{}
            defer func() { <-sem }()
            checkAgreement(t, perl, path)
        })
    }
}
```

Skipping when perl is absent — rather than failing — matches
`precision_test.go` and keeps the suite runnable on a machine without a
toolchain, which for a *version manager* is a realistic case.

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
numbers are ratchets. This ordering is deliberate — it lets you ship a parser
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

**`t/op/signatures.t` (1,613 lines)** and **`t/op/for-many.t`** — see §7.3.4,
these two contain post-5.42 syntax.

### 7.3.3 `t/perf/opcount.t` — a pre-built optree oracle

42K of assertions of the form "this program compiles to exactly these ops in
these counts." Perl's own maintainers wrote an optree-shape test suite, and
its assertions are a **ready-made expected-value table for the parse oracle**
— already reviewed, already maintained, already correct. Mine it early. It is
the highest-value 42K in the tree for this chapter's purposes.

### 7.3.4 Measured parseability, and the version trap

Running `perl -c` over every file with the system perl:

| Directory | Files | `perl -c` clean | Real syntax errors |
|---|---:|---:|---:|
| `base/` | 9 | 9 | 0 |
| `comp/` | 25 | 25 | 0 |
| `opbasic/` | 5 | 5 | 0 |
| `cmd/` | 5 | 5 | 0 |
| `op/` | 228 | 142 | **2** |
| `re/` | 80 | 55 | 0 |
| `mro/` | 73 | 62 | 0 |
| `class/` | 12 | 1 | 0 |
| `io/` | 44 | 17 | 0 |
| `uni/` | 30 | 19 | 0 |

**Read that table carefully — it is mostly a measurement artifact, and the
artifact is the lesson.**

The `class/` row says 1 of 12. Not one of those eleven failures is a parse
failure:

```console
$ perl -c t/class/class.t
Can't locate Config.pm in @INC (@INC entries checked: ../lib) at class/class.t line 7.
```

Every perl5 test opens with the same prelude:

```perl
BEGIN {
    chdir 't' if -d 't';
    require './test.pl';
    set_up_inc('../lib');
    require Config;
}
```

`../lib/Config.pm` exists only in a **built** perl tree. The checkout is not
built. So `BEGIN` dies, and `perl -c` reports failure for a file that parses
perfectly. **`perl -c` conflates "does not parse" with "BEGIN blew up",** and
if you take its exit status at face value you will chase 200 phantom parse
bugs.

Three consequences for the harness:

1. **Classify `perl -c` stderr, never trust the exit status alone.** Match
   `/syntax error|Missing|Unmatched|not allowed/` for real parse failures;
   `/Can't locate/` is an environment failure and must be bucketed as
   *no-answer*, not *WRONG*.
2. **Either build the perl5 tree, or strip the prelude.** Building is
   cleanest. Stripping the `BEGIN` block mechanically is viable and much
   faster, and it is the same file for parse purposes.
3. **3 files carry `-T` on the shebang.** `perl -c` refuses these unless `-T`
   is also on the command line. Detect the shebang and pass it through.

**Now the two genuine syntax errors, which are the more interesting finding.**

```console
$ perl -c t/op/signatures.t
syntax error at op/signatures.t line 927, near "(:"
$ perl -c t/op/for-many.t
syntax error at op/for-many.t line 474, near "( \"
```

Line 927:

```perl
sub tnamed01 (:$alpha, :$beta) { "alpha=$alpha beta=$beta"; }
```

Named parameters in signatures. Line 474:

```perl
foreach my ( \@array ) ( ["A"], ["B"], ["C"] ) {
```

Refaliasing in a multi-var `foreach`. Neither exists in perl 5.42.

```console
$ perl -v
This is perl 5, version 42, subversion 0 (v5.42.0)

$ grep 'define PERL_VERSION' perl5/patchlevel.h
#define PERL_VERSION	45		/* epoch */
```

**The checkout is blead 5.45; the interpreter is 5.42.0.** The corpus is from
a newer Perl than the oracle.

This is not a nuisance, it is a design requirement:

> **The corpus and the oracle must be pinned to the same perl version, and
> the version must be recorded in the ratchet baseline.**

Otherwise a corpus update silently introduces syntax your oracle rejects, the
ratchet fires, and the failure looks like a parser regression when it is a
version skew. Record the perl version in the baseline header and fail loudly
on mismatch (§7.4). Given PVM *is* a version manager, resolving the corpus
perl through PVM itself is both natural and dogfooding.

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

T0–T2 is ~1,240 files and covers the ladder through M4. Do not vendor T5
until M5.

### 7.3.7 Vendoring and licensing

**perl5's tests are dual-licensed Artistic 1.0 / GPL 1.0+** (`perl5/Artistic`,
`perl5/Copying`). **PerlOnJava is the same dual license.**

The implication for a Go repo is not subtle:

> Copying GPL'd test files into a permissively-licensed Go repository puts
> the GPL's terms onto that copy, and — depending on how the build is
> arranged and whom you ask — potentially wider.

Three options:

| Option | Mechanics | Verdict |
|---|---|---|
| **Vendor** | Copy `.t` files into `testdata/` | Simple, hermetic, offline CI, **but imports the license question into the repo** |
| **Fetch** | CI clones perl5 at a pinned SHA into a scratch dir | No third-party code in-repo. Needs network in CI. Pinning is mandatory, and §7.3.4 shows why |
| **Reference** | Point at a local checkout via env var, skip if absent | Zero license exposure, zero CI value |

**Recommendation: fetch, pinned, with a graceful skip.**

```go
// corpusRoot resolves the external perl5 corpus.
//
// perl5's tests are Artistic/GPL. They are NOT vendored into this repo;
// CI clones a pinned revision into a scratch directory. Skipping rather
// than failing keeps the suite green on a machine without a checkout,
// which mirrors the perl-not-found skip in the oracle tests.
func corpusRoot(t *testing.T) string {
    t.Helper()
    if root := os.Getenv("PERL5_CORPUS"); root != "" {
        return root
    }
    t.Skip("PERL5_CORPUS unset; see docs/specs/perl-parser/07-conformance.md §7.3.7")
    return ""
}
```

Pin the corpus revision in a committed file next to the ratchet baseline:

```
# testdata/corpus.pin
perl5     94e5086608bd63e0e1a0f6e6f0f6c31e0e3b2a11   # blead, 2026-07-16
perlonjava <sha>
perl_version 5.45.0                                  # MUST match the oracle
```

**T0 fixtures are yours** — written from scratch, no license question, and
the only tier in `-short`. That is the tier that grows fastest anyway, since
every bug adds one.

---

## 7.4 The ratchet

### 7.4.1 The idea

perl-lsp's proven mechanism (`xtask/src/tasks/parser_matrix.rs`,
`ci/parse_errors_baseline.txt`): **a committed per-file pass/fail baseline
that CI compares against. Any file that regresses fails the build. Any file
that improves requires a baseline update in the same commit.**

The whole value is in one property: **the score cannot silently go down.**
Perl parsers regress constantly, because fixing construct X reorders lexer
state and breaks construct Y three files away. Without a ratchet you discover
that months later; with one you discover it in the PR that caused it.

perl-lsp's `parser_matrix.rs` also carries a **failure taxonomy** worth
copying — a category per failure, with a priority:

```rust
const CATEGORY_TAXONOMY: &[(&str, &str, &str)] = &[
    ("ModernFeature", "P1", "class/try/catch/field/method keywords"),
    ("QuoteLike",     "P2", "q/qq/qw/qx/qr, heredocs, strings"),
    ("Regex",         "P2", "m//, s///, tr///, patterns"),
    ("ControlFlow",   "P2", "given/when/default"),
    ("Dereference",   "P2", "->, postfix deref"),
    ("Subroutine",    "P2", "Signatures, prototypes"),
    ("General",       "P3", "Uncategorized"),
];
```

A raw count says "64 files fail." A taxonomy says "48 of 64 are QuoteLike."
The second is a work plan. Categorise by the **first** error's construct.

A caution from reading their tree: perl-lsp's `parser_ratchet.rs` is a
receipt-emitting scaffold whose measurements are disabled
(`"force-selected (scaffold only; measurements disabled)"`), and
`ci/parse_errors_baseline.txt` contains the single line `0`. The *idea* is
proven and correct; **their implementation of it is aspirational.** Copy the
design, not the code, and make sure yours actually measures something. The
Go version below is complete in ~80 lines.

### 7.4.2 Baseline format

Plain text, one line per file, sorted, greppable, diff-reviewable. Not JSON —
a reviewer must be able to read the diff.

```
# testdata/ratchet/baseline.txt
# Generated by: go test ./internal/parser -run TestRatchet -update
# perl: 5.45.0   perl5: 94e5086608   DO NOT EDIT BY HAND
#
# status  metric  category      path
ok        parse   -             t/base/cond.t
ok        parse   -             t/base/if.t
fail      parse   QuoteLike     t/base/lex.t
ok        agree   -             t/cmd/for.t
wider     agree   Subroutine    t/comp/proto.t
fail      roundtrip -           t/op/tr.t
```

Four statuses (`ok` / `wider` / `fail` / `skip`) across the metrics of §7.2.
The header pins the perl version — mismatch is a hard error, per §7.3.4.

### 7.4.3 The Go implementation

Idiomatic Go: `-update` regenerates, like `golden` files everywhere else.

```go
// ABOUTME: Ratchet gate — a committed per-file baseline that may only improve.
// ABOUTME: Any file regressing from ok to fail fails CI; improvements need a baseline update.

package parser_test

import (
    "bufio"
    "flag"
    "fmt"
    "os"
    "sort"
    "strings"
    "testing"
)

var update = flag.Bool("update", false, "rewrite the ratchet baseline")

const baselinePath = "testdata/ratchet/baseline.txt"

type entry struct {
    Status   string // ok | wider | fail | skip
    Metric   string // parse | agree | roundtrip | incremental
    Category string
    Path     string
}

func (e entry) key() string { return e.Metric + "\x00" + e.Path }

// rank orders statuses so a regression is a numeric decrease.
func rank(status string) int {
    switch status {
    case "ok":
        return 3
    case "wider":
        return 2
    case "fail":
        return 1
    default: // skip
        return 0
    }
}

func TestRatchet(t *testing.T) {
    got := measureCorpus(t) // map[key]entry

    if *update {
        writeBaseline(t, got)
        t.Log("baseline updated; review the diff before committing")
        return
    }

    want := readBaseline(t)
    assertPerlVersionMatches(t, want.perlVersion)

    var regressions, improvements []string
    for k, w := range want.entries {
        g, ok := got[k]
        if !ok {
            // A file left the corpus. Not a regression, but the baseline is stale.
            improvements = append(improvements, "removed: "+w.Path)
            continue
        }
        switch {
        case rank(g.Status) < rank(w.Status):
            regressions = append(regressions, fmt.Sprintf(
                "  %-12s %-24s %s -> %s", g.Metric, g.Path, w.Status, g.Status))
        case rank(g.Status) > rank(w.Status):
            improvements = append(improvements, fmt.Sprintf(
                "  %-12s %-24s %s -> %s", g.Metric, g.Path, w.Status, g.Status))
        }
    }
    for k, g := range got {
        if _, ok := want.entries[k]; !ok {
            improvements = append(improvements, "new: "+g.Path)
        }
    }

    if len(regressions) > 0 {
        sort.Strings(regressions)
        t.Errorf("PARSER REGRESSION — %d file(s) got worse:\n%s\n\n"+
            "If this is intentional, run: go test ./internal/parser -run TestRatchet -update",
            len(regressions), strings.Join(regressions, "\n"))
    }
    if len(improvements) > 0 {
        sort.Strings(improvements)
        t.Errorf("baseline is stale — %d file(s) improved:\n%s\n\n"+
            "Run: go test ./internal/parser -run TestRatchet -update",
            len(improvements), strings.Join(improvements, "\n"))
    }
}
```

**Failing on improvement, not only regression, is deliberate.** A stale
baseline is a baseline nobody trusts, and an untrusted ratchet is worse than
none — people start passing `-update` reflexively. Forcing the update into
the same commit keeps the diff honest and puts the score change in front of a
reviewer.

### 7.4.4 CI shape

```yaml
# .github/workflows/parser-conformance.yml
name: parser conformance
on: [push, pull_request]

jobs:
  fast:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      # Invariants. No perl, no corpus, no excuses. Seconds.
      - run: go test -short ./internal/parser/...

  conformance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }

      - name: pin corpus
        run: |
          read -r _ SHA _ < <(grep '^perl5 ' testdata/corpus.pin)
          git clone --filter=blob:none https://github.com/Perl/perl5 /tmp/perl5
          git -C /tmp/perl5 checkout "$SHA"

      - name: install the pinned perl
        run: |
          VER=$(awk '/^perl_version/{print $2}' testdata/corpus.pin)
          go run ./cmd/pvm install "$VER"      # dogfooding: PVM installs its own oracle

      - name: ratchet
        env:
          PERL5_CORPUS: /tmp/perl5/t
        run: go test ./internal/parser -run TestRatchet

      - name: oracle agreement
        env:
          PERL5_CORPUS: /tmp/perl5/t
        run: go test ./internal/oracle/...
```

Two jobs, deliberately. `fast` gates every push on the invariants and needs
nothing installed. `conformance` costs minutes and needs a perl. Never let
the slow job's flakiness (network, clone, install) mask a broken invariant.

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

## 7.8 The milestone ladder

Ordered so **early milestones deliver usable LSP value before full
conformance**. A parser that handles 40% of Perl correctly, with honest
"unknown" for the rest, powers real editor features. A parser at 95% that
lies about the other 5% powers nothing you can trust.

Every milestone carries the same three non-negotiables, at every level:

- **WRONG = 0** on the oracle. Non-negotiable from M0.
- **Lossless round-trip** on everything that parses. Non-negotiable from M0.
- **Ratchet green.** No file may regress.

### M0 — Lexer round-trips the core

**Corpus:** T2 core (perl5 `base cmd comp opbasic class`, 56 files) +
T0 fixtures.

| Metric | Target |
|---|---|
| Lossless round-trip | **100% of 56 files** |
| No panic on any input | 100% (fuzz, 1 M execs, zero crashers) |
| Position monotonicity | 100% |
| Parse | not required |

**No parser yet.** Tokenize and concatenate back to the identical bytes. It
sounds trivial and is not: it requires heredocs, POD, `__END__`, `q{}` with
nested delimiters, and `s///e` all to be *recognised* — not understood, but
delimited correctly.

**Ships:** syntax highlighting via semantic tokens. Genuinely useful, and it
is what an editor asks for first.

**Exit criterion:** `t/base/lex.t` round-trips. That one file is the gate,
and it is a real gate — see §7.3.2.

### M1 — Parses the core without error

**Corpus:** T2 core (56) + T1 graded, easy tier (~400 of PerlOnJava `unit/`).

| Metric | Target |
|---|---|
| Parse without error | **100% of T2 (56 files)** |
| Parse without error | **≥ 70% of T1-easy (~280 of 400)** |
| Round-trip | 100% of everything that parses |
| Oracle WRONG | **0** |
| Oracle exact | ≥ 40% of markers attempted |

Note the shape: parse-rate is a *ratchet*, WRONG is a *gate*. Constructs not
yet understood must be emitted as explicit `Unknown` nodes preserving their
source text — **not guessed at**. That is what keeps WRONG at zero while
coverage is still low, and it is the discipline the whole ladder rests on.

**Ships:** document symbols, folding ranges, brace matching.

### M2 — Incremental, and correct under edit

**Corpus:** M1's, plus fuzz-generated edit sequences.

| Metric | Target |
|---|---|
| Incremental == full re-parse | **100%. Zero counterexamples in 10 M fuzz execs** |
| Re-parse latency, 5 kLOC file, 1-char edit | < 10 ms p99 |
| Everything from M1 | held |

The prototype-edit and heredoc-edit cases of §7.7.3 must be explicitly
covered by committed seed corpora, not merely "probably reached by the
fuzzer."

**Ships:** a responsive LSP. This is the milestone that makes the thing feel
like a real editor integration rather than a batch tool.

### M3 — The oracle turns on, at scale

**Corpus:** T2 + T3 volume (perl5 `op mro uni lib test_pl`, 347 files).

| Metric | Target |
|---|---|
| Parse without error | **≥ 90% of the compilable corpus (≥ 526 of 584)** |
| Oracle agreement, exact | **≥ 80%** of marker sites |
| Oracle agreement, WRONG | **0** |
| Semantic round-trip (Deparse fixpoint) | ≥ 70% |
| `perf/opcount.t` assertions mined into fixtures | ≥ 200 |

**M3 is where this project overtakes PSC's current parser**, because it is
the first milestone that measures metric (b) at all. Reaching M3 means
something no Perl parser in this ecosystem has demonstrated: a *measured*
claim about parse correctness rather than parse coverage.

**Ships:** trustworthy go-to-definition and PSC type inference on a parse
that has been checked against perl.

### M4 — Prototypes, context, and the hard ambiguities

**Corpus:** M3's, plus T1 graded in full (986), plus `t/comp/proto.t`,
`t/op/signatures.t`.

| Metric | Target |
|---|---|
| Prototype agreement with `prototype(\&f)` | **≥ 95%** where statically determinable |
| `srefgen` marker agreement | **≥ 95%** |
| Oracle exact, all markers | **≥ 90%** |
| Oracle WRONG | **0** |
| T1 graded parse rate | ≥ 90% (≥ 887 of 986) |

The BEGIN-time cases are *expected* to land in `wider` / `no-answer`, and
that is correct behaviour, not failure. What must not happen is committing to
a wrong prototype-driven parse.

**Ships:** correct call-signature help, correct argument-context inference.

### M5 — Regex, Unicode, and the wild

**Corpus:** T4 regex (518) + T5 wild (top-200 CPAN by dependents).

| Metric | Target |
|---|---|
| Parse without error, T4 | **≥ 95%** |
| Parse without error, T5 CPAN | **≥ 95%** |
| Oracle exact, T4+T5 | ≥ 90% |
| Oracle WRONG | **0** |
| UTF-16 position correctness | 100% (fuzz) |

Regex *bodies* need not be parsed — Chapter 2 scopes them as opaque — but
delimiters, modifiers, `s///e` replacements and `(?{ })` blocks must be
handled. `re/pat_advanced.t` at 2,743 lines is the stress case.

**Ships:** a parser that survives real-world code.

### M6 — Conformance

**Corpus:** everything except `porting/` and `win32/`.

| Metric | Target |
|---|---|
| Parse without error, `t/op/` | **≥ 99% (≥ 226 of 228)** |
| Oracle exact, whole corpus | **≥ 95%** |
| Oracle WRONG | **0** |
| Semantic round-trip | ≥ 95% |
| `t/base/lex.t`, `t/comp/parser.t` | full parse agreement |
| `t/japh/` | parses |

**On the 99% figure — set it against a measured floor, not a hope.** The
fix-options plan records exactly this mistake being made and caught:

> The target of 0 in the brief was wrong and unreachable. Upstream's own
> parser errors on ~14% of `t/op`.

Before committing to 99%, run the differential harness (§7.5) and establish
what perl itself, tree-sitter, and PerlOnJava each achieve on the same
corpus. **A target above the best measured implementation is a target set by
guesswork.** Adjust M6 to `max(measured floors) + margin` once that number
exists — and record the measurement next to the number, so the next person
does not have to re-derive it.

### Ladder summary

| M | Theme | Headline gate | LSP value shipped |
|---|---|---|---|
| M0 | Lexer round-trip | 100% lossless on 56 core files; `t/base/lex.t` | Syntax highlighting |
| M1 | Parse the core | 100% of T2, ≥ 70% T1-easy, WRONG = 0 | Symbols, folding |
| M2 | Incremental | incremental == full, 10 M fuzz execs, < 10 ms p99 | Responsive LSP |
| M3 | Oracle at scale | ≥ 90% parse on 584 files, ≥ 80% exact, WRONG = 0 | Trustworthy navigation, PSC |
| M4 | Prototypes | ≥ 95% prototype agreement, ≥ 90% exact | Signature help, context |
| M5 | Regex + wild | ≥ 95% on 518 regex + 200 CPAN | Survives real code |
| M6 | Conformance | ≥ 99% `t/op`, ≥ 95% exact (**re-derive from floors**) | Reference implementation |

---

## 7.9 The first test to write

Before the milestone ladder, before the corpus, before any of it — **write
the prototype oracle test.**

```go
// ABOUTME: The first conformance test: proves a prototype changes the parse, and that we see it.
// ABOUTME: perl reports the difference in its optree; this test asserts we report it too.

func TestPrototypeChangesTheParse(t *testing.T) {
    perl, err := exec.LookPath("perl")
    if err != nil {
        t.Skip("perl not found in PATH — the parse oracle runs a real interpreter")
    }
    ctx := context.Background()

    const withProto = `sub f(\@){} my @a; f(@a)`
    const noProto   = `sub f{}     my @a; f(@a)`

    // 1. Establish ground truth by asking perl. No hand-written expectations.
    withOps, err := oracle.Concise(ctx, perl, withProto)
    require.NoError(t, err)
    noOps, err := oracle.Concise(ctx, perl, noProto)
    require.NoError(t, err)

    // 2. Assert the oracle itself discriminates. If this fails, the harness
    //    is broken and every result built on it is meaningless.
    require.True(t, oracle.Has(withOps, "srefgen"),
        "oracle broken: perl should emit srefgen for a \\@ prototype")
    require.False(t, oracle.Has(noOps, "srefgen"),
        "oracle broken: perl should not emit srefgen without a prototype")

    // 3. Now the actual subject: does OUR parser see the same difference?
    withTree := parser.Parse([]byte(withProto))
    noTree := parser.Parse([]byte(noProto))

    assert.True(t, callPassesReference(withTree, "f"),
        "we parsed f(@a) as a list where perl took a reference — WRONG, not merely imprecise")
    assert.False(t, callPassesReference(noTree, "f"),
        "we parsed f(@a) as a reference where perl took a list")
}
```

**Why this one, ahead of everything else.**

It is nine lines of Perl and it establishes the entire methodology. It proves
the oracle harness works (step 2 is a self-check — the precision test learned
that lesson the hard way, when an injected fault passed because the check was
too weak). It tests the hardest thing a static Perl parser does. It fails
loudly on the current PSC parser, which is the point: it converts a known-but
-unmeasured blind spot into a red test.

And it is the smallest possible thing that answers the question this chapter
exists to answer: **not "did it parse?" but "did it parse the way perl did?"**

Everything else in this chapter is scaffolding around that one question.

---

## 7.10 Checklist

- [ ] `internal/oracle` package: `Compiles`, `Prototype`, `Concise`, `Deparse`
- [ ] Oracle self-check test (§7.9) — **first**
- [ ] Content-hash cache under `testdata/oracle_cache/`, committed
- [ ] Round-trip invariant test over every corpus file
- [ ] `go test -fuzz` targets: lexer, heredoc, quote-operator, incremental
- [ ] Lexer forward-progress assertion (§7.6.2, invariant 4)
- [ ] Ratchet baseline + `-update` flag + `TestRatchet`
- [ ] `testdata/corpus.pin` with perl version, checked at test start
- [ ] `perl -c` stderr classifier (syntax vs `@INC` vs `-T`) — §7.3.4
- [ ] CI split: `fast` (invariants, no perl) and `conformance` (corpus, pinned perl)
- [ ] Failure taxonomy in the ratchet, categories from perl-lsp's list
- [ ] Differential harness with the "perl always wins" triage rule
- [ ] Measured floors for tree-sitter and PerlOnJava before fixing M6's target
