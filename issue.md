---
title: Find the oracle's pinned perl without PATH, and require it in CI
state: in-progress
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a1664d-81b7-7c96-8f48-636ae47d5cb1
- 01a16602-6146-76ab-bb68-f3e8b645ae79
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
created: 2026-10-23T01:36:02.587272751Z
updated: 2026-10-23T14:44:59.756155711Z
sessions:
- start_sha: 56f5c8f664ddd85de4c34bed78394fd1db5a1778
  end_sha: ""
  commits: 0
  started_at: 2026-10-23T14:44:59.756155711Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T14:44:59.756155711Z
---

## Context

- paths: `internal/perltest/pinned.go` (new), `internal/perltest/pinned_test.go` (new), `internal/parseoracle/run.go` (`RunOptions`), `internal/parseoracle/oracle.go:129` (`AskFile`), `internal/parseoracle/testdata/parse_facts.pl` (the inner compiles at ~260 and ~295), `internal/parseoracle/corpus.go` (`BuildShim`, `libraryRoots`' probe at :86, `ObservePin` at :262, shim stamp), `internal/parseoracle/gate_test.go` (TestCIGateRunsTheSweep), `internal/parseoracle/cache.go:244` (the `$]` probe), `internal/parseoracle/ratchet_test.go` (`corpusReport`), `internal/parseoracle/goparser_test.go` (`runGoSubject`), `.github/workflows/parse-fidelity.yml`, `internal/parseoracle/ci_test.go`
- commands: `go test ./internal/perltest/ -run '^TestPinnedPerl' -count=1 -v`
- entrypoints: `internal/perltest/perltest.go` (`Perl542`, `perlExecutable`, `skipOrRequire`), `internal/parseoracle/corpus.go` (`ReadPin`)

The oracle's measurements, the probe corpus, the opcount fixtures and the rv2hv and OPf_PARENS tables are all held to the interpreter in `internal/parseoracle/testdata/corpus.pin`. Today every one of them runs whatever `perl` is first on PATH, which here is /home/perigrin/.local/bin/perl 5.42.0. `make test`, `git zhi verify` and the resolution command would therefore judge 5.44-recorded facts with 5.42, unless every command line carried a PATH prefix, and `make test` cannot carry one per test. This issue adds a helper modelled on `perltest.Perl542`. `perltest.PinnedPerl(t)` reads corpus.pin's `interpreter` and `threads` (via `parseoracle.ReadPin`) and finds a matching perl, so that no command needs a PATH prefix.

Measured 2026-10-23: pvm's registry entry for 5.44.0 (`perl.GetVersionInfo("5.44.0")`) points at a stale scratch path, `.../scratchpad/default-reloc-test`. So `Perl542`'s registry-first lookup does not find 5.44 here. `$HOME/.local/share/pvm/versions/5.44.0/bin/perl` exists and reports `5.044000`, threads `define`.

The lookup order is:
1. pvm's versions directory under xdg's DataHome (`<DataHome>/pvm/versions/<ver>/bin/perl`);
2. pvm's registry;
3. the PATH perl.

It takes the first candidate whose `$]` and `useithreads` match the pin. With no match it skips, and it fails instead when `PVM_REQUIRE_PINNED_PERL` is set. Otherwise it behaves exactly like `Perl542`, which keeps its own `PVM_REQUIRE_PERL542`.

**Every perl process the oracle starts must be the given perl (P16, G1, G2, chain-review round 3).** Measured 2026-10-23: under the 5.44 interpreter, with 5.42 first on PATH, parse_facts.pl's own shell-outs run 5.42. A 5.44 perl that backticks `perl -e 'print $]'` prints `5.042000`, and `parse_facts.pl op/for-many.t` returns 5.42's "syntax error at line 474" instead of 5.44's `ok:1, op_count:2161`. So passing a path to the outer process is not enough. The oracle gains a `Perl` path in its options, defaulting to `perl` on PATH, and it reaches all of these:
- `AskFile` (oracle.go:129), which runs parse_facts.pl;
- parse_facts.pl's inner compiles, which use `$^X` instead of a literal `perl` (`capture` ~260, `capture_with_end` ~295);
- `BuildShim`, and `libraryRoots`' `perl -e` probe (corpus.go:86);
- `ObservePin` (corpus.go:262), which TestRatchetCorpus uses for CheckPin and for the baseline header under `-parseoracle.update`;
- the cache's `$]` probe (cache.go:244).

