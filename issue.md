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
- 01a167da-3323-7ce0-a627-abf5b99603d8
created: 2026-10-22T23:46:33.882832365Z
updated: 2026-10-24T05:57:36.062977032Z
---

## Context

- paths: `internal/parseoracle/parens_test.go` (new), `internal/parseoracle/testdata/parens/` (new), `internal/parseoracle/testdata/parse_facts.pl` (the perl-side B walk, extended to report OPf_PARENS), `internal/parse/` (the `Paren` flag; parser fixes allowed, see Steps), `internal/parseoracle/testdata/parens/gaps.txt` and `defects.txt` (new), `internal/parse/testdata/m3probes/opf-parens.txt` (new), `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/sites.txt`
- commands: `go test ./internal/parseoracle/ -run '^(TestParenComparisonReportsMisattachedParen|TestParenFlagMatchesPerlOPfPARENS)$' -count=1 -v`
- entrypoints: `internal/parseoracle/testdata/parse_facts.pl`, `internal/parseoracle/cvwalk_test.go`, `internal/parse/term.go` (`finishList`)

The canon change writes a `Paren` node's parens, so whether psc attached each paren to the right node now matters, and perl can judge it. perl keeps one boolean bit, OPf_PARENS (`P` in Concise), on the op a source paren enclosed: `my $z = ($x)` is `padsv sP` vs `padsv s`, and `ok(($y eq "x"))` is `seq sKPM` vs `seq sKM`. This check walks perl's optree for a fixture and collects (line, op, PARENS bit). It compares that op by op with psc's nodes and their `Paren` flag, and reports a paren psc put on the wrong node, a paren psc dropped, and a paren psc invented. Tests run perl; production psc never does. Some ops perl folds or optimises away will have no counterpart; the check names those as unmatched rather than guessing. perigrin ruled the check into this milestone on 2026-10-22. Fixing the queue's `dropped-paren` rows is a separate issue, which this one blocks (J4, chain-review round 5): this issue builds the check and fixes only the fixture disagreements it finds.

**Gaps and defects are separate lists (J3).** A fixture whose P bit has no psc counterpart is one of two things:
- a GAP: permanent, because perl folded or optimised the op away and a folded op carries no P bit. It is listed in `testdata/parens/gaps.txt` with the Concise output that shows the fold.
- a DEFECT: psc put `Paren` on the wrong node, dropped it or invented it. It is listed in `testdata/parens/defects.txt`, and that list must be EMPTY when this issue closes (TestParenKnownDefectsEmpty). A defect is fixed in the parser here, or escalated to perigrin. **perigrin's ruling (2026-10-23, chain-review round 6, F15): a paren on an op perl folds away is a known gap, and canon writes it.** Examples measured at 5.44, each with no op carrying P: `$h{(1)}`, `-(1)`, `"a" . ("b")` (const fold) and `($_[0])->[0]` (multideref). psc keeps the source's `Paren` on that node, and canon writes the paren, so the file is faithful. The check lists each such case in gaps.txt with the Concise evidence of the fold. It is never a defect, and never reported as an invented paren.

Measured at 5.44: `return (1,2)` and `return ((1,2))` both compile to `<@> return KP`, so the P sits on the return op that absorbed its parenthesised arguments, and there is no list op. `return (1)` is `const[IV 1] sP`. The comparison maps a P on a list op that absorbed its arguments to psc's List child; it never moves `Paren` onto the Return node. Otherwise canon would write `(return 1, 2)`.

