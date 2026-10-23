---
title: Fix the remaining unfaithful M3-dir files and hold re-emission at 100%
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-60e9-7343-a630-7a93da5ee389
created: 2026-10-23T07:06:50.727978535Z
updated: 2026-10-23T07:06:59.861764122Z
---

## Context

- paths: `internal/parse/perlgitt_faithful_test.go` (TestPerlGitTFaithfulOnM3Dirs), `internal/parse/testdata/faithful_queue.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parse/pin_test.go` (TestCIRunsTheRatchets), `.github/workflows/ci.yml` (the skip-guard loop, ~207), plus canon.go, fidelity.go and the parser files each fix touches
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^(TestPerlGitTFaithfulOnM3Dirs|TestFaithfulQueueMatchesTree)$' -count=1 -v`
- entrypoints: `internal/parse/testdata/faithful_queue.txt` (the `other` rows), `internal/parse/perlgitt_faithful_test.go`

**perigrin's ruling (2026-10-23, chain-review round 4): re-emission is 100% faithful on the M3 dirs, with no defects file.** This is the last re-emission issue. It owns the `other` class: every queue row that no triage issue's class fits, including rows re-classified as `other` by earlier fixes. It fixes each one. It then writes TestPerlGitTFaithfulOnM3Dirs, which asserts that every M3-dir file (`op mro uni lib test_pl`) that parses re-emits faithfully, with NO exception list. The milestone's resolution command runs that test. When this issue closes, the work queue holds no rows (comment lines only), and the M3-dir faithful count is above `faithful_start_m3` (178 at 054e75e6) in m3start.txt. A row that proves unfixable is escalated to perigrin with its perl measurement. It is never parked, and this issue does not close around it.

Size: unknown until the triage issues finish. If more than about 10 `other` rows remain when this issue starts, split it by cause before starting, as the triage issues were split.

## Steps

- [ ] List the queue's `other` rows, and any rows of other classes left by closed owners. For each, reduce the first difference to a minimal snippet and judge it with the pinned perl 5.44.0 (`perl -MO=Deparse,-p` or `perl -MO=Concise`).
- [ ] RED: add each reduced snippet to TestFaithfulOtherRowsFixed in fidelity_test.go, quoting perl's output. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time, in canon, `Faithful` (only where the measurement shows the difference is cosmetic) or the parser. Run the test after each.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/faithful-other.txt`.
- [ ] RED/GREEN TestPerlGitTFaithfulOnM3DirsIgnoresTheQueue: with a fixture tree holding one unfaithful file and a queue listing it, TestPerlGitTFaithfulOnM3Dirs's check still fails.
- [ ] RED: write TestPerlGitTFaithfulOnM3Dirs, which walks every M3-dir file outside `perlRejects` and fails on any file that parses and re-emits unfaithfully. It has no exception list. Run it while rows remain and watch it fail.
- [ ] GREEN: empty the queue by fixing the remaining rows. Run TestPerlGitTFaithfulOnM3Dirs and TestFaithfulQueueMatchesTree and watch both pass, with the queue holding only comments.
- [ ] Re-baseline perlgittfaithful.ratchet in the same commit as each fix, naming each moved file.
- [ ] Add TestPerlGitTFaithfulOnM3Dirs to ci.yml's "Fail if a corpus-backed ratchet skipped" loop and to TestCIRunsTheRatchets' list.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements, writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file so TestM3ProbesHeldToPerl runs them from then on.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Every M3-dir file that parses re-emits faithfully, with no exception list (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] Each reduced `other` case re-emits faithfully (`go test ./internal/parse/ -run '^TestFaithfulOtherRowsFixed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulOtherRowsFixed'`)
- [ ] The work queue holds no rows, only comments (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -q "^[^#]" $f'`)
- [ ] The work queue still matches the tree, now empty (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] The M3-dir faithful count in the committed ratchet is above the start count recorded in m3start.txt (`sh -c 's=$(sed -n "s/^faithful_start_m3 = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(awk "\$1==\"0\" && \$2 ~ /^(op|mro|uni|lib|test_pl)\// {k++} END {print k+0}" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ "$n" -gt "$s" ]'`)
- [ ] CI fails if the 100% check skips: the test is in TestCIRunsTheRatchets' list, and that test passes (`sh -c 'grep -q "\"TestPerlGitTFaithfulOnM3Dirs\"," internal/parse/pin_test.go && go test ./internal/parse/ -run "^TestCIRunsTheRatchets$" -count=1 -v | grep -q "^--- PASS: TestCIRunsTheRatchets"'`)
- [ ] The reduced cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/faithful-other.txt`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)

### Negative Scenarios

- [ ] A queue holding any row fails the 100% check even when that row matches the tree, so the queue cannot serve as an exception list (`go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3DirsIgnoresTheQueue$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3DirsIgnoresTheQueue'`)
