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
`srefgen`/`entersub` counts, `ref_lines`, and every prototype in scope.

## What population is measured

`-MO=Concise,-exec` dumps `PL_main_root` only. A named sub, an anonymous sub
and a `BEGIN` block are each a separate CV and never appear in it, so a count
taken from that output silently omits every prototype-forced reference inside
a sub body -- which is where most of the corpus keeps its calls:

```
$ perl -MO=Concise,-exec -c subbody.pl | grep -c srefgen                    # 1
$ perl -MO=Concise,-exec,-main,-stash=main -c subbody.pl | grep -c srefgen  # 2
```

So the op sequence still comes from Concise, but `srefgen` and `sites` come
from a probe that walks the optree with `B` directly: `main_root`, the
`BEGIN`/`UNITCHECK`/`CHECK`/`INIT`/`END` arrays (perl frees `BEGIN` blocks
after running them unless `B::save_BEGINs` is called first, which the probe
does before any `BEGIN` in the file), every named sub in every package under
`main::`, and each anonymous sub through the `anoncode` op that closes over it.
`-stash=main` was not enough: it cannot reach anonymous subs or `BEGIN` blocks
at all.

`sites` is the per-site form: per marker (`srefgen`, `rv2hv`, `match`,
`readline`, `anonhash`), one entry per op, giving the line of the statement
it belongs to. The same walk finds all five; the reference population is
described below and the other four in the kinds table under "Measuring
another implementation". Perl attributes every op to the
nearest preceding `nextstate`, and a `nextstate` names its statement's FIRST
line (measured on multi-line calls, hash literals and `if`/`elsif`/`for`/
`while` conditions). Ops perl compiled from another file -- the subs a corpus
file pulls in from `t/test.pl` -- name that file in their COP and are not
counted. The comparison is per statement for the same reason: a whole-file
total let a surplus at one statement cancel a deficit at another.

The population is "what the source could have written a backslash for":

| form | op | counted |
|---|---|---|
| `\@a` `\%h` `\&f` `\$x` `\(&f)` `\-1`, every `\`-prototype argument | `srefgen` | yes |
| `\(@a)` `\(@a,@b)` `\my($x,$y)` | one `refgen` per list | yes, once |
| `\1` `\"x"` `\'s'` | folded to `const[IV \1]`, no op | no |
| `goto &NAME` | `srefgen` under `goto` | no |
| `sub { ... }`, `f { ... }` under an `&` prototype | `anoncode`, no srefgen in 5.42 | no |

`goto &NAME` is excluded because perly.y parses `&NAME` as an entersub term
and `newLOOPEX` wraps it in a `REFGEN`: the srefgen is goto's rewrite, not a
reference the parse took at a call site. Counted, every `goto &sub` in the
corpus would score WRONG for a parse both sides agree on.

What the walk does not reach, and is a documented ceiling: `ADJUST` blocks
and `field` initialisers (`feature 'class'` keeps them outside the stash),
`format` bodies, and anything compiled by a string `eval` at run time. And a
folded `\1` on the SAME statement as a prototype-driven reference the subject
missed still sums to zero on that statement, exactly as the whole-file
version did across the file.

## Measuring another implementation

The runner measures our own tree-sitter parser by default. Any other
implementation is measured by supplying a `Subject`: a command the runner
invokes once per file, from the corpus directory, with the file's relative
path appended, within the run's per-file timeout.

```go
parseoracle.Run(ctx, files, parseoracle.RunOptions{
    Dir:     shimT,
    Subject: &parseoracle.Subject{Command: []string{"perl", "testdata/perl_subject.pl"}},
})
```

The subject prints one JSON object on stdout and exits 0:

```json
{"ok": true, "call_sites": [{"kind": "reference", "took_reference": true}]}
```

| Field | Meaning |
|---|---|
| `ok` | Required. `false` for a file the subject rejects; that scores `no-answer`. |
| `call_sites` | Optional. Each site the subject decided something at, and what: for a call, `took_reference` and `unresolved`; for any other kind, the kind itself is the decision. Omitted or `null` means "not computed"; `[]` means "none found". |
| `call_sites[].kind` | Absent for a call. `"reference"` for a `\` the source wrote outside a call. `"hash"`, `"match"`, `"readline"`, `"anonhash"` for the other four parse decisions the harness measures (below). |
| `call_sites[].line`, `end_line` | The span of the statement the site is in. Perl attributes an op to its statement and no finer, so a site without a line accounts for nothing. |
| `call_sites[].unresolved` | The subject could not settle this site. On a call it is the prototype hedge; on any other kind it is a hedge about that kind. |
| `prototypes` | Optional. Sub name to prototype string. Same omitted/null/empty rule. |
| `declined`, `declined_reason` | The subject dropped source it could not handle. Scores `no-answer`. |

**Five questions, five kinds.** Each kind is the subject's side of one
marker op in perl's optree (spec §7.1.2), and each is scored by presence per
statement: every site perl reports must be covered by a site of the matching
kind in the same statement. Perl's optree is post-peephole and only ever
loses marker ops — `$h{a}` folds to `multideref`, a lexical `%h` is `padhv`,
a match under `if (0)` is discarded — so a subject that reports a construct
perl folded away costs nothing. What a subject must never do is stay silent
about a construct perl kept.

| Kind | Report it for | Perl op | The other reading |
|---|---|---|---|
| a call (no kind) + `took_reference` | `f(@a)` under a `\@` prototype, `f(\@a)` | `srefgen` | a flattened list |
| `"reference"` | `\@a`, `\%h`, `\&f`, `\(@a, @b)` outside a call | `srefgen`, `refgen` | — |
| `"hash"` | `%h`, `%$r`, `%{...}`, `->%*`, `$h{k}`, `$r->{k}`, `@h{...}`, `%h{...}` | `rv2hv` | `%` as modulus |
| `"match"` | `/x/`, `m//`, and `=~`/`!~` against anything but `s///` or `tr///` | `match` | `/` as division |
| `"readline"` | `<FH>`, `<$fh>`, `<>`, `<<>>`, `readline(...)` | `readline`, `rcatline` | `<...>` as a glob |
| `"anonhash"` | `{ a => 1 }`, `{}`, `+{ ... }` where the subject read a hash constructor | `anonhash`, `emptyavhv` flagged as a hash | `{` as a block |

