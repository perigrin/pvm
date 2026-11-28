# crochet:review round 1: coverage (chart-perly-w4, V2)

Participant: **coverage**, ids C*. Lens: does the diff `e3432c0f..8d8fff38` cover the decision `docs/plans/2026-11-25-chart-precedence-v2-design.md` (rev 9)?

Everything below was read at 8d8fff38 in a scratch worktree (`wt-coverage`, removed at the end). The `/home/user/pvm` checkout itself sits at c383f861, 209 commits behind origin/pu, so the design, rulings and verdict files are not in its working tree. They had to be read at the commit.

**Result: no blocking findings.** Every requirement, acceptance item and §7 exit item in the design was delivered. I found no §7 stop. The two findings below are both RESIDUAL.

## Commands run (evidence base)

- `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit go test ./internal/chart/... -count=1 -timeout 1800s` at 8d8fff38: every package `ok`, exit 0 (log in `cov-gotest.log`). This covers the chart packages only, not the whole of `make test`.
- `go test -tags precedencereference ./internal/chart/cmd/ratchet/ -run '^(TestPrecedencePerlGitMatchesReference|TestSet1626NeverReachedPerlGit|TestForestEqualityPerlGit)$' -v`: all three PASS. Results: perlgit `compared=617`, `verdict only: 3`; `reaching set 1626: 0` with 5135 `indirob ::= scalar` completions; `compared 620 of 620 files`.
- `go test ./internal/chart/forest ./internal/chart/prectab -run '^(TestFeasibilityShiftsOwnNextTerminal|TestFeasibilityAllNullableRulesAllow|TestFeasibilityNarrowedMixedSets|TestPrectabItemMapping|TestContextCollapseTable)$' -v`: PASS. The logs match the README's numbers: 0 of 674; 0 of 29 (13 empty, 16 non-empty); 1880 / 63 / 1 (set 1626); 1212 core and 12 unmapped items; 46 kept, 0 collapsed by §4.3, 215 by §4.2.
- The verdict runs' answer keys compared with `git show e3432c0f:internal/chart/testdata/ratchet/baseline.txt`, filtered to the 27 answer keys: all six `run*.stdout` files are identical to the baseline. `speed-ratio-aggregate-longlived` runs: before 3.3716 / 3.3255 / 3.3345, after 3.0753 / 3.1492 / 3.0205. The after-median 3.0753 is below the before-minimum 3.3255. The `.meta` files name e3432c0f and 1477098790.
- `git diff --stat 1477098790 8d8fff38`: the verdict's after commit differs from 8d8fff38 only in docs, the README and a Makefile comment. No `.go` file changed, so the reading stands for 8d8fff38.
- `git diff e3432c0f 8d8fff38 -- internal/chart/testdata/ratchet/baseline.txt`: exactly one line removed, `speed-stage-precedence-aggregate-longlived 0.6833`.
- `git show e3432c0f:internal/chart/recognize/testdata/set-digest.golden | cmp - …/set-digest-noprecedence.golden`: identical (F3-4).
- `git diff --stat e3432c0f 8d8fff38 -- …/resolved-forest-digest.golden`: empty, so the golden is held.
- `git log` on the goldens: `67af3e99` touches only `forest-digest.golden` and `dfdc546e` touches only `set-digest.golden`, so each regeneration is its own commit. The brute-force reference is in `a082a556`.
- Every `Test*`, `Fuzz*` and `Benchmark*` name in `internal/chart/README.md` (326 names) has a `func` at 8d8fff38. Every backticked Go identifier in the README (436) occurs in `internal/**/*.go`. The one exception is `callFit`, which predates the diff and is outside V2.
- `git zhi milestone show chart-perly-w4 --format json | jq -r .body` is byte-identical to the earlier participant's `coverage-ms.md`. `git zhi issue list --milestone chart-perly-w4 --all` returns 28 issues, all `done`.

## 1. The design against what was built

