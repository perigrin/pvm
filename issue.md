---
title: Make the oracle answer op/magic.t, which perl compiles
state: pending
urgency: normal
milestone: m3-oracle-at-scale
created: 2026-10-23T01:08:33.371135176Z
updated: 2026-10-23T01:08:33.371135176Z
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
- [ ] GREEN: fix the cause in the oracle. Run the test and watch it pass.
- [ ] Any change to the oracle's perl-side probe or its decline logic gets a per-fix review against perl before merge. The reviewer writes NEW probes outside the corpus, such as a file printing to STDOUT and to STDERR from BEGIN, a file exiting 0 from BEGIN, and a file that dies in BEGIN. Each must still be classified as perl classifies it.
- [ ] Run the sweep (TestRatchetCorpus and TestSiteRatchetCorpus) and re-baseline baseline.txt and sites.txt in the same commit. The message names op/magic.t and every other file that moved.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The oracle answers a file perl compiles that prints from BEGIN, rather than declining it (`go test ./internal/parseoracle/ -run '^TestOracleAnswersFileThatPerlCompiles$' -count=1 -v | grep -q '^--- PASS: TestOracleAnswersFileThatPerlCompiles'`)
- [ ] op/magic.t is no longer a no-answer row in the committed per-file baseline (`sh -c 'grep -q " op/magic.t$" internal/parseoracle/testdata/ratchet/baseline.txt && ! grep -q "^no-answer .* op/magic.t$" internal/parseoracle/testdata/ratchet/baseline.txt'`)
