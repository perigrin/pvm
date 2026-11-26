# chain-review round 2: refinement repair on chart-perly-w4

**Scope.** This round answers the blocking findings that are still open:

- PB6 (held, narrowed), N1 and PB9 (promoted), from `r2-pushback.md`;
- AL12, AL13 and PB9 (promoted), from `r2-alignment.md`;
- FR2-1 to FR2-4, from `r2-fresh2.md`;
- the coordinator's fixture-existence check, which FR2-7 prompted.

No residual finding is answered in the chain: N2 to N4, AL14, FR2-5's per-close key diff beyond PB9, FR2-6, FR2-8, and the residual half of FR2-7. All edits went through the `git zhi` CLI. The repository, git config and branches were not touched.

## The common root, and the re-plan

The chain assumed that once the reduce side landed (d199, 3262), the default recognizer would refuse everything V1's precedence refuses. It does not:

- `print $fh [];` dies on the reduce side.
- `$a <=> $b <=> $c;` and `1 .. 2 .. 3;` keep a reading until the shift side lands (3282).

Pinning guards one by one at d199 would have left intermediate states that are neither V1 nor V2. In those states, production reports a precedence-only death as `scope` (N1). Tests would have to be tuned toward wrong answers and back again (PB6, AL13, FR2-1 to FR2-3).

**Fix: keep the default recognizer unjudged until both sides exist, then switch it in one issue.**

- **One switch-over issue.** `recognize.New` stays exactly the no-precedence recognizer through d199 (reduce side), d1c2 (context) and 3282 (shift side). The two modes are chosen by constructor, not by an `Options` field. Every test in those three issues builds the judged recognizer explicitly, so no existing test, golden or production path moves in them. 01a21463-3262 is now the single switch-over, retitled "Switch the default recognizer to judge precedence: classify precedence refusals by rerun, and pin, rewrite or retire every test that moves". The switch-over runs after 3282: d199 → d1c2 → 3282 → 3262 → {d1eb, d218}.
- **The whole refusal rule lands with the switch.** At the switch-over the default judges, and the rerun rule arrives complete: un-narrowed death gives `grammar`, narrowed death gives `scope`, and otherwise `precedence` at the judged chart's death byte. No state ever reports `scope` for a precedence-only death.
- **Every call site has a stated fate.** 3262's Context gives each recognizer call site one of five fates: pinned, follows production, rewritten, retired, or not expected to move. The list was read from the 25 test call sites of `recognize.New` and the production ones. It has not been run, so 3262's first step is an inventory: flip `New` without committing, run `make test`, and check every failure against the list. A failure not on the list is reported to perigrin before any test is edited.

## How each finding was answered

