# chain-review round 1: refinement repair on chart-perly-w4

Scope: the blocking findings PB1 to PB7 (pushback) and AL1 to AL4 (alignment). No residual finding was answered in the chain. All edits went through the `git zhi` CLI (git-zhi 0.5.2). The repository, git config and branches were not touched.

## How each blocking finding was answered

| Finding | Answer | Issues |
|---|---|---|
| PB1 | **Pin moved, and digest format stated.** d1c2 now says that the no-precedence mode gives every item the context `none`. It also says that the set digest writes a `none`-context item exactly as `{State Dist}` was written before, so `set-digest.golden` does not move until the guards issue regenerates it. The copy of e3432c0f's `set-digest.golden` to `set-digest-noprecedence.golden` moved from d278 to d1c2's first test step, before `Item` changes. d278's criteria #1 and #5 moved with it. d1c2 gained a criterion that runs `TestSetDigestGolden` and `TestSetDigestNoPrecedenceGolden` together. The guard proposed in PB1, "set-digest.golden unchanged since e3432c0f", was not used: d278 regenerates that file, so the guard would go red at milestone verify. The byte-for-byte `cmp` of the no-precedence golden is used instead, because it holds through the milestone. d278 #3's precondition, `test -f set-digest-noprecedence.golden`, would now be green before d278's work, so it was rechained to `TestReferenceRecognizerPrecedence` existing. | 01a21447-d1c2, 01a21447-d278 |
| PB2 | Added the edge d1eb → d278. | 01a21447-d1eb, 01a21447-d278 |
| PB3 | d218 #4 and #6 now cover only the production differentials: toys, seeds, union shortcut, waypoint2 and corpus. `TestForestEqualityRefereeSentences` was removed from #6. d218's prose says the referee-sentence differential becomes real at the referee issue. d247 gained `TestForestEqualityRefereeCandidateNeverConstrains`. It counts entries with the hook, logging 0 on the candidate and at least 1 on the V1 side, and is not a file grep. | 01a21447-d218, 01a21447-d247 |
| PB4 | d2a2 #7 now runs a unit test, `TestCommittedBaselineMatchesKeys`. It parses the committed `baseline.txt` and calls `checkBaseline`, so it does not need perl.git. Added the matching TDD step and a prose line. | 01a21447-d2a2 |
| PB5 | **Stated: the precedence fixtures are Go test fixtures, not port `*-refused` snippets.** `refused-where.txt` lists recognition's refusals, `table` and `grammar` only. The un-narrowed rerun reproduces each one at V1's byte, so the file stays byte-identical through the milestone. d166's dead "scope or precedence rows" criterion became "byte-identical to e3432c0f's once the rerun test passes". The unsatisfiable step "regenerate refused-where.txt if a scope byte moves" was deleted. `refused-where.txt` was removed from the `paths:` of d166 and d199. 01a21463-3262 states the fixture decision and carries a matching byte-identity criterion. | 01a21447-d166, 01a21463-3262 |
| PB6 | **Split d199** into 01a21447-d199 and 01a21463-3262. 01a21447-d199 covers the reduce side, the vacuous case, refusal records and Build, Leo, and the no-precedence mode with `TestParseHasNoPrecedenceSwitch`. It also pins the guards to that mode: `reference_test`, `agree_test`, the brute force in `forest_test`, `setdigest_test`, `forestdigest_test`, `recognizeSource`, and the rerun recognizer. Every guard file is listed in its `paths:`. 01a21463-3262 covers precedence-refusal classification by rerun and both readers of `refused-where.txt`. The edges run d199 → 3262 → d1c2. **One deviation from PB6:** the two retired tests stay in 01a21447-d199, not 3262. `TestRecognitionKeysUnchangedByPipeline` calls `Fatalf` once the default recognizer refuses `$a <=> $b <=> $c;`, so 01a21447-d199 could not close green without retiring it. `TestPrecedenceIgnoredByRecognition` is retired beside it. Until 3262 lands, a death caused only by precedence is classified by d166's rule. No corpus file or port snippet has one (C12), and the prose says so. | 01a21447-d199, 01a21463-3262 |
| PB7 | **Split d1c2** into 01a21447-d1c2 and 01a21463-3282. 01a21447-d1c2 carries the context with no refusal: item layout, collapse table, interning and cache keys, the Leo memo, the measurement, and the PB1 pin. 01a21463-3282 adds the pairing rule, the shift-side refusal, `TestPredictorTakesNonNoneTop`, Build by context, and the judged-recognizer differentials. The edges run d1c2 → 3282 → {d1eb, d218}. | 01a21447-d1c2, 01a21463-3282 |
| AL1 | d166's rerun rule is restated. If the un-narrowed rerun dies anywhere, the reason is `grammar` at its byte, which is V1's. Only if it accepts is the reason `scope`, at the narrowed byte. The prose gives the scope-then-grammar example. Added a negative criterion, `TestRefusalRerunLaterGrammarFaultIsGrammar`, and extended the rerun-test step to name the two-fault file. | 01a21447-d166 |
| AL2 | **Chose option 1.** The referee-sentence comparison removes zero-width slipper `;` leaves, and their places in alternatives' children, from both sides, and removes nothing else. The rule is fixed in d0e1, before any recognizer code. d0e1 gained `TestForestEqualityRefereeIgnoresOnlyZeroWidthSlipperLeaves`, with three subtests: zero-width `;` is equal, a one-byte `;` differs, and a zero-width leaf of another terminal differs. d0e1 now carries "F8-P1 (its effect on the referee-sentence differential)". d247's prose says why its differential criterion can now pass. This is a reading of §5's "node for node" that F8-P1's measurement supports: the leaves move no verdict or grouping. If perigrin will not accept it, AL2 says to take it to him before d0e1 is built. | 01a21447-d0e1, 01a21447-d247 |
| AL3 | Added the edges d02f → d05d → d089 → d0ba. All counters now complete before any differential starts, and the existing d0e1/d10d → d135 edges keep stage 3 after stage 2. The serial counters also answer the shared `feasibility_test.go`, because d02f creates it. They also answer the d089/d10d clash over the Makefile line and `precedence_perlgit_test.go`. The milestone's "touch disjoint files and may be written in parallel" sentence was replaced with the real ordering and its reasons. | 01a21447-d02f, -d05d, -d089, -d0ba; milestone |
| AL4 | 01a21447-d199's `TestRecognitionKeysUnchangedByPipeline` criterion is now gated on `! grep -rq "func TestPrecedenceIgnoredByRecognition" internal/chart`, which is red today. bb9b's `guard:` criterion on `TestPrecedenceShortcutStated` is now gated on the README citing `docs/assessments/chart-perly-w4/verdict`, also red today, and has lost its `guard:` prefix. | 01a21447-d199, 01a2144c-bb9b |

