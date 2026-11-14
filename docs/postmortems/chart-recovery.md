# Milestone chart-recovery Postmortem

**Verdict: the recovery landed inside its projection, and the run stopped at
the verdict reading as planned.** One quiet reading on pu cd855330 (three
ratchet runs, 08:10–08:16 UTC, 2026-11-14), every R and F issue merged:

| Stage (aggregate long-lived) | Before, pu 57caafca | Verdict | Factor |
|---|---|---|---|
| lex | 0.5980 | 0.6023 | ×1.007 |
| recognize | 1.4246 | 0.7194 | ×0.505 |
| forest | 1.1348 | 0.6874 | ×0.606 |
| Sum (min–max) | 3.1329 | **2.0210** (1.9949–2.0249) | ×0.645 |

Before is this host's quiet reading; §4's factors on it project 1.98–2.29
(R1–R3 + R5 + F2) and 1.83–2.14 (with F1).
The lowest run is under 2.57, so nothing was reported under D1. lex-overhead
read 1.144. R4 was dropped by profile (`goTo` 0.47% of samples against 2.05%
noise). Precedence, resolve and readings are unbuilt; the held chart-perly-w2
issues are re-refined next.

All 18 issues are done; all 63 commits carry `Implements:`. Created
2026-11-13 20:22 UTC; last close 08:31 UTC; review fixed point after #588 merged, 10:03 UTC. Budget (D4): 8 working days, due 2026-11-21.

## What Worked Well

- **Stage keys first made every gain attributable.** M1–M3 merged before any R
  or F issue, so each reported its own stage in interleaved pairs: recognize
  R1 1.4507→1.2805, R2 1.2864→0.7821, R5 0.7944→0.7007; forest F1
  1.1902→0.9417, F2 0.9821→0.6630.
- **Goldens held through six optimisation issues, unregenerated.** P5's gate
  (G1) made the forest golden see child rewiring; it then caught all 12,450 of
  F1's gate's rewiring mutations.
- **Gates caught real defects:** P1 G2 (nil-interface panic), P4L G2 (a
  `blocks` cycle exhausting jq), M1 G3 (nil `Root` on a refused chart), R5
  G1/G2 (no gain, traced to a 3,235-entry linear scan, 219.5 steps per
  intern). Nine of 18 issues took one reopen; none took two.

## What Didn't Work

- **The review found a defect no gate saw.** Q1: concurrent forest builds from
  one recognizer differed from sequential ones. Slipped's Recognizer copy
  shared R2's reused buffers, and R1's probe counter made LeoLink write; 17
  races, absent at base. Each gate saw one issue; the defect spanned two.
- **Bounds were written against moving baselines.** R3's bound (memoAt
  whole-set scans) was 0 on pu a5498fa4, because R1's index moved the Leo test, so it could not be red;
  worker rce4e435-16 stopped and perigrin re-pointed it. F2's bound, filled at
  0.586 from F1, passed without F2's work once R2 moved alloc-per-token-forest
  to 0.490; it was re-filled at 0.489.
- **Criteria named tests, not properties.** Review findings P1, P2 and P4
  (two name tests) were kept only because milestone criteria verify them.
- **The machine was unstable.** The 00:58 UTC restart moved hosts (fc-v445 to
  fc-v470) and the quiet sum fell 3.638 → 3.1329 with no recognizer or forest
  change. On a5498fa4, recognize read about 0.77 in one session and 0.90 in
  another. Restarts lost R3's worker 15, R5's worker 17 and F2's first gate.
  R1's first gate stalled an hour with no output.
- **Process slips.** M2's worker never ran `--state start`, so its close ran
  start then `done --force`. The orchestrator's own C6 README fix dropped an
  open item; #587's gate caught it. Unrelated CI flakes needed reruns (macOS
  `internal/cli/docs` hang, R1; Windows `internal/updater` timeout, F2). The merge script refused `skipped` check runs until patched (session,
  not Records).

## What Puzzles Us

- **The forecast missed by about 14×, yet telemetry says RED.** About 14 hours
  against 8.2 days. Median start-to-done 0.8 h (w2: 2.3 h); longest F2 at
  2.4 h; chain review about an hour, not 2.5 days; C1, sized at 3.7 days,
  closed in 28 minutes. Telemetry says fever RED, buffer 16.6 of 9
  burned. MPG 154.5 is P3's session window counting 2,712 commits (the other
  17 total 69). sanbao reports rework_rate 0 and avg_review_cycles 0 despite
  nine gate reopens, because gates run before `done` and no `reopen` is
  written.
- **Why did the chart/RD stage ratios fall about 14% on the new host** with
  unchanged recognizer and forest code?

## Where A Human Had To Act

**Collaboration:** D1, D2, D4, D5, D7 and D11 before the run; R3's
re-pointing during it, since what a criterion measures is perigrin's to
change. Queued, not blocking: D7's ruling on C1, C1's rule-336 nullable stop, the
std5k-cold D12 candidate, the GC-share definition.

**Friction:** the *cause* of R3's ruling. R1 moved the Leo test onto the index,
removing what R3 counted. R1's title ("…for scan, completion and the Leo
test") and its File Structure line said so: answerable at chain review. Restarts, the stalled gate, CI
reruns and the merge script needed hand re-dispatch, not judgment.

## What Will We Change

- **Fill bounds at issue start.** For a key an earlier chain issue can move,
  the decomposer writes the bound as "one unit below pu's reading at start",
  and the orchestrator re-runs its red-on-base check then. P4L ran once, on
  352b3671.
- **Chain review asks, per counted bound,** whether an upstream issue removes
  what is counted (R1 → R3).
- **Add a shared-state criterion.** An issue adding reused buffers or counters
  to a copied type (Recognizer, Chart) gets a `-race` concurrent-use
  criterion; `TestForestConcurrentBuildsOneRecognizer` is the template.
- **Name tests only for goldens and regression guards;** elsewhere assert the
  measured property, so review can delete a redundant test.
- **Record the host kernel with each quiet reading;** a change voids the
  before reading. Compare only interleaved pairs from one session.
- **Watchdog gates:** 15 silent minutes means re-dispatch.
- **The orchestrator runs `--state start` at dispatch** (M2) and records gate
  reopens as `git zhi` transitions, so sanbao sees rework.
- **Size the next budget from this run:** a 0.8 h median and a chain review
  of hours, not w2's 75th percentile.

---
Generated by crochet:postmortem on 2026-11-14.

Data sources: milestone Records and telemetry, issue list, sanbao, merge log
352b3671..origin/pu, trailer counts. `docs health` checked 0 docs. `git zhi
verify`: run by the orchestrator at close.
