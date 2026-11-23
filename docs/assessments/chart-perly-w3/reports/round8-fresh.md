# Round 8, fresh participant

Subject: `docs/plans/2026-11-23-chart-stateless-precedence.md`, blob
`ed91e46239b29283ad5c82cd390411621b52a2cd`. The blob is the same at 336067853 and
at the worktree's HEAD, 146411199, which adds only the minute. Code was read at
`origin/pu` = 8c3317079, the evidence commit. I read the minute and `reports/`
only after writing this.

## What I checked and found to hold

- **§2 figures against the indexed outputs.** I recomputed each from
  `constrain-measurement.md`, `combined-runs/result-run{1,2,3}.txt`,
  `harness/fresh5/result-f5.txt`, `referee.txt` and `toys.txt`:
  - Constrain 2.33–2.45 (43%) and 2.18–2.23 (39%): match.
  - V1 share 0.72–0.74, which is 4.1597/5.6207, 4.0274/5.5886 and 4.1651/5.6781. V1 ratio 4.03–4.17: match.
  - Precedence stage 0.39–0.42, from 0.9322/2.2061, 0.8848/2.1770 and 0.8739/2.2275: match.
  - At most 24 memo entries (HIST max 24): match.
  - V2 3.85–4.02: match.
  - Round 5's 0.727 (4.1237/5.6701), 4.12 and 4.10: match.
  - 1,296 referee sentences, 25,490 toy forests, 4,395 toy sentences and 2,211 aggregate files: match.
  - Six mixed reduce pairs (`mixed reduce pairs 6`): match.
- **Named things that exist:**
  - `(*Forest).Constrain`, and every caller §3.4 names (forest.go:138, precedence.go:327, and both precedence tests).
  - `NewPrecedence`, which already returns an error. `ParseWith` already turns `prec.err` into a refusal.
  - `libraryOperators`, `resolve` and `countReadings`, all unexported as stated.
  - `Protocol.pass`, which starts the clock after GC with nothing per pass, as stated.
  - `waypoint2Refusals`, `CountPrecedence` and `narrow`.
  - `TestForestDigestGolden` and `TestResolvedForestDigestGolden` (the "both forest digests").
  - Baseline `speed-ratio-aggregate-longlived` 4.7915.
  - T1 in the tree, std5k and edit5k.
  - `chart-ratchet.yml` with `timeout-minutes: 20`.
  - The `ci.yml` Ratchets job, with `PERL5_CORPUS`.
  - The README section "The precedence constraint".
  - The reference's need for the unexported `states` is real.
- **Named things that are new are marked new:** `ConstrainReference`, the target, the three keys, the copy API, the hook, the exported access and every new test.
- **Precedent for the target:** `chart-absence-reference` is the same shape, a tag, a Makefile target and a Ratchets step.
- **v2 alignment:**
  - Rule 12 and §8 stand.
  - The state-free keys are closer to §8's "exposed operator context at both edges".
  - They are closer to §10's "stated over the forest from a span's derivations and its neighbouring tokens" than bison's states were.
  - The §11 departure is listed.

## Findings

**M1. §5: the exit and stop 3 can both hold, and nothing orders them.**
- **What can happen:** the whole-ratio exit is met while the precedence stage's median is not below its before-minimum. The stage keys are timed in separate passes, so noise can do this.
- **Consequence if it ships:** the last issue can close on the exit with a stop live, or halt despite the exit, at the implementer's choice.
- **Fix:** one sentence, "a stop overrides the exit".
- **Where:** the decision.

**M2. §5: a §3.4 check can fail without differing from the reference, and no stop covers it.**
- **What can happen:** any of these fails, though production matches the reference:
  - the new test of the reference against bison;
  - the seeded generators' test, on "at least one sentence is refused";
  - the toy grammar that must fail generation;
  - the perl.git target, by timing out or skipping in CI.