| Design item | Delivered by | Status |
|---|---|---|
| §2 G: prune reductions at completion (B) | `recognize/recognize.go` `(*builder).complete`, `judge.go` `(*judgement).verdict` / `completes`, refusal records `refuse*` / `settled` / `keepRefused` | delivered |
| §2 G: prune shifts at the scan (C) | `(*builder).scan` with `(*Recognizer).scanRefused`; slipper (`slipper.go`) and `Expected` judged too | delivered |
| §2 G: `ParseWith` drops `Constrain`; `Constrain` and `ConstrainReference` become test-only | `forest/forest.go` `ParseWith` → `parseStream` → `builtStream` (no Constrain); `TestParseWithNeverConstrains`. The only Constrain calls left are in `forest/equality.go` (the V1 side and the infeasible toys), which only tests use | delivered. "Test-only" holds by use, not by file: `Constrain` and `equality.go`'s helpers still compile into package `forest` |
| §2 G: forest equal to V1's over the corpus and perl.git | `TestForestEqualityCorpus*`, `TestForestEqualityPerlGit` (620 of 620, rerun) | delivered |
| §2 G: speed improves | `verdict/README.md`, recomputed from the run files | delivered (exit met) |
| §2 non-goal: the pins keep their answers | `TestPrecedenceReferenceKnownDivergence`, `TestPrecedenceNullableShiftKnownLimit`: the diff only changes types and package (`prectab.`) | delivered |
| §2 anti-goals: keys, verdicts, forests | 27 keys equal the baseline in all six runs; `refused-where.txt` unchanged; differentials green | delivered |
| §2: `speed-stage-precedence-aggregate-longlived` retired | `cmd/ratchet/main.go` `keys`, `speed.go`, baseline line removed; `TestPipelineStagesHaveNoPrecedenceStage`, `TestCommittedBaselineMatchesKeys` | delivered |
| §4.1 verdict over the FOLLOW∩lookahead terminals; mixed sets kept | `judgement.verdict`; `TestReduceSideMixedVerdictSetKept`, `TestReduceSideIgnoresTerminalsOutsideFollow` | delivered |
| §4.1 the lookahead always exists; `NoLookahead` withholds it only from FOLLOW | `recognize.go` sets `b.la` whatever the options ("judged under every option"); `TestReduceSideNoFollowingTerminal`, `TestReduceSideNoLookaheadSameAcceptance` | delivered |
| §4.1 the six parent-dependent pairs | `judgement.parentItem`, `dep` path in `complete`; `TestLeoTakesNoParentDependentCompletion` | delivered |
| §4.1 infeasible grammars refused; their toys judged by §5 | `newJudgement` → `prectab.Feasible`; `TestJudgedRecognizerRefusesInfeasibleGrammar`, `TestParseRefusesInfeasibleGrammar`, `TestInfeasibleToysJudgedByConstrain` (names ladder, nonassoc, prefix) | delivered |
| §4.2 items (state, distance, context); the do/UNIOP/`my \` rule | `sets.go` `Item.Ctx`; `prectab/contexts.go` `PredictionContexts` / `rawTop`; `TestItemIsStateDistanceContext`, `TestContextCollapseTable`, `TestPredictorTakesNonNoneTop` | delivered |
| §4.3 collapse; pairing under context | Pairing: `predicts(...) != ctx` in `complete`, `TestCompletionPairsUnderContext`. Collapse: removed from production under perigrin's 11-27 ruling F1; kept as an oracle (`forbidsAt` in `contexts_test.go`) | delivered as ruled |
| §4.4 narrowing before recognition | `forest.narrow` / `Narrow`, `recognize/narrow.go`; `TestNarrowBeforeRecognition`, `TestNarrowKeptLibraryOperator`, `TestNarrowedSetIdsStableAcrossFiles` | delivered |
| §4.4 refusal records; Build reads them | `Chart.Refused` / `RuleRefused` / `RefusedAt`; `forest/bycontext.go`; `TestBuildNeverPairsRefusedParent`, `TestRefusalRecordsOutsideSetIdentity`, `TestRefusalRecordsServedByCache` | delivered |
| §4.4 diagnosis by rerun (grammar, then scope, then precedence; precedence at the chart's death byte) | `forest.ChartRefusal`; `TestRefusalRerunOrder`, `TestPrecedenceRefusalByteWhereChartDied`, `TestRefusalRerun*`; `refused-where.txt` unchanged; `pipelineReason` unchanged and already accepting `precedence` | delivered |
| §4.4 interning and cache carry the context and the records | `sets.go` hash includes `Ctx`; cache hit `keepRefused(pos, h.refused)`; `TestTransitionCacheKeysOnContext`, `TestContextCollapseInternsEqualSets` | delivered |
| §4.4 Leo: no skipped refusal; memo carries the context | `leo.go` `leoCrosses`, `leo_memo.go`; `TestLeoCrossesRefusedCompletion` (Leo on and off agree), `TestLeoMemoCarriesContext`, `TestLeoDisagreeZeroUnderPrecedence` | delivered |
| §4.4 `Expected` judged per candidate | `expected.go`; `TestExpectedJudgesPrecedence`, `TestExpectedMatchesRecognizeGivenEachCandidate`, `TestShiftRefusalThroughExpected` | delivered |
| §4.4 guards: no off-switch; reference, fuzz, digests and brute force judged | `ParseWith` panics unless `JudgesAsNew`; `33867dbd` (reference), `665703c5` (fuzz), `dfdc546e` and `67af3e99` (goldens, each alone), `a082a556` (brute force); `resolved-forest-digest.golden` unchanged | delivered |
| §4.5 referee narrows before recognition and judges the recognizer's forest | `internal/chart/precedence.go` `referee.lattice` / `judged` / `judge`; `TestForestEqualityRefereeCandidateNeverConstrains` | delivered |
| §5 forest equality (corpus, waypoint2, referee, seeds 4 and 44; accept/refuse, nodes, alternatives in order) | `ForestDiff` compares verdict, reachable keys, and `slices.Equal` on ordered alternatives; `TestForestEqualityCorpus`, `…Waypoint2`, `…RefereeSentences`, `…GeneratedSentences`, `…Toys` | delivered |
| §5 perl.git reference | `make chart-precedence-reference` (Makefile), CI PASS greps for all three tests (`ci.yml`); rerun green | delivered |
| §5 six counters = 0, and the lister lists exactly set 1626 | `forest/feasibility_test.go`: `TestFeasibilityRuleReduceConflicts`, `…ParentConflicts`, `…ContextConflicts`, `…ShiftsOwnNextTerminal`, `…AllNullableRulesAllow`, `…CoreItemsShiftAgree`, `…NarrowedMixedSets`; observers `TestSet1626NeverReached{Corpus,PerlGit}` | delivered |
| §5 Leo test | `TestLeoCrossesRefusedCompletion` | delivered |
| §5 existing differentials stay; infeasible toys judged by Constrain | `precedence_reference_test.go` (forest, chart, ratchet), `TestInfeasibleToysJudgedByConstrain` | delivered |
| §5 / §7 the 27 keys equal the retaken baseline | `verdict/run*.stdout` against `e3432c0f` `baseline.txt` (rerun) | delivered |
| §6 builder's choices | `Item.Ctx`, `crossKey`, `NewNoPrecedence` and `JudgesAsNew` | — |
| §7 exit: quiet interleaved B A B A B A, after-median below before-minimum, keys unchanged, differentials green | `docs/assessments/chart-perly-w4/verdict/` | met |
| §7 stops | none fired: forests equal; keys and verdicts unchanged; feasibility green; set 1626 reached 0 times (corpus and perl.git); the only item fact is `Ctx`. `RuleRefused` is a record beside the sets (the 3282 gate's note, which perigrin "may confirm" and has not ruled on) | no stop |
| §8 baseline retaken before V2 | e3432c0f is post-retake (#647); this diff removes one line only | delivered |
| §9 process | Rulings recorded on the day (11 `rulings.md` commits, dated); single-owner README rewrite (bb9b, `4fe66ded` / `767bd13c`); chain exported after refinement (`c7157c2e`) and after chain-review (`73f94605`). The close export is not yet made: **C1** | partly pending (C1) |

**Requirements not delivered:** none, apart from the close export in C1, which is not due until close.

## 2. Residual rows spot-check (sample of 16 of the 55 carried)

| Row | Carried by | Checked | Result |
|---|---|---|---|
| A13 | README docs issue | README l.111–115: the reduce side reads the lookahead, an open question for interleaved mode | delivered |
| A21 / F2-4 | README | README l.19–25 name the reversed clauses and rule 12's scope clause | delivered |
| A23 / F3-8 / F8-A1 | README | README l.10–11 name `origin/docs/chart-parser-v2-design` | delivered |
| A24 | README | README l.26–28: rule 2 is about dynamic context, and the spine top is a grammar fact | delivered |
| A25 | README | README l.3233–3237: the set-1626 claim holds on linear lattices; `TestForestLinearLatticesOnly` | delivered |
| A28 | d166 | `#666` merge body, "A28: what the rerun costs" (0.27 ms) | delivered |
| A29 | d199 | `TestReduceSideNoFollowingTerminal` exists | delivered |
| C5 | d0ba | `TestForestEqualityUnionShortcut` | delivered |
| C8 | d278 | `TestFuzzPrecedenceToyIsFeasibleWithLevels`; README l.2243, level-less grammars judge nothing | delivered |
| C11 / F3-9 | d1c2 | the merge body's "Measurements (C11, F3-9)" table: distinct sets, reuse, cache hits | delivered |
| C17 | d089 | `cmd/ratchet/testdata/set1626-inputs.txt` | delivered |
| F2-10 | d135 | `NewPrecedence` body never calls `Feasible` (awk/grep count 0) | delivered |
| F2-11 crumb | d166 | `TestRefusalEverySetAliveAtEnd` | delivered |
| F3-4 | d1c2 | `set-digest-noprecedence.golden` matches e3432c0f's `set-digest.golden` byte for byte (`cmp`) | delivered |
| F8-P1 | d247 / d0e1 | `TestForestEqualityRefereeIgnoresOnlyZeroWidthSlipperLeaves` | delivered |
| P18 | d166 | `ChartRefusal`'s rerun builds `recognize.NewNoPrecedence(...)` of its own; `TestRefusalRerunLeavesCallerRecognizerUntouched` | delivered |

