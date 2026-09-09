<!-- ABOUTME: The parse oracle — measures whether a parser agrees with perl, not just whether it errors. -->
<!-- ABOUTME: Extracts perl's own parse decisions (optree, prototypes) as JSON ground truth. -->

# Parse Oracle

Perl reports how it parsed something. That makes parser *fidelity* measurable
instead of a matter of opinion.

```
$ perl -MO=Concise,-exec -e 'sub f(\@){} my @a; f(@a)'   # srefgen PRESENT
$ perl -MO=Concise,-exec -e 'sub f{}    my @a; f(@a)'    # srefgen ABSENT
```

Identical source shape, different parse, and the optree says which. Verified —
this also holds when the prototype is installed by a string `eval` inside
`BEGIN`, which is the undecidability case.

## Why this exists

Nothing has ever checked whether our tree is the tree perl builds — only
whether it contains ERROR nodes. Those are different questions:

| Metric | Question | Status |
|---|---|---|
| Coverage | Does it parse without error? | Measured today |
| **Fidelity** | **Does it parse the way perl does?** | **Unmeasured** |

A file can score perfectly on coverage and still be parsed wrong. `f @a` read
as a list where perl took a reference produces a clean tree and a false one.

## Usage

```sh
perl testdata/parse_facts.pl FILE            # JSON on stdout
ORACLE_CHDIR=/path/to/t perl testdata/parse_facts.pl FILE
```

Emits `ok`, the linear op sequence (`-exec` order, so it diffs cleanly),
`srefgen`/`entersub` counts, and every prototype in scope.

## What the optree can and cannot answer

The optree is captured **after** the peephole optimiser, so it is not a
faithful record of the parse tree:

```
$ perl -MO=Concise,-exec -e 'my $x = 1 + 2'
3  <$> const[IV 3] s/FOLD
```

The `add` op is gone — constant folding happened before Concise saw it. Two
rules follow.

**Compare for the presence of a marker op, never for op counts or an exact
sequence.** `srefgen` works as a signal because nothing folds it away. A
whole-tree diff would report differences that are the optimiser's, not the
parser's.

**The `s/FOLD` flag is a gift.** Perl marks folded constants rather than
silently rewriting them, so folding is detectable when it matters.

Concise output is byte-identical across runs (verified by md5 on repeated runs
of the same input), so results are safe to cache on a content hash.

## The cache

A cold sweep of all 620 files costs **~6 minutes** across 24 workers, which is
too slow to gate a commit.

**Perl is not where that time goes.** Timing the two phases separately over
619 files (excluding `re/pat_psycho.t`, which hangs the runner either way):

| Phase | Cold | Warm |
|---|---:|---:|
| Oracle (perl) | ~33-39 s | **0.8 s** |
| Our tree-sitter parser | 5 m 28 s | 5 m 28 s |
| Whole sweep | ~6 m 50 s | 6 m 18 s |

The cache does its job completely — it removes essentially all of the perl
cost — and the sweep is still six minutes, because **~87% of it is our own
parser**, which cannot be cached: its output is the thing under test, and a
cached verdict would hide exactly the movement a ratchet exists to detect.

This corrects the attribution the cache was commissioned under. `B::Concise`
was measured at ~600 ms on `t/op/sub.t` and assumed to dominate; across the
whole corpus the oracle averages ~55 ms/file, and our parser averages ~530 ms.
Making the sweep commit-gating is therefore a parser-performance problem, not
an oracle-caching one.

The cache is still worth keeping: it makes the oracle phase free, so it no
longer contributes to whatever the sweep eventually costs, and it removes
620 process spawns from CI.

```go
cache, err := parseoracle.OpenCache("testdata/oracle_cache")
report, err := parseoracle.Run(ctx, files, parseoracle.RunOptions{
    Dir: shimT, Cache: cache,
})
```

The key is `sha256` over everything that can change perl's answer:

| Component | Why it is in the key |
|---|---|
| source bytes | An edited file is a different question. |
| interpreter `$]` | A different perl parses differently. |
| corpus revision | A re-baselined corpus must not inherit old answers. |
| working directory | `Dir` changes what `@INC` finds. |
| taint mode | `-T` changes what perl accepts outright. |
| **`parse_facts.pl`'s hash** | The script is the measuring instrument. Omit it and a change to what is measured silently returns facts the current script would never have produced. |

