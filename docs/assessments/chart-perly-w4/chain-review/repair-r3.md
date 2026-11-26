# chain-review round 3: refinement repair on chart-perly-w4

## Scope

This round answers the seven blocking findings still open:
- N5 / r3-fresh2 FR3-1
- AL14 (promoted by fresh2)
- PB17 (promoted by alignment)
- N6
- FR2-8 (promoted by pushback)
- r3-fresh3 FR3-1
- r3-fresh3 FR3-2

It also adds the coordinator's systemic first step. No residual finding was answered.

All edits went through the `git zhi` CLI (git-zhi 0.5.2). The repository, git config and branches were not touched. No issue was added or split, and no edge or title changed. There are still 19 issues.

## The systemic step

Each issue that changes default behaviour now has the same first step: d166, d218, d247 and 3262. It runs before the README commit, which is safe because the step commits nothing:

1. Apply only this issue's behaviour change, in the working tree, with no test edits.
2. Run `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit make test`.
3. Record every red test in the PR.
4. Check the record against the issue's new Context list, "Tests this change moves".
5. A red test missing from that list is reported to the orchestrator, with its failure text, before anything is edited. It is never fixed silently.
6. Undo the change.

Each issue's stated list:

| Issue | Behaviour change applied | Stated list |
|---|---|---|
| d166 | Narrowing before recognition, plus the rerun rule | No test is expected to go red. The only kind that may move is a test pinning a scope refusal's byte, which perigrin's ruling allows to move; none is known. |
| d218 | `parseStream` no longer calls `Constrain` | The repointed production-versus-reference differentials, and possibly `TestPrecedenceGenerationErrorRefusesEveryFile`. Its stated fate is that `Parse` still refuses with the generation error, now raised from `prectab`'s tables. `TestPipelineStagesReproduceParseWith` and the two known-limit pins are expected to stay green. |
| d247 | The referee narrows the lattice before recognition and judges the recognizer's forest | `TestPrecedenceRefereeMatchesReference` (see r3-fresh3 FR3-2). `TestPrecedenceReferee` is expected to stay green. |
| 3262 | `recognize.New` becomes the judged constructor | The existing call-site fates list, retitled "Tests this change moves", with the two additions below. The old inventory step is replaced by this step, and failures are now reported to the orchestrator. |

## Each finding

| Finding | Answer | Issues |
|---|---|---|
| N5 / r3-fresh2 FR3-1 | The transient clause is dropped. The criterion `TestJudgedRecognizerIsExplicit` now asserts only two lasting facts: the judged constructor's sets differ from the no-precedence constructor's on the perly.y sample, and the V1 side's recognizer comes from the no-precedence constructor. That `New` is still unjudged at d199 stays in the prose. The step says the test must stay true after the switch-over. | 01a21447-d199 |
| AL14 | The "1,296 of 1,299" denominator is corrected in d247 and d0e1. There are 1,296 kept sentences, the `sentences.txt` rows with a rendering. "Carries a leaf" is defined: the slipper inserted a zero-width `;` into that sentence's narrowed-lattice chart. The prose gives the other measures for contrast: on V1's Build path the leaf is present in 1,277 forests and reachable in none. d247's criterion no longer pins a literal for an undefined count. It requires `slipper ; insertions: N of 1296` with N > 0 (F8-P1 measured 1,296), and still requires `ambiguous: 0` and `cut: 0`. | 01a21447-d247, 01a21447-d0e1 |
| PB17 | `recognize.Recognizer` records which constructor built it. `ParseWith` panics on a no-precedence recognizer, so no caller of `Parse` or `ParseWith` can select the mode, including after d218 removes `Constrain`. The no-precedence recognizer is reached only by the rerun inside `parseStream`, by the referees, by the pinned guards and by tests, and none of them goes through `ParseWith`. The guard lands at 3262, because before that `New` is the no-precedence recognizer. The criterion `TestParseHasNoPrecedenceSwitch` now needs three subtests: `no-parameter-or-option` (the reflection check), `parsewith-refuses-no-precedence-recognizer` and `parsewith-accepts-default`. A TDD step was added. A runtime refusal was chosen over a Go `internal/` package, because every caller in this module can import `internal/chart/...`, so visibility alone would not exclude callers such as the ratchet. | 01a21463-3262 |
| N6 | `TestRecognitionKeysUnchangedByPipeline` moves from "Retired" to "Rewritten". It is rewritten over `scope.pl` (`use feature 'class'; no warnings; method foo {}`, from `TestRatchetVerdictIsPipeline`), which every recognizer accepts and the pipeline refuses for scope. The README claim stays. The new criterion checks five things: it passes without skipping; it logs `fixture refused by the pipeline: scope`; the README still names the test; and `TestPrecedenceIgnoredByRecognition` is gone, which is the red-today conjunct. | 01a21463-3262 |
| FR2-8 | d2d1's `paths:` now lists `runN-{before,after}.{stdout,stderr,meta}`. The `drive.sh` step says `drive.sh` writes each run's `.meta` file in the form of `chart-perly-w3/verdict/*.meta`, with one `key: value` per line. The keys are `run`, `state`, `commit` (full hash), `loadavg-before`, `load1`, `cores`, `kernel`, `waited-for-quiet-s`, `start`, `end` and `exit`. The files come from the one quiet session. None is written by hand or by a second session. The criteria are unchanged and already read `commit:` and `exit:`. | 01a21447-d2d1 |
| r3-fresh3 FR3-1 | `TestForestGrammarDataOncePerTables` moves from "Not expected to move" to "Rewritten" in 3262. Once `New` judges, the test counts the derived values: after `New` and `Recognize` it expects the recognizer's and `prectab`'s; after `Build` it expects those plus the forest's. Each is still built once per tables value. `memory_test.go` was added to `paths:` and to the rewrite step. 3262's joint criterion now also runs the test. | 01a21463-3262 |
| r3-fresh3 FR3-2 | d247 now owns `TestPrecedenceRefereeMatchesReference`. It names the test in "Tests this change moves" and rewrites it rather than retiring it. For each kept sentence, the judged referee's verdict and kept alternatives are compared with `ConstrainReference` over the no-precedence narrowed forest. Where the reference refuses (19 sentences at e3432c0f), only verdicts are compared (design §5). There is a new step, and a new criterion requiring `compared 1296 of 1296`, `verdict only: N` with N > 0, and no skip. Today the test logs `compared=1296`, so the criterion is red. | 01a21447-d247 |

