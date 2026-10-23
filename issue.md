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
created: 2026-10-22T23:46:33.882832365Z
updated: 2026-10-23T01:08:52.790449325Z
---

## Context

- paths: `internal/parseoracle/parens_test.go` (new), `internal/parseoracle/testdata/parens/` (new), `internal/parseoracle/testdata/parse_facts.pl` (the perl-side B walk, extended to report OPf_PARENS), `internal/parse/` (the `Paren` flag; parser fixes allowed, see Steps), `internal/parse/testdata/faithful_defects.txt` (the `dropped-paren` rows), `internal/parse/testdata/m3probes/opf-parens.txt` (new), `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/sites.txt`
- commands: `PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^(TestParenComparisonReportsMisattachedParen|TestParenFlagMatchesPerlOPfPARENS)$' -count=1 -v`
- entrypoints: `internal/parseoracle/testdata/parse_facts.pl`, `internal/parseoracle/cvwalk_test.go`, `internal/parse/term.go` (`finishList`)

The canon change writes a `Paren` node's parens, so whether psc attached each paren to the right node now matters, and perl can judge it. perl keeps one boolean bit, OPf_PARENS (`P` in Concise), on the op a source paren enclosed: `my $z = ($x)` is `padsv sP` vs `padsv s`, and `ok(($y eq "x"))` is `seq sKPM` vs `seq sKM`. This check walks perl's optree for a fixture and collects (line, op, PARENS bit). It compares that op by op with psc's nodes and their `Paren` flag, and reports a paren psc put on the wrong node, a paren psc dropped, and a paren psc invented. Tests run perl; production psc never does. Some ops perl folds or optimises away will have no counterpart; the check names those as unmatched rather than guessing. This issue also owns the `dropped-paren` rows of faithful_defects.txt: source parens canon still does not write after the Paren-once change, because psc never set `Paren` on that node. The OPf_PARENS comparison is what judges them. Each is fixed in the parser, or kept with perl's measured bit in its reason. perigrin ruled the check into this milestone on 2026-10-22.

## Steps

- [ ] Measure with perl first. Run `perl -MO=Concise` on each fixture and record which op carries `P`. Use the pinned perl 5.44.0. Include cases where the paren is on a list op (`my ($x) = @_`), on a scalar op (`my $z = ($x)`, with `my $x` declared: for a package `$x` the `P` sits on a nulled `ex-rv2sv sKP`), doubled (`(($y eq "x"))`), and around a call argument. Quote the output in the fixtures.
- [ ] Create `internal/parseoracle/parens_test.go` with two `// ABOUTME:` lines, and fixtures in a readable text format under testdata/parens/.
- [ ] RED: write TestParenComparisonReportsMisattachedParen, a unit test with no perl. Given hand-built perl-side facts and a psc tree whose `Paren` sits on the wrong node, the comparison reports it, and it reports nothing for a matching pair. Run it and watch it fail.
- [ ] GREEN: implement the comparison. Run it and watch it pass.
- [ ] RED: write TestParenFlagMatchesPerlOPfPARENS. It runs the pinned perl (corpus.pin's interpreter, found on PATH) with B on each fixture, reads OPf_PARENS per op, and compares with psc's tree. Run it and watch it fail until the B walk reads the flag.
- [ ] GREEN: extend the walk to read op flags. Run it and watch it pass. A fixture where psc disagrees with perl is a parser defect: fix it here, with the per-fix review below, or file it as an issue in this milestone and mark the fixture as a named known defect, never silently.
- [ ] Work the `dropped-paren` rows of faithful_defects.txt: for each, compare psc's `Paren` with perl's bit, then fix the parser or keep the row with the measured bit quoted in its reason. RED/GREEN TestDroppedParenRowsQuoteParensBit: every kept `dropped-paren` row quotes a Concise `P` measurement.
- [ ] Add the fixtures, with perl's facts, as probe file `internal/parse/testdata/m3probes/opf-parens.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] parse_facts.pl's hash keys the oracle cache (cache.go:52), so extending it re-measures all 620 files.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The comparison reports a paren attached to the wrong node, and nothing for a match (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenComparisonReportsMisattachedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsMisattachedParen'`)
- [ ] psc's Paren flag matches perl's OPf_PARENS bit op by op on every fixture (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenFlagMatchesPerlOPfPARENS$' -count=1 -v | grep -q '^--- PASS: TestParenFlagMatchesPerlOPfPARENS'`)
- [ ] Every kept dropped-paren defect row quotes perl's measured OPf_PARENS bit (`go test ./internal/parse/ -run '^TestDroppedParenRowsQuoteParensBit$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenRowsQuoteParensBit'`)
- [ ] The fixtures are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/opf-parens.txt`)

### Negative Scenarios

- [ ] The perl-side walk reads OPf_PARENS on nulled ops. perl 5.42 compiles `f((1,2))` (with `sub f {}` declared) to an `ex-list lKPM` that carries the only P bit, while `f(1,2)` has none. A walk that skips ex- ops would report nothing on both sides and pass (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenWalkReadsPARENSOnNulledOps$' -count=1 -v | grep -q '^--- PASS: TestParenWalkReadsPARENSOnNulledOps'`)
- [ ] The comparison fails a fixture whose source holds a paren when both sides report zero P ops, rather than calling 0 = 0 a match (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenComparisonRejectsEmptyFactsForParenthesisedSource$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonRejectsEmptyFactsForParenthesisedSource'`)
- [ ] A dropped paren is reported: perl has P on `padsv` in `my $z = ($x)` and psc has no `Paren` on the node (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenComparisonReportsDroppedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsDroppedParen'`)
- [ ] An invented paren is reported: psc sets `Paren` on a node where perl has no P, for example `print (1)+2`, whose `const[IV 1]` is `s` and not `sP` (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenComparisonReportsInventedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsInventedParen'`)
- [ ] A call's argument parens are not taken for a paren on the argument. perl measures `print ((1))+2` as `const[IV 1] sP` and `print (1)+2` as `const[IV 1] s`, and psc's `Paren` matches each (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenFlagCallArgParensAreNotAParen$' -count=1 -v | grep -q '^--- PASS: TestParenFlagCallArgParensAreNotAParen'`)
- [ ] A fixture whose every P op is unmatched (folded or optimised away) fails unless it is listed as a named known gap; "unmatched" never counts as agreement (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenUnmatchedOpsAreNotAgreement$' -count=1 -v | grep -q '^--- PASS: TestParenUnmatchedOpsAreNotAgreement'`)
- [ ] The fixture set cannot pass empty: it holds at least one fixture for each measured class (list op, scalar op, doubled, call argument, nulled `ex-list`), and the test fails if any class has none (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenFixturesCoverEveryClass$' -count=1 -v | grep -q '^--- PASS: TestParenFixturesCoverEveryClass'`)
- [ ] When perl fails to compile a fixture, or prints no ops, TestParenFlagMatchesPerlOPfPARENS fails with perl's stderr, rather than skipping the fixture (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH go test ./internal/parseoracle/ -run '^TestParenFlagFailsWhenPerlProducesNoOps$' -count=1 -v | grep -q '^--- PASS: TestParenFlagFailsWhenPerlProducesNoOps'`)
