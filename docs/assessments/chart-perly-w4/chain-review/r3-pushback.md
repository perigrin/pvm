# chain-review round 3: pushback (plan quality) on chart-perly-w4

## 1. Recommendation

**Modify.** The re-plan is the right shape:
- the default recognizer stays unjudged through d199, d1c2 and 3282;
- 3262 is the single switch-over;
- the V1 side's recognizer is fixed through one function.

It settles PB6, N1 and PB9 at the root, not test by test.

Three blocking findings remain, and each is a small fix:
- **N5:** d199's new criterion states a transient fact. It goes red at 3262 and stays red at milestone verify.
- **N6:** 3262 retires `TestRecognitionKeysUnchangedByPipeline` on a premise the tree contradicts.
- **FR2-8**, promoted: d2d1 cannot close as written without a rerun that its own text forbids.

How this round was checked:
- I read the live chain first: 19 issues. I diffed every body against my round-2 copies; 11 changed.
- I ran `internal/chart/offline/criteria-lint.sh` over all 19 issues. It exits 0, so every non-`guard:` criterion is red at `e3432c0f`.
- I grepped every criterion for transient wording.
- I inventoried the 103 `recognize.New` test call sites by file, and read the four tests from PB6 and the ratchet fixtures.
- Only then did I read `repair-r2.md`.

## 2. Status of my findings

| id | Status | Grounds |
|---|---|---|
| PB1 | **released** | Re-run, with the new order in mind: `set-digest-noprecedence.golden` is still copied in d1c2. `TestSetDigestGolden` runs unjudged through d1c2 and 3282, because `New` is unjudged. 3262 pins `setdigest_test.go` to the no-precedence constructor, and d278 flips it and regenerates the golden. d1c2's criterion holds at verify. |
| PB2 | **released** | The d1eb → d278 edge is still present (d1eb blocks d2d1 and d278). |
| PB3 | **released** | d218 still scopes itself to the production differentials. d247 still owns the hook-counted referee check. d247 now also says it moves `referee.judge` off the no-precedence pin that 3262 gave it, which is consistent with 3262's list. |
| PB4 | **released** | Unchanged since round 2. N2 stays residual. |
| PB5 | **released** | 3262 keeps the full rule: grammar, then scope, then precedence at the death byte. It keeps the `refused-where.txt` byte-identity criterion, and its fixtures are still Go fixtures. |
| PB6 | **released** | All four tests I named now have a fate in 3262. `TestParseAppliesPrecedence`, `TestRatchetVerdictIsPipeline` and `TestPipelineStageKeysRefusedFiles` are rewritten. `TestWaypoint2RefusalsMatchReference` is pinned. One criterion runs all four beside `TestDefaultRecognizerJudgesPrecedence`. The root is gone: no issue before 3262 changes `New`, so no guard moves early. On completeness: the inventory step flips `New` uncommitted, runs `make test`, and escalates any failure not on the list. That is the right control for a list built by reading, and it is the repairer's flagged call. N7 notes one class of test it will escalate needlessly. |
| PB7 | **released** | Unchanged since round 2. |
| PB8 | residual, held | There is still no criterion that the brute force runs judged. 3262 now pins `toyForest`, which carries the brute force, to the no-precedence constructor. So d278 is the only place it is flipped, and nothing checks that it was. I still class this residual: it is a missing check, not a wrong one. |
| PB9 | **released** | d166 now says that `terminals-per-token` is an answer key, counted on the un-narrowed `lex.Lex` tokens, and that a narrowed set id never indexes `tab.TerminalSets`. `TestTerminalsPerTokenCountsUnnarrowedTokens` (subtests `dead-plugin-reading` and `corpus`) is a criterion that is red today and needs no perl.git. d166 and 3262 each also have a step that diffs the 27 keys where perl.git is present. That covers the two issues where a key can move. |
| PB10, PB11 | **released** | The edges are unchanged since round 2, and stage order is enforced. |
| PB12 | residual | The critical chain is still 15 issues: d02f→d05d→d089→d0ba→d0e1→d135→d166→d199→d1c2→3282→3262→d218→d247→d2d1→bb9b. An equal-length path runs 3262→d1eb→d278→d2d1. |
| PB13–PB20 | residual | Unchanged. PB16 was already released in round 2. |
| N1 | **released** | The interim `scope` state is gone. `New` judges first at 3262, and the precedence branch of the rerun rule lands in the same issue. d199 no longer touches `parseStream`, `recognizeSource` or production. |
| N2, N3, N4 | residual | Unchanged. |

## 3. Promotions

| Row | To | Grounds |
|---|---|---|
| **FR2-8** (fresh reader) | **blocking** | Followed literally, d2d1 cannot close. Its criterion #4 requires `run*-before.meta` and `run*-after.meta` files carrying `commit:` and `exit: 0`. Its Steps say only "keep stdout and stderr per run", and its own negative scenario says "Do not rerun until the numbers pass, and do not edit a run's output". A runner who follows the steps gets no meta files. The only ways to green the criterion are a second quiet session, which looks like a rerun, or meta written by hand after the fact, which is the fabrication PB14 worries about. I re-read d2d1: the only mention of `.meta` is criterion #4 (line 37). **Fix:** the `drive.sh` step names the meta file and its two lines, and `drive.sh` writes them. |
| AL14 | not promoted | `carrying zero-width ; leaf: 1296` has an undefined count. Alignment measured 0 reachable and 1,277 present on the V1 Build path, while assess counted 1,296 insertions in the narrowed recognizer, which is d247's own state. The number may well be right for what d247 builds. It is a definition to tighten, so I would stand aside on it alone. |
| AL5–AL11, FR2-5, FR2-6, FR2-7 | not promoted | FR2-5 is answered by PB9's fix. FR2-6, d199's size, is answered: d199 now holds only the reduce side, records, Build and Leo, and adds the constructor. FR2-7 is answered: d05d creates `scancollideToy`, and the feasibility tests stay in forest. The AL rows are unchanged and none makes proceeding wrong. |