- **Why it falls between:** the exit requires "the checks of §3.4 pass", and stop 1 covers only "differs from the reference".
- **Consequence if it ships:** that outcome is neither exit nor stop. The natural repair is the tuning that the stops forbid: reseed, drop a sentence, or loosen an assertion.
- **Fix:** stop 1 reads "a §3.4 check fails, or …".
- **Where:** the decision.

**M3. §3.4: "its bulk moves behind the build tag" names no tag, and the only one the decision names leaves the developer's path.**
- **Precedent:** `make test` already runs the tagged `chart-absence-reference` (Makefile:23).
- **The problem:** `chart-precedence-reference` is explicitly not run by `make test`. If the corpus bulk moves behind that tag, the in-tree agreement check runs only in CI, which undoes "Split it" for in-tree corpora.
- **Fix:** say the bulk goes behind a tag that `make test` runs, as the absence reference does. "Slow" needs a threshold, which the issue can set.
- **Where:** mostly the implementing issue; one clause in the decision.

**M4. §3 "What this does not amend": the v2 §9.5 departure is unlisted.**
- **The rule:** v2 §9.5 says keys "may be added or retired when the reset lands, and never afterwards".
- **The decision:** it adds three keys. That is ruled ("Rule them in", as Recovery Plan D5 ruled its own additions), but the list names only §11.
- **Consequence if it ships:** a reader of the archive checking §9.5 finds an unrecorded departure.
- **Fix:** one bullet.
- **Where:** the decision. Low severity.

**M5. §2: some figures rest only on reports, not on kept outputs.**
- **Which figures:**
  - 105–137 ms (round 7);
  - 20,083 with 185 refused, and 15,000 with 989 refused;
  - round 6's re-run.
- **The gap:** `INDEX.md` has no round-7 harness or output entry. I could not reproduce these figures from `measurement/` or `harness/`.
- **Consequence:** nothing rests on them. Neither the exit nor a stop uses them.
- **Fix:** label them "reported, not kept", as §2 already does for round 6, or add round 7's output to the index.
- **Where:** the decision. Low severity.

**M6. §1 and §5: "≤ 5" is near-vacuous where the baseline was taken.** This is a view, not a defect.
- **On the baseline machine:** today's ratchet reads 4.79, so the absolute half of the exit is already met before any work.
- **On the evidence machine:** the harness read 5.5–5.6, but the harness is different. At ×0.73 either is well under 5.
- **So:** the relative half carries the exit, and the ceiling only guards a slow verdict machine. That is consistent with "Relative and ≤ 5", and no change is needed.

**M7. §4: the stage keys' handling of refused files is undefined.**
- **The cases:** a file that `libraryOperators` or the constraint refuses has no forest for the next stage.
- **What exists:** `chartStages` documents its equivalent for `Build`.
- **Fix:** the stage keys need the same statement, plus the rule that resolve's copies are taken after the constraint.
- **Where:** the implementing issue, not the decision.

**M8. §3.4: the in-tree corpus check runs on every OS in the CI matrix.**
- **The cost:** about 1,650 files, with both judgements and a forest copy each, now inside `go test ./...` on three OSes.
- **The escape:** "proves slow" exists, but see M3.
- **Where:** the implementing issue.

## What I did not examine

- I did not re-run any prototype, generator or timing.
- I did not apply the prototype diffs beyond grepping `combined.diff.gz`.
- I did not verify the 11,574-file mapping, nor the 185 and 989 refusal counts.
- I did not check that the six mixed pairs are `indirob` and `myterm` before `[` and `{`. I checked only the count.
- I did not estimate the ratchet's runtime against the 20-minute timeout.
- I read the Recovery Plan only for D1, D2 and D5.

## Verdict

**Modify.** M1 and M2 are one-sentence edits that close the exit/stop gap. M3 and M4 are one clause each. The rest belongs in issues or needs no change. The decision is otherwise internally consistent:
- every named thing exists or is marked new;
- §2 matches the indexed evidence;
- it fits v2 and the Recovery Plan.
