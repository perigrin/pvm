---
title: Find the oracle's pinned perl without PATH, and require it in CI
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
- 01a1664d-81b7-7c96-8f48-636ae47d5cb1
- 01a16602-6146-76ab-bb68-f3e8b645ae79
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
created: 2026-10-23T01:36:02.587272751Z
updated: 2026-10-23T01:56:39.496897012Z
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

CI: parse-fidelity.yml installs the pinned perl. A step there runs the 5.44-held tests with `PVM_REQUIRE_PINNED_PERL=1`, so they can never skip silently. Each later issue adds its own tests to that step's `-run` list. TestEveryPinnedPerlTestIsRequired keeps that list honest. It statically finds every test function that calls `PinnedPerl` (in the manner of perltest's noskips_test) and asserts each name appears in the REQUIRE step's `-run` pattern, read with the existing `uncommented()` and `stepRunning` helpers. The 5.42 jobs never set the variable, so there the 5.44-held tests skip without failing. Those jobs are ci.yml's Test matrix and "Ratchets (corpus-backed)", which runs internal/parse with `PVM_REQUIRE_PERL542=1`. The tests that skip there include the CI gate tests, TestCIGateRunsTheSweep and (from the site-ratchet issue) TestCISweepRunsOracleOnce. They execute the sweep step, so without 5.44 they would fail with "exited 0 and wrote no receipt" (gate_test.go:147). They call `PinnedPerl` and join the 5.44-held step.

## Steps

- [ ] Create `internal/perltest/pinned.go` and `pinned_test.go` with two `// ABOUTME:` lines each.
- [ ] RED: write TestPinnedPerlPrefersVersionsDir. With a fixture pin and a fake DataHome holding a matching perl, and a non-matching perl first on PATH, the helper returns the versions-dir perl. Run it and watch it fail.
- [ ] GREEN: implement the lookup. Run it and watch it pass.
- [ ] RED/GREEN: write TestPinnedPerlSkipsOnMismatch (no candidate matches and the env var is unset, so the test skips) and TestPinnedPerlFailsUnderRequire (the same with `PVM_REQUIRE_PINNED_PERL=1` fails).
- [ ] RED/GREEN: write TestPinnedPerlChecksThreads. A candidate with the right `$]` but `useithreads` undef is rejected.
- [ ] RED: write TestRunUsesTheGivenPerl in package parseoracle_test. Put a perl first on PATH that exits 1, or a real 5.42. Run `Run` on a fixture whose BEGIN writes `$]` to STDERR, with `RunOptions.Perl` set to the pinned perl. Assert that the INNER compile's reported `$]` is the pinned value and the facts are `ok`. A fake perl that prints canned JSON must fail it, because the inner `$]` comes from the file's own BEGIN. Run it and watch it fail.
- [ ] GREEN: thread the path through `AskFile`, `$^X` in parse_facts.pl's two shell-outs, `BuildShim`, `libraryRoots`, `ObservePin` and the cache probe. Run it and watch it pass.
- [ ] RED/GREEN: write TestObservePinUsesTheGivenPerl: with 5.42 first on PATH, `ObservePin` given the pinned path returns corpus.pin's interpreter. Wire `corpusReport`, `runGoSubject`, TestRatchetCorpus's pin check and the gate tests to `PinnedPerl(t).Path`.
- [ ] RED/GREEN: write TestShimMatchesThePin and TestShimStampMismatchRebuilds. `BuildShim` writes interpreter.stamp. A fixture shim stamped 5.042000 under a 5.044000 pin is rebuilt, or fails if it cannot be rebuilt, and is never used as is.
- [ ] RED/GREEN: write TestEveryPinnedPerlTestIsRequired in ci_test.go, as described above. Add TestCIGateRunsTheSweep to the REQUIRE step.
- [ ] Add a "5.44-held tests" step to parse-fidelity.yml with `PVM_REQUIRE_PINNED_PERL: '1'`. RED/GREEN TestCIRequiresThePinnedPerl in ci_test.go: that step sets the variable, and no ci.yml job does.
- [ ] Run `make test` with the 5.42 PATH perl, confirming nothing that uses the helper fails. Then commit.
- [ ] RED/GREEN TestRatchetPinObservedWithPinnedPerl: the pin check TestRatchetCorpus performs, run on the gate's five-file fixture with a 5.42 first on PATH, observes corpus.pin's interpreter.
- [ ] Before merge, run the ci.yml Test job's command with only 5.42 available (no pvm 5.44 dir; for example `XDG_DATA_HOME` pointed at an empty dir), and confirm the gate tests skip rather than fail.

## Acceptance criteria

### Positive Scenarios

- [ ] The helper finds the pinned perl in pvm's versions directory ahead of PATH (`go test ./internal/perltest/ -run '^TestPinnedPerlPrefersVersionsDir$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlPrefersVersionsDir'`)
- [ ] The oracle's inner compile, not just the outer process, runs the interpreter it is given, with a different perl first on PATH (`go test ./internal/parseoracle/ -run '^TestRunUsesTheGivenPerl$' -count=1 -v | grep -q '^--- PASS: TestRunUsesTheGivenPerl'`)
- [ ] CI's 5.44 step requires the pinned perl, and the 5.42 jobs do not (`go test ./internal/parseoracle/ -run '^TestCIRequiresThePinnedPerl$' -count=1 -v | grep -q '^--- PASS: TestCIRequiresThePinnedPerl'`)
- [ ] The shim records its interpreter, and a shim built by another perl is not used as is (`go test ./internal/parseoracle/ -run '^TestShimMatchesThePin$' -count=1 -v | grep -q '^--- PASS: TestShimMatchesThePin'`)
- [ ] Every test that calls PinnedPerl is in the -run list of CI's REQUIRE step, read from uncommented YAML (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequired$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequired'`)
- [ ] ObservePin asks the perl it is given, not the PATH perl (`go test ./internal/parseoracle/ -run '^TestObservePinUsesTheGivenPerl$' -count=1 -v | grep -q '^--- PASS: TestObservePinUsesTheGivenPerl'`)

### Negative Scenarios

- [ ] With no matching perl and no REQUIRE variable, a held test skips rather than running the wrong perl (`go test ./internal/perltest/ -run '^TestPinnedPerlSkipsOnMismatch$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlSkipsOnMismatch'`)
- [ ] Under PVM_REQUIRE_PINNED_PERL=1 a missing or mismatched perl fails the test (`go test ./internal/perltest/ -run '^TestPinnedPerlFailsUnderRequire$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlFailsUnderRequire'`)
- [ ] A perl with the pinned version but without threads is rejected (`go test ./internal/perltest/ -run '^TestPinnedPerlChecksThreads$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlChecksThreads'`)
- [ ] A shim stamped 5.042000 under a 5.044000 pin is rebuilt or fails, never measured with (`go test ./internal/parseoracle/ -run '^TestShimStampMismatchRebuilds$' -count=1 -v | grep -q '^--- PASS: TestShimStampMismatchRebuilds'`)
- [ ] With 5.42 first on PATH, the sweep's observed pin still equals corpus.pin (`go test ./internal/parseoracle/ -run '^TestRatchetPinObservedWithPinnedPerl$' -count=1 -v | grep -q '^--- PASS: TestRatchetPinObservedWithPinnedPerl'`)
