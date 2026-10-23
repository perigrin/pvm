---
title: Hold psc's Paren flag to perl's OPf_PARENS bit, op by op
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-75d4-76eb-aad2-10b322f366a9
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16602-5fd5-7cc4-9645-4539ca73e839
created: 2026-10-22T23:46:33.882832365Z
updated: 2026-10-23T07:06:22.223394852Z
---

## Context

- paths: `internal/parseoracle/parens_test.go` (new), `internal/parseoracle/testdata/parens/` (new), `internal/parseoracle/testdata/parse_facts.pl` (the perl-side B walk, extended to report OPf_PARENS), `internal/parse/` (the `Paren` flag; parser fixes allowed, see Steps), `internal/parse/testdata/faithful_queue.txt` (the `dropped-paren` rows), `internal/parse/testdata/m3probes/opf-parens.txt` (new), `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/sites.txt`
- commands: `go test ./internal/parseoracle/ -run '^(TestParenComparisonReportsMisattachedParen|TestParenFlagMatchesPerlOPfPARENS)$' -count=1 -v`
- entrypoints: `internal/parseoracle/testdata/parse_facts.pl`, `internal/parseoracle/cvwalk_test.go`, `internal/parse/term.go` (`finishList`)

The canon change writes a `Paren` node's parens, so whether psc attached each paren to the right node now matters, and perl can judge it. perl keeps one boolean bit, OPf_PARENS (`P` in Concise), on the op a source paren enclosed: `my $z = ($x)` is `padsv sP` vs `padsv s`, and `ok(($y eq "x"))` is `seq sKPM` vs `seq sKM`. This check walks perl's optree for a fixture and collects (line, op, PARENS bit). It compares that op by op with psc's nodes and their `Paren` flag, and reports a paren psc put on the wrong node, a paren psc dropped, and a paren psc invented. Tests run perl; production psc never does. Some ops perl folds or optimises away will have no counterpart; the check names those as unmatched rather than guessing. It runs first among the issues that write faithful_queue.txt, before the heredoc/sub-less triage issue, so the two never edit that file and the parser in parallel. **Why the sequence stays (A23, chain-review round 3):** the shared-generated-files rule lets parallel branches regenerate faithful_queue.txt and the ratchets. It does not cover canon.go, fidelity.go and fidelity_test.go, which are hand-written and edited by every one of these issues. Running them in sequence also gives 5f77's handover rule a defined order: a fix that moves a row into a class whose owner has closed is handled by the issue that moved it, which only works if owners close in a known order. The cost is a longer critical chain, accepted over hand-merging four sets of canon/fidelity changes. Dropped-paren fixes that expose a later difference re-classify the row for the triage issues that follow. This issue also owns the `dropped-paren` rows of faithful_queue.txt: source parens canon still does not write after the Paren-once change, because psc never set `Paren` on that node. The OPf_PARENS comparison is what judges them. Each is fixed in the parser; the class must reach zero. perigrin ruled the check into this milestone on 2026-10-22.

## Steps

