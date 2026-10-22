---
title: Ratchet canonical re-emission over perl.git t/ at 342/619
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5ebe-7169-b857-996ae2953394
created: 2026-10-22T23:46:33.694596407Z
updated: 2026-10-22T23:56:40.389207328Z
---

## Context

- paths: `internal/parse/perlgitt_faithful_test.go` (new), `internal/parse/testdata/perlgittfaithful.ratchet` (new)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (`Faithful`), `internal/parse/fidelity_test.go:111` (TestCanonFidelityRatchet), `internal/parse/perlgitt_test.go:114` (TestPerlGitTRatchet)

Canonical re-emission (`Faithful`) has only ever run on T1. A scratch sweep measured perl.git t/ at 342/619 faithful, and the M3 dirs at 178/347. This issue commits that measurement as a per-file 0/1 ratchet over perl.git t/, in the format of t1fidelity.ratchet, so that each later re-emission fix shows up as a recorded gain. It walks the files as TestPerlGitTRatchet does (`Session.ParseFileFrom` over `perlGitTRoots`, skipping `perlRejects`). Test-only; no change to canon or fidelity.

## Steps

- [ ] Create `internal/parse/perlgitt_faithful_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestPerlGitTFaithfulRatchet, which compares per-file results to `testdata/perlgittfaithful.ratchet`, fails on any difference in either direction, and regenerates under the existing update flag. Run it without the file and watch it fail.
- [ ] GREEN: generate the ratchet. Its header states the faithful count, and it must read 342/619; if it does not, explain the difference before committing. Run it and watch it pass.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The faithful ratchet over perl.git t/ is committed (`test -f internal/parse/testdata/perlgittfaithful.ratchet`)
- [ ] Per-file re-emission over perl.git t/ matches the ratchet (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)

### Negative Scenarios

- [ ] The ratchet's stated header count must equal its count of faithful rows; a header reading 342/619 over different rows fails (`go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchetHeaderMatchesRows$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchetHeaderMatchesRows'`)
- [ ] A swap fails: one file flipping 1->0 while another flips 0->1 keeps 342 but fails the per-file comparison, naming both files (`go test ./internal/parse/ -run '^TestFaithfulRatchetDetectsSwap$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRatchetDetectsSwap'`)
- [ ] A ratchet row whose file the walk did not visit (missing from PERL5_CORPUS, or a parse that timed out) fails, rather than dropping out of the 619 (`go test ./internal/parse/ -run '^TestFaithfulRatchetFailsOnUnvisitedFile$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRatchetFailsOnUnvisitedFile'`)
- [ ] A tree holding an Unknown node is never counted faithful, even when the Unknown re-emits its source text verbatim (`go test ./internal/parse/ -run '^TestFaithfulRejectsTreeWithUnknown$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsTreeWithUnknown'`)
- [ ] An empty walk (0 files visited) fails, not matches an empty ratchet (`go test ./internal/parse/ -run '^TestFaithfulRatchetRejectsEmptyWalk$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRatchetRejectsEmptyWalk'`)
