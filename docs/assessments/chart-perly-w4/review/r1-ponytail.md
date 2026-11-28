# crochet:review round 1: ponytail (prefix P)

Lens: is every line needed, and is anything over-engineered?

The `ponytail:ponytail-review` skill loaded and ran here. Its approach is applied below: a concrete case for each finding, the smallest fix, and deletion preferred.

**Subject:** `git diff e3432c0f 8d8fff38`, production Go under `internal/chart/`. I read it in a scratch worktree at 8d8fff38 (`wt-ponytail`). `/home/user/pvm` itself is checked out at an older commit (c383f861), so I did not read the code there.

**What this change does:** V2 moves bison's precedence out of a pass after the forest (`forest.Constrain`) and into the recognizer. The recognizer now refuses forbidden reductions (`judge.go`) and forbidden shifts (`scanRefused`), and each item carries one context. `forest.Build` walks the forest by context (`bycontext.go`). Library operators narrow the lattice before recognition (`narrow`, `recognize.SetID`) instead of pruning the forest after it. `prectab` is the precedence tables and the feasibility counters, split out of `forest`.

**Result: no blocking findings.** Everything below is RESIDUAL. Most of it is reference and test scaffolding that sits in production files. It is not wrong, and nothing calls it in a way that could hurt.

## Method (evidence common to the findings)

- **Exported symbols with no production caller.** For every exported symbol in a changed non-test file, I counted the files that use it outside its own file, split into production and test files (a grep loop). The no-production-caller list is quoted in each finding.
- **Unused code.** `staticcheck -checks U1000 -tests=false ./internal/chart/...` flags the code that only tests reach.
- **Mutations.** I re-ran the two mutations the brief names against `go test ./internal/chart/...`, with `T1_CORPUS` set explicitly. The results are in the "Known items" section.

## Findings

### P1: RESIDUAL. `forest/equality.go` is 265 lines of test-only exported API in a production file, and part of it is duplicated or left over from a finished stage

- **Where:** `internal/chart/forest/equality.go`: `ReduceSideForest`, `ReduceSideToyForest`, `JudgedForest`, `JudgedToyForest`, `CandidateForest`, `CandidateToyForest` and `reduceSideRecognizer`.
- **Evidence:**
  - The export scan found 0 production callers for every symbol in the file: `CandidateForest`, `CandidateToyForest`, `ForestDiff`, `JudgedForest`, `JudgedToyForest`, `NodeKey`, `ReduceSideForest`, `ReduceSideToyForest`, `V1Forest`, `V1Recognizer` and `V1ToyForest`.
  - `ReduceSideForest` and `ReduceSideToyForest` are the interim stage's candidate: the judged recognizer, then `Constrain`. Constrain has now left production, and only `forest/reduceside_test.go` (lines 24–26 and 92–97) still uses them.
  - `CandidateForest` is `builtStream(tab, recognize.New(...))`. `JudgedForest` is `builtStream(tab, recognize.NewJudged(tab, nil, ...))`. `recognize.New` is exactly `NewJudged(tab, nil, opts)` (`recognize.go:199–201`), so the two are the same pipeline.
  - When `Feasible` passes, `CandidateToyForest` is `JudgedToyForest`.
  - `CandidateForest` and `CandidateToyForest` are exported, mutable package variables. Tests reassign them (`forest/judged_test.go:244`, `forest/reduceside_test.go:26`, `cmd/ratchet/forest_equality_test.go:116–218`). None of those tests calls `t.Parallel`, so there is no race today.
- **Grounds:**
  - Two names for one pipeline must change together.
  - The ReduceSide pair keeps alive a configuration that production never runs. Its test, `TestReduceSide…`, pins an intermediate state, not a property V2 ships.
  - The V1 side (`V1Forest`, `V1ToyForest`, `V1Recognizer`) and `ForestDiff` are the referee instrument, which the brief says to keep. I do not flag them.
- **Remedy:**
  - Make `JudgedForest` and `JudgedToyForest` the bodies of `CandidateForest` and `CandidateToyForest`, or drop the Judged pair and point its tests at the Candidate pair.
  - Delete `ReduceSideForest`, `ReduceSideToyForest` and `reduceSideRecognizer`, together with the reduceside test's swap of the candidate.
  - Keep the reduce-side unit tests in `recognize`, which test the recognizer directly.
  - About 60 production lines go.

### P2: RESIDUAL. The precedence referee carries V1-only code in a production file, and the whole tail can move into the test file of the same package

