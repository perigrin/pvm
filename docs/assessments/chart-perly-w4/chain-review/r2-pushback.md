# chain-review round 2: pushback (plan quality) on chart-perly-w4

## 1. Recommendation

**Modify.** The repair is good work. PB1–PB5 and PB7 landed, and I checked each against the code, not the log. The split carries every id it had before, and the edges are acyclic.

Three blocking findings remain.
- **PB6 (held, narrowed).** The guard list in d199 is still incomplete. Four more existing tests fail the moment the default recognizer judges, and neither d199 nor 3262 names them.
- **N1 (new).** The d199/3262 split creates an intermediate state where production reports a precedence-only death as `scope`. Closing d199 green would mean tuning tests toward that wrong reason.
- **PB9 (promoted).** The answer keys are still unchecked until the verdict reading. The chain is now 15 issues long, so a key moved by d166 surfaces 7 issues later.

How this round was checked:
- I read the live chain first: 19 issues and 166 criterion lines.
- I ran `internal/chart/offline/criteria-lint.sh` over all 19 issues. It exits 0: every criterion has a command, every non-`guard:` criterion is red at `e3432c0f`, and it finds no pins.
- I diffed every changed body against my round-1 copies.
- I checked the carries and the milestone mapping rows for the split ids.
- Only then did I read `repair-r1.md`.

## 2. Status of my round-1 findings