A subject answers a question by reporting at least one site of that kind
somewhere in the file; a file with sites of the kind but none at a statement
where perl built the op is `WRONG` there. A subject that reports no site of a
kind anywhere has not answered that question and scores `no-answer` for it,
so an implementation can adopt the kinds one at a time — `ok` alone, then
calls, then hashes — and be measured on what it has.

**Exit status is not a verdict.** Unlike `perl -c`, a subject does not exit
non-zero to say "not Perl"; it says `"ok": false`. A non-zero exit, a timeout,
non-JSON output, output over 16 MiB, or a document without `ok` is a runner
error: the file leaves the denominator, and the subject's stderr is repeated
in the report so the cause is readable. On a timeout the subject's whole
process group is killed, so a shell wrapper that forks its real work does not
leak it.

`testdata/perl_subject.pl` is a working subject — perl's own parser answering
through the contract — and `TestPerlSubjectMeasuresFixture` runs it over the
fixture corpus, where it scores `exact` on every file it compiles.
`docs/specs/perl-parser/07-conformance.md` §7.5.5 is the contract's
specification.

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
Most of the corpus's 19 directories are at 100%; `porting/` (43.2%) tests
perl's own source tree rather than the language.

The per-directory table in `docs/specs/perl-parser/00-findings.md` §0.5 lists
14 of those 19 rows and so does not sum to its own total; an earlier claim
here of "nine of fourteen directories" counted rows in that partial table
rather than directories in the corpus. Neither figure is recomputed by any
test — see §0.5 for what is and is not bound.

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

## In CI

`.github/workflows/parse-fidelity.yml` runs `TestRatchetCorpus` on every push
and PR to `pu`. It installs the pinned interpreter, checks out `perl5` at the
pinned revision, and caches both.

It never passes `-parseoracle.update`. A job that rewrites the baseline it is
checking against launders every regression into the baseline and can never
fail — which is how perl-lsp arrived at a parser ratchet whose baseline file
contains the single line `0`. Re-baselining is a deliberate local act that
lands in the same commit as the change that moved it.

A workflow cannot be run before it is merged, so the file is thin glue over
parts that are tested locally. The gate is two steps. The sweep writes
`$PARSEORACLE_RECEIPT` only after the ratchet has passed -- files swept,
baseline rows, the baseline's sha256, the pin, and a digest of the verdicts
-- and `cmd/receipt` then recomputes every field from the checked-out tree
and refuses fewer than 310 rows. Anything that makes `go test` exit 0
without running the sweep (`-skip`, `-list`, `-count=0`, `-exec /bin/true`,
a second `-run`, `GOFLAGS`, a skipped corpus) leaves no receipt, and the
second step fails on its absence.

`gate_test.go` runs in the normal suite and does not read those two steps;
it executes them as committed -- script, shell, and every `env:` level --
over a five-file shim, and requires a five-file receipt on the matching
baseline, a non-zero exit on a baseline that disagrees, and a non-zero exit
with no corpus. `ci_test.go` holds what execution cannot see: that the
workflow parses, that no job or step is `continue-on-error`, that it reads
`corpus.pin` through `.github/scripts/parseoracle-pin.sh` rather than
carrying a second copy of the revision, and that it checks the interpreter
before spending a sweep on a runner that cannot produce a comparable answer.

What no test here can close: `if: false` or a trigger that never fires
leaves a job that is skipped, and a skipped job counts as success until the
`pu` ruleset requires the "Fidelity ratchet (pinned perl + pinned corpus)"
status check. The workflow header spells that out.

The job spawns perl ~1860 times on every push — `parse_facts.pl` runs `perl
-c` twice per file (`capture` and `capture_with_end`) under one outer perl,
so 620 files cost roughly three processes each. No sweep passes a
`*Cache` — the cache above is built and unit-tested but wired to nothing.
Tracked as `01a095cb`. It is a speedup, not a correctness fix, so the workflow
ships without it.

### Why the sweep gets a longer per-file timeout

`DefaultTimeout` is 60s, calibrated here. Timed across the corpus, the most
expensive file is `op/pack.t` at **37 seconds of CPU, unloaded** — B::Concise
walks a very large optree, and that cost is the file's rather than the load's.
62% of the budget with nothing else running is not enough margin for a shared
runner.

A file that overruns is recorded as a runner error, which moves it out of its
baseline bucket, and the ratchet reports a regression caused by nothing but
load. `$PARSEORACLE_TIMEOUT` raises the bound to 8m in CI so that a timeout
still means what it should: `re/pat_psycho.t` calls `watchdog(5 * 60)` from a
`BEGIN` block, which is why it is baselined as `error`.

That file does **not** wedge on any budget, as this section previously
claimed. `t/test.pl:1983-2011` forks a watchdog that holds the pipe for
exactly `$timeout` (300s) and then `_exit`s; at two `perl -c` per file that is
~10 minutes, so it errors under the 8m budget and would enter a real bucket
under a budget past ~11m — and then fail the ratchet as a regression caused
by nothing. **This row is timeout-coupled: raising `PARSEORACLE_TIMEOUT`
above ~11m requires re-baselining it in the same commit.**
