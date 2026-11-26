# chain-review round 6, review-fresh2: chart-perly-w4

I read the chain live from `/home/user/pvm/.claude/worktrees/v2-refine`: 19 issues, edges unchanged. I diffed every body against my round-5 snapshot (`f2r6-diff.txt`). Only d199, d218, 3262 and the milestone Records changed. I ran my checks before reading `repair-r5.md`.

## 1. Recommendation

**Modify.** The promotion I held, r4-fresh3 FR4-1, is answered and I release it. One finding of my own, FR4-1, now has a concrete break attached, from r5-fresh5 FR5-3. I hold it as blocking. The fix is one paragraph in d218.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| FR2-1 to FR2-8, FR3-1 | **released** (earlier rounds) | Their text is untouched this round. |
| FR3-2 | **held, residual** | 3262 is unchanged in scope, and FR5-5 makes the same point. |
| FR4-1 | **held, now blocking** | **The break is concrete.** d218 still says it will "repoint the existing production-versus-reference differentials at the recognizer's forest". It names none of them, says nothing of inputs the judged chart refuses, and has no criterion for them. One of that class is `TestWaypoint2RefusalsMatchReference`. Its doc comment says it "requires production Constrain and ConstrainReference to agree". It errors "recognition leaves no forest" on any row that does not recognize, and its precedence rows die in the judged chart. 3262 pins it to the no-precedence constructor and runs it in a criterion that `git zhi verify` re-runs at milestone close.<br>**Both readings of d218 fail.**<br>• Read literally, d218 repoints that test, and it, with 3262's criterion, goes red after 3262 closed. That is the same shape as FR3-1, a criterion red at verify because of a later issue.<br>• Read loosely, d218 repoints none of them, and design §5's "'Production' now means the recognizer's forest" goes unmet with nothing noticing.<br>**Nothing catches either.** d218's first step removes `Constrain` from `parseStream`, and none of these tests calls `Parse`, so the step cannot surface them. This is the subject saying something that cannot be followed literally without breaking another issue's criterion, so it is blocking.<br>**Fix:** name the tests d218 repoints, among them the forest and ratchet `precedence_reference_test.go` files, `precedence_perlgit_test.go` and `TestWaypoint2RefusalsMatchReference`. Give each refused input the verdict-only rule d247 gave `TestPrecedenceRefereeMatchesReference`, and add one criterion that runs them. Or state that they stay `Constrain` referees on the no-precedence mode, and reconcile that with design §5. |
| FR5-1 | **released** | 3262's criterion now builds `ParseWith` with a judged recognizer over each failing toy's own precedence. It compares `Feasible`'s refusal as text, and it requires the absence of "not the committed grammar". So it can no longer pass on today's error. d199's criterion carries the same guard. |

## 3. Promotions

| row | status | grounds |
|---|---|---|
| r4-fresh3 FR4-1 (my round-5 promotion) | **released** | **Where precedence comes from is now stated (d199).** The explicit form takes a `*prectab.Precedence`, and every toy with levels is judged through it. The derived form, which `New` becomes, judges with perly.y's precedence when the tables match perly.y rule for rule, by content. It refuses every input when that derivation fails. On any other tables it judges nothing, by the fact that they declare no levels.<br>**The contradiction I promoted on is gone.** The roughly 113 level-less toy `New(` calls now sit in "not expected to move" consistently, with `TestJudgedOverLevellessTablesRecognizesAsBefore` (`recognize-toys`, `fuzz-seeds`) as their check. A failed derivation refuses rather than judging nothing, pinned by `TestJudgedRefusesFailedDerivationOnPerlyTables`.<br>**The match cannot pass silently.** I hunted for a way the stricter right-hand-side match could make the committed tables themselves fall into "judge nothing". It cannot pass unnoticed. `TestJudgedJudgesCopyOfCommittedTables` requires a copy and a regeneration to be judged, with sets different from the no-precedence constructor's, and 3262's `TestDefaultRecognizerJudgesPrecedence` requires `New` to refuse the three fixtures on the committed tables. A representation mismatch between `Tables.Rules` and `G1` goes red at d199. perly.y's G1 has no sequence rules that could expand.<br>**Mirrored where it should be.** d218 and 3262 mirror the text. Every new criterion is a new test, so none is red-before-green in the wrong order. |
| r5-fresh5 FR5-3 | **merged into my FR4-1** | It supplies the concrete red test that makes FR4-1 blocking. It is not a separate row. |

I read the other residual rows in r1-* to r5-*, among them FR5-2, FR5-4 and FR5-5, N12 and N13. They are each true and residual on their own terms. FR5-2 comes nearest. It is the same class as FR4-1, a d166 criterion that 3262 could break, but 3262's rerun prose can be built to keep d166's no-rerun shortcut, so it can be followed without breaking anything. I promote none.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR6-1 | residual | 01a21447-d199 (mirrored in 3262) | The derived form judges nothing on any tables whose rules are not perly.y's rule for rule, including the right-hand sides, "by fact, not by fallback". | **Today's failure was loud.** `buildParsePrecedence` errors when the count or an LHS differs, so `Parse` refuses every recognized file on such tables.<br>**Now it is silent.** A tables value that is perly.y with one right-hand side changed has the same count and LHSs, so today it gets the loud error. Under the derived form it is "not perly.y's", so it is judged by nothing. Once d218 removes `Constrain`, `Parse` accepts on it unjudged. So the edge between "perly.y with a failed derivation: refuse" and "another grammar: judge nothing" falls inside near-perly tables, which d199's own principle calls a precedence off-switch chosen by the grammar's shape.<br>**Why it is not blocking.** Production only loads the committed tables, and `TestWriteTablesAndDrift` in `make test` catches drift between them and perly.y. So no caller reaches this today. | `forest/readings.go:116-133` (count and LHS check, an error). d199's "Derived" bullet. | Keep today's count-and-LHS test as the boundary. Tables that match perly.y on count and every LHS but differ in a right-hand side are a failed derivation, which refuses every input, not a different grammar. Pin it with one subtest. |

## 5. What I did not re-examine

- I ran no criterion, and I did not build anything to compare `Tables.Rules` symbol names against `G1`'s RHS names. I rely on d199's new criterion to catch a mismatch.
- I did not re-read the issues this round left unchanged: d02f to d166, d1c2, 3282, d1eb, d247, d278, d2a2, d2d1 and bb9b.
- I did not check whether `TestJudgedOverLevellessTablesRecognizesAsBefore` can reach `FuzzRecognize`'s seed grammars from the test's package. I assume both live in `recognize`.
