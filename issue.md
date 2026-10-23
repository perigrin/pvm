---
title: Mine at least 200 perf/opcount.t assertions into fixtures held to perl
state: done
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
- 01a16666-9a1b-76b3-9389-e4dbfb01ab9d
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.43792349Z
updated: 2026-10-23T18:09:41.628188031Z
sessions:
- start_sha: 4505f236a5ac6d2c6b78214071acbab509c7bb37
  end_sha: 9fda53a5b2ce81e419a525f3362c647afe4aafeb
  commits: 4
  started_at: 2026-10-23T17:31:22.062437519Z
  ended_at: 2026-10-23T18:09:41.628188031Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T17:31:22.062437519Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-23T18:09:41.628188031Z
observed_paths:
- .github/workflows/ci.yml
- .github/workflows/parse-fidelity.yml
- internal/parse/compiled_test.go
- internal/parse/compiledset_test.go
- internal/parse/perlgitt_faithful_test.go
- internal/parse/pin_test.go
- internal/parse/probes_test.go
- internal/parse/ratchet_test.go
- internal/parse/testdata/compiledclean.txt
- internal/parse/testdata/m3probes/README.md
- internal/parse/testdata/m3probes/baseline.txt
- internal/parse/testdata/perlgittfaithful.ratchet
- internal/parseoracle/opcount_test.go
- internal/parseoracle/testdata/opcount/README.md
- internal/parseoracle/testdata/opcount/opcount.md
- internal/parseoracle/testdata/ratchet/m3start.txt
---

## Context

- paths: `internal/parseoracle/opcount_test.go` (new), `internal/parseoracle/testdata/opcount/` (new), `.github/workflows/parse-fidelity.yml`
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^(TestOpcountFixturesParse|TestOpcountFixturesHeldToPerl)$' -count=1 -v`
- entrypoints: `$HOME/dev/perl5/t/perf/opcount.t` (`test_opcount`, ~line 49)

M3 requires at least 200 `perf/opcount.t` assertions mined into fixtures. opcount.t holds 109 `test_opcount(` calls with 375 op=>count pairs; nothing in the repo mines them today. Each call gives a code snippet and the count of each named op perl compiles it to. This issue writes those as readable text/Markdown fixtures, not JSON, under `testdata/opcount/`. TestOpcountFixturesHeldToPerl asserts there are >= 200 op=>count assertions and runs perl on each snippet to confirm every count. The fixtures then hold their own provenance, and a later issue can grade internal/parse's sites against them. The test runs perl; no production code does. This issue shares no code with the rest of the milestone.

## Steps

