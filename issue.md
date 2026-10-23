---
title: Fix the dropped-paren rows of the re-emission work queue
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5f1a-7d56-a91d-5fa2c478d608
created: 2026-10-23T08:21:55.619658958Z
updated: 2026-10-23T08:21:58.313538944Z
---

## Context

- paths: `internal/parse/testdata/faithful_queue.txt` (the `dropped-paren` rows), the parser files each fix touches (e.g. `internal/parse/term.go`), `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parse/testdata/m3probes/dropped-paren-rows.txt` (new), `internal/parseoracle/testdata/parens/` (fixtures for each fixed row)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v`
- entrypoints: `internal/parse/testdata/faithful_queue.txt`, `internal/parseoracle/parens_test.go` (TestParenFlagMatchesPerlOPfPARENS)

Fix every `dropped-paren` row of the re-emission work queue: a source paren canon still does not write after the Paren-once change, because psc never set `Paren` on that node. The OPf_PARENS check from the issue that blocks this one judges each row. Split out of that issue (J4, chain-review round 5) so the check and the row-fixing are separate deliverables. Under perigrin's round-4 ruling the class must reach zero; an unfixable row is escalated to perigrin, never parked. Under the round-5 ruling `Faithful` is not changed; its forgiven set is closed. It runs first among the issues that write faithful_queue.txt, before the heredoc/sub-less triage issue, so the two never edit that file and the parser in parallel. **Why the sequence stays (A23, chain-review round 3):** the shared-generated-files rule lets parallel branches regenerate faithful_queue.txt and the ratchets. It does not cover canon.go, fidelity.go and fidelity_test.go, which are hand-written and edited by every one of these issues. Running them in sequence also gives 5f77's handover rule a defined order: a fix that moves a row into a class whose owner has closed is handled by the issue that moved it, which only works if owners close in a known order. The cost is a longer critical chain, accepted over hand-merging four sets of canon/fidelity changes. Dropped-paren fixes that expose a later difference re-classify the row for the triage issues that follow.

## Steps

- [ ] For each `dropped-paren` row, reduce it to a fixture, run the pinned perl 5.44.0's `perl -MO=Concise` on it, and record which op carries P.
- [ ] RED: add each fixture to testdata/parens/ so TestParenFlagMatchesPerlOPfPARENS fails on it. Run it and watch it fail.
- [ ] GREEN: fix the parser so psc's `Paren` matches perl's bit and the row leaves the queue. One cause at a time; run the test after each.
- [ ] Add every fixture, with perl's facts, as probe file `internal/parse/testdata/m3probes/dropped-paren-rows.txt`.
- [ ] Remove the fixed rows from faithful_queue.txt, and re-baseline perlgittfaithful.ratchet in the same commit, naming each moved file.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements, writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] No dropped-paren row is left in the work queue (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -Eq "^[^[:space:]#]+[[:space:]]+(dropped-paren)[[:space:]]" $f'`)
- [ ] The work queue still matches the tree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] psc's Paren flag still matches perl's OPf_PARENS bit on every fixture, the fixed rows' fixtures included (`go test ./internal/parseoracle/ -run '^TestParenFlagMatchesPerlOPfPARENS$' -count=1 -v | grep -q '^--- PASS: TestParenFlagMatchesPerlOPfPARENS'`)
- [ ] The fixed rows are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/dropped-paren-rows.txt`)
- [ ] Regression guard: the forgiven set is still exactly the three ruled differences (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)

### Negative Scenarios

- [ ] No row is fixed by inventing a paren: a `Paren` psc sets where perl has no P bit is still reported (`go test ./internal/parseoracle/ -run '^TestParenComparisonReportsInventedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsInventedParen'`)
