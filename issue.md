---
title: Ratchet canonical re-emission over perl.git t/ at 342/619
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5ebe-7169-b857-996ae2953394
created: 2026-10-22T23:46:33.694596407Z
updated: 2026-10-23T07:06:16.487966742Z
---

## Context

- paths: `internal/parse/perlgitt_faithful_test.go` (new), `internal/parse/testdata/perlgittfaithful.ratchet` (new), `internal/parseoracle/testdata/ratchet/m3start.txt` (append), `internal/parse/pin_test.go` (TestCIRunsTheRatchets), `.github/workflows/ci.yml` (the skip-guard loop, ~207)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (`Faithful`), `internal/parse/fidelity_test.go:111` (TestCanonFidelityRatchet), `internal/parse/perlgitt_test.go:114` (TestPerlGitTRatchet)

Canonical re-emission (`Faithful`) has only ever run on T1. A scratch sweep measured perl.git t/ at 342/619 faithful, and the M3 dirs at 178/347. This issue commits that measurement as a per-file 0/1 ratchet over perl.git t/, in the format of t1fidelity.ratchet, so that each later re-emission fix shows up as a recorded gain. It walks the files as TestPerlGitTRatchet does (`Session.ParseFileFrom` over `perlGitTRoots`, skipping `perlRejects`). Test-only; no change to canon or fidelity.

## Steps

- [ ] Create `internal/parse/perlgitt_faithful_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestPerlGitTFaithfulRatchet, which compares per-file results to `testdata/perlgittfaithful.ratchet`, fails on any difference in either direction, and regenerates under the existing update flag. Run it without the file and watch it fail.
- [ ] GREEN: generate the ratchet. Its header line is exactly `# N/619 faithful`, and each row is `<0|1> <path relative to t/>` with 1 = differs, as in t1fidelity.ratchet; later issues' criteria parse both. It must read 342/619; if it does not, explain the difference before committing. Run it and watch it pass.
- [ ] Append `faithful_start_all` (the header's N) and `faithful_start_m3` (rows with 0 whose path is under `op/ mro/ uni/ lib/ test_pl/`, 178 at 054e75e6) to m3start.txt, in the same commit as the ratchet. m3start.txt is append-only: never edit or reorder an existing line. The re-emission issues assert their gains against these two values (perigrin's ruling, chain-review round 4; P21, H1).
- [ ] `Faithful` does not reject an Unknown node (canon.go:631 emits an Unknown as its own text), so the walk checks for Unknown first. That is test code, not a change to fidelity.go.
- [ ] Add TestPerlGitTFaithfulRatchet to ci.yml's "Fail if a corpus-backed ratchet skipped" loop and to TestCIRunsTheRatchets' list (internal/parse/pin_test.go:71), and confirm that the job sets the corpus env it needs.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The faithful ratchet over perl.git t/ is committed (`test -f internal/parse/testdata/perlgittfaithful.ratchet`)
- [ ] The ratchet's header states the faithful count in the form later criteria parse (`grep -Eq '^# [0-9]+/619 faithful' internal/parse/testdata/perlgittfaithful.ratchet`)
- [ ] CI fails if the test skips: the test is in TestCIRunsTheRatchets' list, and that test passes (`sh -c 'grep -q "\"TestPerlGitTFaithfulRatchet\"," internal/parse/pin_test.go && go test ./internal/parse/ -run "^TestCIRunsTheRatchets$" -count=1 -v | grep -q "^--- PASS: TestCIRunsTheRatchets"'`)
- [ ] Per-file re-emission over perl.git t/ matches the ratchet (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] The faithful start counts in m3start.txt equal the ratchet in the commit that recorded them, so typed numbers cannot satisfy it (`sh -c 'f=internal/parseoracle/testdata/ratchet/m3start.txt; r=internal/parse/testdata/perlgittfaithful.ratchet; c=$(git log -1 --format=%H -G "^faithful_start_" -- $f); [ -n "$c" ] || exit 1; a=$(git show $c:$f | sed -n "s/^faithful_start_all = \([0-9][0-9]*\)$/\1/p"); m=$(git show $c:$f | sed -n "s/^faithful_start_m3 = \([0-9][0-9]*\)$/\1/p"); h=$(git show $c:$r | sed -n "s|^# \([0-9][0-9]*\)/619 faithful.*|\1|p"); n=$(git show $c:$r | awk "\$1==\"0\" && \$2 ~ /^(op|mro|uni|lib|test_pl)\// {k++} END {print k+0}"); [ -n "$a" ] && [ "$a" = "$h" ] && [ -n "$m" ] && [ "$m" = "$n" ]'`)

### Negative Scenarios

- [ ] The ratchet's stated header count must equal its count of faithful rows; a header reading 342/619 over different rows fails (`go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchetHeaderMatchesRows$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchetHeaderMatchesRows'`)
- [ ] A swap fails: one file flipping 1->0 while another flips 0->1 keeps 342 but fails the per-file comparison, naming both files (`go test ./internal/parse/ -run '^TestFaithfulRatchetDetectsSwap$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRatchetDetectsSwap'`)
- [ ] A ratchet row whose file the walk did not visit (missing from PERL5_CORPUS, or a parse that timed out) fails, rather than dropping out of the 619 (`go test ./internal/parse/ -run '^TestFaithfulRatchetFailsOnUnvisitedFile$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRatchetFailsOnUnvisitedFile'`)
- [ ] The perl.git t/ walk counts a file whose tree holds an Unknown node as unfaithful, checking for Unknown before it calls `Faithful`, even when the Unknown re-emits its source text verbatim (`go test ./internal/parse/ -run '^TestFaithfulWalkCountsUnknownTreeUnfaithful$' -count=1 -v | grep -q '^--- PASS: TestFaithfulWalkCountsUnknownTreeUnfaithful'`)
- [ ] An empty walk (0 files visited) fails, not matches an empty ratchet (`go test ./internal/parse/ -run '^TestFaithfulRatchetRejectsEmptyWalk$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRatchetRejectsEmptyWalk'`)
