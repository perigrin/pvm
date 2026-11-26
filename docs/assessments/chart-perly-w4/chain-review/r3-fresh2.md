# chain-review round 3, review-fresh2: chart-perly-w4

I read the chain live from `/home/user/pvm/.claude/worktrees/v2-refine` (19 issues) and diffed every body against the round-2 snapshot. I ran my checks before I read `repair-r2.md`.

## 1. Recommendation

**Modify.** The switch-over re-plan answers all of FR2-1 to FR2-4 cleanly. It introduces one new blocking defect, FR3-1, and I promote one residual row, AL14. Both fixes are small.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| FR2-1 | **released** | **Before the switch-over:** `recognize.New` stays unjudged through d199, d1c2 and 3282 (d199, d1c2 and 3282 prose), so `TestWaypoint2RefusalsMatchReference` keeps passing.<br>**At 3262:** the test is in the "pinned to the no-precedence constructor" list. 3262's last negative criterion runs it beside `TestDefaultRecognizerJudgesPrecedence`. The test is self-checking: run on the judged recognizer, it errors "recognition leaves no forest".<br>**The inventory question:** 3262 now owns every recognizer call site, with a `make test` inventory step that stops on any failure it does not list. |
| FR2-2 | **released** | **Ordering:** the edge is now 3282 → 3262, so all three fixtures die in the chart at 3262. 3262 names the side that kills each one: `print $fh [];` dies on the reduce side, and the NCEQOP and DOTDOT chains die on the shift side.<br>**The unowned test:** `TestParseAppliesPrecedence` is now in 3262's "rewritten" list, and a criterion runs it.<br>**The interim `scope` mislabel is gone:** the full rerun rule lands with the switch. |
| FR2-3 | **released** | d199's false sentence is gone. `TestRecognitionKeysUnchangedByPipeline` is retired in 3262 with its README claim, and the rewrite option is dropped with the reason given. The criterion `! grep func … && ! grep README` is red today. |
| FR2-4 | **released** | **One function:** the V1 side gets its recognizer from one function in `equality.go` (d0ba), which d199 points at the explicit no-precedence constructor. d0e1 and d10d reuse it.<br>**The guard:** the new d0ba criterion `TestForestEqualityV1SideRecognizesPrecedenceRefusals` requires the V1 chart to accept all three fixtures and `Constrain` to refuse each. A V1 side on the judged recognizer fails it from 3262 onward, so the vacuity can no longer pass silently. |
| FR2-5 | **released** | d166 now states that `terminals-per-token` is an answer key summed over the un-narrowed tokens. `TestTerminalsPerTokenCountsUnnarrowedTokens` guards it, written before the caller changes. d166 and 3262 each run the resolution command where perl.git is present and treat a moved key as a stop. A smaller residual I am not raising: d218 and d247 also change production or referee paths, and they rely on the verdict reading for the keys, though d247's criterion does check `precedence-disagree` stays 0. |
| FR2-6 | **released** | d199 shrank. The pins and retirements moved out, and it no longer touches production, `parseStream` or `recognizeSource`. The weight moved to 3262: see FR3-2. |
| FR2-7 | **released** | d05d creates `scancollideToy` from F6-C2's grammar, rules and levels spelled out. d135 keeps the stage-1 tests in package `forest`, calling `prectab.Feasible`, and its ≥8 criterion was re-pointed and is red today. |
| FR2-8 | **held, residual** | d2d1 is byte-identical to round 2. The criteria still require `run*-{before,after}.meta` files with `commit:` and `exit:` lines that neither Context nor Steps names. It holds no round open. |

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| AL14 (alignment, d247; also d0e1) | residual | **blocking** | **Contradiction:** d247's prose reads "1,296 of 1,299 sentences carry one", and so does d0e1's. d0e1's own criterion, fixed by AL12, now says the referee has 1,296 kept sentences in all, which I confirmed by running `TestPrecedenceReferee`: "1296 kept sentences agree with bison". Two parts of the subject contradict each other, which is a defect in what it asserts.<br>**Undefined count:** d247's criterion then pins `carrying zero-width ; leaf: 1296` with "carry" undefined. AL14 measured 0 reachable leaves and 1,277 present on V1's Build path. The only reading that yields 1,296 is assess-codebase's "slipper insertions in the narrowed run". The text never says that. So a builder makes the criterion pass by choosing the definition that matches the literal, which §7 forbids as tuning. The alternative is reporting a stop on a correct change.<br>**Fix:** correct "of 1,299" in d0e1 and d247, and define the count, or assert that every kept sentence is compared and that none becomes ambiguous or cut. |