**Only `Facts` are cached, never verdicts.** Facts depend on perl, which does
not change between commits. Verdicts depend on our parser, which changes
constantly, and a cached verdict would hide exactly the movement a ratchet
exists to detect.

There is no eviction, size cap, or TTL. One entry per (content, identity)
means the cache is bounded by the corpus. A corrupt entry is treated as a miss
rather than an error, so a hand-edited or truncated file costs one perl
invocation and can never break a build.

Entries are indented JSON naming the source they answer for, so the cache
reads as a diff of what perl said.

### Generated locally, not committed

The conformance plan assumed the cache would be committed, on the reasoning
that CI would then run zero perl processes. The measurement above removes most
of that argument: regenerating the whole cache from scratch costs ~35 seconds,
against 620 files and ~250 KB (compressed) of perl's opinions carried in the
repo forever and re-churned on every re-baseline of the interpreter or the
corpus pin.

Thirty-five seconds is not worth that. `testdata/oracle_cache/` is therefore
gitignored and regenerated on demand.

The diffability of an entry is not wasted by that choice — it is what lets you
`diff` two locally generated caches across a re-baseline and read what actually
changed in perl's answers, which is the review question the plan cared about.

## Running against perl's own test suite

perl's tests are **not** runnable in place. `t/test.pl:119` does `@INC = ()`
and then unshifts `../lib`, so they only compile inside a tree where `../lib`
is populated — `PERL5LIB` cannot help, since `@INC` is cleared after it is
read. 498 of the 620 files follow this convention.

Populate the shim from **both** library roots. `Config.pm` and every XS
module's `.pm` half live in the architecture-specific directory; leaving it out
costs ~170 files and makes the corpus look far harder than it is.

```sh
mkdir -p shim/lib && cp -r perl5/t shim/t

perl -e 'for (@INC) { print "$_\n" if -d && !/site_perl|vendor_perl/ }' \
  | while read d; do cp -rn "$d"/* shim/lib/ 2>/dev/null; done

cd shim/t && perl -c op/sub.t          # syntax OK
```

## Ground-truth ceiling

Measured with perl 5.42.0 and the shim above: **584/620 (94.2%)** compile.
Nine of fourteen directories are at 100%; `porting/` (43.2%) tests perl's own
source tree rather than the language.

Classifying by stderr rather than exit status — measured against the earlier
*incomplete* shim, when 185 files were failing — shows how little of it was
ever about Perl:

| Cause | Files |
|---|---:|
| Missing module / `@INC` | 167 |
| Version or feature skew | 1 |
| **Genuine syntax error** | **1** (`op/for-many.t`) |
| Other (exit status, timeouts) | 16 |

The completed shim recovers almost all of the first row, which is how 411/620
became 584/620.

**One file in 620 fails for a reason about the language** — `op/for-many.t`,
which uses `foreach my ( \@array ) (...)`, a blead-only refaliasing form that
5.42 cannot parse. The corpus here is blead 5.45 while the interpreter is
5.42.0, so pin both and treat a mismatch as a reason to re-baseline.

The harness must classify stderr. An exit status cannot tell "your parser is
wrong" from "this machine lacks `Config.pm`", and a ratchet built on exit
status encodes the second as though it were the first. A pass rate well under
94% is evidence the setup is broken, not that Perl is hard.

## These counts are bounds, not totals

`perl -c` reports one error and stops, so classification only ever sees each
file's **first** failure. Every number in the table above inherits that:

| Count | Bound | Why |
|---|---|---|
| Genuine syntax error | **LOWER** | A later parse failure is never reached |
| Missing module / `@INC` | **UPPER** | Some of those files fail again once fixed |

`t/op/signatures.t` is the worked example. It aborts at line 32 on a warnings
category 5.42 does not have — version skew — which hides a real parse failure
at line 927. Fix the environment and the file does not become a pass; it
becomes a different failure.

Any report built on these counts has to say so. Reading the syntax-error count
as a total is precisely how a measurement turns into a confident wrong
explanation.

## Pinning

`testdata/corpus.pin` records the interpreter's `$]` and the `perl5` revision.
`ReadPin` checks **both** — a pin that verifies one half is a pin that silently
drifts, and both halves have drifted here already.
