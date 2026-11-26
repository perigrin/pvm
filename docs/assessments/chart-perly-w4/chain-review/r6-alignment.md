# chain-review chart-perly-w4, round 6: review-alignment (resumed)

## 1. Recommendation

**Modify.**

I release r4-fresh3 FR4-1, the finding I held. The judged recognizer's precedence source is now stated in two forms, explicit and derived. The stricter rule-for-rule match holds on the committed tables. I measured that, and it leaves no off-switch.

The repair also added one sentence to 3262 that is false. 3262 says the roughly 113 unqualified `New(` calls in package `recognize`'s tests "run over level-less toy tables". In fact some run over perly.y's own tables. Among them are the cost-bound tests: `TestCompletionProbesBounded`, `TestCacheLookupBounded`, `TestExpectedMemoBounded` and `TestInternOneLookup`. These judge from 3262 on, yet 3262 lists them as "Not expected to move". This is AL17, blocking, and it takes in r3-fresh3 FR3-3. The fix is one Context bullet.

The rest of the subject is consistent with the decision. If AL17 lands, I expect to recommend accept.

**What I checked, live, before reading `repair-r5.md`:**
- **Chain:** 19 issues at `e3432c0f`. Edges are unchanged.
- **Criteria:** I re-ran all 197 against `e3432c0f`, each passed to `sh -c` unmodified. All 197 are red.
- **Residual mapping:** 66 rows plus the crumb. The only misses are AL6's five.
- **Carries:** unchanged.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| AL1–AL4, AL12–AL14, PB17, FR3-4 | released (earlier rounds) | The criteria re-run is all red. The text of the earlier fixes is unchanged, except 3262's FR4-1 sentence, which I examine below. |
| r4-fresh3 FR4-1 (my r5 promotion) | **released** | See "FR4-1: what the fix had to change" below the table. |
| AL5–AL11, AL15, AL16 | residual, held | Unchanged. |

**FR4-1: what the fix had to change, and what I found.**

- **(a) A stated precedence source for the judged `New`.** Done. d199 says that "Explicit" takes a `*prectab.Precedence`, and "Derived" decides by content.
- **(b) What `New` does over tables that are not the committed grammar's.** Done. They judge nothing "by fact". `TestJudgedOverLevellessTablesRecognizesAsBefore` checks two families, `recognize-toys` and `fuzz-seeds`.
- **(c) A real derivation failure on perly.y refuses every input, never judges nothing.** Done, pinned by `TestJudgedRefusesFailedDerivationOnPerlyTables`.
- **(d) The `Feasible` refusal tests cannot pass on the not-committed error.** Done. d199's and 3262's criteria go through the explicit form, compare `Feasible`'s text, and fail if the output contains "not the committed grammar".
- **(e) The `ParseWith` guard is not tripped by `New` over toy tables.** Done. The derived form is the judged constructor, so it does not carry the no-precedence mark.
- **(f) A copy of the committed tables is still judged.** Done. Matching is by content, and `TestJudgedJudgesCopyOfCommittedTables` checks a copy and a regeneration.

**The repairer's flag: the rule-for-rule match also compares RHS names.**

The risk: if the committed `Tables`' rules did not match `grammar.Committed()`'s G1 by RHS name, production would silently judge nothing. That would be the worst off-switch possible.

I measured it at `e3432c0f` (scratch copy, `forest/zz_al6_test.go`):
- 350 table rules against 350 G1 rules.
- 0 LHS, 0 RHS-length and 0 RHS-name differences.
- No sequence rules.
- All 862 RHS symbols are of kind `Name`, so literals and classes cannot make the comparison ambiguous.

So the stricter match holds on production's tables. If a later grammar change broke it, the criterion `TestJudgedJudgesCopyOfCommittedTables` would go red, because the judged sets would equal the no-precedence ones.

`ownTables` in `shared_test.go` is a struct copy of the committed tables, so it is judged too. That keeps those shared-versus-own comparisons symmetric.

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| r3-fresh3 FR3-3 (with pushback N7) | residual | **blocking**, folded into AL17 | The round-5 repair turned the probe-bound risk from an unlisted possibility into a stated falsehood. See AL17. |