**How "op by op" is keyed (K2, round 6).** perl attributes ops only to their statement's COP line (parse_facts.pl:137-160), so the comparison is per statement. For each statement line, perl's side is the multiset of op names that carry P. psc's side is the multiset of op names its `Paren` nodes map to, through a fixed node-kind→op-name table in parens_test.go (K8, chain-review round 7, measured at 5.44):
- a scalar variable maps to `padsv`, or `ex-rv2sv` for a package variable;
- `eq` maps to `seq`, and so on for each comparison;
- a constant maps to `const`;
- a parenthesised List of two or more elements maps to `ex-list` (`my ($a,$b) = @_` gives `ex-list lKPRM*`), or to the list op that absorbed it (`return`, `print`, an assignment's `aassign`);
- a ONE-element parenthesised List maps to its element's op: `my ($x) = @_` puts the P on `padsv[$x] sPRM*/LVINTRO` with the ex-list `lK` and no P, and `local ($x) = 1` on `ex-rv2sv sKPRM*`;
- `qw(a b)` maps to `ex-list` (`ex-list lKP`);
- the other node kinds the fixtures need are added to the table explicitly.

perl's side EXCLUDES the structural ops that carry P with no source paren: `leave vKP` on every block, `if`/`else` and `do`, and `leaveloop vKP` and `lineseq vKP` on every `for` and `while`, plus `scope`. Without that exclusion "dropped" fires on every fixture with a block.

The comparison then reports:
- **misattached:** same statement, a different op name on each side;
- **dropped:** perl has P and psc has no `Paren`;
- **invented:** psc has a `Paren` on a node perl gives no P, unless the case is in gaps.txt;
- **unmapped:** a node kind that is not in the table. That fails the fixture until the table gains a row, and it never counts as agreement.

## Steps

- [ ] Measure with perl first. Run `perl -MO=Concise` on each fixture and record which op carries `P`. Use the pinned perl 5.44.0. Include cases where the paren is on a list op (`my ($a,$b) = @_`, `ex-list lKPRM*`), on a one-element list (`my ($x) = @_`, P on `padsv`, not on the list), on a scalar op (`my $z = ($x)`, with `my $x` declared: for a package `$x` the `P` sits on a nulled `ex-rv2sv sKP`), doubled (`(($y eq "x"))`), on `qw(a b)` (`ex-list lKP`), around a call argument, and inside a block and a loop, to show the structural `leave`, `leaveloop` and `lineseq` P bits the walk must exclude. Quote the output in the fixtures.
- [ ] Create `internal/parseoracle/parens_test.go` with two `// ABOUTME:` lines, and fixtures in a readable text format under testdata/parens/.
- [ ] RED: write TestParenComparisonReportsMisattachedParen, a unit test with no perl. Given hand-built perl-side facts and a psc tree whose `Paren` sits on the wrong node, the comparison reports it, and it reports nothing for a matching pair. Run it and watch it fail.
- [ ] GREEN: implement the comparison. Run it and watch it pass.
- [ ] RED: write TestParenFlagMatchesPerlOPfPARENS. It runs the pinned perl (corpus.pin's interpreter, found through `perltest.PinnedPerl`) with B on each fixture, reads OPf_PARENS per op, and compares with psc's tree. Run it and watch it fail until the B walk reads the flag.
- [ ] GREEN: extend the walk to read op flags. Run it and watch it pass. A fixture where psc disagrees with perl goes in gaps.txt with the Concise output proving perl folded the op, or it is a defect: fix it here, with the per-fix review below, or escalate it to perigrin. defects.txt must be empty at close.
- [ ] RED/GREEN: add the `return (1,2)` fixture (`<@> return KP`, mapped to psc's List child) and TestParenKnownDefectsEmpty, which fails while defects.txt holds any row and requires every gaps.txt row to carry Concise output.
- [ ] RED/GREEN TestParenStructuralOpsAreExcluded and TestParenOneElementListMapsToElement: a fixture with a block and a `for` loop reports no dropped paren from `leave`, `leaveloop`, `lineseq` or `scope`; `my ($x) = @_` matches with P on `padsv`, and a table that mapped it to `ex-list` would report a misattachment.
- [ ] Add the fixtures, with perl's facts, as probe file `internal/parse/testdata/m3probes/opf-parens.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] parse_facts.pl's hash keys the oracle cache (cache.go:52), so extending it re-measures all 620 files.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Every test here that runs perl uses `perltest.PinnedPerl`. Add them to parse-fidelity.yml's 5.44-held step (`PVM_REQUIRE_PINNED_PERL=1`).
- [ ] If a parser fix here moves any of the 619 perl.git t/ files' re-emission, re-baseline perlgittfaithful.ratchet in the same commit; where the file is in the M3 dirs, also update faithful_queue.txt (its rows and `# causes` header) and run TestFaithfulQueueMatchesTree. Name each moved file (K6).
- [ ] RED/GREEN TestParenUnmappedNodeKindFails: a fixture whose `Paren` node kind has no table row fails with "unmapped".
- [ ] RED/GREEN TestParenFoldedOpsAreGaps: each F15 fixture is listed in gaps.txt with Concise evidence, is reported neither invented nor a defect, and re-emits faithfully. **A gap row is re-checked against perl, not against pasted text (F16, chain-review round 7):** the test runs the pinned perl on the row's snippet and requires that the op the table maps the node to is ABSENT from the statement's ops (`$h{(1)}` and `($_[0])->[0]`, both multideref), or that the op which replaced it carries `FOLD` (`-(1)` gives `const[IV -1] s/FOLD`, `"a" . ("b")` gives `const s/FOLD`). Otherwise the row fails: `print (1)+2` gives `const[IV 1] s`, present and unfolded, so listing it as a gap fails and it stays an invented paren.
- [ ] Add the F15 gap fixtures (`$h{(1)}`, `-(1)`, `"a" . ("b")`, `($_[0])->[0]`) to testdata/parens/ and gaps.txt, each with its Concise output showing no P bit, and check that canon writes the paren for each.
- [ ] Run `make test`, then commit.

**Note from the 75d4 review (2026-10-24).** The invented-paren check must test the SINGLE form, not only the doubled one.
- For `$f->(($x))`, psc's one Paren sits on `$x` and perl's P bit sits on `ex-rv2sv`, so the two agree.
- The defect shows in `$f->($x)`, where psc sets Paren and perl sets no P bit (`sKM`). That paren is invented.
- `not ($a == 1)` (`eq sK`) and the list slice `(@a)[0]` behave the same way.
- Add all three as fixtures.

75d4 carries the doubled forms as known-defect rows citing this issue: `$f->(($x))`, `not (($a == 1))` and `((1,2,3))[1]`.

## Acceptance criteria

### Positive Scenarios

- [ ] The comparison reports a paren attached to the wrong node, and nothing for a match (`go test ./internal/parseoracle/ -run '^TestParenComparisonReportsMisattachedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsMisattachedParen'`)
- [ ] psc's Paren flag matches perl's OPf_PARENS bit op by op on every fixture (`go test ./internal/parseoracle/ -run '^TestParenFlagMatchesPerlOPfPARENS$' -count=1 -v | grep -q '^--- PASS: TestParenFlagMatchesPerlOPfPARENS'`)
- [ ] The fixtures are committed as perl-judged probes, held by TestM3ProbesHeldToPerl (`sh -c 'test -s internal/parse/testdata/m3probes/opf-parens.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] Regression guard (passes before this issue's work; it holds the new tests once they exist): every test that calls PinnedPerl, this issue's included, is in the -run list of a CI step that sets the REQUIRE variable (`go test ./internal/parseoracle/ -run '^TestEveryPinnedPerlTestIsRequired$' -count=1 -v | grep -q '^--- PASS: TestEveryPinnedPerlTestIsRequired'`)
- [ ] No known paren defect remains, and every known gap carries perl's evidence that the op was folded (`go test ./internal/parseoracle/ -run '^TestParenKnownDefectsEmpty$' -count=1 -v | grep -q '^--- PASS: TestParenKnownDefectsEmpty'`)
- [ ] A paren on an op perl folds away is a named gap, re-checked against perl (the mapped op absent, or its replacement carrying FOLD), never a defect, and canon still writes it (`go test ./internal/parseoracle/ -run '^TestParenFoldedOpsAreGaps$' -count=1 -v | grep -q '^--- PASS: TestParenFoldedOpsAreGaps'`)
- [ ] The structural ops that carry P without a source paren (leave, leaveloop, lineseq, scope) are excluded on perl's side, so a block or loop never reports a dropped paren (`go test ./internal/parseoracle/ -run '^TestParenStructuralOpsAreExcluded$' -count=1 -v | grep -q '^--- PASS: TestParenStructuralOpsAreExcluded'`)
- [ ] A one-element parenthesised list maps to its element's op, so `my ($x) = @_` matches perl's P on padsv (`go test ./internal/parseoracle/ -run '^TestParenOneElementListMapsToElement$' -count=1 -v | grep -q '^--- PASS: TestParenOneElementListMapsToElement'`)
- [ ] Regression guard: the work queue still matches the tree after any parser fix here (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)

### Negative Scenarios

- [ ] The perl-side walk reads OPf_PARENS on nulled ops. perl 5.42 compiles `f((1,2))` (with `sub f {}` declared) to an `ex-list lKPM` that carries the only P bit, while `f(1,2)` has none. A walk that skips ex- ops would report nothing on both sides and pass (`go test ./internal/parseoracle/ -run '^TestParenWalkReadsPARENSOnNulledOps$' -count=1 -v | grep -q '^--- PASS: TestParenWalkReadsPARENSOnNulledOps'`)
- [ ] The comparison fails a fixture whose source holds a paren when both sides report zero P ops, rather than calling 0 = 0 a match (`go test ./internal/parseoracle/ -run '^TestParenComparisonRejectsEmptyFactsForParenthesisedSource$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonRejectsEmptyFactsForParenthesisedSource'`)
- [ ] A dropped paren is reported: perl has P on `padsv` in `my $z = ($x)` and psc has no `Paren` on the node (`go test ./internal/parseoracle/ -run '^TestParenComparisonReportsDroppedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsDroppedParen'`)
- [ ] An invented paren is reported: psc sets `Paren` on a node where perl has no P and the case is not a gaps.txt fold, for example `print (1)+2`, whose `const[IV 1]` is `s` and not `sP`; a gaps.txt fold such as `$h{(1)}` is not reported (`go test ./internal/parseoracle/ -run '^TestParenComparisonReportsInventedParen$' -count=1 -v | grep -q '^--- PASS: TestParenComparisonReportsInventedParen'`)
- [ ] A call's argument parens are not taken for a paren on the argument. perl measures `print ((1))+2` as `const[IV 1] sP` and `print (1)+2` as `const[IV 1] s`, and psc's `Paren` matches each (`go test ./internal/parseoracle/ -run '^TestParenFlagCallArgParensAreNotAParen$' -count=1 -v | grep -q '^--- PASS: TestParenFlagCallArgParensAreNotAParen'`)
- [ ] A gap row whose mapped op perl still emits, present and unfolded (`print (1)+2`, `const[IV 1] s`), fails the gap check rather than excusing an invented paren (`go test ./internal/parseoracle/ -run '^TestParenGapRowWithUnfoldedOpFails$' -count=1 -v | grep -q '^--- PASS: TestParenGapRowWithUnfoldedOpFails'`)
- [ ] A fixture whose every P op is unmatched (folded or optimised away) fails unless it is listed in gaps.txt with perl's Concise evidence of the fold; "unmatched" never counts as agreement (`go test ./internal/parseoracle/ -run '^TestParenUnmatchedOpsAreNotAgreement$' -count=1 -v | grep -q '^--- PASS: TestParenUnmatchedOpsAreNotAgreement'`)
- [ ] The fixture set cannot pass empty: it holds at least one fixture for each measured class (list op, one-element list, scalar op, doubled, call argument, nulled `ex-list`, `qw`, a list op that absorbed its parenthesised arguments as in `return (1,2)`, a block and a loop for the structural exclusion), and the test fails if any class has none (`go test ./internal/parseoracle/ -run '^TestParenFixturesCoverEveryClass$' -count=1 -v | grep -q '^--- PASS: TestParenFixturesCoverEveryClass'`)
- [ ] When perl fails to compile a fixture, or prints no ops, TestParenFlagMatchesPerlOPfPARENS fails with perl's stderr, rather than skipping the fixture (`go test ./internal/parseoracle/ -run '^TestParenFlagFailsWhenPerlProducesNoOps$' -count=1 -v | grep -q '^--- PASS: TestParenFlagFailsWhenPerlProducesNoOps'`)
- [ ] A node kind with no row in the node-kind→op table is reported as unmapped and fails its fixture, never counted as agreement (`go test ./internal/parseoracle/ -run '^TestParenUnmappedNodeKindFails$' -count=1 -v | grep -q '^--- PASS: TestParenUnmappedNodeKindFails'`)
