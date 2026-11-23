# Chart parser: the precedence constraint without maps

Status: proposed, for crochet:assess (revision 3). Date: 2026-11-23.
Assessment: `docs/assessments/chart-precedence-in-recognizer.md`.

This file keeps its name from revisions 1 and 2, which proposed applying
precedence in the recognizer. Measurement retired that approach (§2), and
perigrin retargeted the work. The governing design is the accepted v2 design,
`docs/plans/2026-10-30-chart-parser-v2-design.md` (branch
`docs/chart-parser-v2-design`), as amended by the Chart Parser Recovery Plan
(`docs/plans/2026-11-13-chart-parser-recovery-plan.md`, branch
`docs/chart-recovery-plan`; perigrin's D2, 2026-11-13).

## 1. Rulings

All perigrin's, of 2026-11-23, recorded here and at refinement in the
milestone's Records.

- After round 1: "Do just B first and let's re-measure".
- After round 1: "Rule them in", for the three stage keys of §5.
- The goal: exit `speed-ratio-aggregate-longlived` at or below 5; budget 7
  days.
- After round 2: "Relative and ≤ 5": the exit is met when the after-reading's
  median is below the before-reading's minimum, taken interleaved in one
  session, and at or below 5.
- After round 2: "Measure first, then decide".
- On the measurement (§2): "We need to eliminate the map".

## 2. The measurement

Taken on `pu` 8c3317079 over the ratchet's own aggregate population: 2,211
files, three quiet runs, long-lived. The full report is in the assessment
minute.

Stage time, as a ratio over recursive descent's time on the same files:

| Stage | Ratio |
|---|---|
| lex | 0.66 |
| recognize | 0.94 |
| build | 0.73 |
| library operators | 0.09 |
| the precedence constraint (`forest.Constrain`) | 2.38 |
| resolve | 0.20 |
| readings | 0.15 |

The constraint is 43% of the pipeline, and costs more than lex, recognize and
build together.

Its cost does not follow the forest's size:

- With every alternative it drops removed beforehand, it still takes 0.88 of
  its time.
- With bison's reduce-side precedence applied first (revisions 1 and 2), it
  takes 0.89.

So no earlier filter can remove much of it.

The CPU profile attributes about 57% of the constraint's own time to Go map
operations:

- the `firsts` and `marked` maps keyed by node and edge context: 48%,
  including the cost of growing maps made with no size hint;
- the goto map in `Precedence.walk`: 9%.

The keys are small and few:

- Over the aggregate, 86% of the nodes the constraint evaluates are evaluated
  in exactly one edge context. No node is evaluated in more than 12 contexts
  or marked in more than 32.
- bison's automaton has 687 states over 237 symbols.

## 3. The decision

**The precedence constraint keeps no hash map on its evaluation path.** It
computes exactly what it computes today; only the data structures it uses
change.

What must become true:

1. **Identical output.** For every file, the constraint keeps and drops the
   same alternatives and gives the same verdict as before.
   `resolved-forest-digest.golden` and `forest-digest.golden` are unchanged,
   and so are every ratchet count key and the referee's `precedence-disagree`
   and `precedence-covered`.
2. **No map on the evaluation path.** No hash map is read or written in
   `Constrain` or in the `Precedence` lookups it calls per derivation.
   Structures sized by the forest are allocated once per call, not grown
   entry by entry.
3. **A slow node stays correct.** The design is chosen for the measured
   shape, few contexts per node, and must stay correct when a node has many.
   A file that makes it slow is a speed finding, never a wrong answer.
4. **The constraint is timed.** The three stage keys ruled in are taken on
   `pu` before the change: `speed-stage-precedence-aggregate-longlived`,
   `speed-stage-resolve-aggregate-longlived` and
   `speed-stage-readings-aggregate-longlived`. They are reported timing keys,
   like every `speed-` key: never gated and never compared with the baseline.

## 4. Exit and stops

Speed is read as chart-recovery's verdict was. Take three quiet measurements
of the before-state and the after-state, interleaved in one session at the
end, with the commit, load, core count and kernel recorded for each. Report
each key's median, minimum and maximum. "Quiet" means a 1-minute load at or
below the core count.

**Exit:** §3 built, its first property holding, and perigrin's relative exit
met by `speed-ratio-aggregate-longlived`.

**Stops,** reported to perigrin and never tuned toward:

- any output in §3.1 changes;
- the after-reading's median `speed-stage-precedence-aggregate-longlived` is
  not below the before-reading's minimum;
- anything that would need a criterion change or an amendment of a started
  issue;
- twice the budget.

## 5. Out of scope

- Precedence in the recognizer, and a value on items: revisions 1 and 2. The
  measurement bounds their gain at about 5% of the pipeline. Round 1's and
  round 2's findings on them stay in the minute.
- Changing what the constraint decides, including its kept-in-one-context
  shortcut.
- Speed work on any other stage.

Only `internal/chart` changes, plus this document and its assessment.
