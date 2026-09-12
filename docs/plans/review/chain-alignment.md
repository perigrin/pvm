<!-- ABOUTME: Alignment review of the parser-fidelity-harness issue chain against conformance-plan §1-§4 and the measured findings. -->
<!-- ABOUTME: Coverage gaps, scope creep, contradictions with 00-findings.md, and dependency correctness; read-only, edits nothing. -->

# Chain alignment review: parser-fidelity-harness

Reviewed 2026-09-08. Inputs: the five issues in the chain; the milestone
context (`docs/plans/2026-09-06-fidelity-harness-milestone.md`); conformance
plan §1-§4 (`docs/plans/2026-09-05-parser-conformance-plan.md`); measured
findings (`docs/specs/perl-parser/00-findings.md`). Plan §5 (M0-M6) is out
of scope by the user's decision and is not counted against the chain.

Every claim below was checked by reading the cited text or by running the
command shown. A shim was built in the scratchpad from the README recipe to
verify the AC-named files actually compile.

Issue IDs, abbreviated:

| Short | ID | Title |
|---|---|---|
| EXPORT | `01a076f1-f779` | Export the parse oracle as a callable API |
| CORPUS | `01a076f2-5fa2` | Build a reproducible, version-pinned corpus |
| COMPARE | `01a076f2-c374` | Bucket results as exact / wider / WRONG / no-answer |
| RUN | `01a076f3-28c6` | Run the oracle across the corpus and report fidelity |
| RATCHET | `01a076f3-84b8` | Freeze verdicts as a ratchet that can actually fail |

## Summary

| Category | Count |
|---|---:|
| Coverage gaps | 6 (2 blocking, 4 minor) |
| Scope creep | 0 |
| Contradictions with measured findings | 0 |
| Deviations from the plan's stated design | 1 |
| Dependency findings | 1 false edge |

**Verdict: not ready as written.** Two gaps (G1, G2) are interface holes
between EXPORT and RUN that would surface as a blocked RUN issue mid-chain:
the exported `Facts` type carries neither the compile stderr that `Classify`
consumes nor a way to set the working directory the corpus requires. Both
are cheap to fix in EXPORT, which is the head of the chain and has not
started. The rest are small and can be folded in as steps or ACs.

---

## Coverage gaps

### G1 (blocking) — `Facts` has no stderr, so RUN cannot call `Classify`

**Plan §3.1:**

> Classify `perl -c` stderr, never trust the exit status alone.

**00-findings §0.5:**

> It must classify stderr: an exit status cannot distinguish "your parser is
> wrong" from "this machine lacks `Config.pm`".

**CORPUS step 5** defines `Classify(stderr string) FailureKind`. **RUN step
2** says "skipping files `Classify` marks Environmental". But RUN gets its
per-file result from EXPORT's `Ask`/`AskFile`, and **EXPORT step 2** fixes
the result type as:

> an exported `Facts` struct (OK, Ops, OpCount, Srefgen, Entersub, Prototypes)

with the AC "`Facts` carries the fields `parse_facts.pl` emits". Those are
the only fields the script emits. `parse_facts.pl` runs `perl ... -c FILE
2>&1`, sets `ok` from `$?`, parses op lines out of the combined text and
discards the rest. Verified on a file that does not compile:

    $ ORACLE_CHDIR=$PWD perl parse_facts.pl op/for-many.t
    {"entersub":0,"file":"op/for-many.t","ok":0,"op_count":0,"ops":[],"prototypes":{...},"srefgen":0}

`ok:0` and nothing else. The message that says *why* — the only thing that
separates Environmental from SyntaxError from VersionSkew — never reaches
Go. RUN has no input to hand `Classify`, and the AC "a parser is never
blamed for a missing `Config.pm`" cannot be met through the oracle path.