| id | Status | Grounds |
|---|---|---|
| PB1 | **released** | d1c2 now copies e3432c0f's `set-digest.golden` to `set-digest-noprecedence.golden` before `Item` changes. It states the digest format: a `none` item prints as `{State Dist}`. It runs `TestSetDigestGolden` and `TestSetDigestNoPrecedenceGolden` together, plus the byte `cmp`. d278 no longer owns any of that, so its #3 single-file-commit check sees only d278's own regeneration. Both criteria stay green at milestone verify: after d278, `TestSetDigestGolden` passes on the regenerated golden. One residual remains: the format statement is checked only by `make test` at each close. That is enough. |
| PB2 | **released** | The d1eb → d278 edge exists. Between d1eb and d278 the guards run in no-precedence mode, so d1eb's `Expected` change does not move `set-digest.golden`. |
| PB3 | **released** | d218 #4 and #6 now cover the production differentials only. d247 gained `TestForestEqualityRefereeCandidateNeverConstrains`, which is hook-counted, logs 0 for the candidate and at least 1 for V1, and is not a grep. That closes SQE #3 as well. |
| PB4 | **released** | d2a2 #7 is now `TestCommittedBaselineMatchesKeys`, which needs no perl.git. New residual N2 is about its TDD step. |
| PB5 | **released** | d166 and 3262 now state one rule: if the un-narrowed rerun dies anywhere, the reason is `grammar`. Both assert that `refused-where.txt` is byte-identical to e3432c0f's, and the precedence fixtures are Go fixtures. I checked the premise: all 336 listed rows are `table` or `grammar`. The only 3 refused snippets the file does not list are the `isa` snippets, and those are refused by resolve's feature scope, not by a library operator or precedence. So the file can stay identical. |
| PB6 | **held, narrowed** | The sizing half is released: d199 now has 14 criteria and 3262 has 4, and the guard files are listed. The part still wrong: once `recognize.New(tab, Options{})` judges the reduce side, these existing tests fail. They are in neither d199's paths nor its guard list nor 3262. `forest/readings_test.go:189` `TestParseAppliesPrecedence` wants `Refusal{Reason: "precedence"}` at Byte 0 for `$a <=> $b <=> $c;`. `cmd/ratchet/verdicts_test.go:132` `TestRatchetVerdictIsPipeline` wants `prec.pl` (`$a <=> $b <=> $c;`) recognized. `cmd/ratchet/stages_test.go:143` `TestPipelineStageKeysRefusedFiles` wants `prec.pl` (`1 <=> 2 <=> 3;`) to reach the precedence stage; the harness builds `recognize.New(tab, recognize.Options{})` at `speed.go:147,286,380,443`, and d2a2, which rewrites it, comes eight issues later. `waypoint2_refusals_test.go:95` `TestWaypoint2RefusalsMatchReference` calls "recognition leaves no forest" an error on rows such as `$a <=> $b <=> $c;`. The repairer moved `TestRecognitionKeysUnchangedByPipeline` into d199 on exactly this ground (`Fatalf` at `verdicts_test.go:196`). These four fail for the same reason. **Fix:** list them in d199, pinned to the no-precedence mode or rewritten as each needs. Also grep every `recognize.New(..., Options{})` caller in test files that feeds a forest to `Constrain`, `ConstrainReference` or the stage harness, since the four are what I confirmed, not necessarily all. |
| PB7 | **released** | d1c2 (8 criteria) refuses nothing and holds the goldens and differentials. 3282 (8 criteria) adds the shift side, the pairing rule, Build by context and the judged differentials. Carries were checked: C6, F6-R4 and F7-A2 are on 3282, and C11, F3-9, P12, A32, C10(d), C13 and F3-4 are on d1c2. The mapping rows were updated. |
| PB8 | residual, not promoted | There is still no brute-force criterion. It is sharper now: d199 explicitly pins the brute force to the no-precedence mode, so if d278 fails to flip it, the brute force goes on passing silently. It remains a criterion the chain does not yet have, not one that says something wrong, so it stays residual. I would add `TestBruteForceJudged` with a non-zero judged count to d278. |
| PB9 | **promoted to blocking** | See section 3. |
| PB10 | **released** | Repaired by AL3's edges d02f → d05d → d089 → d0ba. `feasibility_test.go` is created by d02f, and the d089/d10d Makefile clash is now ordered. |
| PB11 | **released** | The stage order is enforced by d089 → d0ba, and the rationale was rewritten. |
| PB12 | residual | The critical chain grew from 10 to 15: d02f→d05d→d089→d0ba→d0e1→d135→d166→d199→3262→d1c2→3282→d218→d247→d2d1→bb9b, with an equal-length path through d1eb→d278. Its only parallelism is d0e1/d10d and d247/d278/d2a2. The cost is accepted. It is the main reason PB9 matters more now. |
| PB13 | residual | Unchanged. The log-line criteria can be met by typing the logged string. |
| PB14 | residual | Unchanged. The six verdict runs can be fabricated. |
| PB15 | residual | Unchanged for d135 #10 and d218 #8. d278's precondition was rechained to `func TestReferenceRecognizerPrecedence`, which is the same shape. The new d2a2 #7 is gated by its `main.go` grep in the same way. |
| PB16 | **released** | AL4's gating made d199's criterion red today, and the lint confirms it. |
| PB17 | residual | Unchanged. The no-precedence mode is reached from forest (the rerun), chart (the referee, `recognizeSource`) and every pinned guard. `TestParseHasNoPrecedenceSwitch` has to tolerate whatever exported handle provides that. d199 still does not name the mechanism. |
| PB18 | residual | Unchanged. This is the same finding as AL11. |
| PB19 | residual | Unchanged. This is the same finding as AL9. |
| PB20 | residual | Unchanged. git-zhi is still 0.5.2, and the repair itself ran on it. |

## 3. Promotions

| Row | Promoted to | Grounds |
|---|---|---|
| **PB9** (mine) | **blocking** | d166's step, followed literally, does the wrong thing. "The ratchet's direct `Recognize` callers that hold a scope pass narrowed lattices" includes `chartInputs`' `Chart` in `cmd/ratchet/verdicts.go:269-277`. That caller holds `scope` and sums `len(tab.TerminalSets[t.TerminalSet])` per lexed token into `terminals-per-token`, which is one of the 27 answer keys (baseline 1.243). Counting after narrowing lowers that key, which is a §7 stop. A narrowed set id is one the committed tables lack, so indexing `tab.TerminalSets` with it is wrong in any case. The plan causes this by its own instruction, and the first check is d2d1 #7, now 7 issues and two splits later. **Fix:** d166 states that `terminals-per-token` is counted from the un-narrowed lex stream. d166 also adds a criterion pinning the corpus's terminals-per-token. A corpus-only unit test over the in-tree T1 is enough and needs no perl.git. |
| AL5–AL11 | none promoted | I read each row. AL5 (labels in the decline count), AL6 (README-first unchecked), AL7 (A14 generation cost), AL8 (measurements in PR text), AL9 (bb9b negatives), AL10 (chain export) and AL11 (CI grep) are each true, and none makes proceeding wrong. AL8 comes nearest: the C11 and A28 measurements live only in PR text. But the design makes them reports, not gates, so it stays residual. |

