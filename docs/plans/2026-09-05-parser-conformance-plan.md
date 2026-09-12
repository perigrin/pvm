<!-- ABOUTME: Project plan for proving the Go Perl parser conforms: harness code, ratchet, CI shape, corpus logistics, milestone ladder. -->
<!-- ABOUTME: Split out of the specification's chapter 7; the spec defines what "correct" means, this plan schedules and scaffolds the measuring of it. -->

# Perl parser conformance plan

Split out of `docs/specs/perl-parser/07-conformance.md` on 2026-09-05. The
specification keeps the definitions: the parse oracle and its limits (spec
§7.1), the four metrics (§7.2), the corpus and its tiers (§7.3), differential
testing (§7.5), the fuzz invariants (§7.6) and the incremental-parse property
(§7.7). This plan holds everything that is a project decision rather than a
statement about Perl — code sketches, CI mechanics, licensing logistics, and
the milestone targets — because those change as the work progresses and a
specification should not.

Section numbers below are this plan's own; "spec §7.x" means the chapter.

---

## 1. The oracle harness in Go

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


---

## 2. Making the oracle fast enough for CI

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

## 3. Corpus logistics

### 3.1 Classifying `perl -c` output

Spec §0.5 measured that a bare `perl -c` exit status conflates "does not
parse" with "`BEGIN` blew up", and that almost every failure in perl5's suite
is environmental. Two consequences for the harness:

1. **Classify `perl -c` stderr, never trust the exit status alone.** Match
   `/syntax error|Missing|Unmatched|not allowed/` for real parse failures;
   `/Can't locate/` is an environment failure and must be bucketed as
   *no-answer*, not *WRONG*.
2. **3 files carry `-T` on the shebang.** `perl -c` refuses these unless `-T`
   is also on the command line. Detect the shebang and pass it through.

Spec §0.7 draws the third: the corpus and the oracle must be pinned to the
same perl version, and the version recorded in the baseline header (§4.2
below). A mismatch is a reason to re-baseline, not a regression.

### 3.2 Vendoring and licensing

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
| **Fetch** | CI clones perl5 at a pinned SHA into a scratch dir | No third-party code in-repo. Needs network in CI. Pinning is mandatory, and spec §0.7 shows why |
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
    t.Skip("PERL5_CORPUS unset; see docs/plans/2026-09-05-parser-conformance-plan.md §3.2")
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

## 4. The ratchet

### 4.1 The idea

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

### 4.2 Baseline format

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

Four statuses (`ok` / `wider` / `fail` / `skip`) across the metrics of spec §7.2.
The header pins the perl version — mismatch is a hard error, per spec §0.7 and §3.1 above.

### 4.3 The Go implementation

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

### 4.4 CI shape

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

## 5. The milestone ladder

Corpus tiers T0–T6 are defined in spec §7.3.6; the metrics and buckets in
spec §7.1.5 and §7.2.

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
and it is a real gate — see spec §7.3.2.

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

Spec chapter 6 §6.11 ships the full-reparse server first (step 3) and adds
the token cache and damage/repair only after measurement (steps 6-7). This
milestone is the gate that flips step 7's flag; it does not move that work
earlier.

| Metric | Target |
|---|---|
| Incremental == full re-parse | **100%. Zero counterexamples in 10 M fuzz execs** |
| Re-parse latency, 5 kLOC file, 1-char edit | < 10 ms p99 |
| Everything from M1 | held |

The prototype-edit and heredoc-edit cases of spec §7.7.3 must be explicitly
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

Before committing to 99%, run the differential harness (spec §7.5) and establish
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

## 6. The first test to write

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

And it is the smallest possible thing that answers the question spec chapter 7
exists to answer: **not "did it parse?" but "did it parse the way perl did?"**

Everything else in chapter 7, and all of this plan, is scaffolding around that one question.

---

## 7. Checklist

Acceptance list. Build order is spec chapter 6 §6.11; the test of §6 is its
step 0.

- [ ] `internal/oracle` package: `Compiles`, `Prototype`, `Concise`, `Deparse`
- [ ] Oracle self-check test (§6) — **first**
- [ ] Content-hash cache under `testdata/oracle_cache/`, committed
- [ ] Round-trip invariant test over every corpus file
- [ ] `go test -fuzz` targets: lexer, heredoc, quote-operator, incremental
- [ ] Lexer forward-progress assertion (spec §7.6.2, invariant 4)
- [ ] Ratchet baseline + `-update` flag + `TestRatchet`
- [ ] `testdata/corpus.pin` with perl version, checked at test start
- [ ] `perl -c` stderr classifier (syntax vs `@INC` vs `-T`) — §3.1
- [ ] CI split: `fast` (invariants, no perl) and `conformance` (corpus, pinned perl)
- [ ] Failure taxonomy in the ratchet, categories from perl-lsp's list
- [ ] Differential harness with the "perl always wins" triage rule
- [ ] Measured floors for tree-sitter and PerlOnJava before fixing M6's target
