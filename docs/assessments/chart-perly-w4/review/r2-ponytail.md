# crochet:review round 2: ponytail (prefix P)

The subject is unchanged: `git diff e3432c0f 8d8fff38`. I read all 13 residual rows: C1–C2, Q1–Q4 and P1–P7.

**Result: I hold no blocking finding. I promote nothing and raise no new finding. I dispute one fact in Q4.**

## Rows read

| Row | My view | Promote? |
|---|---|---|
| C1 (close export not made) | Agree. It is due at close, so it is pending rather than a defect. | No |
| C2 (#649 in the range) | Agree. It is an accounting note for the postmortem. | No |
| Q1 (`NewJudged` does not check `p.Tab == tab`) | Agree. I read `newJudgement` at 8d8fff38: `Feasible(p)` and `PredictionContexts(p)` read `p.Tab`, while the loops index `tab`, and nothing compares the two. The one-line guard is the right size. No shipped path passes a foreign `p`, because `New` derives it and `ParseWith` panics on any other. | No |
| Q2 (the slipper closure runs no reduce side) | Agree, and confirmed (see below). | No |
| Q3 (refusal records keyed by parent state) | Agree it is unproven, not a fault. It found 0 cases over 986 files, the differentials are green, and I could not construct a trigger either. | No |
| Q4 (`constrainSource` is dead) | It is dead, but the claim that the delivery removed its last caller is wrong (see below). | No |
| P1–P7 | Unchanged from round 1. | No |

## Q2: re-checked, kept RESIDUAL

- **Evidence:**
  - `git show 8d8fff38:internal/chart/recognize/slipper.go`: `semiAt` builds `v := &builder{r: r, c: c, pos: pos, interned: true}` and calls `v.complete(la)` with `v.la` unset. `complete` therefore sets `j = nil`, and nothing is judged in that closure.
  - `git show 8d8fff38:internal/chart/recognize/recognize.go | grep -n 'judged under every option'` finds line 319. That line's comment says the reduce side is judged under every option, which is true of the main closure but not of the slipper's.
- **Why it does not block:**
  - Quality's count found 0 non-allow `ReduceAt` entries before `;`, `}` or EOF in perly.y's tables, so the judged and unjudged closures agree today.
  - A grammar change that broke this would make the recognizer accept what V1 refuses. The corpus and perl.git forest-equality differentials compare against V1's Constrain and would go red on any file that reaches it. So the failure is detectable, not silent, though only as far as the corpus reaches.
- **From my lens:**
  - The comment at `recognize.go:319` overstates what is judged. It belongs with P3's stale or overreaching comments.
  - The smaller fix is Q2's second option: pins for `}` and EOF beside the `;` pin, plus a `shortcut:` comment in `semiAt`. It is three test rows and one comment, with no new code path.

## Q4: dispute of fact (not mine to release)

- **Claim:** "The function exists at e3432c0f with callers. The delivery removed its last caller."
- **Evidence:**
  - `git grep -n constrainSource e3432c0f -- .` matches only the definition (`precedence_test.go:83–84`) and an archive mention in `docs/assessments/chart-perly-w3/reports/round6-fresh.md`. There is no call site.
  - `git log -S'constrainSource(t' -- internal/chart` finds `bae3ee09` (2026-11-16), which is before e3432c0f.
- **Conclusion:** the helper was already uncalled at the baseline, which matches my round 1 note ("Noted, out of this delivery"). Deleting it is still right, but it is pre-existing debt, not the delivery's.
- **Classification:** I leave this to Q4's author. It changes the attribution only, not the classification.

## Blocking held

None.