## 4. New findings

| id | Class | Issue | Claim | Consequence | Evidence | Suggested action |
|---|---|---|---|---|---|---|
| N1 | **blocking** | d199, 3262 | d199: "Until the precedence-refusal issue that follows, a chart that dies only under precedence is classified by the narrowing issue's grammar-or-scope rule; no corpus file or port snippet has such a death." | d199 pins the rerun's recognizer to the no-precedence mode. Under d166's rule, a death caused only by precedence therefore has a rerun that accepts, so production reports `scope` at the judged death byte. That reason is wrong: V1 says `precedence`. The prose excuses it on the ground that no corpus file or port snippet has such a death. But the unit tests do: `TestParseAppliesPrecedence` and the `prec.pl` ratchet fixtures (PB6). To close d199 green, someone must edit them to expect `scope`, which tunes tests toward a known-wrong answer. 3262 must then edit them back to `precedence` at a non-zero byte. Neither issue says so. | `forest/forest.go` reasons; `forest/readings_test.go:191`; d199 and 3262 bodies | Move the final branch of the rule into d199: when both no-precedence reruns accept, the reason is `precedence` at the judged death byte. 3262 keeps `TestRefusalRerunOrder`, the two readers' one rule, and the fixtures. Add `TestParseAppliesPrecedence`'s rewrite (byte not 0) to d199. |
| N2 | residual | d2a2 | The step reads "Write the failing test … Run it and see it fail for the intended reason", and then removes the key | `TestCommittedBaselineMatchesKeys` passes on today's tree, because `baseline.txt` and `keys` already agree. It cannot fail first, so the step cannot be followed literally. The criterion's red-before comes only from its `main.go` grep (the PB15 shape). | d2a2 Steps | Name it a guard written before the edit, or have the failing half assert that the retired key is absent from both. |
| N3 | residual | d0e1 (AL2 option 1) | The referee comparison drops zero-width slipper `;` leaves on both sides, and nothing else | From the plan-quality side I accept it. The rule is fixed before any recognizer code, and three subtests bound it against widening. One gap: dropping a leaf can make two V2 alternatives identical. If the comparison deduplicates, a doubled alternative V1 lacks would compare equal. F8-P1 measured "none becomes ambiguous", so this is a guard, not a known defect. | d0e1's new criterion | Add a fourth subtest: two alternatives differing only by a zero-width `;` against V1's one alternative are reported different. |
| N4 | residual | d199 | The repairer's call to keep the two retired tests in d199 | I agree. `TestRecognitionKeysUnchangedByPipeline` `Fatalf`s at `verdicts_test.go:196` as soon as the default recognizer refuses its fixture, and `TestPrecedenceIgnoredByRecognition` asserts the opposite of d199. Recorded only so the flag is answered. PB6 extends the same reasoning to four more tests. | — | None. |

## 5. What I did not re-examine

- I did not run `make test` at any intermediate state. PB6 and N1 come from reading the tests named and their recognizers, not from running d199's state. The list of four is confirmed, but it may not be complete.
- I did not re-read the unchanged issue bodies line by line. I diffed them against round 1, and for d02f, d05d, d089, d0ba, d10d, d135, d1eb, d2d1 and bb9b only the edges or nothing changed. bb9b's gated criterion I saw only through the lint.
- I did not judge AL2 against the design text. That is alignment's question. N3 covers only whether the test bounds the rule.
- I did not run the perl.git targets, the fuzz target, or `make chart-ratchet`.
- I did not re-check the milestone's residual-mapping rows other than C6, C10–C13, F3-4, F3-9, F5-3, F6-R4 and F7-A2.