- [ ] The fixtures are held to the oracle's pinned perl, 5.44.0 (corpus.pin), found through `perltest.PinnedPerl`, so the tests skip under the 5.42 jobs and fail under `PVM_REQUIRE_PINNED_PERL=1`. Under 5.42, perl segfaults compiling perf/opcount.t whole (exit 139, first crash between lines 1120 and 1130, the builtin::indexed foreach snippets). Re-check that under 5.44.0, which compiles the file, and quote the result in the test's comment. Either way, mine and check snippets ONE AT A TIME, never by compiling the whole file. opcount.t's counts are blead's, so some may legitimately differ on 5.44.0. Each dropped fixture is quoted with perl's output, and the >= 100 nonzero floor must still hold.
- [ ] Read opcount.t's `test_opcount` and its callback to learn how it counts ops, including which ops it skips and how it treats nested subs. Quote that in the test's comment.
- [ ] Create `internal/parseoracle/opcount_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestOpcountFixturesParse. It reads every fixture in testdata/opcount/ into (name, snippet, op=>count) records, and asserts at least 200 op=>count pairs and the source line in opcount.t for each. Run it with no fixtures and watch it fail.
- [ ] GREEN: mine the fixtures from opcount.t, by hand or with a small one-off script that is not committed as production code. Each fixture names its opcount.t line. Run the test and watch it pass.
- [ ] RED: write TestOpcountFixturesHeldToPerl. It runs the pinned perl with B on each snippet separately, counting ops the way opcount.t does, and compares against the fixture. Flip one count by hand, watch the test fail, and restore it.
- [ ] GREEN: run the test and watch every fixture pass. Drop a fixture perl disagrees with only after quoting perl's output in the commit message.
- [ ] Add internal/parseoracle/testdata/opcount/README.md describing the fixture format (snippet, op=>count pairs, the opcount.t line each came from), like internal/infer/testdata/oracle/README.md
- [ ] CI's ratchet step runs only ./internal/parse and ./internal/lexer, so this test would never run there. Add TestOpcountFixturesHeldToPerl and its perl-running siblings to parse-fidelity.yml's 5.44-held step (`PVM_REQUIRE_PINNED_PERL=1`), so they fail rather than skip there.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The fixture directory is committed (`test -d internal/parseoracle/testdata/opcount`)
- [ ] The fixtures hold at least 200 op=>count assertions, each traced to its opcount.t line (`go test ./internal/parseoracle/ -run '^TestOpcountFixturesParse$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesParse'`)
- [ ] The pinned perl 5.44.0 confirms every mined op count (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestOpcountFixturesHeldToPerl$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesHeldToPerl'`)
- [ ] Regression guard (passes before this issue's work; it holds the new tests once they exist): every test that calls PinnedPerl, this issue's included, is in the -run list of a CI step that sets the REQUIRE variable (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequired$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequired'`)

### Negative Scenarios

- [ ] Zero counts cannot carry the floor. Of opcount.t's op=>count pairs, 177 are `=> 0`, and those pass for any snippet that lacks the op, including a truncated or mis-mined one. Every fixture must hold at least one nonzero count, and at least 100 of the >= 200 mined assertions must be nonzero (`go test ./internal/parseoracle/ -run '^TestOpcountFixturesNotCarriedByZeroCounts$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesNotCarriedByZeroCounts'`)
- [ ] A fixture whose snippet perl fails to compile (a strict error, or a missing `use feature` from the enclosing block in opcount.t, such as `signatures` at line 269 or `try` at line 677) fails the held-to-perl test with perl's stderr; it is never scored as all-zero counts (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestOpcountFixtureCompileFailureIsAFailure$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixtureCompileFailureIsAFailure'`)
- [ ] Counts are taken over the snippet sub's ROOT, the way `test_opcount` walks `B::svref_2object($coderef)->ROOT`, naming null ops `ex-<ppname>` by targ. They are not taken over the main program, where `leavesub` and the `nextstate` counts differ (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestOpcountCountsSubRootLikeOpcountT$' -count=1 -v | grep -q '^--- PASS: TestOpcountCountsSubRootLikeOpcountT'`)
- [ ] A fixture with one count altered from perl's fails the held-to-perl check and names the op, the expected count and perl's count (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestOpcountDetectsDriftedCount$' -count=1 -v | grep -q '^--- PASS: TestOpcountDetectsDriftedCount'`)
- [ ] Each fixture's snippet and counts still match the opcount.t line it names, so a fixture cannot drift from its source while still agreeing with perl (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestOpcountFixturesMatchTheirSourceLines$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesMatchTheirSourceLines'`)
- [ ] No two fixtures claim the same opcount.t line, and the 200 floor counts distinct (line, op) pairs, so duplicating a fixture cannot reach it (`go test ./internal/parseoracle/ -run '^TestOpcountFixturesHaveDistinctSources$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesHaveDistinctSources'`)
- [ ] An empty or missing testdata/opcount/ fails the parse test; it does not report 0 assertions as a pass (`go test ./internal/parseoracle/ -run '^TestOpcountFixturesRejectEmptyDir$' -count=1 -v | grep -q '^--- PASS: TestOpcountFixturesRejectEmptyDir'`)
