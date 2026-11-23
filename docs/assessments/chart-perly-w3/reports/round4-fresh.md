# Round 4, fresh participant: chart-stateless-precedence

Subject: `docs/plans/2026-11-23-chart-stateless-precedence.md` at 6dc1d9558, blob 123012c1a5765883ded36de157c8d0e8ac764f0a.
I read only "Experiments after round 3" from the assessment. All experiments below ran in a scratch worktree on `pu` 8c3317079 with `combined.diff` applied. I used the prototype's own `PrepareCombined` harness, which checks V1 and V2 against V0 on the ParseWith reason, the readings, recognition acceptance, the Constrain verdict and the exact kept set. The test files I added are named zz_fresh4*.

## What held

None of these inputs produced a kept-set, verdict or readings difference, in V1 or V2:

- **11,574 files.** Every .t, .pl and .pm under /opt/corpora (perl5 and PerlOnJava), plus every Perl file tracked in the repo, `std5k.pl` included. 11,110 of them reached the constraint.
- **20,083 random and handcrafted operator sentences.** V0 refuses 185 of them by precedence.
- **15,000 Leo-chain sentences.** These are right-recursive chains 1 to 25 deep, wrapped in nullable contexts such as `for (;;)`, `() =`, `:lvalue`, `do {} until` and trailing commas. V0 refuses 989 of them by precedence.

The state-free filter (V1) never disagreed with `Constrain`.

## Findings

**H1. Section 3, "Identical output" (refusal reason), against section 3.3. A crafted input breaks it.**
The reduce-side filter in recognition turns precedence refusals into grammar refusals:

- `print $fh [];` gives V0 `{Byte:0 Reason:precedence}` and V2 `{Byte:11 Reason:grammar}`.
- `sort $f [];`, `{ print $fh [] }`, `print $fh [] if 1;` and `my @x = (print $fh []);` behave the same way.
- 57 of the 15,000 chain sentences differ, and each one ends in `$fh []`.

The cited evidence already showed this: `combined/toys.txt` reports "recognition acceptance differs 30", for indirob, indirobsort and mixed on "LSTOP VAR LB RB". Section 2 reports only "unchanged on every aggregate file". The aggregate contains no precedence refusals, because recursive descent accepts all of its files; my 11,574-file run found zero V0 precedence refusals.

Consequence if it ships: the differential test as specified compares "verdict and exact kept set", not the reason, so the stop never fires and the change ships silently. It breaks:

- the decision's own promise;
- the README's refusal classes;
- `TestPrecedenceIgnoredByRecognition`, which asserts that recognition accepts `print $fh [];`, once Reduce is on by default.

**H2. Section 3, "Expected is computed with lookahead off, so it is unchanged". This is false.**

- `print $fh [ ) ;`, `sort $f [ ) ;` and `{ print $fh [ } ;`: V0, V2 and no-lookahead all die at the same node, but V2's `Expected` lacks `PERLY_BRACKET_CLOSE`.
- `print $fh [ ] ) ;` and `print $fh [] ]`: V2 dies one token earlier (node 4 against 5), so the refusal byte moves.

Why: `Expected` closes only the dead set again. `expected.go` justifies that with "a pruned completion is never needed by a later set". That holds for follow-set lookahead. It fails for a precedence refusal of a terminal that is in the follow set: items refused at an earlier node are absent from every later set.

Consequence: error text and refusal bytes change, and the refused-where and port fixtures may move.

**H3. Governing design.** Section 3.3 reverses v2 design §8: "precedence leaves recognition … recognition uses perly.y's rules as written, is Boolean" (perigrin, 2026-10-30). It also reverses the README line "Recognition ignores both (spec §8)". The decision names three live documents to change. It names neither §8's amendment nor that test. perigrin's "V2: both" ruling may authorise the reversal, but the record should say that it amends §8.

**H4. Section 3.1, "fails … refused when the tables are built, never mis-parsed". This overclaims.**

- The prototype's guard panics in `NewStateless` in two cases: a reduce key still mixed by its parent item, and a shift key mixed by its spine-top item. So it checks only that keys agree across bison's LR(0) states.
- It does not check that the evaluator supplies the key that model implies. Four pieces of evaluator logic decide that: `canon` dropping context, m==1 passing context through unchanged, the empty-right-child shift exemption, and `rootTop`. Only the oracle covers them, so the oracle is the real guard.
- No test or toy trips the generation failure. Add one.
- `ReduceSideOf`, which builds recognition's tables, does not fail on mixed keys. It stays conservative. The decision should say so.
- The slipper is safe only because `;`, `}` and EOF are in no Never set. I checked this: 0, 0 and 0 rules. Nothing guards it: give one of them precedence and the inserted-`;` paths become filterable.

**H5. Section 3.3, "A Leo memo never completes a link the filter refuses". This path is untested.**
`RefusedLinks` was 0 on the aggregate (`combined/result-compare.txt`) and 0 on my 15,000 chains. The Build-side branch is either dead or unreached. Recognition skips Leo whenever any link is refused or has a parent table. Either add a forcing test, as §3.2 does for multi-entry memos, or drop the branch.

**H6. Section 5, exit and stops.**

- Nothing separates what §3.3 adds. V1 to V2 moves the pipeline from 0.73 to 0.69, the recognize stage "straddles noise", and every output change I found comes from §3.3.
- perigrin needs the final reading to report V1 alone beside V2, or §3.3 as a separate commit with its own reading, so he can weigh about 4% against a contract change.
- The stop "any output differs from the oracle" is only as wide as the differential test. Widen the test to cover precedence-refused inputs: `waypoint2Refusals`, the referee's refused sentences and the port refused fixtures. Compare reason, byte and `Expected`, not only verdict and kept set.

## Not examined

- Timing.
- `make test` with the diff applied. Reduce is off in `Parse` in the prototype, so committed tests do not see V2.
- edit5k.pl, which is not in the repo.
- Branch or interleaved lattices. Build refuses them, and interleaved lexing is not built.
- `stateless.diff` on its own.
- Coverage of recognition's Leo-decline path and of the transition cache under Reduce, beyond what the corpora exercise.
- Earlier rounds.

My corpus runs passed `Scope{}`, so 2,839 files were refused by scope after the constraint. They were still compared at the constraint.

## Recommendation: modify

1. perigrin rules on one of two paths:
   - "Identical output" and the recognition contract explicitly allow precedence refusals to become grammar refusals, with a byte and a changed `Expected`.
   - §3.3 does not ship. V1 alone gives most of the gain, at 0.73.
2. Record the amendment of design §8.
3. Widen the differential test as H6 describes.
4. State H4's limits, and add a toy that trips the generation failure and a test that forces a Leo refusal.
5. Report V1 and V2 separately in the final reading.
