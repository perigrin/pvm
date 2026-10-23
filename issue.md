---
title: Ratchet oracle agreement per marker site, with the M3 floor, in the CI sweep
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-3e74-7a1a-b7bf-090cd8bceadd
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
blocks:
- 01a12059-30a8-722e-bbc3-52066b8610cb
- 01a16602-61a2-78de-a798-da4cef3208c5
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:33.509926758Z
updated: 2026-10-23T01:08:58.90055854Z
---

## Context

- paths: `internal/parseoracle/sites_test.go`, `internal/parseoracle/ratchet_test.go` (`corpusReport`), `internal/parseoracle/receipt.go`, `internal/parseoracle/gate_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt` (new), `internal/parseoracle/testdata/ratchet/m3start.txt`, the gate's fixture directory (a sites.txt for its five-file shim, new), `.github/workflows/parse-fidelity.yml`
- commands: `PATH=$HOME/.local/perl-5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`
- entrypoints: `internal/parseoracle/ratchet_test.go:43` (`corpusReport`, TestRatchetCorpus), `internal/parseoracle/gate_test.go:105-208` (TestCIGateRunsTheSweep, `baselineEnv`), `internal/parseoracle/receipt.go`

Commit M3's oracle condition as a ratchet, at the 5.44.0 pin. TestSiteRatchetCorpus takes the report from `corpusReport(t)` and writes per-marker x bucket site counts to `testdata/ratchet/sites.txt`, one `scope marker bucket count` row each (scope `all` or `m3`, the M3 dirs `op mro uni lib test_pl`). It fails in both directions against that file and enforces the floor: M3-dir exact >= 80% of sites, and WRONG = 0 everywhere. The numbers are re-measured here. The 5.42 scratch figures were 5340 sites / 4933 exact / 400 wider / 7 no-answer, but they silently dropped op/magic.t's 46 perl-declined sites. Counted as no-answer, as the counting rule requires, they were 5386 / 4933 / 400 / 53 (M3 dirs 3244 / 2896 / 302 / 46), with rv2hv at 389 wider. 5.44.0 will move them again. This issue also records the starting site counts in m3start.txt (`rv2hv_wider`, `m3_exact_pct`). Later issues compare against that file, not against 389.

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
- [ ] RED: write TestCISweepRunsOracleOnce. It runs the gate's sweep step and asserts from the receipt that parseoracle.Run walked each file once. Change the sweep step's `-run TestRatchetCorpus` in `.github/workflows/parse-fidelity.yml` to a pattern running both tests, point the gate at a fixture sites.txt via the override, and require the receipt's `sites` line. Run TestCIGateRunsTheSweep and TestCISweepRunsOracleOnce, and watch them pass.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The floor rejects M3-dir exact below 80% and any WRONG site (`go test ./internal/parseoracle/ -run '^TestSiteFloorRejectsBelowEightyOrWrong$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorRejectsBelowEightyOrWrong'`)
- [ ] The corpus report is computed once per process (`go test ./internal/parseoracle/ -run '^TestCorpusReportComputesOnce$' -count=1 -v | grep -q '^--- PASS: TestCorpusReportComputesOnce'`)
- [ ] The starting rv2hv wider count is recorded beside the pin (`grep -Eq '^rv2hv_wider = [0-9]+$' internal/parseoracle/testdata/ratchet/m3start.txt`)
- [ ] The CI sweep step runs the site ratchet (`grep -q TestSiteRatchetCorpus .github/workflows/parse-fidelity.yml`)
- [ ] The CI gate executes the sweep step, site ratchet and receipt line included (`go test ./internal/parseoracle/ -run '^TestCIGateRunsTheSweep$' -count=1 -v | grep -q '^--- PASS: TestCIGateRunsTheSweep'`)
- [ ] One CI sweep walks each file once, for both ratchets (`go test ./internal/parseoracle/ -run '^TestCISweepRunsOracleOnce$' -count=1 -v | grep -q '^--- PASS: TestCISweepRunsOracleOnce'`)
- [ ] Regression guard: CI never regenerates the ratchets it checks, now that the sweep step also runs the site ratchet (`go test ./internal/parseoracle/ -run '^TestCINeverRebaselines$' -count=1 -v | grep -q '^--- PASS: TestCINeverRebaselines'`)
- [ ] The per-file and per-site ratchets both match the 620-file sweep at the 5.44.0 pin (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v | grep -cE '^--- PASS: Test(Site)?RatchetCorpus ' | grep -qx 2`)

### Negative Scenarios

- [ ] The site comparison fails when sites.txt is missing or empty, rather than passing with nothing to compare (`go test ./internal/parseoracle/ -run '^TestSiteRatchetFailsWithoutBaselineFile$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetFailsWithoutBaselineFile'`)
- [ ] The site comparison fails on an empty report (0 files or 0 sites), so a corpus that was never walked cannot match a zeroed file (`go test ./internal/parseoracle/ -run '^TestSiteRatchetRejectsEmptyReport$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetRejectsEmptyReport'`)
- [ ] The floor fails when the M3 dirs have 0 sites; a 0/0 ratio is not >= 80% (`go test ./internal/parseoracle/ -run '^TestSiteFloorRejectsZeroSites$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorRejectsZeroSites'`)
- [ ] The floor's denominator includes no-answer and wider sites; 80 exact of 100 sites with 30 no-answer is not scored as 80/70 (`go test ./internal/parseoracle/ -run '^TestSiteFloorCountsNoAnswerInDenominator$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorCountsNoAnswerInDenominator'`)
- [ ] One WRONG site outside the M3 dirs fails the floor too; WRONG = 0 holds over the whole corpus (`go test ./internal/parseoracle/ -run '^TestSiteFloorRejectsWrongOutsideM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorRejectsWrongOutsideM3Dirs'`)
- [ ] A swap fails the ratchet: one file's site moving exact->wider while another's moves wider->exact leaves the per-marker totals unchanged, and the comparison still reports both files (`go test ./internal/parseoracle/ -run '^TestSiteRatchetDetectsSwapAtEqualTotals$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetDetectsSwapAtEqualTotals'`)
- [ ] A report missing files the per-file baseline lists (a sweep timeout or a dropped directory) fails, rather than shrinking the denominator (`go test ./internal/parseoracle/ -run '^TestSiteRatchetFailsOnDroppedFiles$' -count=1 -v | grep -q '^--- PASS: TestSiteRatchetFailsOnDroppedFiles'`)
- [ ] The M3-dir subset matches only the five dirs `op mro uni lib test_pl` by path prefix, so a lookalike such as `opbasic/` (which starts with `op`) is not counted as an M3 dir (`go test ./internal/parseoracle/ -run '^TestM3DirSubsetMatchesExactDirs$' -count=1 -v | grep -q '^--- PASS: TestM3DirSubsetMatchesExactDirs'`)
