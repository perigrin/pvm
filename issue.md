---
title: Ratchet oracle agreement per marker site, with the M3 floor, in the CI sweep
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-3e74-7a1a-b7bf-090cd8bceadd
blocks:
- 01a12059-30a8-722e-bbc3-52066b8610cb
- 01a16602-61a2-78de-a798-da4cef3208c5
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:33.509926758Z
updated: 2026-10-22T23:47:11.481918828Z
---

## Context

- paths: `internal/parseoracle/sites_test.go`, `internal/parseoracle/testdata/ratchet/sites.txt` (new), `.github/workflows/parse-fidelity.yml`
- commands: `PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestSiteRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v`
- entrypoints: `internal/parseoracle/ratchet_test.go` (`corpusReport`, TestRatchetCorpus), `internal/parseoracle/gate_test.go:105` (TestCIGateRunsTheSweep)

Commit M3's oracle condition as a ratchet. TestSiteRatchetCorpus takes the report from `corpusReport(t)` and writes per-marker x bucket site counts to `testdata/ratchet/sites.txt`, for the whole corpus and for the M3 dirs (`op mro uni lib test_pl`). It fails in both directions against that file, and it enforces the floor: M3-dir exact >= 80% of sites, and WRONG = 0 everywhere. Expected at 054e75e6: whole corpus 5340 sites / 4933 exact / 400 wider / 7 no-answer; M3 dirs 3198 / 2896 / 302 / 0; rv2hv 540 sites, 389 wider. Every later rv2hv change is judged by this file, not by a scratch measurement. The CI sweep step runs TestRatchetCorpus today, so this test joins that same `go test` invocation (one `-run` pattern sharing `corpusReport`) rather than running a second sweep.

## Steps

- [ ] RED: write TestSiteFloorRejectsBelowEightyOrWrong, a unit test of the floor check. 79.9% M3-dir exact fails, one WRONG site fails, and 80% with 0 WRONG passes. Run it and watch it fail.
- [ ] GREEN: implement the floor check. Run the test and watch it pass.
- [ ] RED: write TestSiteRatchetCorpus. Its counts come from the per-site helper; it compares them to sites.txt and fails on any difference, gain or loss; it regenerates the file under `-parseoracle.update`. Run it without sites.txt and watch it fail.
- [ ] GREEN: generate sites.txt with `-parseoracle.update`. Check its totals against the measurement above, and explain any difference in the commit message. Run the sweep and watch it pass.
- [ ] Change the sweep step's `-run TestRatchetCorpus` in `.github/workflows/parse-fidelity.yml` to a pattern running both tests in one invocation. Run TestCIGateRunsTheSweep and watch it pass.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The floor rejects M3-dir exact below 80% and any WRONG site (`go test ./internal/parseoracle/ -run '^TestSiteFloorRejectsBelowEightyOrWrong$' -count=1 -v | grep -q '^--- PASS: TestSiteFloorRejectsBelowEightyOrWrong'`)
- [ ] The sites ratchet file is committed (`test -f internal/parseoracle/testdata/ratchet/sites.txt`)
- [ ] The CI sweep step runs the site ratchet (`grep -q TestSiteRatchetCorpus .github/workflows/parse-fidelity.yml`)
- [ ] The CI gate still executes the sweep step (`go test ./internal/parseoracle/ -run '^TestCIGateRunsTheSweep$' -count=1 -v | grep -q '^--- PASS: TestCIGateRunsTheSweep'`)
- [ ] The per-site ratchet matches sites.txt over the 620-file sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestSiteRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v | grep -q '^--- PASS: TestSiteRatchetCorpus'`)