## 4. New findings

| id | Class | Issue | Claim | Consequence | Evidence | Suggested action |
|---|---|---|---|---|---|---|
| N5 | **blocking** | d199 (with 3262) | Criterion: "The default stays unjudged until the switch-over: on the perly.y sample `recognize.New`'s sets equal the no-precedence constructor's, the judged constructor's differ from them in at least one set, and the differential's V1 side reports [the no-precedence constructor]" (`TestJudgedRecognizerIsExplicit`) | A transient fact has been written as a criterion. Verify runs every criterion at milestone completion. From 3262 on, `New` is the judged constructor, so the first clause is false. The test fails at 3262's close, and the d199 criterion is red at verify. 3262 does not list the test among its rewrites or retirements, so its inventory step escalates it as an unnamed failure. Fixing it then means changing a criterion, which is a §7 stop. This is the shape SQE #8 warned about. | Of the 19 bodies, only d199 contains `TestJudgedRecognizerIsExplicit`. 3262 lists no fate for it. | Keep the permanent clauses as the criterion: the V1 side's recognizer comes from the no-precedence constructor, and the judged constructor's sets differ from the no-precedence constructor's on the sample. Keep "New stays unjudged until 3262" in d199's prose, or make that clause a separate test that 3262 rewrites to "`New` equals the judged constructor", and list it in 3262's Context. |
| N6 | **blocking** | 3262 | Context: retire `TestRecognitionKeysUnchangedByPipeline` because "once narrowing and the shift side are in, no file is both accepted by the judged recognizer and refused by the pipeline, so it has no rewrite" | The premise is false. A scope refusal made by resolve is accepted by any recognizer and refused by the pipeline. `TestRatchetVerdictIsPipeline`'s own fixture `scope.pl` (`use feature 'class'; no warnings; method foo {}`) must be recognized and not accepted, and d166's narrowing touches only `PLUGIN_*` library operators, not `method`. Retiring the test drops the only guard that `leo-disagree`, `items-per-token`, `set-reuse` and `lex-unmapped` stay recognition's keys whatever the pipeline does. That property still holds under V2, and `verdicts.go:87` still relies on it. | `cmd/ratchet/verdicts_test.go:136-151` (scope.pl, `!r.Recognized` is an error); `verdicts_test.go:186-221` (the retired test's body, keyed on one fixture) | Rewrite it over `scope.pl`, or any resolve-scope fixture, instead of retiring it. Then 3262's criterion becomes "passes without skipping, over a scope fixture", and the README claim stays. |
| N7 | residual | 3262 | "Not expected to move. These are the recognizer-internal tests over toy grammars without levels … If one of them moves, a forest differs from V1's." | Several tests on the unchanged list run on the committed perly.y tables and assert cost bounds, not forests: `cache_test.go` (`TestCacheLookupBounded`, `TestExpectedMemoBounded`), `probes_test.go` (`TestCompletionProbesBounded`, at most 0.64 probes per token), and `shared_test.go`. Once `New` judges, items carry contexts and the sets can multiply (C11 measures exactly this), so a bound can move with no forest difference. The inventory would escalate it as a forest stop, which is the wrong diagnosis. | `recognize/cache_test.go:19,54,130`; `recognize/probes_test.go:22,28` | Add a fate: "cost bounds over perly.y. A move is a cost finding for the verdict, reported, not a forest stop". Or have d1c2's C11 measurement cover these three bounds on the judged constructor. |
| N8 | residual | 3262 | One issue holds the inventory, about 20 call-site pins, the switch, the full rerun rule, both readers, 3 rewrites and 2 retirements | It has 9 criteria but a large mechanical surface. The pins change nothing while `New` is unjudged (its own step says "run make test with New still unjudged and see it pass unchanged"). So they could land earlier, in d199 beside the constructors, and leave 3262 to the switch and its consequences. | 3262 Steps | Optionally move "pin the listed call sites" into d199. |
| N9 | residual | d199 | The repairer's flag: `TestJudgedRecognizerIsExplicit` assumes the judged reduce side refuses at least one completion on the perly.y sample | I judge it low risk. Any sample statement mixing additive and multiplicative operators, such as `$a + $b * $c`, has a completion of the additive rule before `*` whose verdict forbids it on a single-terminal lookahead set. I did not measure it. d278's "refused completions: N" on the same sample carries the same assumption. | — | Measure it once at d199. If it is 0, that is a fact to report, not a reason to change the criterion. |

## 5. What I did not re-examine

- I did not run `make test` with `New` flipped. N7's list comes from reading, as 3262's does.
- I did not re-verify the 1,296 referee count (AL12) or AL14's definition. Those belong to alignment.
- I did not re-read d02f, d089, d1eb, d218, d278, d2a2, d2d1 or bb9b beyond confirming they are byte-identical to round 2 (d2d1 apart from FR2-8).
- I did not re-check every residual-mapping row. I read the milestone diff: A16, A19, A29, C10, C12, C13, F3-6, F3-7, F5-3 and F6-R2 now point to the right halves, and no carried id was lost. d199's carries (A16, A29, F3-6, C13 reduce, F6-R2) and 3262's (F5-3, C12, F3-7 switch, C10 a) together cover what the pair held in round 2.
- I did not run the perl.git targets, the fuzz target or `make chart-ratchet`.