`corpusReport`, `runGoSubject` and the gate tests pass `PinnedPerl(t).Path`. Changing parse_facts.pl changes its hash, which keys the oracle cache, so every file is re-measured once. The re-pin issue's re-baseline absorbs that.

**Shim stamp (A21).** The local shim's lib/Config.pm is from 5.42 (`version => '5.42.0'`), and 5.44 run against it mis-measures silently: base/if.t passes, while class/method.t fails with "Perl lib version (5.42.0) doesn't match executable". `BuildShim` writes the building interpreter's `$]` and `useithreads` to `<shim>/interpreter.stamp`. `corpusReport` rebuilds the shim when the stamp is missing or differs from corpus.pin, and it fails if a rebuilt shim still disagrees.

**Package rule (A24).** perltest imports parseoracle (for `ReadPin`). So every caller of `PinnedPerl` under internal/parseoracle lives in package `parseoracle_test`, never package `parseoracle`, or the import cycles. That includes magic_test.go, opcount_test.go, sites_test.go and parens_test.go.

CI: parse-fidelity.yml installs the pinned perl. Two steps set `PVM_REQUIRE_PINNED_PERL=1`, and each ends in a fail-unless-PASS loop over its own `-run` names, as ci.yml:203-218 does for the ratchets, so a held test that skips or never runs fails the job.
- The existing receipt-guarded sweep step runs TestRatchetCorpus with `-parseoracle.corpus`, the restored shim and its timeouts. It sets the variable too, so there is still one sweep (G12, F13, H8, chain-review round 5). At this issue its `-run` and its fail-unless-PASS loop name TestRatchetCorpus ONLY: TestSiteRatchetCorpus does not exist until the site-ratchet issue, which adds it to both (G13, F14, K1, round 6). The loop reads a log the step writes with `go test ... 2>&1 | tee "$RUNNER_TEMP/sweep.log"`, so the step sets `shell: bash` and starts with `set -o pipefail`, or the tee would hide go test's exit status.
- The "5.44-held tests" step runs every other held test, with `PERL5_CORPUS` and `PARSEORACLE_SHIM` from the restored shim. It is placed after the shim restore, so TestGoParserExactFloor, TestGoParserSubjectMeasures and TestGoParserWrongIsZero, which skip without a shim, run there.

Each later issue adds its own tests to the 5.44-held step's `-run` list.

**How a test reaches the pinned perl (P20, H2, F12, chain-review round 4).** Helpers take the perl as a parameter: `corpusReport(t, p)`, `runGoSubject(t, p)`, `runFixture(t, p)`, `buildFixtureShim(t, p)` and the gate's sweep runner receive a `perltest.Perl`, and only test functions call `PinnedPerl`. A helper that calls `PinnedPerl` itself is a defect. TestEveryPinnedPerlTestIsRequired keeps that rule and the step honest. It parses every `_test.go` in internal/parse AND internal/parseoracle, and finds every test function that calls `PinnedPerl`. It asserts each name appears in the `-run` pattern of SOME step that sets `PVM_REQUIRE_PINNED_PERL=1`, read with the existing `uncommented()` helper, and that that step's package list covers the test's package, since `go test ./internal/parseoracle/ -run X` for an X in internal/parse runs nothing and exits 0. And it fails on any non-test function that calls `PinnedPerl`. Under that rule TestRatchetCorpus, TestSiteRatchetCorpus, TestRatchet, TestGoParserExactFloor, TestGoParserSubjectMeasures and TestGoParserWrongIsZero (M1's WRONG = 0 gate) all call `PinnedPerl` themselves, so each must be named in a REQUIRE step. The two sweep tests are named in the sweep step, and the rest in the 5.44-held step. G11: `runFixture` takes the pinned perl, so fixture.txt is judged by the pinned interpreter only, and TestRatchet moves from the 5.42 matrix to the 5.44 step.

