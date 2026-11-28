# crochet:review round 2: fresh2 (prefix F)

Participant: **fresh2**, a fresh reader in round 2. Subject: `git diff e3432c0f 8d8fff38`. Every command below ran with `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit`. They ran in scratch worktrees `wt-fresh2` and `wt-fresh2-v2` (both at 8d8fff38) and `wt-fresh2-v1` (at e3432c0f), which are removed at the end. The repository was not edited.

(This file replaces an older `r2-fresh2.md` from the chain-review of 2026-11-25 that had the same name. That report is kept in the repository at `docs/assessments/chart-perly-w4/chain-review/r2-fresh2.md`.)

**Result: no blocking findings.** I promote nothing. I raise one new RESIDUAL finding, F1, and correct one round-1 row, Q4.

## What I looked for that the three round-1 lenses could all have missed

### 1. Is "the forest equals V1's" checked against V1 itself?

Every round-1 participant relied on the in-tree differential, `ForestDiff(CandidateForest, V1Forest)`. That differential's V1 side, `forest/equality.go` `V1Forest`, is not e3432c0f's code. At 8d8fff38 it is built from:
- `recognize.NewNoPrecedence`;
- V2's `Build`, which now walks by context through `bycontext.go`;
- `libraryOperators`;
- `Constrain`.

The delivery rewrote the first two: `forest.go` +170 lines, `recognize.go` +213, `sets.go`, `leo*.go`. A change to that shared code could move both sides of the differential together and leave it green.

What anchors the comparison to e3432c0f today:
- `set-digest-noprecedence.golden`, for the recognizer's sets;
- `resolved-forest-digest.golden`, for Parse's forest **after resolve**, on the 458-file sample;
- the 27 answer keys.

Nothing anchors the pre-resolve forest after Constrain to V1's own code. `67af3e99` regenerated `forest-digest.golden`, which was the one pre-resolve forest anchor.

**Probe: a cross-commit check, which no round-1 participant ran.** It is the same test file in two worktrees, with only `probeForest` differing. The sources and outputs are kept in `f2-probe/`.
- **At e3432c0f:** V1's own `parseStream` up to and including Constrain. That is `lex.Lex`, then `Linear`, then `recognize.New` (unjudged at that commit), then `Build`, then `libraryOperators`, then `ParsePrecedence`, then `Constrain`.
- **At 8d8fff38:** production's `builtStream(tab, recognize.New(...), src, scope)`. That is narrow, then the judged recognizer, then `Build`, with no Constrain.
- **Both sides:** the scope is each file's imports from `parse.ParseWithLoader` with the guacamole stubs. Each accepted forest gets two digests:
  - `forestText`, which sorts alternatives;
  - an ordered variant that keeps each node's alternatives in `Alt` order, because §5 compares alternatives in order.

Results:

| Population | Files | Accepted, both | Digest differs (sorted or ordered) | Refused set |
|---|---|---|---|---|
| forest sample plus the 996 T1 `.t` files | 1780 | 1439 | **0** | the same 341 files |
| perl.git `t/**/*.t` | 620 | 617 | **0** | the same 3 files |

The 55 refusal-reason differences in the T1 run are mine, not the delivery's. My V1 probe did not map `lex-unmapped` and reports `grammar` instead. The set of refused files is identical.

**Conclusion.** The §7 stop "a forest differs from V1's" does not fire when the check is made against V1's own code. The verdict here is independent of the shared code. This is the strongest evidence I can give that the exit's "differentials green" means what it says. It leaves F1.

### 2. Does the exit rest on one unreplicated measurement?

The `exit: met` comes from one six-run session, `verdict/`, at 1477098790. Coverage showed that 1477098790 differs from 8d8fff38 only in docs.

I did not re-time it. The machine was not quiet: load1 was 2.77, with other participants' test runs and a `git zhi verify` live. A reading under that load would be worth less than the kept one.

I did test whether the kept readings replicate. There are three sessions, read from the `run*.stdout` files with `grep speed-ratio-aggregate-longlived`:

| Session | After commit | Before runs | After runs | Exit |
|---|---|---|---|---|
| `early-reading` | 5b890890 | 3.2143, 3.2489, 3.3005 | 3.3983, 3.4617, 3.3991 | unmet |
| `early-rereading` | 6a555065 (after 4792) | 3.2481, 3.2591, 3.3143 | 3.0461, 3.1869, 3.0685 | met |
| `verdict` | 1477098790 | 3.3716, 3.3255, 3.3345 | 3.0753, 3.1492, 3.0205 | met |

- The exit is met in two independent sessions, at two commits that both include the 4792 rework.
- In the verdict session, every after run is below every before run: the after-maximum is 3.1492 and the before-minimum is 3.3255.
- Pooled across the two met sessions, the after-maximum is 3.1869 and the before-minimum is 3.2481.
- The before side drifts by about 0.1 from session to session. That drift is larger than the spread within a session, so the within-session interleaved protocol the design specifies is the right comparison.

The exit is replicated in substance, so I raise no finding. One caveat: the formal reading of the final code was taken once, by design (`drive.sh` refuses to overwrite a reading).

### 3. Has the whole of `make test` been run at 8d8fff38?

