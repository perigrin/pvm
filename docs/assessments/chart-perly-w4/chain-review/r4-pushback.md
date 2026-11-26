# chain-review round 4: pushback (plan quality) on chart-perly-w4

## Drafter: growth check

I diffed the live chain against my round-3 copy, `pb3/`, before reading `repair-r3.md`:
- 7 issue bodies changed: 3262 (29 lines), d247 (14), d199 (4), d2d1 (4), d166 (3), d218 (3) and d0e1 (2).
- The milestone body gained 10 lines.
- No edge, title or label changed. The DAG is identical to round 3, and there are still 19 issues.

Every changed line maps to a row of `repair-r3.md`, and I found no edit that the log omits.

| Edit | Answers | Blocking? |
|---|---|---|
| d199: drop the transient clause from the criterion and the step | N5 / r3-fresh2 FR3-1 | yes |
| 3262: the `ParseWith` guard paragraph, a TDD step, and the 3-subtest `TestParseHasNoPrecedenceSwitch` criterion | PB17 (promoted by alignment) | yes |
| 3262: `TestRecognitionKeysUnchangedByPipeline` rewritten over `scope.pl`, with a new criterion | N6 | yes |
| 3262: a fate for `TestForestGrammarDataOncePerTables`, `memory_test.go` added to `paths:`, and the test added to the joint criterion | r3-fresh3 FR3-1 | yes |
| d247: defined slipper count, and the criterion literal replaced by "N of 1296" with N > 0 | AL14 (promoted by fresh2) | yes |
| d0e1: the F8-P1 quote corrected | AL14 | yes. It is prose only, and keeps d0e1 consistent with d247. |
| d247: own and rewrite `TestPrecedenceRefereeMatchesReference`, with a step and a criterion | r3-fresh3 FR3-2 | yes |
| d2d1: `.meta` and `.stderr` added to `paths:`; `drive.sh` writes the meta files | FR2-8 (my promotion) | yes |
| d166, d218, d247 and 3262: a standard first step and a "Tests this change moves" list | the coordinator's systemic step | **No finding of any lens asked for it in d166, d218 or d247.** In 3262 it replaces the existing inventory step. |
| d218: the stated fate of `TestPrecedenceGenerationErrorRefusesEveryFile` | FR3-5 | **No.** FR3-5 is a residual. |
| Milestone: C10 and F3-7 rows, and a round-3 Records entry | N6, PB17, mapping | yes |

**Verdict on growth.** Two kinds of edit answer nothing blocking:
- the systemic first step in d166, d218 and d247;
- d218's fate for FR3-5.

Neither adds a criterion, a path, a stop condition or a dependency. Each is one Steps line, plus a Context paragraph naming what the issue expects to go red. Its cost is one extra `make test` per behaviour-changing issue. Its benefit is that a list reasoned from code gets checked against a run before any test is edited. That is the control the repairer's fourth flag asks for, and the pattern 3262's inventory already had. I would keep both kinds of edit. I would not count them as answers to anything, and they should not be cited as having closed FR3-5, which stays residual.

## 1. Recommendation

**Modify.** All three findings I was asked to settle are released: N5, N6 and FR2-8. Nothing I raised is still blocking.

One residual from another lens, r3-fresh3 FR3-7, I now judge blocking (section 3), so I cannot recommend accept. Its fix is one clause in 3282.

How this round was checked:
- I read the live chain and diffed every body.
- I ran `internal/chart/offline/criteria-lint.sh` over all 19 issues. It exits 0, so every criterion is red at `e3432c0f`.
- I counted the bison-refused kept sentences in `sentences.txt`: 1,296 kept, 19 refused. That is what d247's new `verdict only: N>0` needs.
- I listed every caller of `ParseWith`, which is what the PB17 guard depends on.
- I grepped the test suite for pinned scope-refusal bytes, which is what d166's "no test is expected to go red" depends on.
- Then I read `repair-r3.md`.

## 2. Status of my findings

