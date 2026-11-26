# chain-review chart-perly-w4, round 7: review-alignment (resumed)

## 1. Recommendation

**Accept.**

- **AL17 is released.** It was the only blocking finding I held.
- **No promotions, no new blocking finding.** I read every residual row from r1 to r6. I promote none and raise no new blocking finding.
- **Coverage is complete.** Against the design, the rulings and the 66 residual rows, the chain stands as it did in earlier rounds: every row is carried or declined with a reason, and no issue is scope creep.
- **One new residual, AL18.** It is an incomplete list, which the run-first step already covers.

**Checks, made before reading `repair-r6.md`.** I read the live chain first: 19 issues at `e3432c0f`, with edges taken from `blocked_by`.

- **Edges:** unchanged from rounds 3 to 6.
- **Criteria:** I re-ran all 198 against `e3432c0f`, each passed to `sh -c` unmodified. All 198 are red.
- **Residual mapping:** the 66 rows plus the crumb are present. The only mismatches are AL6's five.
- **Carries lines:** unchanged.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| AL1–AL4, AL12–AL14, PB17, FR3-4, r4-fresh3 FR4-1 | released (earlier rounds) | The criteria re-run is 198 of 198 red. The texts that answered them are unchanged, except 3262's guard paragraph. That paragraph now widens PB17's guard; see below. |
| AL17 | **released** | See "AL17 in detail" below the table. |
| AL5–AL11, AL15, AL16 | residual, held | Unchanged. |

**AL17 in detail.**

What the fix had to change as well:
- (a) Narrow the "level-less toys" claim to the calls that really use toy tables.
- (b) Name the perly.y-backed tests in package `recognize` that now judge.
- (c) Give the cost bounds a fate that is not "retune".
- (d) Keep `agree_test` and `setdigest_test` pinned.

What I found:
- **(a)** The first bullet now covers only calls over level-less toy tables, and `TestJudgedOverLevellessTablesRecognizesAsBefore` covers them. That claim is now true.
- **(b)** A "Judge from here, expected to hold" bullet names the equivalence tests over perly.y. It says a red forest or acceptance there is a §7 stop.
- **(c)** A cost-bounds bullet names `TestCompletionProbesBounded`, `TestLeoMemoEntriesBounded`, `TestCacheLookupBounded`, `TestExpectedMemoBounded` and `TestInternOneLookup`. A red bound is "reported to perigrin with the measured number", and the bound is not changed without his ruling. That is the right disposition: a cost move is information for the verdict reading, not a forest difference.
- **(d)** `agree_test` and `setdigest_test` stay pinned.

I re-measured both quoted numbers at `e3432c0f`. `TestCompletionProbesBounded` logs "22217 goTo binary searches over 34924 tokens: 0.6362 per token" against a bound of 0.64. `TestLeoMemoEntriesBounded` logs "1410 … over 35718 nodes: 0.0395 per node" against 0.040. Both match.

**Bias check:** this was my bar. Against that bias, I listed the perly.y-backed tests from the code rather than from the issue. That listing is how I found AL18.

**The rest of the repair, checked.**

- **`ParseWith` guard, widened (r6-fresh4/5/6 FR6-1).**
  - A recognizer now records the precedence it judges with.
  - `ParseWith` panics unless that precedence is the one `New(tab)` would derive. Over perly.y-shaped tables, a copy included, that rejects any explicit precedence, the empty one included.
  - The criterion has two new named subtests: an empty explicit precedence over the committed tables, and the same over a copy of them.
  - The only `ParseWith` caller outside `Parse` is still `TestPipelineStagesReproduceParseWith`, and it passes `New`.
  - `TestParseRefusesInfeasibleGrammar` now goes through `parseStream` inside `forest`, so it does not hit the guard.
  - I found no route left by which a caller of `Parse` or `ParseWith` can bypass perly.y's precedence over perly.y's rules. This is design §4.4.
- **Toys with levels in d218's production differential (N14).**
  - Over the toys that `Feasible` passes, "production" is the explicit judged form, then Build, with no `Constrain`. d199 mirrors this.
  - This follows design §5: "'Production' now means the recognizer's forest". `Parse` cannot see a toy's levels, so the explicit form is the only faithful reading.
  - The phrase "13 toys, less `empty` and `emptyred`" is correct against `toyNames`, which has 13 entries.
- **Brute force (FR6-2).**
  - d278 says the brute force's level-less toys judge nothing, and adds a feasible toy with levels.
  - Its new criterion requires `refused derivations: [1-9]…` and the "judges nothing" note in `forest_test.go`. It is red today.
  - This satisfies "Regenerate everything" for the brute-force outside reference in substance. A regeneration over level-less toys alone would be a no-op.

## 3. Promotions

None.

**Rows I weighed most closely.**

The family raised as N15, r6-fresh2 FR6-1, r6-fresh3 FR6-1, r6-fresh4 FR6-2 and r6-fresh6 FR6-7: the derived form "judges nothing" over tables that are not perly.y's, where V1's `Parse` refused them with "not the committed grammar's tables". It is a real behaviour change, but not one the decision's anti-goals cover:

- No answer key or corpus verdict is computed over such tables.
- The one route by which production's own tables could fall into "not perly.y's" is caught twice. The first is staleness of the committed tables against `grammar.Committed()`, which `tables-drift` holds. The second is `TestJudgedJudgesCopyOfCommittedTables`, which requires judged sets that differ from the no-precedence ones.
- I measured the rule-for-rule match in round 6: 350 rules against 350, with no differences.

It is worth stating in d199, but it does not make proceeding wrong. I leave it residual.

**Other rows read and not promoted:**
- FR6-3 to FR6-9 (r6-fresh6). FR6-9's hook-owner wording is real and worth fixing, but it names the wrong issue for a hook that d218 does add. It is residual.
- PB8, PB12–PB15, PB18–PB20, N2, N3, N7, N9–N13.
- FR4-2–FR4-4, FR5-2–FR5-5.
- AL5–AL11, AL15, AL16.

## 4. New findings

| id | class | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| AL18 | residual | 01a21463-3262 | The "Judge from here, expected to hold" list names the perly.y-backed equivalence tests in package `recognize`. | The list is incomplete. It omits `TestRecognizersExtendSharedTablesConcurrently`, `TestRecognizersOverDistinctTablesShareNothing` and `TestUnsetTokenRefuses` (`shared_test.go`, all built over `committedTables` or `ownTables`). Each compares recognizers with one another, so each should hold judged. The run-first step catches a red one unlisted. Nothing claims the list is exhaustive. | I listed every `func Test…` in `recognize/*_test.go` whose body uses `committedTables`, `corpusSample` or `ownTables`. `agree_test` and `setdigest_test` are pinned as files. | Add the three names to the bullet, or say the bullet is by file (`shared_test.go`). |

## 5. What I did not re-examine

- **Whether any cost bound actually goes red under a judged recognizer.** I did not build one. The chain now says what happens if one does.
- **Unchanged issues.** Only d199, d218, d278 and 3262 changed this round. I did not re-read the other 15 beyond the criteria re-run and the mapping cross-check.
- **perl.git targets.** Their criteria are gated red today.
- **`ParseWith`'s guard against a recognizer whose precedence equals the derived one in content but is a different pointer.** The issue says "the one `New(tab)` would derive". Equality by pointer or by value is the builder's choice under §6. Either rejects the empty explicit precedence, which is what the criterion checks.
