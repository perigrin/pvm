---
title: Find the oracle's pinned perl without PATH, and require it in CI
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
created: 2026-10-23T01:36:02.587272751Z
updated: 2026-10-23T01:36:14.636161974Z
---

## Context

- paths: `internal/perltest/pinned.go` (new), `internal/perltest/pinned_test.go` (new), `internal/parseoracle/run.go` (`RunOptions`), `internal/parseoracle/corpus.go` (`BuildShim`, `libraryRoots`, the `perl -e` probe), `internal/parseoracle/cache.go:244` (the `$]` probe), `internal/parseoracle/ratchet_test.go` (`corpusReport`), `internal/parseoracle/goparser_test.go` (`runGoSubject`), `.github/workflows/parse-fidelity.yml`, `internal/parseoracle/ci_test.go`
- commands: `go test ./internal/perltest/ -run '^TestPinnedPerl' -count=1 -v`
- entrypoints: `internal/perltest/perltest.go` (`Perl542`, `perlExecutable`, `skipOrRequire`), `internal/parseoracle/corpus.go` (`ReadPin`)

The oracle's measurements, the probe corpus, the opcount fixtures and the rv2hv and OPf_PARENS tables are all held to the interpreter in `internal/parseoracle/testdata/corpus.pin`. Today every one of them runs whatever `perl` is first on PATH, which here is /home/perigrin/.local/bin/perl 5.42.0. `make test`, `git zhi verify` and the resolution command would therefore judge 5.44-recorded facts with 5.42, unless every command line carried a PATH prefix, and `make test` cannot carry one per test. This issue adds a helper modelled on `perltest.Perl542`. `perltest.PinnedPerl(t)` reads corpus.pin's `interpreter` and `threads` (via `parseoracle.ReadPin`) and finds a matching perl, so that no command needs a PATH prefix.

Measured 2026-10-23: pvm's registry entry for 5.44.0 (`perl.GetVersionInfo("5.44.0")`) points at a stale scratch path, `.../scratchpad/default-reloc-test`. So `Perl542`'s registry-first lookup does not find 5.44 here. `$HOME/.local/share/pvm/versions/5.44.0/bin/perl` exists and reports `5.044000`, threads `define`.

The lookup order is:
1. pvm's versions directory under xdg's DataHome (`<DataHome>/pvm/versions/<ver>/bin/perl`);
2. pvm's registry;
3. the PATH perl.

It takes the first candidate whose `$]` and `useithreads` match the pin. With no match it skips, and it fails instead when `PVM_REQUIRE_PINNED_PERL` is set. Otherwise it behaves exactly like `Perl542`, which keeps its own `PVM_REQUIRE_PERL542`.

The oracle (`parseoracle.Run`, `BuildShim`, the cache's version probe) gains a `Perl` path in its options, defaulting to `perl` on PATH. `corpusReport` and `runGoSubject` pass `PinnedPerl(t).Path`.

CI: parse-fidelity.yml installs the pinned perl. A step there runs the 5.44-held tests with `PVM_REQUIRE_PINNED_PERL=1`, so they can never skip silently. Each later issue adds its own tests to that step's `-run` list. The 5.42 jobs never set the variable, so there the 5.44-held tests skip without failing. Those jobs are ci.yml's Test matrix and "Ratchets (corpus-backed)", which runs internal/parse with `PVM_REQUIRE_PERL542=1`.

## Steps

- [ ] Create `internal/perltest/pinned.go` and `pinned_test.go` with two `// ABOUTME:` lines each.
- [ ] RED: write TestPinnedPerlPrefersVersionsDir. With a fixture pin and a fake DataHome holding a matching perl, and a non-matching perl first on PATH, the helper returns the versions-dir perl. Run it and watch it fail.
- [ ] GREEN: implement the lookup. Run it and watch it pass.
- [ ] RED/GREEN: write TestPinnedPerlSkipsOnMismatch (no candidate matches and the env var is unset, so the test skips) and TestPinnedPerlFailsUnderRequire (the same with `PVM_REQUIRE_PINNED_PERL=1` fails).
- [ ] RED/GREEN: write TestPinnedPerlChecksThreads. A candidate with the right `$]` but `useithreads` undef is rejected.
- [ ] RED/GREEN: write TestRunUsesTheGivenPerl in internal/parseoracle. `RunOptions.Perl` names the interpreter for the facts walk, and `BuildShim` and the cache probe use it too. Wire `corpusReport` and `runGoSubject` to `PinnedPerl(t).Path`.
- [ ] Add a "5.44-held tests" step to parse-fidelity.yml with `PVM_REQUIRE_PINNED_PERL: '1'`. RED/GREEN TestCIRequiresThePinnedPerl in ci_test.go: that step sets the variable, and no ci.yml job does.
- [ ] Run `make test` with the 5.42 PATH perl, confirming nothing that uses the helper fails. Then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The helper finds the pinned perl in pvm's versions directory ahead of PATH (`go test ./internal/perltest/ -run '^TestPinnedPerlPrefersVersionsDir$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlPrefersVersionsDir'`)
- [ ] The oracle runs the interpreter it is given, not the PATH perl (`go test ./internal/parseoracle/ -run '^TestRunUsesTheGivenPerl$' -count=1 -v | grep -q '^--- PASS: TestRunUsesTheGivenPerl'`)
- [ ] CI's 5.44 step requires the pinned perl, and the 5.42 jobs do not (`go test ./internal/parseoracle/ -run '^TestCIRequiresThePinnedPerl$' -count=1 -v | grep -q '^--- PASS: TestCIRequiresThePinnedPerl'`)

### Negative Scenarios

- [ ] With no matching perl and no REQUIRE variable, a held test skips rather than running the wrong perl (`go test ./internal/perltest/ -run '^TestPinnedPerlSkipsOnMismatch$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlSkipsOnMismatch'`)
- [ ] Under PVM_REQUIRE_PINNED_PERL=1 a missing or mismatched perl fails the test (`go test ./internal/perltest/ -run '^TestPinnedPerlFailsUnderRequire$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlFailsUnderRequire'`)
- [ ] A perl with the pinned version but without threads is rejected (`go test ./internal/perltest/ -run '^TestPinnedPerlChecksThreads$' -count=1 -v | grep -q '^--- PASS: TestPinnedPerlChecksThreads'`)