The 5.42 jobs never set the variable, so there the 5.44-held tests skip without failing. Those jobs are ci.yml's Test matrix and "Ratchets (corpus-backed)", which runs internal/parse with `PVM_REQUIRE_PERL542=1`. The tests that skip there include the CI gate tests, TestCIGateRunsTheSweep and (from the site-ratchet issue) TestCISweepRunsOracleOnce. They execute the sweep step, so without 5.44 they would fail with "exited 0 and wrote no receipt" (gate_test.go:147). They call `PinnedPerl` and join the 5.44-held step.

**Fixture shim stamp (A25, G9, H5).** `buildFixtureShim` (gate_test.go:363) builds the gate's five-file shim by hand, with no lib/ and no stamp. Under the stamp rule, `corpusReport` would rebuild it from a corpus the gate deliberately withholds (`PERL5_CORPUS=<temp>/no-corpus`). So `buildFixtureShim` writes interpreter.stamp from the `perltest.Perl` it is given (its `$]` and threads), and TestFixtureShimCarriesTheStamp holds it.

**Pin tests that can fail at this point in the chain (P19).** This issue runs before the re-pin, while corpus.pin still names 5.042000, and the PATH perl here is also 5.042000. So a "5.42 first on PATH" decoy is the pinned perl itself, and it proves nothing. Every pin test writes its OWN temp corpus.pin, naming a version different from its decoy, and never reads the committed pin. The decoy is either a stub `perl` script that prints a fake `$]` or exits 1, or a different real perl (/usr/bin/perl reports 5.038002 here; pvm also holds 5.40.x and 5.38.0). This covers TestRunUsesTheGivenPerl, TestObservePinUsesTheGivenPerl and TestRatchetPinObservedWithPinnedPerl.

## Steps