Every sampled row was delivered. The uncarried rows A14 and P9 have recorded reasons. A14 is "left to the milestone postmortem", which the postmortem should pick up; it is already in the Records, so it is not raised here.

## 3. Live documents against the diff

### `CLAUDE.md`: unchanged by the diff, and it falsifies nothing

- Its chart-relevant claims are the test rules: pass 100% before committing, and `make test` before and after. The diff breaks these in practice in the golden-alone commits (one red commit before each golden) and in d1c2's head-only `make test`. Both are already in the Records: d278 gate F2, "perigrin to reconcile", and d1c2 gate F5. They are not re-raised here.
- Its line that non-doc Markdown is test data still holds for `internal/chart/README.md`: `forest/precedence_statefree_test.go` `TestPrecedenceShortcutStated` reads it, and the test passes.
- Its "Individual package testing" list names no chart package. That was true before the diff and stays true.

### `internal/chart/README.md`: rewritten whole by bb9b, with no claim falsified

- No claim I checked is falsified, and no enumeration is left incomplete.
- Every test name it cites exists, and so does every Go identifier the diff touched.
- The numbers I re-measured match: 674; 29 (13/16); 1880/63/1; 1212/12; 46/0/215; 620 of 620; 3.0753 below 3.3255.
- Its key-retirement history says the precedence stage key is retired with its baseline line, and `TestCommittedBaselineMatchesKeys` holds the baseline to the printed keys. The baseline diff is exactly that line.
- No sentence describes V2 work as still pending. Searches for `V2 will`, `pending`, `not yet`, `later issue` and `TestPrecedenceIgnoredByRecognition` turn up only intentional future-work statements: interleaved lexing (spec §13).
- One README-adjacent claim describes pending work, the chain export's (**C1**).

