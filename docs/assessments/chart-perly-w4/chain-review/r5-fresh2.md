# chain-review round 5, review-fresh2: chart-perly-w4

I read the chain live from `/home/user/pvm/.claude/worktrees/v2-refine`: 19 issues, with edges unchanged since round 4. I diffed every body against the round-4 snapshot (`f2r5-diff.txt`). The issues that changed are d166, d199, d1eb, 3262 and 3282, plus the milestone's Records. I ran my checks before I read `repair-r4.md`.

## 1. Recommendation

**Modify.** I hold nothing open of my own, but I promote r4-fresh3 FR4-1 to blocking. Round 4's repair turned an open question into a contradiction between two statements in the chain. The fix is one paragraph.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| FR2-1 to FR2-8 | **released** (rounds 3 and 4) | Their text is untouched by this repair. 3262's criterion running `TestWaypoint2RefusalsMatchReference` is unchanged. |
| FR3-1 | **released** (round 4) | d199's `TestJudgedRecognizerIsExplicit` criterion is unchanged and still asserts only durable facts. |
| FR3-2 | **held, residual** | 3262 gained two more criteria (`TestRefusalRerunPrecedenceThenUnmappedIsLexUnmapped`, `TestParseRefusesInfeasibleGrammar`). The pinning commit could still close on its own. |
| FR4-1 | **held, residual** | d218 is unchanged. It still says "repoint the existing production-versus-reference differentials" without naming them or giving a criterion. Its first step cannot surface them, because none calls `Parse`. |

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| r4-fresh3 FR4-1 (d218, 3262; now also d199) | residual | **blocking** | **Precedence is not in `grammar.Tables`.** Today it is derived only for the committed grammar. `buildParsePrecedence` (`forest/readings.go:116-133`) reads `grammar.Committed()`'s levels and `%prec` list by rule id, and returns "not the committed grammar's tables" for any other tables value. d135 moves that into `prectab` unchanged, and 3262 makes `recognize.New(tab, Options{})` the judged constructor. Neither says where the judged `New` gets precedence for a tables value that is not the committed grammar's.<br>**Round 4 made it a contradiction.** d199 now says the judged constructor's table generation "returns `Feasible`'s refusal as an error, and `Recognize` returns that error, unwrapped, for every input". d218 says the generation error is "now raised where the judged recognizer reads `prectab`'s tables". 3262 says `parseStream` "turns that error into the refusal of every file". Read together, the judged `New` on any toy tables value hits `prectab`'s "not the committed grammar's tables" error and returns it for every input. Yet 3262 files "the recognizer-internal tests over toy grammars without levels" under **Not expected to move**. That is about 113 `New(` calls in 14 `recognize` test files, `FuzzRecognize`'s decoded grammars and its replayed seeds among them. Both cannot hold. Two parts of the subject contradicting each other is a defect in what it asserts.<br>**The other reading is no escape.** Suppose instead the judged `New` judges nothing on non-committed tables. Then once d218 removes `Constrain`, `Parse(toyTab, …)` returns unjudged forests where today it refuses. That is the "precedence off-switch chosen by the grammar's shape" which d199 itself forbids.<br>**The repairer agrees.** It predicts that 3262's first `make test` step will report a batch of these tests. A list the issue's own author expects to be wrong cannot be accepted as stated, and the stop it causes is certain, not contingent.<br>**Fix:** one paragraph in d199, mirrored in 3262 and d218. Say how the judged constructor gets its precedence: per tables value through `prectab`'s derived value, from the committed grammar's levels, or passed explicitly to the judged constructor. Say what `New` does on tables that are not the committed grammar's: judge nothing, and keep `Parse` refusing them by a check that does not run through the recognizer. Then move the toy-tables tests into whichever fate that choice gives them. |

I read every other residual row in r1-* to r4-*, including r4-fresh4 FR4-2 to FR4-4, and promote none of them. They are true and fixable later, and none makes proceeding wrong.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR5-1 | residual | 01a21463-3262 | `TestParseRefusesInfeasibleGrammar`: "Parse with the judged default refuses every input on each grammar `prectab.Feasible` refuses, with the generation error … one subtest per toy". | **`Parse` takes only tables.** It cannot see a toy's levels, so on emptyred's tables it already refuses every recognized input today, with "not the committed grammar's tables" (`parseStream` → `ParsePrecedence`).<br>**Two ways to pass.** The test passes for the right reason only if it injects `Feasible`'s refusal through a `grammar.Derived` hook on committed-shape tables, as `TestPrecedenceGenerationErrorRefusesEveryFile` injects `NewPrecedence`'s error. The phrase "in the shape of" hints at this but does not require it. A version that compares no error text passes through the not-committed error, before and after the work.<br>Not blocking: the test is new, so the criterion is red today, and the promotion's fix settles the mechanism. | `forest/forest.go` `parseStream` (lines 133-136). `forest/precedence_test.go:369-392`. | Say the subtests inject each toy's `Feasible` refusal as the committed-shape tables' derived precedence, and compare `Refusal{Reason: thatErr.Error()}`. |

### Round-4 fixes re-derived and found sound

- **FR3-4, `lex-unmapped`.**
  - **d166:** the un-narrowed rerun dying at the first unmapped byte gives `lex-unmapped`, and otherwise `grammar`. The narrowed chart dying at the unmapped token skips the rerun, which is sound: narrowing only removes readings, so the un-narrowed run can die no earlier, and it must die at an empty set.
  - **3262:** the same rule, under precedence. Both carry measured fixtures and new criteria, which are red today because the tests do not exist.
  - I did not run the two byte claims (16 and 19).
- **FR3-7.** 3282's differential and criterion now except the toys `Feasible` refuses, matching d218.
- **r4-fresh4 FR4-1.**
  - d199 owns the refusal of infeasible grammars, with `TestJudgedRecognizerRefusesInfeasibleGrammar` (5 subtests).
  - d1eb's replacement criterion is consistent with it.
  - The milestone Records entry is accurate, and the residual mapping is unchanged.

## 5. What I did not re-examine

- I did not run any criterion, and I did not measure the 16 and 19 byte claims.
- I did not check how d0ba's toy differential builds its candidate on non-committed tables. That depends on the mechanism the promotion asks the chain to name.
- I did not re-read issues outside the diff: d02f to d135, d1c2, d218, d247, d278, d2a2, d2d1 and bb9b.
- The figure of about 113 in-package `New(` calls is a grep count over `recognize/*_test.go`. I did not check each one's tables.