## Change log

| Issue | Change | Finding |
|---|---|---|
| 01a21447-d199 | Rewrote the `TestJudgedRecognizerIsExplicit` criterion's description, dropping the `New`-equals-no-precedence clause. Reworded its step. The command is unchanged. | N5 / r3-fresh2 FR3-1 |
| 01a21463-3262 | Added the `ParseWith` guard paragraph, a guard TDD step, and the three-subtest `TestParseHasNoPrecedenceSwitch` criterion. | PB17 |
| 01a21463-3262 | Moved `TestRecognitionKeysUnchangedByPipeline` from Retired to Rewritten, over `scope.pl`. Updated the rewrite step and replaced its criterion. | N6 |
| 01a21463-3262 | Added the `TestForestGrammarDataOncePerTables` fate, `memory_test.go` in `paths:`, and the test in the joint criterion. Narrowed "Not expected to move". | r3-fresh3 FR3-1 |
| 01a21463-3262 | The inventory step became the standard first step, reporting to the orchestrator. The call-site list is titled "Tests this change moves". | systemic |
| 01a21447-d166 | Added "Tests this change moves" and the standard first step. | systemic |
| 01a21447-d218 | Added "Tests this change moves", including `TestPrecedenceGenerationErrorRefusesEveryFile` with its fate, and the standard first step. | systemic |
| 01a21447-d247 | Gave the 1,296 kept sentences, defined "carries a leaf", and replaced the slipper-pin criterion's literal with a defined non-zero count. | AL14 |
| 01a21447-d247 | Added "Tests this change moves" naming `TestPrecedenceRefereeMatchesReference` and its rewrite. Added a step and a criterion. | r3-fresh3 FR3-2 |
| 01a21447-d247 | Added the standard first step. | systemic |
| 01a21447-d0e1 | Corrected the F8-P1 quote: 1,296 of 1,296 kept sentences in the narrowed run, and 1,277 present but none reachable on V1's path. | AL14 |
| 01a21447-d2d1 | `paths:` gains the `.stderr` and `.meta` files. The `drive.sh` step spells out the `.meta` form and says who writes the files. | FR2-8 |
| Milestone | The C10 row says (a) rewrites `TestRecognitionKeysUnchangedByPipeline` instead of retiring it. The F3-7 row adds the `ParseWith` guard. Records gained a "Chain-review round 3 repair" entry. | N6, PB17, mapping |

Every other line of each edited body is byte-identical to the live body read at the start of this round, including Negative Scenarios, the README-first steps and the close steps. All 7 stored bodies and the milestone body were re-read and match their source files. Labels and the resolution command are unchanged.

## Verification

- All 6 new or changed criterion commands were run at e3432c0f with `PERL5_CORPUS=/nonexistent`, and all are red.
- Every test the new text names exists today: `TestForestGrammarDataOncePerTables` (`forest/memory_test.go:46`), `TestPrecedenceRefereeMatchesReference` (`precedence_test.go:275`, logging `sentences=1296 compared=1296 refused=19 pruned=1277`), `TestPrecedenceGenerationErrorRefusesEveryFile` (`forest/precedence_test.go:370`), `TestPipelineStagesReproduceParseWith`, and the `scope.pl` fixture (`verdicts_test.go:137`). The README names `TestRecognitionKeysUnchangedByPipeline` at line 485.
- `internal/chart/offline/criteria-lint.sh` exits 0 over all 19 issues.

## Unchanged

- **Issue order and critical chain.** The critical chain is 15 issues: d02f → d05d → d089 → d0ba → d0e1 → d135 → d166 → d199 → d1c2 → 3282 → 3262 → d218 → d247 → d2d1 → bb9b.
- **Residual findings, which this round left alone:**
  - N7 / r3-fresh3 FR3-3: the cost bounds over perly.y. If they move at 3262, the systemic step reports them; they are not fixed.
  - N8 and r3-fresh2 FR3-2: the size of 3262.
  - N9 and AL15: the logged count for `TestJudgedRecognizerIsExplicit`.
  - AL16.
  - r3-fresh3 FR3-4 and FR3-7.
  - FR3-5 got only a stated fate in d218's list, not a criterion.
