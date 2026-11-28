# Milestone chart-perly-w4 Postmortem

This milestone delivered chart precedence V2: the recognizer judges precedence itself, and production no longer runs Constrain. The decision is `docs/plans/2026-11-25-chart-precedence-v2-design.md`, revision 9.

| Measure | Value |
|---|---|
| Issues | 28 |
| Commits on the issues' sessions | 155 |
| Pull requests | #655–#692, plus docs PRs |
| Execution window | 2026-11-26 00:21 to 2026-11-27 22:24 UTC |

**The exit was met.** The formal reading (#690) ran quiet, interleaved B A B A B A. It gave a before-minimum of 3.3255 and an after-median of 3.0753 for `speed-ratio-aggregate-longlived`.

- All 27 answer keys were identical across the six runs.
- `resolved-forest-digest.golden` was byte-identical.
- At close, the verify gate passed 210 of 210 criteria at `7aa30362`.
- crochet:review reached a fixed point at round 2, with 0 blocking findings.

## What Worked Well

- **Differentials first, recognizer second.** The no-Constrain forest differential was green before the switch-over (#676 → #679). So every later recognizer change was judged against a standing V1 comparison:
  - corpus: 1881 of 1881 files;
  - perl.git: 620 of 620 files;
  - seeds 4 and 44.
- **No §7 forest stop fired across 28 issues.** The review's fresh reader confirmed it independently: 0 of 1439 and 0 of 617 forests differ from e3432c0f's own pipeline.
- **The early speed reading paid for itself.** perigrin ruled "Take a quiet reading now" after the d218 gate flagged risk. The reading caught the unmet exit (3.3991 against 3.2143) while four issues were still unstarted:
  1. A profile named the cause, `byContext`'s per-node hash map.
  2. One sonnet issue (`4792`) fixed it.
  3. The re-reading met the exit.

  Without the early reading, d247, d278, d2a2 and bb9b would have shipped on a path that missed.
- **Gates with mutation probes.** Gates reopened 15 issues. Several reopens were tests that could not fail, and a gate found each one by reverting or mutating the code:
  - d199, d35d, 4792, d2a2, and d247's F1;
  - d1eb's `Expected` seeding defect, found on a feasible toy.

## What Didn't Work

- **Cost was out of sight until late.** No issue's criteria bounded `forest.Build`'s time or allocations. 3282 doubled the forest stage (0.67 → 1.45), and every gate passed it, because the only speed signal was the milestone exit.
- **The bounds had no headroom.** The judged probes bound was set at 0.6378 against 0.64, the Leo links bound at one link of headroom, and Build allocations at 0.4293 against 0.43. Each was set from a single measurement, so a toolchain bump can turn any of them red.
- **Tests that could not fail kept reaching the gate.** Of the 17 reopen cycles, at least 6 had a vacuous test or a vacuous log line as their main finding:
  - d35d (twice);
  - 4792;
  - d2a2;
  - d247;
  - d218's README pass came close too.

  Workers ran mutation proofs but did not always aim them at the claim the criterion makes.
- **The environment misled workers.** The shell's default `T1_CORPUS` (`/opt/corpora/PerlOnJava`) is wrong, which produced 4792's false mutation proof. Three parallel workers also tripped the ratchet's load gate (d2a2).
- **Bulk rewrites dropped criterion phrases.** bb9b's README rewrite removed the phrase 1251's criterion greps for. Only `git zhi verify` at review caught it, which took #692.

## What Puzzles Us

- **The forest stage's rise was stable but not dominant.** It rose from 0.67 to 1.13, yet the whole pipeline fell, because dropping Constrain saved more. The design's §7 exit gates only the aggregate, so a stage could get worse indefinitely while the exit still held.
- **d35d took 32 commits,** against at most 10 for any other issue, after a ruling widened its scope mid-issue.
- **Q3 is unresolved:** refusal records are keyed by parent state, not context. It showed no effect on T1 (0 of 51), and no input triggering it was built.

## Where A Human Had To Act

- **Collaboration: decisions reserved for perigrin.**
  - Criterion repairs: d10d.
  - Scope changes: d199's out-of-scope items, d35d's widening, d1c2's dead clause.
  - Amendment 1 for 3262, adding two unlisted tests.
  - The three speed rulings: read early, profile, then fix A.
  - Each was a §7 stop or a started-issue amendment, which the protocol rightly reserves for a human.
- **Friction.**
  - A subagent stalled for 15 minutes on an MCP permission prompt during review.
  - Two container restarts killed in-flight workers. The agent recovered them from scratch reports without human input.
  - Four questions to perigrin stayed unanswered across the milestone:
    - O1, Leo walk union keys;
    - the `internal/cli/docs` stub;
    - the trailer versus "no model identifiers";
    - golden-alone commits against CLAUDE.md's "pass before every commit".

  None blocked progress, but each still sits in the Records.
- **What would have let the agent proceed:** a brief rule against MCP tools for review subagents, the correct `T1_CORPUS` in the environment, and a standing ruling on ordering golden-regeneration commits.

## What Will We Change

1. **Put a cost criterion on every recognizer or forest issue.** For example, Build allocations per token on the in-tree sample, with a named bound and stated headroom. Then refinement catches a doubled stage at the issue that causes it, not at the exit.
2. **Set bounds with headroom.** A bound is the measured value plus a stated margin (say 3%), or the measured maximum over five runs. Never a single run's value rounded up.
3. **Give every log-count criterion a counted source.** A `compared N of M` criterion must name where N is incremented, and the worker must mutate that site. Add this to the SQE prompt's negative scenarios.
4. **Run `git zhi verify <milestone>` after any docs-only rewrite issue,** before its gate. Rewrites remove the phrases other issues' criteria grep for.
5. **Fix the environment default.** Set `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit` in the cloud environment, and drop the per-brief warning.
6. **Take the early speed reading as a planned step.** Refinement should place a speed checkpoint after the switch-over issue, so it is no longer an ad-hoc ruling.

---
Generated by crochet:postmortem on 2026-11-28
Data sources: `chain/issues.json` at close, the milestone Records, `git zhi verify` (210 of 210), the readings and the review record. Not used: sanbao, and `docs health`, which observes no document covering `internal/chart`.