Round 1 ran `go test ./internal/chart/...` only. `early-rereading/README.md` records a `make test` pass at 6a555065, not at 8d8fff38. The range also carries non-chart Go code (#649, `internal/version`), and CLAUDE.md requires 100%.

I ran `T1_CORPUS=… make test` at 8d8fff38 (log: `f2-maketest.log`). Result: **exit 0.** All 54 packages with tests are `ok` with no FAIL, `internal/version` among them. `chart-absence-reference` passed both of its tests: `TestAbsenceCandidatesMatchFullMap` (3430 keys, 0 mismatches) and `TestTDTAbsentReasonsMatchPerClaimScan`. The CLAUDE.md rule holds at 8d8fff38.

### 4. Does anything outside `internal/chart` depend on the changed API?

`grep -rln 'internal/chart' --include=*.go cmd internal test`, excluding `internal/chart/`, finds four test files in `internal/lexer` and `internal/parse`. They mention a chart testdata path in a comment and import nothing. So `ParseWith`'s new panic, the move of `prectab` and the export changes reach no caller outside the chart. No production binary imports the chart.

## New findings

### F1: RESIDUAL. The forest-equality differential's V1 side shares the code the delivery changed, and nothing in the tree anchors pre-resolve forests to V1's own code any more

- **Where:**
  - `internal/chart/forest/equality.go`: `V1Forest`, `V1ToyForest`, `V1Recognizer`;
  - `forest/forestdigest_test.go`: `TestForestDigestGolden`, regenerated in `67af3e99`.
- **Evidence:**
  - `V1Forest` calls the current `Build` over `recognize.NewNoPrecedence`, then `libraryOperators` and `Constrain`, all at HEAD.
  - `git diff --stat e3432c0f 8d8fff38` shows `forest.go`, `recognize.go`, `sets.go`, `leo.go` and `leo_memo.go` changed.
  - The cross-commit probe in §1 above: 0 of 1439 and 0 of 617 forests differ from e3432c0f's own pipeline, so **the claim holds for this delivery**.
  - The milestone Records (`coverage-ms.md`) pin only that the V1 side's *recognizer* is unjudged (FR2-4). They do not pin that its `Build` or recognizer code is V1's.
- **Grounds:**
  - The differential is the §7 "forest differs from V1's" stop. It is meant to outlive this milestone: the README says the V1 side and `ForestDiff` are the referee instrument that is kept.
  - A later change to `Build`'s context walk, or to the unjudged path of `complete`, moves both sides together and stays green.
  - The goldens that are still anchored, set digests and the resolved forest, would not see a pre-resolve change that resolve then prunes away.
  - Not blocking, because I re-ran the claim independently and it holds.
- **Remedy:** do one of these.
  - Commit a golden of `V1Forest`'s digests over the sample, generated from e3432c0f's pipeline (the `f2-probe` V1 side), and a test that holds `V1Forest` to it, so the instrument has a fixed point.
  - At least, state in the README's forest-equality section that the V1 side is V1's *pipeline* over HEAD's shared code, and record what anchors it.

## Round-1 residual rows: every row read

| Row | My view | Notes |
|---|---|---|
| C1 close export not made | RESIDUAL, agree | Due at close. It is the only durable copy of the Records, because `refs/zhi` cannot be pushed, so the postmortem should not start without it |
| C2 #649 in the range | RESIDUAL, agree | §3 above runs `make test` over it too |
| Q1 `NewJudged` with no tables check | RESIDUAL, agree | Not reachable from `New`, `Parse` or `ParseWith` |
| Q2 slipper closure's reduce side unjudged; only `;` pinned | RESIDUAL, agree | The README (l.2818–2822, l.3285–3288) honestly says nothing judges the zero-width `;`. The comment at `recognize.go:319` ("judged under every option") is what overstates. The `}` and EOF pins remain the cheap remedy |
| Q3 refusal records keyed without context; `alt` checks the first state | RESIDUAL, agree | F1's probe adds independent evidence that no corpus or perl.git forest is affected. Still unproven in general |
| Q4 dead `constrainSource` | **Correction:** not the delivery's | At e3432c0f, `grep -rn constrainSource internal/chart/*.go` finds only its definition (`precedence_test.go:84`) and no caller. Q4 says "the delivery removed its last caller". It did not: it was already dead, as P's "Noted, out of this delivery" says. Q4 is a pre-existing cleanup, not a delivery residual. Only Q can release it |
| P1 `equality.go` test-only API, Candidate and Judged duplication | RESIDUAL, agree | F1 bears on the V1 half of the same file |
| P2 referee's V1-only code in `precedence.go` | RESIDUAL, agree | |
| P3 two stale comments (`LibraryOperators`, `measurePipelineStages`) | RESIDUAL, agree | Same kind as the four known stale comments |
| P4 three "is this perly.y?" predicates | RESIDUAL, agree | Reachable only with tables a test builds |
| P5 duplicated pairing and refusal step in `complete` | RESIDUAL, agree | |
| P6 test-only exports and counters | RESIDUAL, agree | |
| P7 `LeoLinkContext` repeats `LeoLink`'s lookup | RESIDUAL, agree | |
| P, known: mutations A and B survive | RESIDUAL, agree | No promotion |

## Release

I hold no blocking finding. F1 is residual and holds no round open.

## Not checked

- I did not re-time the speed exit, because the machine was loaded.
- I did not attempt a full `-race` pass with a raised timeout.
- I did not build an input for Q3.