- [ ] Create `internal/perltest/pinned.go` and `pinned_test.go` with two `// ABOUTME:` lines each.
- [ ] RED: write TestPinnedPerlPrefersVersionsDir. With a fixture pin and a fake DataHome holding a matching perl, and a non-matching perl first on PATH, the helper returns the versions-dir perl. Run it and watch it fail.
- [ ] GREEN: implement the lookup. Run it and watch it pass.
- [ ] RED/GREEN: write TestPinnedPerlSkipsOnMismatch (no candidate matches and the env var is unset, so the test skips) and TestPinnedPerlFailsUnderRequire (the same with `PVM_REQUIRE_PINNED_PERL=1` fails).
- [ ] RED/GREEN: write TestPinnedPerlChecksThreads. A candidate with the right `$]` but `useithreads` undef is rejected.
- [ ] RED: write TestRunUsesTheGivenPerl in package parseoracle_test. Write a temp corpus.pin naming the target perl's version. Put a decoy first on PATH: a stub `perl` that prints a fake `$]` or exits 1, or /usr/bin/perl (5.038002). Run `Run` on a fixture whose BEGIN writes `$]` to STDERR, with `RunOptions.Perl` set to the target perl. Assert that the INNER compile's reported `$]` is the target's and the facts are `ok`. A fake perl that prints canned JSON must fail it, because the inner `$]` comes from the file's own BEGIN. Run it and watch it fail.
- [ ] GREEN: thread the path through `AskFile`, `$^X` in parse_facts.pl's two shell-outs, `BuildShim`, `libraryRoots`, `ObservePin` and the cache probe. Run it and watch it pass.
- [ ] RED/GREEN: write TestObservePinUsesTheGivenPerl: with a temp pin and a decoy first on PATH whose version differs from it, `ObservePin` given the target path returns the temp pin's interpreter. Wire `corpusReport`, `runGoSubject`, TestRatchetCorpus's pin check and the gate tests to `PinnedPerl(t).Path`.
- [ ] RED/GREEN: write TestShimMatchesThePin and TestShimStampMismatchRebuilds. `BuildShim` writes interpreter.stamp. A fixture shim stamped 5.042000 under a 5.044000 pin is rebuilt, or fails if it cannot be rebuilt, and is never used as is.
- [ ] Change `corpusReport`, `runGoSubject`, `runFixture`, `buildFixtureShim` and the gate's runner to take a `perltest.Perl` parameter, and have each test function call `PinnedPerl` and pass it down.
- [ ] RED/GREEN: write TestFixtureShimCarriesTheStamp: `buildFixtureShim` writes interpreter.stamp matching the perl it was given, so the gate's sweep uses the five-file shim rather than rebuilding it.
- [ ] RED/GREEN: write TestEveryPinnedPerlTestIsRequired in ci_test.go, as described above, plus TestEveryPinnedPerlTestIsRequiredRejectsHelperCaller: a fixture source where a test reaches `PinnedPerl` only through a helper is flagged. Add TestCIGateRunsTheSweep, TestRatchet and the TestGoParser* tests to the 5.44-held step, with `PERL5_CORPUS`, `PARSEORACLE_SHIM` and the fail-unless-PASS loop. Give the sweep step `PVM_REQUIRE_PINNED_PERL=1`, `shell: bash`, `set -o pipefail`, a log at `$RUNNER_TEMP/sweep.log`, and its own fail-unless-PASS loop over TestRatchetCorpus only. The 5.44-held step logs to `$RUNNER_TEMP/held.log` the same way. TestEveryPinnedPerlTestIsRequired also asserts that each REQUIRE step's loop names exactly its `-run` names, so the two cannot drift.
- [ ] RED/GREEN: make `stepRunning`'s match unambiguous. The gate finds the sweep by the step's `name:` or by `-parseoracle.corpus` in its `run:`, never by a test name the 5.44-held step may also mention. Add TestGateFindsTheSweepStep (H8): in a fixture workflow where a REQUIRE step naming TestRatchetCorpus comes before the sweep step, the gate still picks the sweep step. Add TestEveryPinnedPerlTestIsRequiredFlagsUnlistedSweepTest: in a fixture workflow, a sweep test named in neither REQUIRE step is flagged.
- [ ] Add a "5.44-held tests" step to parse-fidelity.yml with `PVM_REQUIRE_PINNED_PERL: '1'`. RED/GREEN TestCIRequiresThePinnedPerl in ci_test.go: each REQUIRE step (the sweep step and the 5.44-held step) sets the variable, and no ci.yml job does.
- [ ] RED/GREEN: extend TestCIGateCannotBeSkipped (gate_test.go:235-239) from the steps matching TestRatchetCorpus and cmd/receipt to EVERY step that sets `PVM_REQUIRE_PINNED_PERL`, so no REQUIRE step can carry an `if:` that turns it off (G14). Add TestCIGateCannotBeSkippedCoversRequireSteps: a fixture workflow whose 5.44-held step has `if: false` is flagged.
- [ ] Run `make test` with the 5.42 PATH perl, confirming nothing that uses the helper fails. Then commit.
- [ ] RED/GREEN TestRatchetPinObservedWithPinnedPerl: the pin check TestRatchetCorpus performs, run on the gate's five-file fixture against a temp pin, with a decoy of a different version first on PATH, observes the temp pin's interpreter.
- [ ] Before merge, run the ci.yml Test job's command with only 5.42 available (no pvm 5.44 dir; for example `XDG_DATA_HOME` pointed at an empty dir), and confirm the gate tests skip rather than fail.

## Acceptance criteria

### Positive Scenarios

