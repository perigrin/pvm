# chain-review round 4: fresh3, as a participant (chart-perly-w4)

## 1. Recommendation

**Accept, with residuals.** Both of my blocking findings, r3-fresh3 FR3-1 and FR3-2, landed. The edits they needed elsewhere also landed, and I found nothing they broke. I have no promotions. I have one new finding, and it is residual: the repair's stated fate for `TestPrecedenceGenerationErrorRefusesEveryFile` at d218 is incomplete.

**Method.**

- I read the live chain at `e3432c0f` before reading the log: 19 issues, with edges unchanged.
- I diffed every body and the milestone against my round-3 snapshot. Seven issues changed: d0e1, d166, d199, d218, d247, d2d1 and 3262, plus the milestone. 3282 is unchanged.
- I re-ran the tests behind my findings. The results match round 3:
  - `TestCompletionProbesBounded` reads `0.6362` against the bound `0.64`;
  - `TestPrecedenceRefereeMatchesReference` logs `sentences=1296 compared=1296 refused=19 pruned=1277`;
  - `TestForestGrammarDataOncePerTables` and `TestPrecedenceGenerationErrorRefusesEveryFile` both pass today.
- After that I read `repair-r3.md` and the residual rows of every report from r1 to r3.

## 2. Status of my findings (r3-fresh3)

| id | status | grounds |
|---|---|---|
| r3-fresh3 FR3-1 | **released** | 3262 now lists `memory_test.go` in its paths. It moves `TestForestGrammarDataOncePerTables` to "Rewritten", where after `New`+`Recognize` the test expects the recognizer's and `prectab`'s derived values, and after `Build` it expects those plus the forest's. `memory_test.go` is narrowed out of "Not expected to move" except for its other tests. The joint criterion runs the test, and that criterion is red today because `TestDefaultRecognizerJudgesPrecedence` does not exist yet. I also checked the knock-on: the rewrite step names the test. One residual note: the fate pins exactly one `prectab` derived value. If d1c2's collapse table or d135's item mapping is cached as a second one, the count moves again, but 3262's new first step would catch that. |
| r3-fresh3 FR3-2 | **released** | d247 names `TestPrecedenceRefereeMatchesReference` under "Tests this change moves". Its fate is a rewrite: the judged referee is compared against `ConstrainReference` over the no-precedence narrowed forest, with verdicts only where the reference refuses. The issue adds a TDD step and a criterion requiring `compared 1296 of 1296` and `verdict only: N>0`. That criterion is red today, because the log reads `compared=1296`. Knock-on checks: the rewrite needs a no-precedence narrowed forest, which `builtForest`, pinned at 3262, still gives. The criterion `\.Constrain(` does not match `.ConstrainReference(`, so the "no production file calls Constrain" criterion stays satisfiable. |
| r3-fresh3 FR3-3 | **held, residual** | The probe bound still has 0.6% headroom (re-run: `22217/34924 = 0.6362` against `0.64`). The test is perly-backed, and it is still in none of 3262's categories. "Not expected to move" still says "recognizer-internal tests over toy grammars … If one of them moves, a forest differs from V1's." The new first step now reports an unlisted red test to the orchestrator rather than to perigrin, so it is a pause, not a halt. That is why I keep this residual and do not promote it. |
| r3-fresh3 FR3-4 | **held, residual** | Unchanged. With "never rerun" for lex-unmapped, a file whose narrowed or judged chart dies before an unmapped token gets `grammar` at the unmapped byte, where V1 gives `lex-unmapped`. d166's new list claims every `refused-where.txt` row keeps its reason, which is true of the port rows but not of this two-fault case. No answer key moves, under the ruling "Accept/refuse only". |
| r3-fresh3 FR3-5 | **released into new finding FR4-1** | d218 now names the test and gives it a fate. The fate as stated creates a new problem; see FR4-1. |
| r3-fresh3 FR3-6 | **released** | d2d1's paths now list `runN-{before,after}.{stdout,stderr,meta}`. The `drive.sh` step specifies the `.meta` keys, including `commit:` with the full hash and `exit:`, which is what the criteria read. A full hash starts with `e3432c0f`, so the check `^commit: e3432c0f` holds. |
| r3-fresh3 FR3-7 | **held, residual** | 3282's `TestForestEqualityJudgedRecognizer` still covers "the toys" with no `Constrain` and no exception for toys that `Feasible` refuses. d218 states that exception, but 3282 does not. |

## 3. Promotions

None.

I read every residual row from r1 to r3: AL5–AL11, PB8–PB20, AL14–AL16, FR2-5–FR2-8, N2–N4, N7–N9, and r3-fresh2 FR3-2. The ones I weighed:

- **AL11 / PB18.** d10d says "CI's Ratchets job runs that target", yet Records say CI has no perl.git, so that job exits 2. P10's "found by CI" is therefore not true today. But the differential runs, and fails rather than skips, wherever perl.git is present, and every criterion that depends on perl.git is guarded. Residual.
- **N8 / r3-fresh2 FR3-2: the size of 3262.** The issue grew this round, gaining the guard TDD and two more rewrites. It is now about 11 criteria, and a full `make test` runs twice. Most of the added work is mechanical test edits with stated fates, so I judge it large but one session. Residual.
- **N7.** This is the same as my FR3-3, held above.

## 4. New findings

| id | blocking/residual | issue id | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR4-1 | residual | 01a21447-d218 (with 01a21463-3262) | d218's fate for `TestPrecedenceGenerationErrorRefusesEveryFile` says the generation error is "now raised where the judged recognizer reads `prectab`'s tables". But `ParsePrecedence`'s builder errors for every tables value that is not the committed grammar's: `forest/readings.go:121`, "N table rules for perly.y's M: not the committed grammar's tables". The default constructor builds every toy-grammar recognizer: `New(toyTables(...))` across `recognize`'s tests, which 3262 files under "Not expected to move". So at 3262 the judged `New` must swallow that error. Raising it literally at d218 would turn every toy recognizer test red. Raising it would require telling "not the committed grammar" apart from a real generation error, and neither issue says so. | d218's first step would surface many unlisted red tests and pause for the orchestrator. Or the worker invents the distinction. The test's exact-message check, `Refusal{Reason: genErr.Error()}`, also requires the recognizer to pass the error through unwrapped. | `forest/readings.go:116-133`. `forest/precedence_test.go:369-392`, which sets the error through `grammar.Derived` on committed-shape tables. `recognize/*_test.go` toy call sites. 3262 "Not expected to move". | One sentence in d218, mirrored in 3262: the judged `New` judges nothing for tables that are not the committed grammar's. For the committed grammar's tables, a generation error is returned unwrapped by `Recognize`, and `parseStream` turns it into the refusal. |

## 5. Not re-examined

- I did not re-examine anything outside the seven changed bodies and my own findings, beyond the residual rows listed above.
- I did not rerun `criteria-lint.sh`. The repairer reports that it exits 0, and its pin half is still vacuous on this chain.
- I ran no V2 code, so FR3-3's direction and FR4-1's practical effect are reasoned from the code, not observed.
- I did not check the other participants' released blockers (N5, AL14, PB17, N6, FR2-8) beyond PB17's interaction with `ParseWith`'s callers. The only caller outside `forest` is `TestPipelineStagesReproduceParseWith`, which uses `recognize.New`, so the guard breaks nothing.