- [ ] Measure with perl first. Run `perl -MO=Concise` on each fixture and record which op carries `P`. Use the pinned perl 5.44.0. Include cases where the paren is on a list op (`my ($x) = @_`), on a scalar op (`my $z = ($x)`, with `my $x` declared: for a package `$x` the `P` sits on a nulled `ex-rv2sv sKP`), doubled (`(($y eq "x"))`), and around a call argument. Quote the output in the fixtures.
- [ ] Create `internal/parseoracle/parens_test.go` with two `// ABOUTME:` lines, and fixtures in a readable text format under testdata/parens/.
- [ ] RED: write TestParenComparisonReportsMisattachedParen, a unit test with no perl. Given hand-built perl-side facts and a psc tree whose `Paren` sits on the wrong node, the comparison reports it, and it reports nothing for a matching pair. Run it and watch it fail.
- [ ] GREEN: implement the comparison. Run it and watch it pass.
- [ ] RED: write TestParenFlagMatchesPerlOPfPARENS. It runs the pinned perl (corpus.pin's interpreter, found through `perltest.PinnedPerl`) with B on each fixture, reads OPf_PARENS per op, and compares with psc's tree. Run it and watch it fail until the B walk reads the flag.
- [ ] GREEN: extend the walk to read op flags. Run it and watch it pass. A fixture where psc disagrees with perl is a parser defect: fix it here, with the per-fix review below, or file it as an issue in this milestone and mark the fixture as a named known defect, never silently.
- [ ] Work the `dropped-paren` rows of faithful_queue.txt: for each, compare psc's `Paren` with perl's bit, then fix the parser so psc's `Paren` matches perl's bit and the row leaves the queue. The class must reach zero.
- [ ] Add the fixtures, with perl's facts, as probe file `internal/parse/testdata/m3probes/opf-parens.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] parse_facts.pl's hash keys the oracle cache (cache.go:52), so extending it re-measures all 620 files.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Every test here that runs perl uses `perltest.PinnedPerl`. Add them to parse-fidelity.yml's 5.44-held step (`PVM_REQUIRE_PINNED_PERL=1`).
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue, and never close this issue with a row of its class left (perigrin's ruling, chain-review round 4).
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The comparison reports a paren attached to the wrong node, and nothing for a match (`go test ./internal/parseoracle/ -run '^TestParenComparisonReportsMisattachedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsMisattachedParen'`)
- [ ] psc's Paren flag matches perl's OPf_PARENS bit op by op on every fixture (`go test ./internal/parseoracle/ -run '^TestParenFlagMatchesPerlOPfPARENS$' -count=1 -v | grep -q '^--- PASS: TestParenFlagMatchesPerlOPfPARENS'`)
- [ ] The fixtures are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/opf-parens.txt`)
- [ ] Every test that calls PinnedPerl, this issue's included, is in the -run list of CI's REQUIRE step (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequired$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequired'`)
- [ ] No dropped-paren row is left in the work queue (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -Eq "^[^ #]+ (dropped-paren) " $f'`)
- [ ] The work queue still matches the tree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)

### Negative Scenarios

- [ ] The perl-side walk reads OPf_PARENS on nulled ops. perl 5.42 compiles `f((1,2))` (with `sub f {}` declared) to an `ex-list lKPM` that carries the only P bit, while `f(1,2)` has none. A walk that skips ex- ops would report nothing on both sides and pass (`go test ./internal/parseoracle/ -run '^TestParenWalkReadsPARENSOnNulledOps$' -count=1 -v | grep -q '^--- PASS: TestParenWalkReadsPARENSOnNulledOps'`)
- [ ] The comparison fails a fixture whose source holds a paren when both sides report zero P ops, rather than calling 0 = 0 a match (`go test ./internal/parseoracle/ -run '^TestParenComparisonRejectsEmptyFactsForParenthesisedSource$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonRejectsEmptyFactsForParenthesisedSource'`)
- [ ] A dropped paren is reported: perl has P on `padsv` in `my $z = ($x)` and psc has no `Paren` on the node (`go test ./internal/parseoracle/ -run '^TestParenComparisonReportsDroppedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsDroppedParen'`)
- [ ] An invented paren is reported: psc sets `Paren` on a node where perl has no P, for example `print (1)+2`, whose `const[IV 1]` is `s` and not `sP` (`go test ./internal/parseoracle/ -run '^TestParenComparisonReportsInventedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsInventedParen'`)
- [ ] A call's argument parens are not taken for a paren on the argument. perl measures `print ((1))+2` as `const[IV 1] sP` and `print (1)+2` as `const[IV 1] s`, and psc's `Paren` matches each (`go test ./internal/parseoracle/ -run '^TestParenFlagCallArgParensAreNotAParen$' -count=1 -v | grep -q '^--- PASS: TestParenFlagCallArgParensAreNotAParen'`)
- [ ] A fixture whose every P op is unmatched (folded or optimised away) fails unless it is listed as a named known gap; "unmatched" never counts as agreement (`go test ./internal/parseoracle/ -run '^TestParenUnmatchedOpsAreNotAgreement$' -count=1 -v | grep -q '^--- PASS: TestParenUnmatchedOpsAreNotAgreement'`)
- [ ] The fixture set cannot pass empty: it holds at least one fixture for each measured class (list op, scalar op, doubled, call argument, nulled `ex-list`), and the test fails if any class has none (`go test ./internal/parseoracle/ -run '^TestParenFixturesCoverEveryClass$' -count=1 -v | grep -q '^--- PASS: TestParenFixturesCoverEveryClass'`)
- [ ] When perl fails to compile a fixture, or prints no ops, TestParenFlagMatchesPerlOPfPARENS fails with perl's stderr, rather than skipping the fixture (`go test ./internal/parseoracle/ -run '^TestParenFlagFailsWhenPerlProducesNoOps$' -count=1 -v | grep -q '^--- PASS: TestParenFlagFailsWhenPerlProducesNoOps'`)