| Finding | Answer | Issues |
|---|---|---|
| PB6 (narrowed) | The guard list is now a full inventory in the switch-over issue, which owns every test that moves.<br>**Pinned to no-precedence:**<br>• `toyForest`, which covers the brute force, every toy `Constrain` test and `TestPrecedenceForbiddenInnerSpanRefusesFile`<br>• `precedence_statefree_test.go`, `precedence_pinned_test.go`, `forest/precedence_reference_test.go` and `forestdigest_test.go`<br>• `reference_test.go`, `agree_test.go` and `setdigest_test.go`<br>• the ratchet's `precedence_reference_test.go` and `precedence_perlgit_test.go`<br>• `TestWaypoint2RefusalsMatchReference`, the rerun recognizer, and `referee.judge` (until d247)<br>**Following production:** `Parse`, `facts.go`, the benches, `chartInputs` and the stage harness.<br>**Rewritten:** `TestParseAppliesPrecedence`, `TestRatchetVerdictIsPipeline` and `TestPipelineStageKeysRefusedFiles`.<br>**Retired:** `TestPrecedenceIgnoredByRecognition` and `TestRecognitionKeysUnchangedByPipeline`.<br>**Not expected to move:** the recognizer-internal toy tests, plus startforms, memory, copy and resolve. An inventory step checks the list against `make test`. | 01a21463-3262, 01a21447-d199 |
| N1 | The interim `scope` state no longer exists. The precedence branch of the rerun rule lands in the same issue that makes production judge. d199 no longer touches production, `parseStream`'s rerun or `recognizeSource`. | 01a21463-3262, 01a21447-d199 |
| AL13 | `TestRatchetVerdictIsPipeline` and `TestParseAppliesPrecedence` are named in 3262 with their new expectations: `prec.pl` is not recognized, and the refusal is `precedence` at the chart's death byte rather than Byte 0. A 3262 criterion runs both, plus `TestPipelineStageKeysRefusedFiles` and `TestWaypoint2RefusalsMatchReference`, beside the new `TestDefaultRecognizerJudgesPrecedence`. | 01a21463-3262 |
| FR2-1 | `TestWaypoint2RefusalsMatchReference` is pinned to the no-precedence constructor at the switch-over, and a 3262 criterion runs it. Before the switch-over it is untouched and passes, because `New` is unjudged. | 01a21463-3262 |
| FR2-2 | 3262 now comes after 3282, so all three fixtures die in the chart there. 3262 says which side kills each one: `print $fh [];` dies on the reduce side (indirob before `[`); `$a <=> $b <=> $c;` and `1 .. 2 .. 3;` die on the shift side. `TestParseAppliesPrecedence` is rewritten in 3262. | 01a21463-3262 (edges) |
| FR2-3 | d199's false sentence is gone. `TestRecognitionKeysUnchangedByPipeline` is retired in 3262 with its README claim, and the rewrite option is dropped, with the reason why. The old AL4-gated criterion is replaced by `! grep func … && ! grep README`, which is red today. | 01a21447-d199, 01a21463-3262 |
| FR2-4 | **The V1 side is the V1 path from the moment it is built.**<br>• **d0ba:** the V1 side takes its recognizer from one function in `equality.go` that only ever builds the unjudged recognizer.<br>• **New d0ba criterion:** `TestForestEqualityV1SideRecognizesPrecedenceRefusals` has subtests `indirob`, `nceqop-chain` and `dotdot-chain`. The V1 side's chart must accept `print $fh [];`, `$a <=> $b <=> $c;` and `1 .. 2 .. 3;`, and its `Constrain` must refuse each. Measured at e3432c0f in a scratch copy: all three are recognized and refused by `Parse` as `precedence` at Byte 0. If the V1 side ever ran the judged recognizer, the chart would die and the test would fail. The test is written at d0ba and re-run at milestone verify.<br>• **d199:** points that function at the explicit no-precedence constructor before the switch-over. Its criterion `TestJudgedRecognizerIsExplicit` checks that the V1 side reports the no-precedence constructor, that `New` equals the no-precedence constructor on the perly.y sample, and that the judged constructor's sets differ there.<br>• **d0e1 (referee V1 path) and d10d (corpus V1):** both now say they take the recognizer from that same function. | 01a21447-d0ba, 01a21447-d199, 01a21447-d0e1, 01a21447-d10d |
| PB9 | d166 now states that `terminals-per-token` is an answer key, not a chart-shape key. `chartInputs` sums it over the un-narrowed `lex.Lex` tokens' committed sets, and a narrowed set id never indexes `tab.TerminalSets`. A new d166 criterion, `TestTerminalsPerTokenCountsUnnarrowedTokens`, has two subtests:<br>• `dead-plugin-reading`: the un-narrowed sum must be larger than the logged narrowed sum.<br>• `corpus`: the count must equal an independent un-narrowed sum over T1.<br>It needs no perl.git. Its step writes the test as a guard before the caller changes. d166 and 3262 each gain a step: where perl.git is present, run the resolution command and see the 27 keys unchanged. These two are the issues where a key can move. | 01a21447-d166, 01a21463-3262 |
| AL12 | d0e1's referee count is now `compared 1296 of 1296`. Measured today: `TestPrecedenceReferee` logs "1296 kept sentences agree with bison". The criterion text says where 1,296 comes from. | 01a21447-d0e1 |
| Fixtures (coordinator, FR2-7) | Every test, helper and file the new text names was checked with `grep` or `ls`. The only missing one was `scancollide`:<br>• **d05d** now creates `scancollideToy` in `precedence_pinned_test.go`. It is built with `pinnedToy`, the way `nullableShiftToy` is, from minute row F6-C2's grammar, whose rules and levels are spelled out. The `scancollide` subtests in d135 and d218 use it.<br>• **d135:** the stage-1 tests stay in package `forest`, because `loadToy`, `pinnedToy` and `scancollideToy` are forest test helpers that `prectab` cannot import. They call `prectab.Feasible`. d135's "prectab runs ≥8 TestFeasibility" criterion became "forest runs ≥8, and `feasibility_test.go` calls `prectab.Feasible`", which is red today.<br>• **Not created by this repair:** `equality.go`, `forest_equality_test.go` and the union-shortcut fixture are made by their own issues, which say "(new)" or "add". | 01a21447-d05d, 01a21447-d135 |