## Change log

One line per change: issue, what changed, and the finding it answers.

- 01a21447-d199: split by `--split`. Retitled to "Refuse forbidden completions in the recognizer (reduce side), with refusal records Build reads and the guards pinned to a no-precedence mode". The body was rewritten for the first half, with 15 criteria. **PB6**
- 01a21447-d199: `paths:` gains the five guard test files, `port_lex_core_test.go` and the rerun recognizer, and drops `refused-where.txt`. The prose says which guards are pinned to the no-precedence mode. **PB6, PB5**
- 01a21447-d199: the retired-test step and criteria stay here. Carries are now A16, A29, F3-6, C13 (reduce half), F6-R2, F3-7 (switch half) and C10 (a). **PB6**
- 01a21447-d199: the `TestRecognitionKeysUnchangedByPipeline` criterion is gated on `TestPrecedenceIgnoredByRecognition` being gone. **AL4**
- 01a21463-3262 (new, `model:opus`): "Classify precedence refusals by rerun at the byte where the chart died, one rule for both readers of refused-where.txt". It holds three criteria moved from d199: `TestPrecedenceRefusalByteWhereChartDied`, `TestRefusedWhereReadersAgree` and `TestRefusalRerunOrder`. It carries F5-3 and C12 (precedence-refusal half). **PB6**
- 01a21463-3262: states that the precedence fixtures are Go test fixtures, not port snippets. Adds the criterion "refused-where.txt byte-identical to e3432c0f's". **PB5**
- 01a21447-d1c2: split by `--split`. Retitled to "Carry the canonical spine top as item context, with the collapse table, interning, cache keys and the Leo memo, refusing nothing yet". The body was rewritten for the first half, with 8 criteria. Carries are now C11, F3-9, P12, A32, C10 (d), C13 (shift and memo half) and F3-4. **PB7**
- 01a21447-d1c2: a set-digest paragraph says no-precedence items are `none` and the digest writes them as before. A first step copies the golden. The criterion `TestSetDigestGolden`+`TestSetDigestNoPrecedenceGolden` and the byte-for-byte `cmp` were added, the `cmp` moved from d278. `paths:` gains `setdigest_test.go` and `set-digest-noprecedence.golden` (new). F3-4 moved here. **PB1**
- 01a21463-3282 (new, `model:opus`): "Refuse forbidden scans (shift side) and pair completions under their context, with Build walking by context". It holds eight criteria moved from d1c2 and carries C6, F6-R4 and F7-A2. **PB7**
- 01a21447-d278: the "keep the current set-digest golden" step, the `set-digest-noprecedence.golden` path, criteria #1 and #5, and the F3-4 carry were all removed, since they moved to d1c2. The prose now says the pin already exists. **PB1**
- 01a21447-d278: #3's precondition `test -f set-digest-noprecedence.golden` became `grep -rq "func TestReferenceRecognizerPrecedence" internal/chart/recognize`, which keeps #3 red before d278's work. **PB1**
- 01a21447-d1eb → 01a21447-d278: edge added. **PB2**
- 01a21447-d218: the prose limits "for real" to the production differentials. #4 is reworded to cover the production differentials. #6 drops `TestForestEqualityRefereeSentences` from both its `-run` pattern and its grep. **PB3**
- 01a21447-d247: added a step and a criterion for `TestForestEqualityRefereeCandidateNeverConstrains`, with hook counts of 0 on the candidate and at least 1 on V1. The prose was updated to match. **PB3**
- 01a21447-d247: the prose says the differential's comparison drops zero-width slipper leaves on both sides. **AL2**
- 01a21447-d2a2: #7 was replaced by `TestCommittedBaselineMatchesKeys`, which needs no perl.git. Added a TDD step and a prose sentence. **PB4**
- 01a21447-d166: the rerun rule was restated: the rerun dying anywhere means `grammar`. Added the two-fault example and the criterion `TestRefusalRerunLaterGrammarFaultIsGrammar`, and the rerun step now names the two-fault file. **AL1**
- 01a21447-d166: criterion #10 is now byte-identity of `refused-where.txt` with e3432c0f's. The "regenerate refused-where.txt if a scope byte moves" step was removed, `refused-where.txt` left `paths:`, and the prose says the file does not change. **PB5**
- 01a21447-d0e1: the prose fixes the zero-width slipper `;` leaf rule. Added a step and the criterion `TestForestEqualityRefereeIgnoresOnlyZeroWidthSlipperLeaves` with three subtests. Carries gains "F8-P1 (its effect on the referee-sentence differential)". **AL2**
- 01a21447-d02f → 01a21447-d05d, 01a21447-d05d → 01a21447-d089, 01a21447-d089 → 01a21447-d0ba: edges added. **AL3**
- 01a2144c-bb9b: the `guard:` `TestPrecedenceShortcutStated` criterion is now gated on the verdict citation, and the `guard:` prefix was dropped. **AL4**
- Milestone body, Design Rationale: the "disjoint files … in parallel" sentence was replaced with the sequential, hard-edged stage order and its reasons. The critical-chain length was added. **AL3**
- Milestone body, Records: added "Chain-review round 1 repair" with the splits and the 19-issue count. **PB6, PB7, PB1, AL3**
- Milestone body, residual mapping: the d199 and d1c2 titles were updated wherever they appear. These rows were rewritten:
  - C6, F6-R4 and F7-A2 now point to 01a21463-3282.
  - F5-3 now points to 01a21463-3262.
  - C12 now points to d166 and 01a21463-3262.
  - F3-4 now points to 01a21447-d1c2.
  - F8-P1 now points to d247 and d0e1.

  **PB6, PB7, PB1, AL2**