I read every other residual row in r1-* and r2-* and promote none of them:
- **AL5 to AL11:** true, and none makes proceeding wrong.
- **PB8 (brute force):** now sharper, because 3262 pins `toyForest`, which carries the brute force, and d278 must flip it back with no criterion. But it is still a missing criterion rather than a wrong one.
- **PB12 to PB15, PB18 to PB20, N2, N3:** unchanged, and residual on their own terms.
- **N4:** moot.
- **PB17:** largely settled by choosing the mode by constructor.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR3-1 | **blocking** | d199 (01a21447-d199), 3262 (01a21463-3262) | d199's criterion `TestJudgedRecognizerIsExplicit` asserts "on the perly.y sample `recognize.New`'s sets equal the no-precedence constructor's". | **The criterion is true only until 3262**, whose whole purpose is to make `recognize.New` judge.<br>**3262 does not list it:** the test appears in none of 3262's fates (pinned, follows production, rewritten, retired, not expected to move).<br>**The inventory step stops on it:** flipping `New` makes this test fail. A failure not named in Context "is reported to perigrin before any test is edited", so the switch-over is certain to stop on the chain's own test.<br>**Milestone verify goes red:** if the test is instead edited at 3262, d199's criterion, which `git zhi verify` re-runs over done issues, goes red. | `grep TestJudgedRecognizerIsExplicit` over the chain and milestone finds only d199's criterion line. 3262 Context, "What happens to each existing recognizer call site". | Drop the transient clause from d199's test and criterion: keep "the judged constructor's sets differ" and "the V1 side uses the no-precedence constructor". Or name the test in 3262's "rewritten" list, with its new assertion that `New` equals the judged constructor, and key d199's criterion only on the parts that stay true. |
| FR3-2 | residual | 3262 (01a21463-3262) | One issue now holds the inventory, about 12 pinned call sites, the default flip, the full three-reason rerun rule, the readers' agreement, three rewrites, two retirements and an answer-key run. That is 10 criteria and about 6 commits. | This is the largest issue in the chain. Round 1 split the rerun rule out of d199 for size (PB6), and this re-plan puts it back beside the flip. The pinning is mechanical, and keeping the switch in one issue is the point of the re-plan, so I would not split the flip. | 3262 Steps and criteria. | Consider moving the pinning commit, "pin the listed call sites; make test passes with New unjudged", into its own issue just before 3262. It changes no behaviour and closes green on its own. |

### On the repairer's two flags

- **3262's call-site list:** my FR3-1 is exactly one unlisted failure the inventory would hit. I found no other through criterion text. I did not re-walk the 25 call sites.
- **"The judged reduce side refuses ≥1 completion on the perly.y sample":** this is unmeasured, but I judge it safe. `std5k.pl` holds dozens of `= … ||`, `&&` and `or` lines (55 matches for `' \|\| | && | or | and '`). In each, completing the assignment before the lower-precedence operator is forbidden, so the outer item is never advanced and the sets differ. This is reasoning, not a run.

### Re-derived and found sound

- **Edges:** d199 → d1c2 → 3282 → 3262 → {d1eb, d218}. The stage order is intact.
- **The mapping:** C10 (a) and F3-7 (the switch half) moved to 3262 in both the `Carries:` line and the milestone rows. No carried id was lost.
- **Named tests exist:** `TestRatchetVerdictIsPipeline` (`verdicts_test.go:132`) and `TestPipelineStageKeysRefusedFiles` (`stages_test.go:144`).

## 5. What I did not re-examine

- I did not run any criterion. I did not run 3262's inventory or walk the 25 call sites again.
- d02f, d089, d1eb, d218, d278, d2a2, d2d1 and bb9b are unchanged since round 2, apart from title references. I did not re-read them beyond checking that none asserts `recognize.New` is unjudged.
- I did not check the number behind d0ba's new subtests or d166's corpus subtest.
- I did not check whether the set-1626 observers (d089) or the equality candidates silently change meaning when `New` starts judging at 3262. My reasoning says set 1626's mixed rule keeps the completion, but I did not run it.