## 4. perigrin's rulings of 2026-11-26 and 2026-11-27

| Ruling | Carried out? | Evidence |
|---|---|---|
| 11-26 "Apply fixes, accept" (FR7-1, FR6-6, FR6-3) | yes | FR6-3: the corpus differentials go through `builtStream`, which narrows (`JudgedForest`, `CandidateForest` → `narrow`). FR7-1 and FR6-6 sit in issue bodies, and their outcomes are checked by the bb9b and d166 gates |
| 11-26 "Upgrade to v0.7.2" | yes, in the environment | Records; no repository artifact is expected |
| 11-26 "Fix name, drop EOF" (9e3a) | yes | the 9e3a body has criterion 3 as `TestFeasibilityAllNullableRulesAllow`, the EOF half dropped with the reason given; the test exists and logs `verdict-2 toy: 1 of 1` |
| 11-26 "Drop 2nd alternative" (d10d) | yes | the d10d body's criterion is `compared ([1-9][0-9]*) of \1( |$)` alone |
| 11-26 hooks: "Yes", "Add chart, then install", "Leave pre-push off", "Add to setup script" | yes where the repository can show it | `a29300ff ci: accept chart as a commit type` (`commitlint.config.js`); the rest is environment |
| 11-26 O1 "Amend d218 now" | yes | the d218 body says 8 of 13; `TestInfeasibleToysJudgedByConstrain` appends ladder, nonassoc and prefix |
| 11-26 O2 "New issue blocking d218" (d1e6) | yes | `prectab/semicolon_test.go` `TestSlipperSemicolon{Reduce,Shift}NeverForbidden` |
| 11-26 O4 "Measure before switch" (d35d) | yes | `TestCompletionProbesBoundedJudged`, `TestLeoMemoEntriesBoundedJudged`; d35d merged before 3262 |
| 11-26 "Fix it in d35d" and "Add it with the fix" | yes | `leoCrosses` keyed per (memo, `crossKey`); `TestJudgedLeoWalkLinear`, `TestLeoLinksPerTokenBoundedJudged`, `TestLeoCrossingVerdictPerLookahead` |
| 11-27 F1 "Delete it, keep as oracle" | yes | `PredictionContexts` has no §4.3 clause; `forbidsAt` is only in `prectab/contexts_test.go`; README l.3276–3279 states the granularity |
| 11-27 F2 "Add to 3282" (and F3 comments) | yes | `TestShiftRefusalThroughSlipper`, `TestShiftRefusalThroughExpected`; the `sets.go` comment says "sorted by state, distance and context", and the `leo_memo.go` comment says `{(0, 0, start)}` |
| 11-27 "Share one copy" | yes | `recognize/shared_test.go`: `alone := ownTables(tab)`, made once for every baseline |
| 11-27 "Pin it, amend 3262" | yes | `forest/readings_test.go` `builtReadings` uses `recognize.NewNoPrecedence` |
| 11-27 "Take a quiet reading now" | yes | `docs/assessments/chart-perly-w4/early-reading/`, with six runs |
| 11-27 "Profile, then new issue" and "Commit as a docs PR" | yes | early-reading README with medians, minimum and `exit: unmet`; issue 4792 blocked d2d1. The profile itself is reported in `rulings.md`, not committed, and the ruling did not ask for it to be |
| 11-27 "A, then re-read" | yes | `forest/bycontext.go`: no per-node map (a lazily made `more` map for rare extra pairs only); `keep` made once per walk; `leoAt` pre-sized `make([]leoFact, 0, n/4)` in `forest.go`; `early-rereading/README.md` `exit: met` |

