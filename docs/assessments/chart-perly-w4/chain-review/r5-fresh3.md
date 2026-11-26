# chain-review round 5: fresh3, as a participant (chart-perly-w4)

## 1. Recommendation

**Modify.** I am promoting my own r4-fresh3 FR4-1 to blocking. The round-4 repair changed what the residual meant. Before, it was a gap in one fate at d218. Now d199 and 3262 state that the judged constructor refuses, for every input, any grammar whose table generation fails. With that rule, the switch-over, `01a21463-3262`, would refuse the level-less toy grammars that it lists under "Not expected to move".

The fix is one sentence plus one criterion. Everything else I hold is released or residual.

**Method.**

- I read the live chain at `e3432c0f` first: 19 issues, with edges and labels unchanged.
- I diffed every body and the milestone against my round-4 snapshot. Five issues changed: d166, d199, d1eb, 3262 and 3282, plus Records.
- I re-checked the code behind FR4-1:
  - `forest/readings.go:116-133` errors for any tables value that is not the committed grammar's;
  - `grammar/tables.go` carries no levels or `%prec`, so a toy tables value has no other source of precedence;
  - the `recognize` tests make 113 `New(` calls across 11 files that build toy tables.
- After that I read `repair-r4.md` and every row from r4-*.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| r3-fresh3 FR3-1 | released (in round 4) | Unchanged this round. |
| r3-fresh3 FR3-2 | released (in round 4) | Unchanged this round. |
| r3-fresh3 FR3-3 | **held, residual** | The probe bound is unchanged: `0.6362` against `0.64`, from my round-4 re-run, with no code changed since. `TestCompletionProbesBounded` is still in none of 3262's categories, and 3262's first step only pauses on it. The same point is r4-fresh4 FR4-2. |
| r3-fresh3 FR3-4 | **released** | d166 and 3262 now decide lex-unmapped on the death byte of the un-narrowed, no-precedence rerun, which is V1's rule applied to V1's run. Two cases are pinned, each by a new criterion that is red today: a scope fault followed by an unmapped token is `lex-unmapped` at byte 16 (d166), and a precedence fault followed by an unmapped token is `lex-unmapped` at byte 19 (3262). The knock-on change also landed: d166's `TestRefusalLexUnmappedNotRerun` now covers only a narrowed chart dying at the unmapped token. That case is sound, because the un-narrowed run admits a superset of readings and must die at that same empty set. I checked the byte offsets by hand: `1 .. 2 .. 3;\n` is 13 bytes and `print ` is 6, which gives 19. I did not run the V1 measurements quoted for these fixtures. |
| r3-fresh3 FR3-5 | released into FR4-1 | |
| r3-fresh3 FR3-6 | released (in round 4) | |
| r3-fresh3 FR3-7 | **released** | 3282's prose, step and criterion now cover only the toys `Feasible` passes. The five toys it refuses are excepted and judged by `Constrain` over the no-precedence mode, matching d218. |
| r4-fresh3 FR4-1 | **promoted to blocking**, see §3 | |

## 3. Promotions

**r4-fresh3 FR4-1 → blocking.**

**What it was.** In round 4, d218's fate for `TestPrecedenceGenerationErrorRefusesEveryFile` said the generation error is raised where the judged recognizer reads `prectab`. But `prectab`'s builder also raises an error for every tables value that is not the committed grammar's. The chain never said how the judged `New` should tell those apart.

**Why it is blocking now.** The round-4 repair answered r4-fresh4 FR4-1 by making the refusal rule explicit and general.

- d199 now says: "The judged constructor's table generation returns `Feasible`'s refusal as an error, and `Recognize` returns that error, unwrapped, for every input." Then: "An infeasible grammar is therefore never recognized unjudged by the judged constructor, which would be a precedence off-switch chosen by the grammar's shape."
- 3262 now says that from the switch-over, `parseStream` turns that error into the refusal of every file.

The tables give a judged constructor no precedence other than the committed grammar's: `grammar.Tables` carries no levels. For a toy tables value, the only table generation that exists, `ParsePrecedence`'s builder, fails with "N table rules for perly.y's M: not the committed grammar's tables". So the literal rule, "table generation fails → `Recognize` returns the error for every input", makes `recognize.New(toyTables(...), Options{})` refuse every input once 3262 makes `New` the judged constructor.