Nothing else in any body changed. Negative Scenarios sections, README-first steps and close steps were kept, and every body was diffed against its original before writing. Each stored body was re-read and compared with its source file, and all 11 match. The milestone body also matches, and its resolution command is unchanged.

## Verification

- All 13 new or changed criterion commands were run against the tree at e3432c0f with `PERL5_CORPUS=/nonexistent`. All 13 are red. They fail because the work is missing: the harness itself works, so `TestSetDigestGolden` passes and the new test is absent. None of them needs perl.git.
- `internal/chart/offline/criteria-lint.sh` was run over all 19 issues and exits 0. It finds no line without a command, no non-guard command already green, and no pin on a file a later issue edits.

## Issue list (19)

| Issue | Model | Title | Blocks |
|---|---|---|---|
| 01a21447-d02f | opus | Pin feasibility counters 1-3 | d05d, d135 |
| 01a21447-d05d | opus | Pin feasibility counters 4-6 and the closing terminals' levels | d089, d135 |
| 01a21447-d089 | opus | List the post-narrowing mixed terminal sets (set 1626) | d0ba, d135 |
| 01a21447-d0ba | opus | Build the forest-equality differential (toys, seeds 4 and 44) | d0e1, d10d |
| 01a21447-d0e1 | sonnet | Differential over waypoint2 rows and referee sentences | d135 |
| 01a21447-d10d | sonnet | Differential over the corpus and perl.git t/ | d135 |
| 01a21447-d135 | opus | Move the precedence tables into package prectab | d166 |
| 01a21447-d166 | opus | Narrow library operators before recognition; grammar/scope by rerun | d199 |
| 01a21447-d199 | opus | Reduce side, refusal records, Build; guards pinned to a no-precedence mode | 01a21463-3262 |
| 01a21463-3262 | opus | Classify precedence refusals by rerun; one rule for both readers | d1c2 |
| 01a21447-d1c2 | opus | Item context, collapse table, interning, cache keys, Leo memo; refusing nothing | 01a21463-3282 |
| 01a21463-3282 | opus | Shift side, pairing under context, Build by context | d1eb, d218 |
| 01a21447-d1eb | opus | Expected reports only allowed continuations | d2d1, d278 |
| 01a21447-d218 | opus | Stop calling Constrain from ParseWith | d247, d278, d2a2 |
| 01a21447-d247 | opus | Referee narrows the lattice and judges the recognizer's forest | d2d1 |
| 01a21447-d278 | opus | Every recognizer guard with precedence on | d2d1 |
| 01a21447-d2a2 | sonnet | Retire speed-stage-precedence | d2d1 |
| 01a21447-d2d1 | opus | V2 verdict reading | bb9b |
| 01a2144c-bb9b | opus | README shared sections rewrite | — |

