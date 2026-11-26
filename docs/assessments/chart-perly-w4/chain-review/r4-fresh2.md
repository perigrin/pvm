# chain-review round 4, review-fresh2: chart-perly-w4

I read the chain live from `/home/user/pvm/.claude/worktrees/v2-refine`: 19 issues, with edges unchanged since round 3. I diffed every body against the round-3 snapshot (`f2r4-diff.txt`) and ran my checks before I read `repair-r3.md`.

## 1. Recommendation

**Accept.** None of my findings is still blocking, I promote nothing, and my one new finding is residual.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| FR2-1 to FR2-7 | **released** (round 3) | They are unchanged since round 3. I re-checked that the 3262 criterion running `TestWaypoint2RefusalsMatchReference` is still there. I also confirmed that `referee.narrowed`, `TestPrecedenceKeepsCorpusAccepted` and the `constrained` helper all reach the recognizer through `builtForest` (`precedence.go:245`), so pinning `referee.judge` at 3262 covers them too. |
| FR2-8 | **released** | d2d1's `paths:` now lists `runN-{before,after}.{stdout,stderr,meta}`. Its `drive.sh` step says `drive.sh` writes each `.meta` in w3's form, including `commit:` (the full hash) and `exit:`. The criterion's `^commit: e3432c0f` prefix match and its `exit: 0` exact match both fit that form. |
| FR3-1 | **released** | **The fix:** d199's `TestJudgedRecognizerIsExplicit` criterion now asserts only two things. The judged constructor's sets differ from the no-precedence constructor's on the perly.y sample, and the V1 side's recognizer comes from the no-precedence constructor. Both stay true after 3262. The transient "`New` equals no-precedence" clause is gone. The step says the test must survive the switch-over.<br>**What else it would have to change:** I looked for any other criterion asserting that `New` is unjudged. I grepped all 191 criterion lines, and every unjudged/no-precedence assertion is a durable one: d0ba V1 side, d199 infeasible grammar, d1eb infeasible Expected, d278 reference.<br>**Still red today:** the test does not exist yet. |
| FR3-2 | **held, residual** | 3262 has grown further. It now also carries the `ParseWith` guard with its TDD cycle and two more rewrites (`TestRecognitionKeysUnchangedByPipeline`, `TestForestGrammarDataOncePerTables`). The point stands: the pinning commit changes no behaviour and could close on its own earlier, as N8 also says. It holds no round open. |
| AL14 (my round-3 promotion) | **released** | **The contradiction is gone:** d0e1 and d247 no longer say "of 1,299". Both give 1,296 kept sentences, which I confirmed today: `TestPrecedenceRefereeMatchesReference` logs `sentences=1296 compared=1296 refused=19 pruned=1277`.<br>**The count is defined:** "carries a leaf" means a slipper-inserted zero-width `;` in the narrowed-lattice chart. The criterion is now `slipper ; insertions: [1-9][0-9]* of 1296`, a defined non-zero count rather than a literal that only one unstated definition yields, and it keeps `ambiguous: 0` and `cut: 0`.<br>**The rest of the fix landed:** d247 now owns the rewrite of `TestPrecedenceRefereeMatchesReference`. Its criterion `verdict only: [1-9]` agrees with the 19 refused sentences in that same run. |

## 3. Promotions

None. I read every residual row in r1-* to r3-*.

- **r3-fresh3 FR3-7** (3282's "over the toys" without the infeasible-toy exception that d218 states) comes closest. `emptyred` and `empty` are in `toyNames`, so a literal reading puts toys in the judged-recognizer differential that no recognizer judges. But the criterion only greps a test name the builder writes, and design §5 states the exception. A builder who misses it reports a stop rather than building the wrong thing. I would add one clause to 3282, but I do not hold it.
- **N7 / r3-fresh3 FR3-3** (cost bounds over perly.y on 3262's "not expected to move" list): 3262 still says "If one of them moves, a forest differs from V1's", which is untrue for a probe-count bound. The systemic first step now routes any such red test to the orchestrator before anything is edited, so the false sentence cannot produce a silent wrong fix. Residual.
- **FR3-4, FR3-5, AL15, AL16, N8, N9 and PB8 to PB20:** each is residual on its own terms.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR4-1 | residual | 01a21447-d218 (with 01a21463-3262) | d218: "repoint the existing production-versus-reference differentials at the recognizer's forest", following design §5's "'Production' now means the recognizer's forest". Its "Tests this change moves" names them only as a class. | **3262 pins every one of them to the no-precedence constructor:**<br>• `forest/precedence_reference_test.go`<br>• `cmd/ratchet/precedence_reference_test.go`<br>• `precedence_perlgit_test.go`<br>• `TestWaypoint2RefusalsMatchReference`<br>**None calls `Parse`.** So d218's systemic step, which removes `Constrain` from `parseStream`, turns none of them red, and the step surfaces nothing.<br>**Nothing names or checks the repoint:** a worker can close d218 with them still comparing a V1 forest with `ConstrainReference`. That keeps them green and useful as a `Constrain` referee check, but it does not meet design §5's sense of "production". d247 did this properly for the referee test: it named the test, stated verdict-only handling for refused inputs, and added a criterion. d218 does none of that for the rest.<br>**The forest-equality differentials** (d0ba family) carry the real V2-against-V1 evidence, which is why this is residual. | `grep ParseWith\|Parse(` over those files: no call. d218's Context and Steps. 3262's pinned list. | Name the tests d218 repoints. For each, say what happens on an input the judged chart refuses, either verdict-only as in d247 or skipped with a count, and add one criterion that runs them. Or record that they stay `Constrain` referees on the no-precedence mode, and say so in the design's terms. |

## 5. What I did not re-examine

- I did not run any criterion other than re-running `TestPrecedenceRefereeMatchesReference` and reading its log.
- I did not check `ParseWith`'s new panic against the speed harness's recognizer construction in `speed.go`. 3262 lists the speed harness as "follows production", so it should hold a judged recognizer.
- I did not check whether `scope.pl` is refused with reason `scope`, which the rewritten `TestRecognitionKeysUnchangedByPipeline` criterion's log line needs, as opposed to some other resolve reason. `TestRatchetVerdictIsPipeline` labels it scope for perl but asserts only refusal.
- I did not re-read d02f to d10d, d135, d1c2, 3282, d1eb, d278, d2a2 or bb9b. They are byte-identical to round 3.
- I did not check the milestone's residual mapping beyond the C10 and F3-7 rows the repair touched, plus a diff of the body, which shows only those rows and a new Records entry changed.
