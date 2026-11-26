# chain-review round 7: fresh3, as a participant (chart-perly-w4)

## 1. Recommendation

**Accept, with residuals.** The round-6 repair settled my last held finding, r3-fresh3 FR3-3. 3262 now names `TestCompletionProbesBounded` and the other cost bounds, with a fate that does not misread a cost move as "a forest differs". I found nothing the repair broke, and I have no promotion and no new blocking finding.

**Method.**

- I read the live chain at `e3432c0f` first: 19 issues, with edges and labels unchanged.
- I diffed every body and the milestone against my round-6 snapshot. Four issues changed (d199, d218, d278 and 3262), plus Records.
- I checked every test 3262 now names against the tree. All 15 exist: `cache_test`, `recognize_test`, `probes_test`, `shared_test`, `slipper_test`, `leo_test` and `intern_test`.
- I listed every test in `recognize` and `forest` that runs the recognizer over perly.y's committed tables, its sample or the corpus. I compared that list with 3262's three new groups and its pinned and rewritten lists. Everything is covered except a few tests in `forest_test.go` and `readings_test.go`, which 3262's paths and "Follow production" cover. The only bound among those, `TestForestRightRecursionLinear`'s 2.5× growth ratio, is not something judging can break.
- I checked d218's claim of "13 toys, less `empty` and `emptyred`" against `toyNames` (`forest/precedence_test.go:23`). It matches, and the minute (`chart-precedence-v2.md:832`) records that the other eleven pass every counter.
- After that I read `repair-r6.md` and every row from r6-*.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| r3-fresh3 FR3-1, FR3-2, FR3-4, FR3-6, FR3-7 | released (earlier rounds) | Unchanged this round. |
| r3-fresh3 FR3-3 | **released** | 3262's list is split three ways. (a) Level-less toys, covered by d199's `TestJudgedOverLevellessTablesRecognizesAsBefore`. (b) Perly.y-backed tests that judge from 3262 on, each named, including `TestRecognizersShareGrammarTables`, which I raised alongside the probe bound. (c) Named cost bounds, which are `TestCompletionProbesBounded` (0.6362 against 0.64, the number I measured), `TestLeoMemoEntriesBounded` (0.0395 against 0.040), `TestCacheLookupBounded`, `TestExpectedMemoBounded` and `TestInternOneLookup`. A red bound "is reported to perigrin with the measured number, and not changed without his ruling", and it is "a cost finding, not 'a forest differs'". That was my complaint: the old text routed a cost move to a false §7 stop. 3262 can still pause on a bound, but it now pauses with the right diagnosis and a named owner, which is the right outcome for a performance bound. |
| r3-fresh3 FR3-5, r4-fresh3 FR4-1, r5-fresh3 FR5-1 | released (earlier rounds) | Rechecked against the new `ParseWith` guard. A failed derivation on perly.y's tables, as in `TestPrecedenceGenerationErrorRefusesEveryFile`, is still what `New(tab)` would derive, so the guard accepts it and `parseStream` turns the error into a refusal. `TestParseRefusesInfeasibleGrammar` now runs through `parseStream` inside `forest`, so it does not go through the guard. |
| r6-fresh3 FR6-1 | **held, residual** | The repair kept this case deliberately. 3262's guard now says that "for tables whose rules are not perly.y's, the explicit form stays allowed, because those tables have no derived levels to bypass". So a grammar that has levels but differs from perly.y by one rule is still parsed without judgement, where V1 refused it loudly. No production caller builds such tables, since `ParseWith`'s only caller outside `forest` is `TestPipelineStagesReproduceParseWith`, over the committed tables. Four other participants hold the same point as residual: r6-fresh2 FR6-1, r6-fresh4 FR6-2, r6-fresh6 FR6-7 and r6-pushback N15. |

## 3. Promotions

None. I read every residual row from r1 to r6. The ones I weighed this round:

- **r6-fresh6 FR6-3.** Neither d199's corpus differential nor 3282's says that the candidate's lattice is narrowed. Design §4.1 says that over un-narrowed input the recognizer "keeps completions V1 drops", and §3 counts 1,110 tokens carrying `PLUGIN_*` readings. So an un-narrowed candidate would differ from V1, and that would look like a §7 stop. I keep it residual: design §4.1 states that "forest equality is claimed for narrowed input, which is what production gives it", and from d166 onward production narrows. A worker who follows the design builds the narrowed candidate, so the gap is one of wording, not plan. One clause in each issue would close it.
- **r6-fresh6 FR6-4, FR6-5, FR6-6, FR6-8, FR6-9:** classification wording, path lists, issue size and naming of the hook. None of them leaves a state that cannot close green.
- **r6 FR6-1, FR6-2, FR6-7, N15:** the same point as my FR6-1, held residual above.

## 4. New findings

None.

| id | blocking/residual | issue id | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| (none) | | | | | | |

## 5. Not re-examined

- I ran no V2 code. The behaviour of the `ParseWith` guard and of the explicit and derived forms is reasoned from the chain and the tree.
- I did not re-run `criteria-lint.sh`. The new and changed criteria (`TestBruteForceJudgedOverLeveledToy` and the two new `TestParseHasNoPrecedenceSwitch` subtests) name tests or subtests that do not exist today, so they are red.
- I did not check d278's brute-force toy choice beyond its criterion. It reuses the fuzz oracle's feasible toy, which d278 also creates, and it requires a non-zero refusal count.
- I did not re-examine the 15 issues left unchanged this round beyond confirming they are byte-identical to my round-6 snapshot.