The edges d02f/d05d/d089 → d135 are now implied by the stage chain. They were left in place because they are redundant, not wrong.

## Critical chain (15 issues, from `git zhi list --critical`)

01a21447-d02f → 01a21447-d05d → 01a21447-d089 → 01a21447-d0ba → 01a21447-d0e1 → 01a21447-d135 → 01a21447-d166 → 01a21447-d199 → 01a21463-3262 → 01a21447-d1c2 → 01a21463-3282 → 01a21447-d218 → 01a21447-d247 → 01a21447-d2d1 → 01a2144c-bb9b

A path of equal length runs 01a21463-3282 → 01a21447-d1eb → 01a21447-d278 → 01a21447-d2d1. The chain grew from 10 to 15 issues: 2 links come from the splits and 3 from serialising the counters. PB12 was accepted as residual.

## Judgements a reviewer should check in round 2

- **AL2:** whether option 1 is acceptable under design §5. Option 1 drops zero-width slipper `;` leaves on both sides of the referee-sentence comparison, and nothing else.
- **PB6:** the two retired tests are in the first half of the split, against PB6's suggestion.
- **PB1:** the digest-format statement is what keeps `set-digest.golden` stable between d1c2 and d278. No criterion checks it except `TestSetDigestGolden` passing at each close.