Every ruling dated 2026-11-26 or 2026-11-27 was carried out as ruled.

## Findings

### C1: RESIDUAL. The chain export at close has not been made

- **Where:** `docs/assessments/chart-perly-w4/chain/README.md` (final sentence: "They are exported again at close."); design §9, first bullet.
- **Evidence:** `git log e3432c0f..8d8fff38 -- docs/assessments/chart-perly-w4/chain` lists two exports only: `c7157c2e`, after refinement, and `73f94605`, after chain-review. `chain/milestone.json` predates every execute Record.
- **Grounds:** §9 requires three exports. The third is due at close, so this is pending work a live document says is pending, not a defect in the delivery. The chain cannot be pushed (`refs/zhi` gets a proxy 403), so this export is the only durable copy of the execute Records and the gate history.
- **Remedy:** make the close export (milestone.json, issues.json, a README update) as a docs PR after review and before or with the postmortem.

### C2: RESIDUAL. The delivery range includes an unrelated non-chart PR

- **Where:** `internal/version/github.go`, `internal/version/ratelimit_test.go` (`01d467cb`, merged as #649).
- **Evidence:** `git diff --stat e3432c0f 8d8fff38` lists both files. `git log --merges` shows `d59dd367 Merge detect go test on Windows… (#649)`. The commit has no `Implements:` trailer.
- **Grounds:** the brief describes the delivery as PRs #655–#691, but the range also carries #649, a Windows rate-limit fix outside V2. It does not touch V2's coverage. It only means "210 files" overstates the milestone's footprint, and the README's isolation rule (§11) is not engaged, because #649 is not a chart issue.
- **Remedy:** none for the code. The postmortem should state the delivery as the trailer set, or as the range less #649.

## Classification count

BLOCKING: 0. RESIDUAL: 2.