- **Where:** `internal/chart/precedence.go`: `builtForest` (243), `referee.narrowed` (327), `narrow` (412), the `referee.p` field, and the `*recognize.Chart` result of `referee.judged`.
- **Evidence:**
  - `staticcheck -checks U1000 -tests=false` reports `builtForest`, `(*referee).narrowed` and `narrow` as unused.
  - `r.p` is set in `newReferee` (286) and read only by tests (`forest_equality_test.go:126` and `:281`, `precedence_test.go:254`).
  - `judge` throws away `judged`'s chart (`f, _, skip := r.judged(s)`, line 370). Only `precedence_test.go:384` reads it.
- **Grounds:**
  - The milestone Records already list `referee.narrowed` (d247 gate). This finding adds what hangs off it: `builtForest`, `narrow` (about 30 lines) and `r.p`.
  - These are all in package `chart`, the same package as `precedence_test.go`. That is unlike `forest/equality.go`, so nothing stops them from living in the test file.
- **Remedy:**
  - Move `builtForest`, `narrow` and `referee.narrowed` into `precedence_test.go`.
  - Have the tests build the precedence that `r.p` holds through a test helper, and keep `newPrecedence`'s rule-count check as the referee's guard.
  - Let the tests that need the chart call `recognize.Recognize` themselves.
  - About 45 production lines go. Behaviour is unchanged.

### P3: RESIDUAL. Two comments that V2 made false and that are not on the known list

- **Where:**
  - `internal/chart/forest/resolve.go:810`, the doc of `LibraryOperators`;
  - `internal/chart/cmd/ratchet/speed.go`, the doc of `measurePipelineStages`.
- **Evidence:**
  - The `LibraryOperators` comment reads: "LibraryOperators is ParseWith's library-operator stage (libraryOperators), exported for the ratchet's stage keys." It is unchanged from e3432c0f (`git show e3432c0f:internal/chart/forest/resolve.go`).
  - ParseWith no longer runs that stage: it narrows the lattice (`narrow`, `builtStream`).
  - The ratchet's production code no longer calls it either. `stageInputs` lost the call in this delivery. Its callers are now only tests and `V1Forest` (through `v1LibraryOperators`).
  - `measurePipelineStages` still says: "Library operators have no key, so the stage keys do not sum to the pipeline." The library-operator work is now narrowing, and narrowing runs inside `lexOf` (`stageFuncs` → `narrowedLattice`). `measureStages` times `lexOf` as `speed-stage-lex-aggregate-longlived`, so it does have a key.
- **Grounds:** These are the same kind as the four stale V1 comments already known, but they are not among them. A reader who trusts the first comment will think production prunes the forest for library operators.
- **Remedy:**
  - Reword the `LibraryOperators` comment to "V1's library-operator pass over a built forest, kept for the V1 side and its tests".
  - Drop the sentence in `measurePipelineStages`, or say that narrowing is timed inside the lex stage.

### P4: RESIDUAL. "Are these perly.y's tables?" is decided in three places, with three different strictnesses

- **Where:**
  - `recognize/judge.go` `deriveJudgement`: rule count, then the LHS and RHS names of every rule against `grammar.Committed()` G1;
  - `prectab/prectab.go` `buildParsed`: rule count and LHS names only;
  - `chart/precedence.go` `newPrecedence`: rule count only.
- **Evidence:** I read the three functions. `deriveJudgement` calls `ParsePrecedence` only after its own check passes.
- **Grounds:** Take a tables value with perly.y's LHS names but one RHS changed.
  - `deriveJudgement` says "not perly.y's", so the candidate judges nothing.
  - `ParsePrecedence` accepts it, so `V1Forest` constrains with perly.y's precedence.
  - The two sides of the differential would then judge different grammars.

  Only tests can build such tables today, so this is a maintenance trap in waiting, not a live fault.
- **Remedy:** Make one predicate in `prectab`, for example a `NotCommitted` flag on `Parsed` set by the full LHS and RHS comparison. `deriveJudgement` then reduces to reading `ParsePrecedence`'s result.

### P5: RESIDUAL. `complete` now holds two copies of the pairing-and-refusal step

- **Where:** `internal/chart/recognize/recognize.go`, `(*builder).complete`. These are the two parent loops: the core-indexed `awaiting` path and the branch-set `shortcut:` path.
- **Evidence:** `git diff e3432c0f 8d8fff38 -- internal/chart/recognize/recognize.go`. Both loops gained the same three steps: the `predicts(it.State, it.Ctx) != ctx` check, the `dep && !j.completes(...)` → `b.refuse` step, and `b.add(..., it.Ctx)`.
- **Grounds:** The branch-set path is rare (the shortcut comment says interleaved mode would make it common). A later change to §4.3's pairing rule made in one loop only would go unseen until a lattice with a branch hits it.
- **Remedy:** Write one local closure, `advance(it Item, target int)`, that does the pairing check, the parent-dependent refusal and the add, and call it from both loops. That saves about 8 lines and removes the copy.