## Change log

| Issue | Change | Finding |
|---|---|---|
| 01a21447-d199 | Retitled with a single-document `--split`. Edges and labels were kept. New title: "Refuse forbidden completions (reduce side) in a judged recognizer built beside the default, with refusal records Build reads". | root, N1, PB6 |
| 01a21447-d199 | Body rewritten:<br>• The prose says `New` stays unjudged, gives the constructor-based mode choice, and says the V1-side function is repointed.<br>• Removed: the guard-pin list, the interim-`scope` sentence and the false "fixture is no longer recognized" sentence.<br>• Steps: the retire step was removed and a constructor step added.<br>• Criteria: the retire, `TestRecognitionKeysUnchangedByPipeline` and `TestParseHasNoPrecedenceSwitch` criteria moved to 3262, and `TestJudgedRecognizerIsExplicit` was added.<br>• Carries: F3-7 (switch half) and C10 (a) moved to 3262.<br>• `paths:` dropped the guard and test files and gained `equality.go`. | root, N1, PB6, FR2-3, FR2-4 |
| 01a21463-3262 | Retitled with a single-document `--split`. New title: "Switch the default recognizer to judge precedence: classify precedence refusals by rerun, and pin, rewrite or retire every test that moves". | root |
| 01a21463-3262 | Body rewritten:<br>• Context gives the switch-over, which side kills each fixture, the full rerun rule, and the call-site inventory with its fates.<br>• Steps: inventory first, then pin, then switch, rerun, readers, the rewrite/retire step, and a step to check the answer keys where perl.git is present.<br>• Criteria added: `TestDefaultRecognizerJudgesPrecedence`; a joint run of the rewritten and pinned tests; and `TestRecognitionKeysUnchangedByPipeline` retired with its README claim.<br>• Criteria moved in from d199: `TestPrecedenceIgnoredByRecognition` retired, and `TestParseHasNoPrecedenceSwitch`.<br>• Criteria kept: rerun-byte, readers, rerun-order and `refused-where.txt` identity.<br>• Carries gain F3-7 (switch half) and C10 (a). | root, N1, PB6, AL13, FR2-1, FR2-2, FR2-3, PB9 |
| Edges | Removed d199 → 3262, 3262 → d1c2, 3282 → d1eb and 3282 → d218. Added d199 → d1c2, 3282 → 3262, 3262 → d1eb and 3262 → d218. | FR2-2, root |
| 01a21447-d0ba | Prose: the V1 side uses one unjudged-recognizer function. Added a step and the criterion `TestForestEqualityV1SideRecognizesPrecedenceRefusals` with 3 subtests. | FR2-4 |
| 01a21447-d0e1 | The referee V1 path uses the V1 side's recognizer function. The count `1299` became `1296`, in both the description and the grep. | FR2-4, AL12 |
| 01a21447-d10d | The corpus V1 is the exported V1 side, using its unjudged-recognizer function. | FR2-4 |
| 01a21447-d166 | Added the `terminals-per-token` paragraph, `verdicts_test.go` in `paths:`, guard-test steps, a step to check the answer keys where perl.git is present, and the criterion `TestTerminalsPerTokenCountsUnnarrowedTokens`. | PB9 |
| 01a21447-d1c2 | One sentence: `New` stays no-precedence until the switch-over, and the tests use the judged constructor. | root |
| 01a21463-3282 | "The guards keep the no-precedence mode…" was replaced with the explicit-constructor and unjudged-default statement. | root |
| 01a21447-d247 | One sentence: `referee.judge`, pinned to no-precedence at the switch-over, moves onto the judged `New` here. | PB6 |
| 01a21447-d05d | Creates `scancollideToy` from F6-C2's grammar. Changes to `paths:` and to the counter-6 step. | Fixtures, FR2-7 |
| 01a21447-d135 | The feasibility tests stay in forest and call `prectab.Feasible`. Changes to the `paths:` note and the prose, and the ≥8-tests criterion was re-pointed to package forest. | Fixtures, FR2-7 |
| Milestone body | • Mapping titles for d199 and 3262 were updated wherever they appear.<br>• The C10 row now gives (a) to 3262, and the F3-7 row gives the switch half to 3262.<br>• Records gained "Chain-review round 2 repair".<br>• Design Rationale stage 3 says the change is built beside the default and switched in one issue. | root, mapping |