- [ ] The helper finds the pinned perl in pvm's versions directory ahead of PATH (`go test ./internal/perltest/ -run '^TestPinnedPerlPrefersVersionsDir$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlPrefersVersionsDir'`)
- [ ] The oracle's inner compile, not just the outer process, runs the interpreter it is given, with a different perl first on PATH (`go test ./internal/parseoracle/ -run '^TestRunUsesTheGivenPerl$' -count=1 -v | grep -q '^--- PASS: TestRunUsesTheGivenPerl'`)
- [ ] CI's 5.44 step requires the pinned perl, and the 5.42 jobs do not (`go test ./internal/parseoracle/ -run '^TestCIRequiresThePinnedPerl$' -count=1 -v | grep -q '^--- PASS: TestCIRequiresThePinnedPerl'`)
- [ ] The shim records its interpreter, and a shim built by another perl is not used as is (`go test ./internal/parseoracle/ -run '^TestShimMatchesThePin$' -count=1 -v | grep -q '^--- PASS: TestShimMatchesThePin'`)
- [ ] Every test in internal/parse and internal/parseoracle that calls PinnedPerl is in the -run list of a CI step that sets PVM_REQUIRE_PINNED_PERL=1 and whose package list covers it, read from uncommented YAML (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequired$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequired'`)
- [ ] ObservePin asks the perl it is given, not the PATH perl (`go test ./internal/parseoracle/ -run '^TestObservePinUsesTheGivenPerl$' -count=1 -v | grep -q '^--- PASS: TestObservePinUsesTheGivenPerl'`)
- [ ] The gate's hand-built fixture shim carries an interpreter stamp, so the gate's sweep does not rebuild it (`go test ./internal/parseoracle/ -run '^TestFixtureShimCarriesTheStamp$' -count=1 -v | grep -q '^--- PASS: TestFixtureShimCarriesTheStamp'`)
- [ ] Regression guard: the CI gate still runs the committed sweep step over the fixture shim (`go test ./internal/parseoracle/ -run '^TestCIGateRunsTheSweep$' -count=1 -v | grep -q '^--- PASS: TestCIGateRunsTheSweep'`)

### Negative Scenarios

- [ ] With no matching perl and no REQUIRE variable, a held test skips rather than running the wrong perl (`go test ./internal/perltest/ -run '^TestPinnedPerlSkipsOnMismatch$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlSkipsOnMismatch'`)
- [ ] Under PVM_REQUIRE_PINNED_PERL=1 a missing or mismatched perl fails the test (`go test ./internal/perltest/ -run '^TestPinnedPerlFailsUnderRequire$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlFailsUnderRequire'`)
- [ ] A perl with the pinned version but without threads is rejected (`go test ./internal/perltest/ -run '^TestPinnedPerlChecksThreads$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlChecksThreads'`)
- [ ] A shim stamped 5.042000 under a 5.044000 pin is rebuilt or fails, never measured with (`go test ./internal/parseoracle/ -run '^TestShimStampMismatchRebuilds$' -count=1 -v | grep -q '^--- PASS: TestShimStampMismatchRebuilds'`)
- [ ] With a decoy of another version first on PATH, the sweep's observed pin equals the pin it was given (`go test ./internal/parseoracle/ -run '^TestRatchetPinObservedWithPinnedPerl$' -count=1 -v | grep -q '^--- PASS: TestRatchetPinObservedWithPinnedPerl'`)
- [ ] A test that reaches PinnedPerl only through a helper is flagged by the coverage check (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequiredRejectsHelperCaller$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequiredRejectsHelperCaller'`)
- [ ] The CI gate finds the sweep step by its name or -parseoracle.corpus, even when a REQUIRE step naming TestRatchetCorpus comes first (`go test ./internal/parseoracle/ -run '^TestGateFindsTheSweepStep$' -count=1 -v | grep -q '^--- PASS: TestGateFindsTheSweepStep'`)
- [ ] A sweep test named in neither REQUIRE-setting step is flagged by the coverage check (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequiredFlagsUnlistedSweepTest$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequiredFlagsUnlistedSweepTest'`)
- [ ] A REQUIRE step with an `if:` that can turn it off is flagged, the 5.44-held step included (`go test ./internal/parseoracle/ -run '^TestCIGateCannotBeSkippedCoversRequireSteps$' -count=1 -v | grep -q '^--- PASS: TestCIGateCannotBeSkippedCoversRequireSteps'`)
