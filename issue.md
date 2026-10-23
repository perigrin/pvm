---
title: Ratchet oracle agreement per marker site, with the M3 floor, in the CI sweep
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-3e74-7a1a-b7bf-090cd8bceadd
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
- 01a16666-9a1b-76b3-9389-e4dbfb01ab9d
blocks:
- 01a12059-30a8-722e-bbc3-52066b8610cb
- 01a16602-61a2-78de-a798-da4cef3208c5
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16602-5ebe-7169-b857-996ae2953394
- 01a1664d-6fdb-7a09-8211-882e814db87c
created: 2026-10-22T23:46:33.509926758Z
updated: 2026-10-23T01:55:58.746869153Z
---

## Context

- paths: `internal/parseoracle/sites_test.go`, `internal/parseoracle/ratchet_test.go` (`corpusReport`), `internal/parseoracle/receipt.go`, `internal/parseoracle/gate_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt` (new), `internal/parseoracle/testdata/ratchet/m3start.txt`, the gate's fixture directory (a sites.txt for its five-file shim, new), `.github/workflows/parse-fidelity.yml`
- commands: `PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`
- entrypoints: `internal/parseoracle/ratchet_test.go:43` (`corpusReport`, TestRatchetCorpus), `internal/parseoracle/gate_test.go:105-208` (TestCIGateRunsTheSweep, `baselineEnv`), `internal/parseoracle/receipt.go`

Commit M3's oracle condition as a ratchet, at the 5.44.0 pin. TestSiteRatchetCorpus takes the report from `corpusReport(t)` and writes `testdata/ratchet/sites.txt` with two kinds of row. `site <path> <marker> <bucket> <count>` gives one row per file, marker and bucket. `total <scope> <marker> <bucket> <count>` gives the scope `all` or `m3`, the M3 dirs being `op mro uni lib test_pl`. Only nonzero rows are written, sorted by kind, path, marker and bucket, so the diff stays readable (A22/P18). The per-file rows are what make the ratchet see a swap between two files at equal totals. A swap between two sites inside one file, at equal per-file counts, stays invisible; that is accepted, and baseline.txt's per-file status does not see it either. the totals rows are what later issues and the docs quote (A14, chain-review round 2). The test fails in both directions against that file and enforces the floor: M3-dir exact >= 80% of sites, and WRONG = 0 everywhere. The numbers are measured here, at the 5.44.0 pin, with op/magic.t's sites counted as no-answer until the op/magic.t issue fixes it. The 5.42 scratch figures are not a target. This issue records the starting site counts in m3start.txt: `rv2hv_wider`, whose scope is `all` (the whole 620-file corpus, sites.txt's `total all rv2hv wider`), and `m3_exact_pct`. The op/magic.t issue records `rv2hv_wider_after_magic` under its own key once that file is answered, and the arrow-read issue compares against that.

The CI gate (gate_test.go) runs the committed sweep step over a five-file fixture shim, pointing the per-file baseline at a fixture through `baselineEnv`. It checks a receipt that only a completed TestRatchetCorpus writes. For TestSiteRatchetCorpus to join that step:
- it needs its own baseline override and a fixture sites.txt for the gate's five files;
- `corpusReport` must compute once per process (sync.Once), so that two tests in one `go test` walk the corpus once;
- the receipt must say the site ratchet ran. The decision: TestSiteRatchetCorpus adds a `sites` line to the receipt TestRatchetCorpus writes, and the gate's receipt check requires that line.

## Steps