**Rows I read and did not promote.** I read every row in r1–r5, from every participant.
- **Not promoted:** N12, FR5-1 (fresh2 and fresh3), N13, FR5-2, FR5-3, FR5-4 and FR5-5. FR5-4 is the hook-counted toy candidate, which must except the infeasible toys. It is a gap in what the issue says, not a false claim.
- **Unchanged and not promoted:** PB8, PB12–PB15, PB18–PB20, N2, N3, N9–N11, FR4-2–FR4-4 and FR2/FR3 residuals.

## 4. New findings

| id | class | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| AL17 | **blocking** | 01a21463-3262 | See AL17 in detail below the table. | See below. | See below. | See below. |

**AL17 in detail.**

**What 3262 says.** Under "Not expected to move" it says: "The roughly 113 unqualified `New(` calls inside package `recognize`'s tests run over level-less toy tables, so under the derived form they recognize exactly as before. They are covered by … `TestJudgedOverLevellessTablesRecognizesAsBefore`". The list it leads also names "the rest" of the recognizer-internal tests.

**Why it is false.** Several of those calls use perly.y's own tables:
- `cache_test.go`: `TestCacheLookupBounded`, `TestExpectedMemoBounded` and the warm/cold test, via `committedTables`;
- `probes_test.go`: `TestCompletionProbesBounded`, `TestPostdotIndexMatchesSetScan` and `TestPostdotIndexConcurrentRecognizers`, via `corpusSample`;
- `intern_test.go`: `TestInternOneLookup`, via `corpusSample`;
- `shared_test.go`: via `committedTables` and `ownTables`.

Under the derived form, every one of these judges from 3262 on.

**Consequence.** Three of them are cost bounds, and the context adds sets. C11 measured +9.3% distinct sets.
- `TestCompletionProbesBounded` reads 0.6362 against a bound of 0.64, which is 0.6% headroom (r3-fresh3 FR3-3).
- `TestCacheLookupBounded` and `TestExpectedMemoBounded` are likewise bounds on sets the context changes (pushback N7).

A red bound at 3262 is not on the list. Its run-first step therefore sends it to the orchestrator, and amending a started issue is a §7 stop. The prose also tells the builder they "recognize exactly as before", which is false.

**Evidence.** I ran `grep -n 'New(tab' -B12` over the tests. The tables come from `committedTables` or `corpusSample` (perly.y) in `cache_test.go` lines 55, 103 and 131, `probes_test.go` 29, `intern_test.go` 11, and `shared_test.go`. `completionProbesBound = 0.64`, and r3-fresh3 measured 0.6362. The 113 count includes 82 `New(tab` calls, and `tab` is not always a toy.

**Suggested action.** Narrow the sentence to the toy-table calls (`toyStmt`, `toyTables`). Then add a bullet that names the perly.y-tables tests in package `recognize` and gives each one a fate:
- **Equivalence tests** (`shared_test`, `TestPostdotIndex*`, `TestInternOneLookup`) follow production, judged on both sides.
- **The cost bounds** (`TestCompletionProbesBounded`, `TestCacheLookupBounded`, `TestExpectedMemoBounded`) are re-measured under the judged default. Two fates are possible:
  - pin them to the no-precedence constructor, which keeps their V1 meaning;
  - keep them on production, and treat a bound that no longer holds as a cost finding for the verdict reading, reported rather than retuned.

  The chain must say which.

## 5. What I did not re-examine

- **Whether the cost bounds actually fail under a judged recognizer.** I did not build one. AL17 holds whether they fail or not, because the claim that they "recognize exactly as before" is false either way.
- **Every one of the 82 `New(tab` call sites.** I classified them by the file's table source, not one by one.
- **Unchanged bodies.** Only d199, d218 and 3262 changed this round. I did not re-read the other 16 beyond the criterion re-run.
- **perl.git targets.** Their criteria are gated red today.