### P6: RESIDUAL. Exports and counters kept only for tests

- **Where:**
  - `forest.ConstrainEntries` and its atomic counter (`forest/precedence.go:30–34`);
  - `(*recognize.Recognizer).Entries` (`narrow.go:88`, one test file, no production caller);
  - `Recognizer.leoSteps`, `leoRefused` and `leoLinks`, incremented on the `leo` and `leoCrosses` path and read only by `leo_test.go` and its neighbours;
  - in `prectab`: `ReduceDiffers`, `RuleReduceConflicts`, `ParentDiffers`, `Awaiting`, `ParentConflicts`, `JudgeShifts`, `PredictorTopsConflict`, `RuleTopsDisagree`, `ShiftRec`, and `ShiftVerdict.ContextConflict`.
- **Evidence:**
  - The export scan found one production reference for each `prectab` helper, the call inside `Feasible`. The only other user is `forest/feasibility_test.go`.
  - `ConstrainEntries` has no production caller and 4 test files.
- **Grounds:**
  - The `prectab` helpers are exported only because their test lives in another package (`forest`).
  - The counters cost nothing that matters, but they are API and struct state that exist only so a test can watch.
  - This is not harmful. The `ConstrainEntries` counter in particular backs the "production enters Constrain 0 times" evidence, so it pays for itself.
- **Remedy:** Move `forest/feasibility_test.go` into `prectab` and unexport the ten helpers. Leave the counters, or put them behind one test-hook struct. Low priority.

### P7: RESIDUAL (nice to have). `LeoLinkContext` repeats the lookup `LeoLink` has just done

- **Where:**
  - `internal/chart/recognize/finished.go`: `LeoLink` and `LeoLinkContext`, which both call `awaitingOne`;
  - `forest/forest.go:452–462`: the caller calls both, one after the other, on the same `(at, sym)`.
- **Grounds:** There are two exported accessors and two scans where one would do. The cost is once per memo, so speed is not the point.
- **Remedy:** Have `LeoLink` also return the item's context and delete `LeoLinkContext`. That is one accessor fewer.

## Known items: confirmed, not promoted

- **The Leo-link context check in `completedUnder`** (`l.ctx == ctx`, `bycontext.go`) and **the prefix-context check in `under`** (`int(p.ctx) != ctx`). The Records list these at the 3282 gate as "effectively dead" and "untested defence". I re-ran both mutations against `go test ./internal/chart/...` at 8d8fff38. Results:
  - **Mutation A** deleted `l.ctx == ctx &&`. Every chart package passed, with no FAIL line and no build failure.
  - **Mutation B** deleted the three-line `if int(p.ctx) != ctx { continue }` from `under`. Every chart package passed again.
  - Both runs are in `mut-ponytail.log`, and the worktree was restored after each.

  So the Records are confirmed: no test reaches either check.

  They are defence that is cheap and correct by §4.3, so they are not harmful. They stay RESIDUAL, and I do not promote them.
  - The prefix check costs one int compare per placed item.
  - Removing it would make `under` accept a prefix item of another context, which is wrong in principle, even though no fixture reaches that case.

  The honest fix is a test that reaches each case, not deletion.
- **The four stale V1 comments, and `referee.narrowed`:** already known. P2 and P3 extend them, and do not repeat them.
- **V1's `Constrain`, `ConstrainReference` and the state-free evaluator** (`forest/precedence.go`): kept by design as the reference and referee instrument. Not flagged.

## Noted, out of this delivery

`staticcheck` also reports `constrainSource` in `precedence_test.go:63` as unused. It is present at e3432c0f and was already uncalled there. The same goes for `ownDir`, `(*Tables).itemText`, `labelsDir`, `(*Recognizer).lookup` and `(*Recognizer).store`. None is the delivery's.

## Summary

- **Blocking findings:** none.
- **Residual findings:** 7 (P1–P7).
- **Lean:** about 120 production lines could go or move into test files (P1 about 60, P2 about 45, P5 about 8, P7 about 6), with no change in behaviour.
- **Not checked:**
  - The §7 stops: forest equality, answer keys and set 1626. Other participants re-run those.
  - The ratchet's timing.