- [ ] RED: write TestSiteFloorRejectsBelowEightyOrWrong, a unit test of the floor check. 79.9% M3-dir exact fails, one WRONG site fails, and 80% with 0 WRONG passes. Run it and watch it fail.
- [ ] GREEN: implement the floor check. Run the test and watch it pass.
- [ ] RED/GREEN: write TestCorpusReportComputesOnce, then memoise `corpusReport` with sync.Once (its error included).
- [ ] RED: write TestSiteRatchetCorpus. Its counts come from the per-site helper. It compares them to sites.txt, honouring a `PARSEORACLE_SITES_BASELINE` override like the per-file `baselineEnv`, and fails on any difference, gain or loss. It regenerates the file under `-parseoracle.update`, and on a pass it adds the `sites` line to the receipt. Run it without sites.txt and watch it fail.
- [ ] GREEN: generate sites.txt with `-parseoracle.update` at the 5.44.0 pin. Write `rv2hv_wider` and `m3_exact_pct` into m3start.txt from it. Run the sweep and watch both tests pass.
- [ ] TestCISweepRunsOracleOnce and TestCIGateRunsTheSweep execute the sweep step, so they call `perltest.PinnedPerl`, skip on the 5.42 jobs, and are in parse-fidelity.yml's 5.44-held step. They live in package parseoracle_test.
- [ ] RED: write TestCISweepRunsOracleOnce. It runs the gate's sweep step and asserts from the receipt that parseoracle.Run walked each file once, and that the receipt carries the `sites` line. Change the sweep step's `-run TestRatchetCorpus` in `.github/workflows/parse-fidelity.yml` to a pattern running both tests, point the gate at a fixture sites.txt via the override, and require the receipt's `sites` line. Run TestCIGateRunsTheSweep and TestCISweepRunsOracleOnce, and watch them pass.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The floor rejects M3-dir exact below 80% and any WRONG site (`go test ./internal/parseoracle/ -run '^TestSiteFloorRejectsBelowEightyOrWrong$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorRejectsBelowEightyOrWrong'`)
- [ ] The corpus report is computed once per process (`go test ./internal/parseoracle/ -run '^TestCorpusReportComputesOnce$' -count=1 -v | grep -q '^--- PASS: TestCorpusReportComputesOnce'`)
- [ ] The starting rv2hv wider count is recorded beside the pin (`grep -Eq '^rv2hv_wider = [0-9]+$' internal/parseoracle/testdata/ratchet/m3start.txt`)
- [ ] Regression guard (passes today; the issue extends it): the CI gate executes the sweep step, site ratchet and receipt line included (`go test ./internal/parseoracle/ -run '^TestCIGateRunsTheSweep$' -count=1 -v | grep -q '^--- PASS: TestCIGateRunsTheSweep'`)
- [ ] One CI sweep walks each file once for both ratchets, and its receipt carries the sites line (`go test ./internal/parseoracle/ -run '^TestCISweepRunsOracleOnce$' -count=1 -v | grep -q '^--- PASS: TestCISweepRunsOracleOnce'`)
- [ ] Regression guard: CI never regenerates the ratchets it checks, now that the sweep step also runs the site ratchet (`go test ./internal/parseoracle/ -run '^TestCINeverRebaselines$' -count=1 -v | grep -q '^--- PASS: TestCINeverRebaselines'`)
- [ ] The per-file and per-site ratchets both match the 620-file sweep at the 5.44.0 pin, found through the pinned-perl helper (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v | grep -cE '^--- PASS: Test(Site)?RatchetCorpus ' | grep -qx 2`)

### Negative Scenarios

- [ ] The site comparison fails when sites.txt is missing or empty, rather than passing with nothing to compare (`go test ./internal/parseoracle/ -run '^TestSiteRatchetFailsWithoutBaselineFile$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetFailsWithoutBaselineFile'`)
- [ ] The site comparison fails on an empty report (0 files or 0 sites), so a corpus that was never walked cannot match a zeroed file (`go test ./internal/parseoracle/ -run '^TestSiteRatchetRejectsEmptyReport$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetRejectsEmptyReport'`)
- [ ] The floor fails when the M3 dirs have 0 sites; a 0/0 ratio is not >= 80% (`go test ./internal/parseoracle/ -run '^TestSiteFloorRejectsZeroSites$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorRejectsZeroSites'`)
- [ ] The floor's denominator includes no-answer and wider sites; 80 exact of 100 sites with 30 no-answer is not scored as 80/70 (`go test ./internal/parseoracle/ -run '^TestSiteFloorCountsNoAnswerInDenominator$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorCountsNoAnswerInDenominator'`)
- [ ] One WRONG site outside the M3 dirs fails the floor too; WRONG = 0 holds over the whole corpus (`go test ./internal/parseoracle/ -run '^TestSiteFloorRejectsWrongOutsideM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorRejectsWrongOutsideM3Dirs'`)
- [ ] A swap fails the ratchet: one file's site moving exact->wider while another's moves wider->exact leaves the per-marker totals unchanged, and the comparison still reports both files (`go test ./internal/parseoracle/ -run '^TestSiteRatchetDetectsSwapAtEqualTotals$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetDetectsSwapAtEqualTotals'`)
- [ ] A report missing files the per-file baseline lists (a sweep timeout or a dropped directory) fails, rather than shrinking the denominator (`go test ./internal/parseoracle/ -run '^TestSiteRatchetFailsOnDroppedFiles$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetFailsOnDroppedFiles'`)
- [ ] The M3-dir subset matches only the five dirs `op mro uni lib test_pl` by path prefix, so a lookalike such as `opbasic/` (which starts with `op`) is not counted as an M3 dir (`go test ./internal/parseoracle/ -run '^TestM3DirSubsetMatchesExactDirs$' -count=1 -v | grep -q '^--- PASS: TestM3DirSubsetMatchesExactDirs'`)
