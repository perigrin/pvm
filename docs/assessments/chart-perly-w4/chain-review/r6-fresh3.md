# chain-review round 6: fresh3, as a participant (chart-perly-w4)

## 1. Recommendation

**Accept, with residuals.** I am releasing r4-fresh3 FR4-1, which I promoted in round 5. d199 now states where the judged recognizer's precedence comes from, and 3262 and d218 repeat it. With that rule, the level-less toy suites cannot be refused, and a copy or regeneration of perly.y's tables cannot switch judging off.

I found no blocking problem in the repair. I have one new finding, and it is residual. Tables that declare levels but whose rules are not perly.y's are now parsed silently with no judgement, where V1 refused them loudly. No production path builds such tables.

**Method.**

- I read the live chain at `e3432c0f` first: 19 issues, with edges and labels unchanged.
- I diffed every body and the milestone against my round-5 snapshot. Three issues changed (d199, d218 and 3262), plus Records.
- I re-checked the code the repair rests on:
  - `grammar/lr0.go:154-175` makes one BNF rule per non-sequence G1 rule, and a sequence rule expands to two;
  - `forest/readings.go:121`: today's count check passes on perly.y, so perly.y's G1 has no sequence rules, and comparing right-hand sides rule for rule is well defined on the committed tables;
  - `grammar.Tables` (`lr0.go:58-83`) carries no levels.
- After that I read `repair-r5.md` and every row from r5-*.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| r3-fresh3 FR3-1, FR3-2, FR3-4, FR3-6, FR3-7 | released (earlier rounds) | Unchanged this round. I re-checked 3262's memory_test fate against the derived form. That test copies the committed tables (`*tab = *committedTables(t)`), so the copy is judged by content, and the stated fate still holds: the recognizer's derived value plus `prectab`'s. |
| r3-fresh3 FR3-3 | **held, residual** | `TestCompletionProbesBounded` reads `0.6362` against the bound `0.64`, and it is still in none of 3262's categories. The new "Not expected to move" text covers the unqualified `New(` calls on toys, but this test runs on the perly.y corpus sample, and so does `TestRecognizersShareGrammarTables`. If either moves, 3262's first step only pauses. |
| r3-fresh3 FR3-5 | released into FR4-1 | |
| r4-fresh3 FR4-1 (promoted in r5) | **released** | Three things the repair had to change, each checked against the live chain. (a) **Where precedence comes from.** d199 defines an explicit form, which takes a `*prectab.Precedence` and is used by every toy with levels, and a derived form, which `New` becomes. The derived form uses perly.y's precedence only when the rules match perly.y's rule for rule by content. Otherwise it judges nothing, because those tables declare no levels. If the derivation fails on perly.y's tables, `Recognize` refuses every input. (b) **Criteria for each case.** d199 adds three, each red today because the test names are new. `TestJudgedOverLevellessTablesRecognizesAsBefore` covers the recognize toys and the fuzz seeds. `TestJudgedJudgesCopyOfCommittedTables` covers a copy and a regeneration, and requires their sets to differ from the no-precedence sets, so the copy check cannot pass by judging nothing. `TestJudgedRefusesFailedDerivationOnPerlyTables` covers the failure case. (c) **The 3262 and d218 mirrors.** 3262's "Not expected to move" now covers the roughly 113 unqualified toy `New(` calls through d199's criterion. d218's fate for `TestPrecedenceGenerationErrorRefusesEveryFile` now names the derived form's refusal. Both infeasible-toy criteria, d199's and 3262's, now compare against `Feasible`'s text and grep that `not the committed grammar` is absent, which also answers my r5 FR5-1. I hunted for a toy suite this could still break. The toys with levels that go through `recognize.New` (statefree, pinned, forest reference and `toyForest`) are pinned to the no-precedence constructor at 3262. `FuzzRecognize`'s precedence toy uses the explicit form (d278). I found none left. |
| r5-fresh3 FR5-1 | **released** | Answered by the text comparison and the `! grep "not the committed grammar"` guard in 3262's `TestParseRefusesInfeasibleGrammar` criterion. That criterion now goes through `ParseWith` with an explicit judged recognizer, so it can no longer pass on today's non-committed-tables refusal. |

## 3. Promotions

None. I read every residual row from r1 to r5. The ones I weighed:

- **r5-fresh5 FR5-3** (`TestWaypoint2RefusalsMatchReference`): the rows are precedence refusals, and the test errors when recognition leaves no forest. If d218 repoints it at the judged recognizer, it goes red, and d218's first step would not catch that, because removing `Constrain` from `parseStream` leaves this self-contained test green. But 3262 pins the test to the no-precedence mode, so a d218 worker who leaves it pinned sees nothing red. d0e1's `TestForestEqualityWaypoint2` already compares V2 with V1 over the same rows. The worst outcome is ambiguity, not a red state. Residual. One line in d218, "stays pinned; waypoint2 coverage is d0e1's", would close it.
- **r5-fresh5 FR5-4** (whether the candidate enters `Constrain` on the infeasible toys): this resolves through d218's stated exception. Residual.
- **FR2-6 / N8 / FR5-5: issue size.** d199 is now 16 criteria and 7 TDD cycles, and 3262 is 11 criteria and 4 cycles. Both are large. Each added piece is bounded and has its criteria stated, so I keep this residual and leave the sizing call to pushback.

## 4. New findings

| id | blocking/residual | issue id | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR6-1 | residual | 01a21447-d199 (and 3262, d218) | The derived form reads "rules are not perly.y's" as "the tables declare no levels". But `grammar.Tables` records nothing about levels. A tables value generated from a grammar that has `%left`/`%prec` declarations and differs from perly.y by one rule is therefore judged with no precedence at all. In V1, `Parse` over such tables refused every file with `ParsePrecedence`'s error. After d218 drops that call, V2's `Parse` and `ParseWith`, given the derived recognizer that the guard accepts, parse those tables silently and unjudged. d199's own argument, that an unjudged grammar is "a precedence off-switch chosen by the grammar's shape", applies here too, though only to a different grammar. | A loud V1 refusal becomes a silent result. No production caller builds such tables: they all use `committed()` or `committedTables`. So this matters only to a developer who edits the grammar without regenerating against `grammar.Committed()`. No test pins it either way. | `grammar/lr0.go:58-83`, where `Tables` has no levels. `forest/readings.go:116-133`, today's loud error. d199's "Derived" bullet. 3262's guard paragraph. | Either have `grammar.Generate` record whether the grammar declared levels, and have the derived form refuse every input when levels are declared but the rules are not perly.y's. Or say in d199 that a levelled non-perly grammar is unjudged by design, and pin it with a test, so that the change from V1 is recorded rather than silent. |

## 5. Not re-examined

- I ran no V2 code, so the derived form's behaviour is reasoned from the chain and `lr0.go`.
- I did not re-run `criteria-lint.sh` over the new d199 and 3262 criteria. They name new tests, so they are red today.
- I did not re-examine the 16 issues left unchanged this round beyond confirming they are byte-identical to my round-5 snapshot.
