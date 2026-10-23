---
title: Make the oracle answer op/magic.t, which perl compiles
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
- 01a16602-5da6-79c8-aea5-e13d8f14c679
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-23T01:08:33.371135176Z
updated: 2026-10-23T01:35:08.813616255Z
---

## Context

- paths: `internal/parseoracle/run.go`, `internal/parseoracle/oracle.go`, `internal/parseoracle/testdata/parse_facts.pl`, `internal/parseoracle/magic_test.go` (new), `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/sites.txt`
- commands: `go test ./internal/parseoracle/ -run '^TestOracleAnswersFileThatPerlCompiles$' -count=1 -v`
- entrypoints: `internal/parseoracle/run.go` (how a perl decline is decided), `internal/parseoracle/testdata/parse_facts.pl`

**perigrin's ruling (2026-10-23): op/magic.t is an oracle defect, and it is fixed in this milestone.** perl compiles op/magic.t (`perl -c`: `syntax OK`, exit 0), and the file runs about 60 of its tests inside BEGIN blocks. The oracle still reports "perl declined to compile the file", so it loses the file's 46 marker sites (at 5.42: rv2hv 21, srefgen 3, anonhash 6, match 13, readline 3). The cause is unverified. One candidate is TAP printed during compilation, which may confuse how the oracle reads perl's output. Until this is fixed, the site ratchet counts those 46 sites as no-answer. That is why this issue follows the site ratchet and re-baselines it.

## Steps

- [ ] Reproduce at the 5.44.0 pin: run the oracle on op/magic.t alone, and capture perl's stdout, stderr and exit status as the oracle sees them. Find the line of code that turns that output into "declined". Write the cause in the test's comment.
- [ ] Create `internal/parseoracle/magic_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestOracleAnswersFileThatPerlCompiles, a minimal fixture reduced from op/magic.t that reproduces the cause. perl compiles it, and the oracle must answer with facts rather than decline. Run it and watch it fail.
- [ ] GREEN: fix the cause in the oracle. Run the test and watch it pass. The test finds perl through `perltest.PinnedPerl`; add it and the negatives below to parse-fidelity.yml's 5.44-held step (`PVM_REQUIRE_PINNED_PERL=1`).
- [ ] RED/GREEN the negatives the fix must not break, each a fixture judged by the pinned perl: TestOracleStillDeclinesFileDyingInBEGIN (a file that dies in BEGIN is still declined), TestOracleExitInBEGINIsNotExact (a file that exits 0 from BEGIN builds no optree and is never graded exact), and TestOracleStderrFromBEGINIsNotAFailure (a file that compiles while printing to STDERR from BEGIN is answered, not classified as a compile failure).
- [ ] Any change to the oracle's perl-side probe or its decline logic gets a per-fix review against perl before merge. The reviewer writes NEW probes outside the corpus, such as a file printing to STDOUT and to STDERR from BEGIN, a file exiting 0 from BEGIN, and a file that dies in BEGIN. Each must still be classified as perl classifies it.
- [ ] Run the sweep (TestRatchetCorpus and TestSiteRatchetCorpus) and re-baseline baseline.txt and sites.txt in the same commit. The message names op/magic.t and every other file that moved.
- [ ] RED/GREEN TestM3StartRv2hvRecordedAfterMagic: m3start.txt has an `rv2hv_wider_note = after op/magic.t` line, and a non-negative `rv2hv_wider`.
- [ ] Re-record m3start.txt's `rv2hv_wider` from sites.txt's `total all rv2hv wider` row now that op/magic.t is answered, and say why in the commit message. The rv2hv issues, which follow this one, measure their fall from this value.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The oracle answers a file perl compiles that prints from BEGIN, rather than declining it (`go test ./internal/parseoracle/ -run '^TestOracleAnswersFileThatPerlCompiles$' -count=1 -v | grep -q '^--- PASS: TestOracleAnswersFileThatPerlCompiles'`)
- [ ] op/magic.t is graded exact or wider in the committed per-file baseline, so excluding it cannot satisfy this (`grep -Eq '^(exact|wider) .* op/magic.t$' internal/parseoracle/testdata/ratchet/baseline.txt`)
- [ ] m3start.txt's rv2hv starting count was re-recorded after op/magic.t was answered, and says so (`go test ./internal/parseoracle/ -run '^TestM3StartRv2hvRecordedAfterMagic$' -count=1 -v | grep -q '^--- PASS: TestM3StartRv2hvRecordedAfterMagic'`)
- [ ] CI runs the op/magic.t fixture under the pinned perl (`grep -q TestOracleAnswersFileThatPerlCompiles .github/workflows/parse-fidelity.yml`)

### Negative Scenarios

- [ ] A file that dies in BEGIN is still declined, so the fix does not grade files perl rejects (`go test ./internal/parseoracle/ -run '^TestOracleStillDeclinesFileDyingInBEGIN$' -count=1 -v | grep -q '^--- PASS: TestOracleStillDeclinesFileDyingInBEGIN'`)
- [ ] A file that exits 0 from BEGIN builds no optree and is never graded exact (`go test ./internal/parseoracle/ -run '^TestOracleExitInBEGINIsNotExact$' -count=1 -v | grep -q '^--- PASS: TestOracleExitInBEGINIsNotExact'`)
- [ ] A file that prints to STDERR from BEGIN but compiles is answered, not taken for a compile failure (`go test ./internal/parseoracle/ -run '^TestOracleStderrFromBEGINIsNotAFailure$' -count=1 -v | grep -q '^--- PASS: TestOracleStderrFromBEGINIsNotAFailure'`)
- [ ] Regression guard (passes today): comp/final_line_num.t, which perl rejects through a BEGIN handler, is still not graded exact (`sh -c 'f=internal/parseoracle/testdata/ratchet/baseline.txt; test -s $f && grep -q " comp/final_line_num.t$" $f && ! grep -Eq "^exact .* comp/final_line_num.t$" $f'`)