**Action:** in EXPORT, have `parse_facts.pl` emit the compile diagnostics
(e.g. a `stderr` field, populated when `ok` is 0) and add the field to
`Facts`. This means dropping "existing, unchanged" from EXPORT's paths table
for `parse_facts.pl`. The alternative — RUN spawning a second `perl -c` per
file just to capture stderr — violates plan §2 item 3 ("one `perl
-MO=Concise` run per source file... Do not spawn per assertion").

### G2 (blocking) — `AskFile` has no working-directory parameter, so it cannot compile corpus files

**Milestone context:**

> `t/test.pl:119` clears `@INC`, so these tests only compile inside a tree
> where `../lib` is populated

**README (the shim recipe the milestone points at):**

    ORACLE_CHDIR=/path/to/t perl testdata/parse_facts.pl FILE
    ...
    cd shim/t && perl -c op/sub.t          # syntax OK

The corpus files begin `chdir 't' if -d 't'; require './test.pl'`, so perl
must be run with cwd = `shim/t` and the file given as `op/sub.t`.
`parse_facts.pl` supports this only via the `ORACLE_CHDIR` env var.

**EXPORT step 2** defines the API as `Ask(ctx, src []byte)` and
`AskFile(ctx, path string)` — no directory, no options. Its AC says the
oracle "locates `parse_facts.pl` relative to the package, not the caller's
working directory, so `AskFile` works from any test", which is about finding
the *script*, not about where perl runs. Nothing in EXPORT, CORPUS or RUN
says how `ORACLE_CHDIR` (or an equivalent) gets set for a corpus file. As
specified, `AskFile(ctx, "shim/t/op/sub.t")` runs perl from the Go test's
cwd and every `test.pl`-using file (498 of 620 per §0.5) reports a `BEGIN`
failure that is about `@INC`, not syntax — exactly the failure mode §0.5
warns about.

**Action:** give `AskFile` a directory argument (`AskFile(ctx, dir, rel
string)`) or an options struct, and have it set `ORACLE_CHDIR`. Add an
EXPORT AC: "`AskFile` compiles `t/op/sub.t` from a shim tree." That AC also
gives the EXPORT→CORPUS edge a reason to exist (see D1).

### G3 — `-T` shebang files are not handled

**Plan §3.1 item 2:**

> 3 files carry `-T` on the shebang. `perl -c` refuses these unless `-T` is
> also on the command line. Detect the shebang and pass it through.

**Plan §7 checklist:** "`perl -c` stderr classifier (syntax vs `@INC` vs
`-T`)".

No issue mentions `-T`. Verified the three files and the failure:

    $ grep -l '^#!.* -T' -r t --include='*.t'
    t/op/taint.t  t/op/utftaint.t  t/perf/taint.t
    $ perl -c op/taint.t
    "-T" is on the #! line, it must also be used on the command line at op/taint.t line 1.
    $ perl -T -c op/taint.t
    op/taint.t syntax OK

Through the oracle path these three files return `ok:0` (verified for
`op/taint.t`) and, with G1 fixed, would classify as Other. Three files out
of 584 is a small number, but it is a stated §3.1 requirement and the fix is
a few lines in `parse_facts.pl` (read line 1, add `-T` to the flags).

**Action:** add a step to CORPUS or EXPORT; the natural home is
`parse_facts.pl`'s `capture` since that is where the flags live. Add an AC
naming `t/op/taint.t`.

### G4 — Corpus location, licensing, and the graceful skip are unstated

**Plan §3.2:**

> **Recommendation: fetch, pinned, with a graceful skip.** ... perl5's tests
> are Artistic/GPL. They are NOT vendored into this repo; CI clones a pinned
> revision into a scratch directory. Skipping rather than failing keeps the
> suite green on a machine without a checkout

**Plan §4.4** has explicit CI steps: `git clone` perl5 and `checkout "$SHA"`
from `corpus.pin`, then `go run ./cmd/pvm install "$VER"` for the pinned
perl.

CORPUS step 1 takes `BuildShim(dir)` but never says where `dir` comes from
(`PERL5_CORPUS` in the plan) or what happens when it is absent. RUN step 5
makes the full run opt-in but does not say the corpus is external. RATCHET
step 7 says "CI workflow runs the full corpus on push, caching the corpus
build" without the clone-at-pin or install-pinned-perl steps.

The perl install matters more than it looks. RATCHET step 6 makes a pin
mismatch report "re-baseline needed". `ubuntu-latest` ships a perl that is
not 5.42.0. If the CI job does not install the pinned perl, every CI run is a
pin mismatch — either permanently red or permanently "re-baseline needed",
and in the second case the ratchet in CI never compares anything.

The two-job `fast`/`conformance` split of §4.4 is largely moot for this
scope: `fast` runs parser invariants (`go test -short ./internal/parser/...`)
that belong to §5, and the existing `.github/workflows/ci.yml` already runs
`go test ./...` on every push, which is what will exercise the fixture tier.
No finding there.

**Action:** CORPUS adds "corpus resolved from `PERL5_CORPUS`; tests skip
with a pointer to §3.2 when unset; `.t` files are never copied into
`testdata/`". RATCHET step 7 spells out the two CI steps from §4.4: clone
perl5 at the pinned SHA, and install the pinned perl with `pvm` before
running. Decide whether the ratchet job is `on: push` or nightly and say so.

### G5 — The cache's location and whether it is committed are undecided

**Plan §2 item 1:**

> Key a `testdata/oracle_cache/<sha256>.json` file on `sha256(perlVersion +
> "\x00" + src)`. **Commit the cache.** CI then runs zero perl processes on
> unchanged inputs, and the cache doubles as a review artifact

**Plan §7 checklist:** "Content-hash cache under `testdata/oracle_cache/`,
committed".

RUN step 3: "cache keyed on the file's content hash plus the corpus pin, so a
re-run costs nothing for unchanged files." No path, no statement about
committing, and the AC does not mention the cache at all. The milestone's
File Structure puts the cache in `oracle.go`; RUN puts it in `cache.go` —
harmless, but a sign the cache was not thought through.

Committing 584 JSON files each carrying a full op sequence may be many MB;
that is a real trade-off against the plan's "review artifact" argument and
should be decided rather than left to whoever writes `cache.go`. Either
answer is defensible; silence is not.

Note on the key: neither the plan's key nor RUN's includes a hash of
`parse_facts.pl` itself. Editing the oracle script (as G1 and G3 require)
would then serve stale facts from cache. Fold the script's hash into the key.

**Action:** RUN states the cache path, whether it is committed, and includes
the script hash in the key. If not committed, say why (size) and note that
CI pays ~20 s per run, which RUN's own numbers say is acceptable.

### G6 — The ratchet's `category` column has no source

**Plan §4.1:**

> perl-lsp's `parser_matrix.rs` also carries a **failure taxonomy** worth
> copying — a category per failure, with a priority ... A raw count says "64
> files fail." A taxonomy says "48 of 64 are QuoteLike." The second is a
> work plan. Categorise by the **first** error's construct.

**Plan §7 checklist:** "Failure taxonomy in the ratchet, categories from
perl-lsp's list".

RATCHET step 2 defines the baseline line as `status metric category path`,
so the column exists, but no step in any issue produces a category. COMPARE's
`Verdict{Bucket, Marker, Detail}` supplies a marker name for WRONG/wider,
which could stand in for category on those rows; nothing categorises a
no-answer (error-node) row by construct.

**Action:** the ponytail answer is to populate `category` from
`Verdict.Marker` for WRONG/wider, write `-` for the rest, and record in
RATCHET that construct-taxonomy for error nodes is deferred — and to where.
An unlabeled deferral becomes drift.

### Deliberate narrowings, recorded so they do not read as gaps

- **§1 `Deparse`** (fixpoint round-trip) and §7's "`Deparse`" checklist item
  are not in any issue. Deparse feeds metric (c), which only §5's M3 uses.
  Consistent with the milestone's "What exists" scoping to `parse_facts.pl`.
  Should be written down as deferred to the parser milestone.
- **§1's `Op{Seq, Class, Name, Arg, Flags}`** is richer than the chain's
  `Ops []string`. COMPARE's body says "Perl marks folded constants with
  `s/FOLD`, so folding is detectable where it matters" — but
  `parse_facts.pl` keeps only the op *name* (`/^\s*\S+\s+<[^>]*>\s+(\w+)/`),
  so the flag is not in `Facts` and nothing in the chain can detect FOLD.
  No step needs it; COMPARE's constant-folding test (step 5) works on marker
  presence alone. Strike the sentence from COMPARE or extend the script; the
  first is smaller.
- **§3.2's `perlonjava <sha>` pin line** is not needed; PerlOnJava is used
  only by differential testing (spec §7.5), outside §1-§4.
- **§2's "skip when perl is absent"** is the current behaviour of `askPerl`
  (`t.Skipf`). EXPORT's AC "pass unchanged in intent" should preserve it once
  `Ask` returns an error instead of skipping; worth one line in the AC.

---

## Scope creep

None found. Every step in every issue traces to §1-§4 or the milestone
context. Specifically checked and cleared:

- COMPARE reads our parser's tree (`*parser.Tree`) — that is measurement,
  not parser work. No issue modifies `internal/parser`.
- RUN step 7 writes the measured number into `00-findings.md` — that is the
  milestone's stated deliverable ("turns that from an anecdote into a
  number").
- RATCHET step 7 adds a CI workflow — §4.4.

One dependency note, not creep: RUN step 4 says "bounded errgroup".
`golang.org/x/sync` is not in `go.mod` (checked), and plan §1 opens with
"Stdlib plus `os/exec`. No dependencies." §2's own sketch uses a channel
semaphore. Use that.

---

## Contradictions with measured findings

None found. Every number the issues quote was checked against
`00-findings.md`:

| Issue claim | Findings | Match |
|---|---|---|
| CORPUS: 498 of 620 use the `@INC = ()` convention | §0.5 "498 of 620 files" | yes |
| CORPUS: omitting the arch root costs ~170 files; 66.3% vs 94.2% | §0.5 "~170 files", "411/620 (66.3%)", "584/620 (94.2%)" | yes |
| CORPUS: blead 5.45 vs 5.42.0; `for-many.t:474` | §0.7 verbatim | yes |
| CORPUS step 6: `Classify` sees only the first error; count is a lower bound | §0.5 "`perl -c` reports **one** error and stops" | yes |
| CORPUS AC: `Unknown warnings category` is version skew | §0.5 `signatures.t` line 32 | yes |
| CORPUS: four failure kinds | §0.5's four-row table | yes |
| RUN: 44 files, 29 clean, 15 error nodes; coverage not fidelity | §0.6 verbatim, including "All 29 'clean' files are clean only in the sense of having no error node" | yes |
| RUN: 584-file compilable corpus; ~9 ms / ~25 ms; Concise byte-identical | §0.5 table; README timing and md5 note | yes |
| COMPARE: prototype pair is WRONG/exact; optree is post-peephole | §0.3, §0.4; README | yes |
| RATCHET: pin mismatch is re-baseline, not regression | §0.7 "treat a mismatch as a reason to re-baseline rather than as a regression" | yes |

Also verified by building the shim: CORPUS's AC names `t/op/sub.t` and
`t/class/class.t`; both report `syntax OK` under a two-root shim on 5.42.0,
so the AC is achievable (§0.5 lists `class/` at 10/12 without naming the
two failures, which is why this was worth checking).

---

## Deviation from the plan's stated design

### V1 — RATCHET does not fail on improvement; §4.3 says it must

**Plan §4.3:**

> **Failing on improvement, not only regression, is deliberate.** A stale
> baseline is a baseline nobody trusts, and an untrusted ratchet is worse
> than none — people start passing `-update` reflexively. Forcing the update
> into the same commit keeps the diff honest and puts the score change in
> front of a reviewer.

and the code: `t.Errorf("baseline is stale — %d file(s) improved ...")`.

**Plan §4.1:** "Any file that improves requires a baseline update in the
same commit." Spec §7.4 says the same.

**RATCHET step 5 and AC:**

> Improvements do not fail the build, but the run reports them so the
> baseline gets refreshed deliberately rather than drifting.

This is the opposite decision, without a stated reason. A report inside a
green test is invisible in CI — nobody reads the log of a passing job — so
"refreshed deliberately rather than drifting" is precisely what §4.3 argues
this choice prevents. Not a contradiction with measured findings, so not
counted there, but it is the one place the chain overrides the plan rather
than implements it.

**Action:** align RATCHET to §4.3 (improvements fail with the `-update`
instruction), or record in the issue why the plan's argument is wrong here.

Two smaller RATCHET underspecifications while here: the baseline's four
statuses (`ok / wider / fail / skip`, §4.2) are never mapped onto COMPARE's
four buckets (`exact / wider / WRONG / no-answer`), and the `metric` column
(§4.2: `parse | agree | roundtrip | incremental`) is never given a value —
only `agree` exists in this scope. One line each.

---

## Dependency correctness

Claimed chain: EXPORT → (CORPUS ∥ COMPARE) → RUN → RATCHET.

| Edge | Needed? | Evidence |
|---|---|---|
| EXPORT → COMPARE | yes | COMPARE step 2 takes `facts Facts`, defined in EXPORT |
| EXPORT → CORPUS | **no, as written** — see D1 | |
| CORPUS → RUN | yes | RUN uses `BuildShim`, `Classify`, the pin |
| COMPARE → RUN | yes | RUN's `Report` is per-file `Verdict`s |
| RUN → RATCHET | yes | RATCHET compares `Report`s; "comes last deliberately" |

### D1 — EXPORT → CORPUS is a false dependency as the issues stand

CORPUS's steps call `BuildShim`, `ReadPin`, and `Classify`. None of them
calls `Ask` or `AskFile` or mentions `Facts`. The only thing CORPUS needs
from perl is `perl -c` on two files (step 1's assertion), and `Classify`
takes a string. The milestone says corpus logistics "are independent of the
comparison logic and can proceed in parallel with it", and they are
independent of the export too.

Two ways to resolve it, either is fine:

- Drop the edge. CORPUS starts immediately, in parallel with EXPORT.
- Keep it, and make CORPUS's compile check go through `AskFile` (which
  requires G2 fixed first). Then the edge is real and the corpus test
  doubles as the end-to-end check that the exported oracle works on a shim.

The second is the better use of the edge, since it is the only place before
RUN where `AskFile` meets a real corpus file.

No missing edges. RATCHET's CI step depends on CORPUS's pin file and
`BuildShim`, but that is transitive through RUN.

---

## Coverage table

| Plan section | Requirement | Covered by | Status |
|---|---|---|---|
| §1 | Run perl, decode facts, exported API | EXPORT | covered, minus G1/G2 (stderr, cwd) |
| §1 | `Compiles` (yes/no gate) | EXPORT `Facts.OK` | covered |
| §1 | `Prototype` | EXPORT `Facts.Prototypes` | covered |
| §1 | `Concise` + `Has` (marker presence) | EXPORT `Facts.Ops`; COMPARE marker-presence rule | covered; `Flags`/`s/FOLD` not captured (deliberate, note it) |
| §1 | `Deparse` fixpoint | — | UNCOVERED, deliberate (§5 metric only); record the deferral |
| §1 | "No dependencies" | RUN step 4 says errgroup | see scope note; use a channel |
| §2 item 1 | Content-hash cache, committed, under `testdata/oracle_cache/` | RUN step 3 | partial — G5 |
| §2 item 2 | Parallelise to core count | RUN step 4 | covered |
| §2 item 3 | One perl run per file, not per assertion | EXPORT (`parse_facts.pl` is per-file) | covered; G1's alternative would break it |
| §2 item 4 | Fixture tier fast by default; full sweep opt-in | RUN step 5, AC "< 10 seconds" | covered (`-corpus=full` flag in place of a build tag; equivalent) |
| §2 | Skip when perl absent | current `askPerl` | implicit; add to EXPORT AC |
| §3.1 item 1 | Classify stderr, never trust exit status | CORPUS steps 5-6; RUN step 2 | covered in CORPUS; unreachable from RUN until G1 |
| §3.1 item 2 | `-T` shebang pass-through | — | UNCOVERED — G3 |
| §3.1 | Pin corpus and oracle together; mismatch = re-baseline | CORPUS steps 3-4; RATCHET step 6 | covered |
| §3.2 | Fetch at pinned SHA, never vendor; graceful skip via `PERL5_CORPUS` | — | UNCOVERED — G4 |
| §3.2 | `corpus.pin` with perl version and perl5 SHA | CORPUS step 4 | covered |
| §3.2 | T0 fixtures, own-authored, in the fast tier | RUN step 1 "5-file fixture" | covered |
| §4.1 | Committed per-file baseline; regression fails the build | RATCHET steps 1-4 | covered, with a real fail-on-synthetic-regression test |
| §4.1 | Improvement requires baseline update in the same commit | RATCHET step 5 | DEVIATES — V1 |
| §4.1 | Failure taxonomy / category | RATCHET step 2 (column only) | partial — G6 |
| §4.1 | "Make sure yours actually measures something" | RATCHET step 4 | covered |
| §4.2 | Plain-text format with header (perl version, perl5 SHA, DO NOT EDIT) | RATCHET step 2 | covered; bucket-to-status mapping unstated |
| §4.3 | `-update` regenerates; read-only otherwise | RATCHET step 3, AC | covered |
| §4.3 | Pin mismatch is a distinct outcome | RATCHET step 6 | covered |
| §4.4 | CI job: clone perl5 at pin, install pinned perl, run ratchet | RATCHET step 7 | partial — G4 (no clone/install steps) |
| §4.4 | `fast` job without perl | existing `ci.yml` `go test ./...` | covered by existing CI; §4.4's `fast` job content is §5 work |

---

## Notes for the pushback lens (not alignment findings)

- With marker-presence comparison, a file with no prototype call sites
  scores `exact` vacuously. The headline fidelity rate from RUN will be
  dominated by files that were never really tested. Spec §7.1.5 warns about
  the mirror-image trap ("assert a floor on `exact`"); RUN's report should
  probably separate "exact, N markers checked" from "exact, 0 markers", or
  the number the milestone exists to produce will be flattering.
- `parse_facts.pl`'s prototype probe prepends a `CHECK` block to a copy of
  the file, which moves the `#!` off line 1. That is why `op/taint.t`
  reports `ok:0` yet a full `prototypes` map. Harmless today, but G3's fix
  needs to apply `-T` to both spawns or the probe and the compile will
  disagree about whether the file compiles.