| id | Status | Grounds |
|---|---|---|
| PB1–PB7 | **released** | Unchanged since round 3, and still holding. d1c2, d1eb, d278 and d218's scope are untouched. d247's referee move is consistent with 3262's pin of `referee.judge`. `refused-where.txt` identity is in both d166 and 3262. |
| PB8 | residual, held | There is still no criterion that the brute force runs judged at d278. I would still stand aside on it alone. |
| PB9 | **released** | It is unchanged since round 3. d166's new list says `terminals-per-token` is counted before narrowing, which agrees with its criterion. |
| PB10, PB11 | **released** | Edges unchanged. |
| PB12–PB15, PB18–PB20 | residual | Unchanged. The critical chain is 15 issues. |
| PB16 | **released** | Released in round 2. |
| PB17 | **released** | It was promoted by alignment, and I check it here because it began as my row. **What the fix had to cover:** a runtime guard has to reject every route by which a caller could hand `ParseWith` a no-precedence recognizer. It also has to leave the legitimate users of the no-precedence recognizer working: the rerun, the referees, the pinned guards and the tests. **Checked:** `ParseWith(` has exactly one caller outside `forest.go`, which is `cmd/ratchet/stages_test.go:275`, and it passes `recognize.New(...)`. That recognizer is judged after 3262, so the panic cannot fire there. `Parse` passes `recognize.New` too. None of 3262's pinned sites calls `ParseWith`, which matches its claim. The guard lands at 3262, when `New` first judges, so no earlier state can panic. The criterion's three subtests cover both directions: the panic fires, and the default is accepted. The runtime guard is a defensible choice over `internal/`, because every importer is inside the module. |
| N1 | **released** | Released in round 3. |
| N2, N3, N4 | residual | Unchanged. |
| **N5** | **released** | **What the fix had to change:** d199's criterion text, the test itself (its step), and 3262's fate list, if the test still asserted the transient fact. **Checked:** the criterion now asserts only two lasting facts. The judged constructor's sets differ from the no-precedence constructor's on the sample, and the V1 side reports the no-precedence constructor. The step says the test "must stay true after the switch-over". The transient "`New` is unjudged" lives only in prose. Because the test no longer moves at 3262, 3262 rightly gives it no fate. |
| **N6** | **released** | **What the fix had to change:** 3262's fate (retire becomes rewrite), its criterion, its rewrite step, the README claim (keep it), the milestone's C10 row, and bb9b, if bb9b deleted the README claim. **Checked:** all five are done. The criterion requires a PASS with no SKIP, the log line `fixture refused by the pipeline: scope`, and the README still naming the test. It is red today through its `TestPrecedenceIgnoredByRecognition` conjunct. bb9b's "No V1-only claim" regex does not name this test, so there is no conflict. The premise holds in the tree: `scope.pl` is required to be recognized and not accepted (`verdicts_test.go:136-151`). |
| N7 | residual | 3262's "Not expected to move … If one of them moves, a forest differs from V1's" is unchanged, and `probes_test.go`'s bound has 0.6% headroom (r3-fresh3 FR3-3). The new first step now reports a red test *with its failure text* to the orchestrator, before any edit. So a probe-bound failure reaches perigrin labelled as what it is, even though Context would misname it. I would stand aside on it alone. |
| N8, N9 | residual | Unchanged. N9 is the same finding as AL15. |
| **FR2-8** (my promotion) | **released** | **What the fix had to change:** d2d1's `paths:`, the `drive.sh` step, who writes the meta files, and agreement between the meta keys and what criterion #4 reads. **Checked:** `paths:` lists `runN-{before,after}.{stdout,stderr,meta}`. `drive.sh` itself writes `commit:` (the full hash) and `exit:` among its keys, which satisfies the criterion's `^commit: e3432c0f` prefix match and its `-qx "exit: 0"`. The step says the files come from the one quiet session and are never written by hand. This also resolves FR3-6. |

## 3. Promotions

| Row | To | Grounds |
|---|---|---|
| **r3-fresh3 FR3-7** | **blocking** | It is a step that does the wrong thing when followed literally, and it would fire a false §7 stop. 3282's criterion `TestForestEqualityJudgedRecognizer` reads "the judged recognizer and Build, with no Constrain, equal V1 over the toys and seeds 4 and 44". The infeasible toys `emptyred`, `nullableShift`, `MarkFirstFilter`, `empty` and `scancollide` are, by d199's own `TestRecognizerDoesNotJudgeInfeasibleGrammar`, judged by nothing. With no `Constrain`, their forests keep the alternatives V1 drops, so a correct implementation differs from V1 on exactly those toys. d218 excepts them explicitly ("except every toy `prectab.Feasible` refuses"). 3282, the first issue whose differential runs with no `Constrain`, does not. A worker reading 3282 literally sees "a forest differs from V1's", which every issue lists as a stop, and halts on correct work. 3282 did not change this round. **Fix:** one clause in 3282's prose and criterion: "the toys `prectab.Feasible` passes", as d218 says. |
| r3-fresh3 FR3-3 / N7 | not promoted | See N7 above. The new first step routes the likely failure to perigrin with its text. That is the right destination for a cost bound. |
| r3-fresh3 FR3-4 | not promoted | When a lex-unmapped file has an earlier scope or precedence fault, its reason moves from `lex-unmapped` to `grammar`. That is a real gap in the rule. But the `lex-unmapped` answer key is computed from lexing (`lexUnmapped` in `verdicts_test.go`'s key set and `measure`), not from refusal reasons, and no corpus file or port row has both faults that I know of. It stays residual. |
| r3-fresh3 FR3-5 | not promoted | It now has a stated fate in d218 (see the growth check). It is still not a criterion, and that is acceptable. |
| r3-fresh2 FR3-2, AL15, AL16, and all earlier AL/FR residual rows | not promoted | They are unchanged, and none makes proceeding wrong. |

## 4. New findings

| id | Class | Issue | Claim | Consequence | Evidence | Suggested action |
|---|---|---|---|---|---|---|
| N10 | residual | d166, d218, d247 | The systemic first step: "record, in the PR, every test that goes red" | This is the same outside-the-tree recording AL8 flagged. A reviewer or postmortem cannot read the record from the repository. The step's value is the check against the list, which happens in-session, so this is minor. | The four Steps lines | Optionally commit the red-test record beside the verdict under `docs/assessments/chart-perly-w4/`. |
| N11 | residual | d166 | "Tests this change moves: none is expected to go red … No such test is known" | I checked it. No test pins a scope refusal's byte: the only `Refusal{...}` byte assertions are lex-unmapped/grammar (`forest_test.go:583`) and accept checks (`resolve_test.go`). `TestPipelineStagesReproduceParseWith` compares reasons only. The claim stands as far as reading can take it. | grep over `*_test.go` for `Refusal{`, `ref.Byte` and `"scope"` | None. Recorded so the next reader need not repeat it. |

## 5. What I did not re-examine

- I did not run `make test` under any issue's behaviour change. The lists are the repairer's reading, and N11 is mine.
- I did not re-read the unchanged bodies d02f, d05d, d089, d0ba, d10d, d135, d1c2, d1eb, d278, d2a2, 3282 and bb9b, beyond confirming they are byte-identical to round 3. FR3-7 comes from 3282's unchanged text and d218's.
- I did not re-check the full residual mapping. The milestone diff touches only the C10 and F3-7 rows and the Records entry, and those match the issue bodies.
- I did not run the perl.git targets, the fuzz target or `make chart-ratchet`.