**Which tests that hits.** 3262 files these, by name, under "Not expected to move: … the recognizer-internal tests over toy grammars without levels … If one of them moves, a forest differs from V1's." They include the toy cases in `cache_test`, `intern_test`, `lattice_test`, `leo_test`, `leodisagree_test`, `recognize_test`, `setids_test`, `shared_test` and `slipper_test`, and `FuzzRecognize`'s oracle on `toyStmt`.

**Where the chain contradicts itself.** The one place that says what a level-less grammar should get is d278, which comes later: "the level-less fuzzed grammars judge nothing". That means they are recognized, not refused. d199's own refusal rule and d199's "off-switch chosen by the grammar's shape" argument point the other way. Nothing reconciles them.

**Consequence.** At 3262, the first step will see a large batch of unlisted red tests, which the repairer also predicts in `repair-r4.md`. The issue then pauses for the orchestrator. That intermediate state cannot close green until someone makes a design call the chain does not make: either "no precedence for these tables" is the same as "judge nothing", or it is a refusal, which d199's argument favours. This fits the brief's definition of a blocker, a dependency order that leaves an intermediate state unable to close green. It is also a defect in what d199 and 3262 assert.

**Suggested action.** Add one sentence to d199, mirrored in 3262, that separates three cases:

1. **Tables that are not the committed grammar's, with no precedence supplied.** These declare no levels, so the judged constructor judges nothing. Recognition is unchanged, and this is what d278 already says of the fuzzed grammars.
2. **Precedence present but refused by `Feasible`.** `Recognize` returns the refusal for every input.
3. **A generation error on the committed grammar's tables.** `Recognize` returns the error unwrapped.

Then add a d199 criterion, red before the work, showing that on `toyStmt` the judged constructor's sets equal the no-precedence constructor's.

That leaves an open question about the off-switch. A caller could build a non-committed tables value to escape judging. But `Parse` over such tables already refuses every file through `ParsePrecedence` today, so `ParseWith`'s guard is unaffected.

**Other rows read and not promoted.**

- r4-fresh4 FR4-2 overlaps my FR3-3 and this promotion. My promotion carries its blocking half.
- FR4-3 and FR4-4, r4-fresh2 FR4-1, and N10 and N11 are process-scope and coverage points. In each case, the step that catches a missed test or the check that is missing produces a pause, not a wrong result.
- The round-1 to round-3 residuals I weighed in round 4 are unchanged.

## 4. New findings

| id | blocking/residual | issue id | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR5-1 | residual | 01a21463-3262 | `TestParseRefusesInfeasibleGrammar` checks that `Parse`, with the judged default, "refuses every input on each grammar prectab.Feasible refuses, with the generation error". But `Parse` on any toy tables value already refuses every input at `e3432c0f`, with `ParsePrecedence`'s "not the committed grammar's tables" error. | A test that forgets to supply the toy's precedence, through the `grammar.Derived` hook as `TestPrecedenceGenerationErrorRefusesEveryFile` does, passes on today's behaviour without `Feasible` running at all. The criterion only greps for PASS lines. | `forest/forest.go:134-137`. `forest/readings.go:121-122`. 3262's new criterion. | Require each subtest to assert that the refusal is `Feasible`'s error, naming the failing counter, and not the non-committed-tables error. Or require it to log that error and grep for it. |

## 5. Not re-examined

- I did not run the V1 measurements quoted for the new lex-unmapped fixtures (`1 plus 2;…` at byte 16, and `1 plus 2;` as `scope` at byte 2). I checked only the byte arithmetic.
- I did not re-run `criteria-lint.sh` over the edited criteria.
- I ran no V2 code, so FR4-1's batch of red tests is reasoned from `readings.go`, `tables.go` and the toy call sites, not observed.
- I did not check the four issues changed only in round 4 (d0e1, d218, d247 and d2d1) beyond confirming they are byte-identical to my round-4 snapshot.
