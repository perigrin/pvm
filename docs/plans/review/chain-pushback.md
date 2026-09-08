<!-- ABOUTME: Pushback review of the parser-fidelity-harness chain: is it executable as written? -->
<!-- ABOUTME: One section per check; every finding quotes the issue and gives the concrete rewrite. -->

# Chain pushback: parser-fidelity-harness

Reviewed 2026-09-08 against the five issues in milestone `parser-fidelity-harness`
(0/5 done) and the repo at `internal/parseoracle/` (README.md, fidelity_test.go
147 lines, testdata/parse_facts.pl 107 lines). Everything below that says
"verified" was run on this machine; the commands are in the text.

Issue IDs are abbreviated: **1 export** (01a076f1-f779), **2 corpus**
(01a076f2-5fa2), **3 compare** (01a076f2-c374), **4 run** (01a076f3-28c6),
**5 ratchet** (01a076f3-84b8).

## Summary

1. **The verify gate will extract nothing from this chain.** The installed
   `git-zhi` (the build in `~/.local/bin`, which contains the
   `\(`([^`]+)`\)` regex and `verify.Dropped`) reads only `- [ ]` checkbox
   lines and runs only the first paren-wrapped span on each. Every AC line in
   the chain is a plain `- ` bullet with bare backticks. Result today:
   `0/0 acceptance criteria passing`, exit 0, milestone completes with nothing
   checked. If anyone converts the bullets to checkboxes without adding
   `(`cmd`)`, every line becomes "unverifiable" and the gate blocks instead.
   The brief's count of "1 runnable command" in issues 1 and 4 is wrong in the
   optimistic direction: `go test ./internal/parseoracle/` there is a bare
   span, so it is ignored too. **Real count: zero runnable commands in five
   issues.** Full table and rewrites in check 5.
2. **Issue 1's API is missing two things the rest of the chain needs**, and
   since it is the head of the chain, both must be added there before
   execution starts: a working-directory option (perl's tests only compile
   from `t/`; `parse_facts.pl` already supports `ORACLE_CHDIR`, the Go API
   does not expose it) and the compile diagnostics (`parse_facts.pl` throws
   stderr away; `Classify(stderr)` in issue 2 and the Environmental skip in
   issue 4 have no input). Details in check 6.
3. Issue 4 is oversized and issue 5 bundles a CI job with a different failure
   domain. Splits in check 1.

## Check 1: Sizing

**Issue 4 (run) is too large for one session.** Steps 2 through 6 are five
separate GREENs: sequential walk plus Environmental skip, content-hash cache,
parallel workers plus a determinism test, opt-in flag, and two output formats,
then a findings-doc update. Two of those have no justified need yet:

- *JSON output* (step 6, "emit the report as both a human table and JSON").
  No consumer is named anywhere in the chain. The ratchet in issue 5 consumes
  `Report` in-process and its baseline is plain text by design. Cut it.
- *Cache* (step 3, plus the whole of `cache.go`). Measured here:
  20 sequential `perl -c op/sub.t` runs take 0.72 s wall (36 ms each on a
  14 KB file, process spawn included). 584 files at the README's ~25 ms
  Concise cost is ~15-20 s serial and ~3 s across the 8 workers step 4 adds.
  The cache also has an unspecified on-disk location and size (the plan's
  version commits one JSON per file, each carrying the full op list). Defer
  it to a follow-up issue gated on a measurement of the parallel run.

Split: **4a** walk + Environmental skip + opt-in flag + human table + findings
entry (the deliverable); **4b** cache, opened only if 4a's full run is measured
slow enough to matter. With 4b deferred, 4a fits a session.

**Issue 5 (ratchet) bundles a CI workflow with the ratchet.** Step 7 needs a
perl 5.42.0 on `ubuntu-latest` (plan §4.4 does it with
`go run ./cmd/pvm install`; `cmd/pvm` exists), a pinned clone of perl5, and a
cache of the built shim. That is network, install, and runner concerns, none
of which the local verify gate can exercise (see check 4). Move step 7 and the
last AC into their own issue, blocked by 5. The ratchet is fully useful run
locally.

Issues 1, 2, 3 are sized correctly. Issue 2 is three functions but they share
one fixture (the shim) and one test file.

## Check 2: Missing QA

**Issue 1**
- AC "`Ask` accepts a `context.Context` and honours cancellation" has no step
  that tests it. Step 3 only says to use `exec.CommandContext`. Add a RED:
  source `BEGIN { sleep 10 }` under a 200 ms timeout must return within ~1 s
  with a context error. This test will *fail* with the naive implementation;
  see check 6 item 3 for why.
- AC "locates `parse_facts.pl` relative to the package, not the caller's
  working directory" has no test. Go 1.24 (go.mod says 1.24.3) has
  `t.Chdir`; the test is `t.Chdir(t.TempDir())` then `AskFile`.
- AC "`Facts` carries the fields `parse_facts.pl` emits" is only checked for
  two fields (`Srefgen`, `Prototypes`) in step 1. `OK`, `Ops`, `Entersub`
  and the missing `stderr` (check 6) are untested.

**Issue 2**
- Step 3 tests only an interpreter-version mismatch. The AC says `ReadPin`
  "returns an error naming both on mismatch": the corpus-revision half has no
  test and no step says where the checkout's revision comes from.
- AC "a test asserts `Config.pm` is present in the built `lib/`" appears in
  the AC but in no step. Add to step 2.
- AC "`Classify` distinguishes the four kinds on real stderr strings captured
  from the corpus" needs those strings committed as fixtures in
  `corpus_test.go`; no step captures them. Say so in step 5.
- `t/class/class.t` is named in the AC but not in step 1. Verified both
  compile under the README shim on this machine (`op/sub.t syntax OK`,
  `class/class.t syntax OK`; `op/for-many.t` fails at line 474 as claimed).

**Issue 3**
- **The `wider` bucket has no step and no test.** Steps 1-6 exercise exact,
  WRONG, and no-answer only. Spec §7.1.5 defines wider as "a generic
  `CallExpression` where perl resolved a specific prototype-driven form, and
  we mark it unresolved". Our parser has no unresolved marking, so wider is
  unreachable today. Either add one characterisation test asserting `Compare`
  never emits wider yet (with a comment saying why) or drop it from this
  issue's AC. Do not leave a bucket that nothing can produce and nothing
  checks.
- **A false-WRONG case is missing.** `sub f{} my @a; f(\@a);` produces
  `srefgen` in perl *and* an explicit reference in our tree. If `Compare`
  keys on "perl has srefgen, we have none" without looking for the explicit
  `\`, every explicit-reference call in the corpus scores WRONG. Add a RED:
  this case buckets exact.
- AC "a test or a compile-time structure makes that hard to violate
  accidentally" is not a test. The mechanical version: `Compare` returns the
  same `Verdict` when `Facts.Ops` is shuffled and duplicated. That is one
  table row.

**Issue 4**
- AC "finishes in under 10 seconds" is checkable but no step checks it;
  `go test ./internal/parseoracle/ -timeout 10s` does, and belongs in the AC
  command (check 5).
- Step 2 "skipping files `Classify` marks Environmental" cannot be tested
  until `Facts` carries stderr (check 6 item 2).

**Issue 5**
- Step 1 says `TestRatchet` compares "a fixture report against a checked-in
  baseline"; the AC says `baseline.txt` holds the corpus verdicts. If the
  committed baseline is the full corpus, `TestRatchet` needs `PERL5_CORPUS`
  and will skip without it, so an AC command of
  `go test -run TestRatchet` passes vacuously on a machine without the
  corpus. State explicitly: the synthetic-regression test (step 4) uses an
  in-memory report and a temp baseline so it always runs; the corpus
  comparison skips when the corpus is absent.
- AC "`-update` is the only way to rewrite the baseline" has no test. The
  proxy: run without `-update`, assert the file is byte-identical after.

## Check 3: Dependency cycles and critical chain

No cycle. From the `BLOCKED_BY` fields: 1 has none; 2 and 3 depend on 1; 4 on
2 and 3; 5 on 4. That is a DAG with critical chain 1 → {2 | 3} → 4 → 5,
length 4, matching the claim.

Nothing is serialised that need not be:
- 3 needs only the `Facts` type from 1, so it is correctly not behind 2.
- 5 could technically start once 4's `Report` type exists (its own tests run
  on a fixture report), but the milestone doc's reason for putting it last
  ("a baseline committed earlier freezes noise") is sound. Leave it.

The real constraint on the chain is not its length but **issue 1's API
shape**. Both 2 and 4 will stall and re-open 1 unless 1 ships the
working-directory option and the stderr field (check 6, items 1 and 2). Fix
the issue text now rather than discover it at 4.

## Check 4: Untestable acceptance criteria

Quoted verbatim; each with the checkable replacement.

| Issue | AC as written | Why it cannot be checked | Replacement |
|---|---|---|---|
| 1 | "pass unchanged **in intent**" | "Intent" is a judgment. | "pass, and their assertion lines are unchanged" (the test names still exist and pass; `git diff` on the two functions is empty apart from the helper call). |
| 2 | "documented **where a reader of the result will see it**, not only in the commit message" | Location is a judgment. | "the `Classify` doc comment and the `FailureKind` String() output both say the syntax-error count is a lower bound" (grep-able). |
| 3 | "a test or a compile-time structure makes that **hard to violate accidentally**" | "Hard" is unmeasurable. | The op-order invariance test from check 2. |
| 3 | "`Verdict` carries **enough detail** to explain a WRONG result without re-running perl" | "Enough" is a judgment. | "every WRONG verdict has a non-empty `Marker`" — a table test. |
| 4 | "**in the style of** the existing findings" | Style is a judgment. | Keep the checkable half: the findings entry contains the exact command line that produced the number (grep). |
| 5 | "`baseline.txt` is committed, **human-readable**" | Judgment. | "one line per file, `status metric category path`, header lines start with `#`" — checkable with `head`/`grep`. |
| 5 | "CI runs the full corpus and completes within its timeout; the corpus build is cached rather than rebuilt per job" | Not checkable on the machine the gate runs on; needs a GitHub runner. | Move to the split-out CI issue; there, the local check is at most `test -f .github/workflows/parser-conformance.yml`, and the real check is the Actions run URL recorded in the issue. |
| 5 | "`-update` is the **only way** to rewrite the baseline" | Universal negative. | "running `TestRatchet` without `-update` leaves `baseline.txt` byte-identical". |

## Check 5: AC-command executability

### What the gate actually does

Read from the git-zhi source (`internal/verify/extract.go`,
`internal/issue/sections.go`, `internal/verify/cmd.go`, pu worktree) and
confirmed against the installed binary with `strings`:

1. `## Acceptance Criteria` is split into lines; only lines matching
   `^- \[([xX ])\] (.+)$` become items. **A plain `- ` bullet is not an item.
   A wrapped continuation line is not part of the item.**
2. For each item, the first `(`...`)` span is the command. A bare
   `` `...` `` span with no parens is *not* run; an item that has bare
   backticks and no paren-wrapped span is reported as **Unverifiable**, and
   `unverifiableCount > 0` makes the gate return an error.
3. An issue with zero items is skipped silently.

The whole section is what `git-zhi issue show` prints (confirmed for issue 1:
plain `- ` bullets, wrapped lines). So:

- **As stored:** all five issues yield zero items → gate prints
  `parser-fidelity-harness: 0/0 acceptance criteria passing`, exit 0. No
  false regressions, but *nothing is verified*; the milestone closes
  unchecked.
- **If bullets are checkboxed without adding `(`cmd`)`:** every line with a
  bare backtick becomes Unverifiable and the gate blocks completion.

Either way every AC line has to be rewritten to the form the extractor
reads: one line, `- [ ] <behaviour> (`<runnable command>`)`, code names and
paths in bare backticks only.

### Every backticked span, classified

Classification is of what the span *is*; the "gate sees" column is what the
installed extractor does with it as written.

| Issue | Span | Class | Gate sees today |
|---|---|---|---|
| 1 | `Ask` | identifier | ignored (bare, no checkbox) |
| 1 | `AskFile` | identifier | ignored |
| 1 | `Facts` | identifier | ignored |
| 1 | `parse_facts.pl` | filename | ignored |
| 1 | `Ask` | identifier | ignored |
| 1 | `context.Context` | identifier | ignored |
| 1 | `TestOracleDiscriminatesPrototypes` | identifier | ignored (and on a wrapped line) |
| 1 | `TestPrototypeParsesAreIndistinguishableToUs` | identifier | ignored (continuation line) |
| 1 | `go test ./internal/parseoracle/` | **runnable** | **ignored: bare, not paren-wrapped** |
| 1 | `parse_facts.pl` | filename | ignored |
| 1 | `AskFile` | identifier | ignored |
| 2 | `BuildShim` | identifier | ignored |
| 2 | `t/op/sub.t` | path | ignored |
| 2 | `t/class/class.t` | path | ignored |
| 2 | `perl -c` | fragment (no file; verified `perl -c </dev/null` exits 0, so it would even pass vacuously) | ignored |
| 2 | `Config.pm` | filename | ignored |
| 2 | `lib/` | path | ignored |
| 2 | `corpus.pin` | filename | ignored |
| 2 | `ReadPin` | identifier | ignored |
| 2 | `Classify` | identifier | ignored |
| 2 | `Unknown warnings category` | perl diagnostic text | ignored |
| 2 | `Classify` | identifier | ignored |
| 3 | `fidelity_test.go` | filename | ignored |
| 3 | `Compare` | identifier | ignored |
| 3 | `Verdict` | identifier | ignored |
| 4 | `Config.pm` | filename | ignored |
| 4 | `go test ./internal/parseoracle/` | **runnable** | **ignored: bare, not paren-wrapped** |
| 4 | `00-findings.md` | filename | ignored |
| 5 | `baseline.txt` | filename | ignored |
| 5 | `-update` | flag fragment | ignored |

Totals: 30 spans, 2 runnable, **0 extractable**. Issues 2, 3, 5 have no
runnable span at all; issues 1 and 4 have one each in a form the gate does
not read.

### Rewrites

Each block replaces the issue's `## Acceptance Criteria` wholesale. Every
line is a single line. Test names are proposals; whatever the executor names
them, the `-run` pattern must match. Commands run from the repo root via
`sh -c`, so the corpus path is inlined with a fallback to the checkout that
exists on this machine (`/home/perigrin/dev/perl5`, revision 94e5086608,
v5.45.0-68; verified).

**Issue 1 (export)**

```markdown
## Acceptance Criteria

- [ ] `Ask` and `AskFile` are exported and callable from an external test package (`go test ./internal/parseoracle/ -run TestAsk -v`)
- [ ] `Facts` decodes every field `parse_facts.pl` emits, including `stderr`, with matching JSON tags (`go test ./internal/parseoracle/ -run TestAskDecodesFacts`)
- [ ] `Ask` returns within one second when its context is cancelled, even with a perl that sleeps in `BEGIN` (`go test ./internal/parseoracle/ -run TestAskCancel -timeout 30s`)
- [ ] `AskFile` accepts a working directory so corpus files compile from `t/` (`go test ./internal/parseoracle/ -run TestAskFileDir`)
- [ ] `AskFile` finds `parse_facts.pl` after the test changes directory (`go test ./internal/parseoracle/ -run TestAskFileFromOtherDir`)
- [ ] both characterisation tests pass with their assertions unchanged (`go test ./internal/parseoracle/ -run 'TestOracleDiscriminatesPrototypes|TestPrototypeParsesAreIndistinguishableToUs'`)
- [ ] the package is green (`go test ./internal/parseoracle/`)
```

**Issue 2 (corpus)**

```markdown
## Acceptance Criteria

- [ ] `BuildShim` yields a tree in which `t/op/sub.t` and `t/class/class.t` compile under `perl -c` (`PERL5_CORPUS=${PERL5_CORPUS:-$HOME/dev/perl5/t} go test ./internal/parseoracle/ -run TestBuildShimCompiles`)
- [ ] the shim's `lib/` contains `Config.pm`, proving both library roots were copied (`PERL5_CORPUS=${PERL5_CORPUS:-$HOME/dev/perl5/t} go test ./internal/parseoracle/ -run TestBuildShimHasConfigPm`)
- [ ] `ReadPin` returns an error naming both the recorded and the running interpreter version on mismatch (`go test ./internal/parseoracle/ -run TestReadPinVersionMismatch`)
- [ ] `ReadPin` returns an error naming both the recorded and the checked-out corpus revision on mismatch (`PERL5_CORPUS=${PERL5_CORPUS:-$HOME/dev/perl5/t} go test ./internal/parseoracle/ -run TestReadPinRevisionMismatch`)
- [ ] `Classify` separates Environmental, VersionSkew, SyntaxError and Other on committed stderr fixtures, with `Unknown warnings category` as VersionSkew (`go test ./internal/parseoracle/ -run TestClassify`)
- [ ] the lower-bound caveat is in `Classify`'s doc comment (`grep -q 'lower bound' internal/parseoracle/corpus.go`)
```

**Issue 3 (compare)**

```markdown
## Acceptance Criteria

- [ ] a parse with an error node buckets no-answer, never WRONG (`go test ./internal/parseoracle/ -run TestCompareErrorNodeIsNoAnswer`)
- [ ] the prototype pair from `fidelity_test.go` buckets WRONG and exact respectively (`go test ./internal/parseoracle/ -run TestComparePrototypePair`)
- [ ] an explicit reference `f(\@a)` with no prototype buckets exact even though perl emits srefgen (`go test ./internal/parseoracle/ -run TestCompareExplicitRefIsExact`)
- [ ] `my $x = 1+2` does not bucket WRONG although perl's optree has no `add` (`go test ./internal/parseoracle/ -run TestCompareConstantFold`)
- [ ] `Compare` returns the same verdict when `Facts.Ops` is shuffled and duplicated (`go test ./internal/parseoracle/ -run TestCompareIgnoresOpOrderAndCount`)
- [ ] every WRONG verdict carries a non-empty `Marker` (`go test ./internal/parseoracle/ -run TestVerdictNamesMarker`)
```

**Issue 4 (run), after the split in check 1**

```markdown
## Acceptance Criteria

- [ ] the default run touches only the fixture and the package finishes in under ten seconds (`go test ./internal/parseoracle/ -timeout 10s`)
- [ ] a full run over the compilable corpus reports exact, wider, WRONG and no-answer totals and names each WRONG file (`PERL5_CORPUS=${PERL5_CORPUS:-$HOME/dev/perl5/t} go test ./internal/parseoracle/ -run TestCorpusRun -corpus=full -timeout 20m`)
- [ ] files `Classify` marks Environmental are excluded from the denominator and counted separately (`go test ./internal/parseoracle/ -run TestRunExcludesEnvironmental`)
- [ ] two consecutive runs over the fixture produce identical reports (`go test ./internal/parseoracle/ -run TestRunDeterministic`)
- [ ] the measured rate is in `00-findings.md` with the command that produced it (`grep -q 'go test ./internal/parseoracle/ -run TestCorpusRun -corpus=full' docs/specs/perl-parser/00-findings.md`)
```

**Issue 5 (ratchet), without the CI step**

```markdown
## Acceptance Criteria

- [ ] `baseline.txt` is committed with a header carrying interpreter version, corpus revision and a DO NOT EDIT marker (`head -5 internal/parseoracle/testdata/ratchet/baseline.txt | grep -q 'DO NOT EDIT BY HAND'`)
- [ ] the ratchet FAILS on a synthetic exact-to-WRONG regression against a temp baseline (`go test ./internal/parseoracle/ -run TestRatchetFailsOnRegression`)
- [ ] an improvement is reported and does not fail (`go test ./internal/parseoracle/ -run TestRatchetReportsImprovement`)
- [ ] a pin mismatch reports re-baseline needed, not a regression (`go test ./internal/parseoracle/ -run TestRatchetPinMismatch`)
- [ ] running without `-update` leaves `baseline.txt` byte-identical (`go test ./internal/parseoracle/ -run TestRatchetReadOnly`)
```

After rewriting, `git-zhi verify parser-fidelity-harness --dry-run` will still
print nothing (no issue is done), so the check that the rewrite parses is:
`git-zhi issue show <id>` and confirm every AC line starts with `- [ ]` and
contains exactly one `(`...`)` span on that same line.

## Check 6: Steps that cannot be followed

Ordered by how much of the chain each one breaks.

**1. Issue 1's API has no working-directory option; issue 4 cannot run the
corpus through it.** perl's tests begin with `chdir 't' if -d 't'; require
'./test.pl'` and `t/test.pl` clears `@INC` (verified: `t/test.pl:117-120`).
`parse_facts.pl` handles this with `ORACLE_CHDIR` (README "Usage"; script
lines 66-70), but issue 1 specifies `AskFile(ctx, path)` with nothing to set
it. Called with an absolute path from Go's cwd, `require './test.pl'` fails
in `BEGIN` and the oracle reports `ok:0` for the 498 files that use the
convention. Issue 4 step 2 would then classify ~80% of the corpus
Environmental and the headline number would be measured on the remainder.
*Fix in issue 1:* `AskFile(ctx, dir, rel string)` (or an `Options{Dir}`)
that sets `ORACLE_CHDIR=dir` and passes `rel`; one test compiling a copy of
`t/op/sub.t` from a shim-shaped temp dir.

**2. `Facts` has no stderr, so `Classify(stderr)` has no input.**
`parse_facts.pl` captures `perl -MO=Concise -c FILE 2>&1` into `$concise`
and keeps only lines matching the op pattern; the diagnostics are discarded,
and the JSON has no error text. Issue 2 step 5 and issue 4 step 2 both
consume "stderr" that nothing produces. *Fix in issue 1:* the script emits
`"stderr"` (the non-op lines of that capture) and `Facts` gains
`Stderr string \`json:"stderr"\``. That also makes issue 2's fixture strings
obtainable from the oracle itself.

**3. Step 3 of issue 1 does not achieve its own AC.** `exec.CommandContext`
kills only the outer `perl parse_facts.pl`. The script spawns
`sh -c "cd ... && perl -MO=Concise -c ..."` through backticks; that child
keeps the stdout pipe open, and `cmd.Output()` keeps waiting for EOF after
the kill. A hung inner perl still wedges the suite. *Fix:* set
`cmd.WaitDelay = 2 * time.Second` (stdlib, one line; Go 1.20+), and write
the cancellation test from check 2 so this is caught rather than assumed.

**4. Issue 2 never says where the corpus is.** `BuildShim(dir)` has one
argument and the text never names the source. The README recipe uses a
relative `perl5/t`; the checkout is actually at `/home/perigrin/dev/perl5`
(outside the repo, not vendored). Plan §3.2 already decided this:
`PERL5_CORPUS` env var, skip when unset. *Fix:* `BuildShim(corpusT, dst
string)`; tests read `PERL5_CORPUS` and `t.Skip` when empty; the same
variable feeds `ReadPin`'s revision check (read `$PERL5_CORPUS/../.git/HEAD`
or run `git -C` there) and issue 4's full run.

**5. Issue 4 step 4 names `errgroup`; the module does not have it.**
`golang.org/x/sync` is absent from `go.mod` (only `x/text`, `x/net`, `x/sys`,
`x/term`). Either add the dependency or use the semaphore-channel plus
`sync.WaitGroup` pattern the plan (§2) already shows. Stdlib is the shorter
diff.

**6. Issue 4's cache is unspecified.** Step 3 says "cache keyed on the
file's content hash plus the corpus pin" with no location, format, or
whether it is committed. Plan §2 proposes committing one JSON per file under
`testdata/oracle_cache/`; at 584 files each carrying the full op list that
is a review-noise question nobody has answered. Per check 1, defer.

**7. Issue 3 step 2 cannot produce `wider`.** See check 2. The step list
never mentions the bucket after the table in the description.

**8. Three `-T` files are unhandled.** `op/taint.t`, `op/utftaint.t`,
`perf/taint.t` carry `-T` on the shebang (verified with grep on the shim);
`perl -c` refuses them with "Too late for -T". Plan §3.1 says pass `-T`
through; no issue does, and `Classify` has no kind for it, so they land in
Other and shift the denominator by three. Two lines in `parse_facts.pl`
(read the shebang, add `-T`) or a fixed Environmental classification; say
which in issue 2.

**9. Issue 5 step 7's CI shape disagrees with the chain's paths.** Plan §4.4
reads `testdata/corpus.pin` at the repo root and runs
`go test ./internal/parser -run TestRatchet`; the chain puts the pin at
`internal/parseoracle/testdata/corpus.pin` and the test in
`internal/parseoracle`. Whoever writes the workflow will copy the plan
snippet and get a red job. Moot if step 7 is split out (check 1); otherwise
fix the paths in the step.

**10. Issue 5 step 1 has two baselines in mind.** Covered in check 2: say
which one `baseline.txt` is (the corpus) and that the regression test uses
a temp file.

Things I checked that are fine: `internal/parser.Tree`, `Parse`, `RootNode`,
`HasError`, `NamedChild`, `Kind` all exist with the signatures the steps
assume; `internal/types` exists for issue 3's Unknown-vs-Any reference;
`cmd/pvm` exists for the plan's `pvm install` step; the 44-file
base/comp/cmd/opbasic count and the 620-file total are correct; the shim
recipe in the README works as written (Config.pm present from the
`x86_64-linux-thread-multi` root); `go test ./internal/parseoracle/`
currently passes in 1.3 s.

## The single most important thing to fix before execution

**Rewrite every Acceptance Criteria line in all five issues to
`- [ ] <behaviour> (`<command>`)`, one line each, using the blocks in
check 5.** As stored, the completion gate extracts nothing from this chain,
so the milestone would close with `0/0` verified and the harness that exists
to keep a number honest would itself be unchecked.

Immediately behind it, and required before issue 1 is picked up because it
is the head of the chain: **add the working-directory option and the
`stderr` field to issue 1's API** (check 6, items 1-3). Without them issue 2
has nothing to classify and issue 4 measures the wrong 20% of the corpus.