**Body preservation.** Every other line of each edited body is byte-identical to the live body read at the start of this round, including Negative Scenarios, the README-first steps and the close steps. All 11 stored bodies and the milestone body were re-read and match their source files. Labels survived both retitles, and the milestone resolution is unchanged.

## Verification

- **New and changed criteria are red.** All 8 new or changed criterion commands were run at e3432c0f with `PERL5_CORPUS=/nonexistent`, and all are red. The criteria moved verbatim from d199 to 3262 were already red.
- **Named tests exist.** The existing tests 3262 names (`TestParseAppliesPrecedence`, `TestRatchetVerdictIsPipeline`, `TestPipelineStageKeysRefusedFiles` and `TestWaypoint2RefusalsMatchReference`) exist and pass today.
- **The V1 premise holds.** In a scratch copy of the tree (`repair/scratch/`, not the repository), `print $fh [];`, `$a <=> $b <=> $c;`, `1 .. 2 .. 3;` and `$a == $b <=> $c;` are each recognized by `recognize.New` and refused by `Parse` as `precedence` at Byte 0.
- **Lint is clean.** `internal/chart/offline/criteria-lint.sh` over all 19 issues exits 0: no missing command, no non-guard criterion green, and no pin on a file a later issue edits.

## Issue order and critical chain (15, from `git zhi list --critical`)

01a21447-d02f → 01a21447-d05d → 01a21447-d089 → 01a21447-d0ba → 01a21447-d0e1 → 01a21447-d135 → 01a21447-d166 → 01a21447-d199 → 01a21447-d1c2 → 01a21463-3282 → **01a21463-3262 (switch-over)** → 01a21447-d218 → 01a21447-d247 → 01a21447-d2d1 → 01a2144c-bb9b

There are still 19 issues. d10d runs parallel to d0e1. After 3262, d1eb, then d278, and d247 and d2a2 run in parallel before d2d1.

## For round 3 to check

- **The call-site inventory in 3262 comes from reading, not from running.** The inventory step is the guard: an unlisted failure is reported, not fixed.
- **Recognizer internals in the explicit-constructor design.** The judged and no-precedence constructors share one recognizer, so d1c2's interning and cache keys are exercised by the judged constructor's tests, while `New` stays unjudged until 3262.
- **PB17 (residual)** is partly settled by the constructor choice. It is not closed: `ParseWith`'s `*Recognizer` argument can still carry a no-precedence recognizer, and `TestParseHasNoPrecedenceSwitch` reflects only over parameters and option fields.
- **`TestJudgedRecognizerIsExplicit` (d199)** relies on the judged reduce side refusing at least one completion on the perly.y sample. d278 assumes the same of the reference, but this repair has not measured it.
