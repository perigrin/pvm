# Milestone chart-perly-w2 Postmortem

**Verdict: killed at waypoint 2 by spec §10's long-lived speed kill.** With the
forest added to the chart side's timed span, using the ratchet's own protocol
at 3487d258:

| Reading | Forest timed: aggregate long-lived | Forest not timed: aggregate long-lived | Forest timed: std5k long-lived |
|---|---|---|---|
| 1 | 3.51 | 2.10 | 2.23 |
| 2 | 3.64 | 2.20 | 2.22 |
| 3 | 3.47 | 2.22 | 2.18 |
| Median | **3.51** | 2.20 | 2.22 |

The limit is 3. All nine measurements were between 3.42 and 3.70. Lex-overhead
stayed within its limit throughout (medians 1.14 to 1.19).

Before the stop, waypoint 2 merged and closed 13 issues. 85 commits went
into the issues' branches, and all of them carry an `Implements:` trailer. 15
issues are pending, 3 are cancelled, and PERLGITPARITY stopped after measuring,
by design.

## What Worked Well

- **The kill was measured early and properly, not projected.** FOREST's gate
  projected 3.9 to 4.2. perigrin ruled "Measure properly first". The answer came
  from three quiet readings under the unchanged protocol, 6 to 12 minutes after
  FOREST merged (01:44 UTC; readings at 01:50, 01:53 and 01:56). It came before precedence, resolve or readings were built on a
  forest that cannot fit.
- **Gates caught real defects before merge.** Examples:
  - CTORONCE: a `weak.Make` crash on non-heap tables.
  - FOREST: superlinear forest construction (if/elsif ×2000 took 5.2 s; after
    the fix, 10.6 ms) and exponential enumeration.
  - REFRENDER: two terminals falsely listed as unrenderable, which put 20
    resolutions back under the referee.
  - PERLGITLABELS: two lexer sites that were never listed.

  No issue went past three reopens.
- **Mutation checking stayed the norm.** In CACHEGAPS, CTORONCE, XWORD and
  FOREST each new guard was shown red under a named mutation, re-run by the
  gate. FOREST's corpus forests were compared node for node against a
  reference built with no pruning, 2,678 files in all.

## What Didn't Work

- **The kill was predictable from waypoint 1.** w1 left lex plus recognize at a
  long-lived aggregate of about 2.2. That left 0.8 for the forest, precedence,
  resolve and readings together. Refinement put the speed checkpoint at the end
  of the chain (01a1c324-2009) and timed nothing per stage. The forest alone
  took 1.3.
- **The ratchet times less than the waypoint builds.** Its speed keys still time
  lex plus recognize. A forest-timed reading needed a scratch harness. The
  issue that would have put the whole pipeline in the timed path
  (01a1c324-1f72) sat behind readings.
- **The w1 close took eight attempts.**
  - Two were killed at the session's 2-hour limit.
  - One was killed by a container restart.
  - One failed on the 60 s suite, which ABSENTFAST then fixed.
  - Three were refused by criterion defects. Each needed a ruling: "Fix the
    check", "Repair the 3 checks", and "Convert 1–2, move 3–4".

  Four merged w2 issues carry the same unrunnable-criterion defect (4202, 1b0f,
  41c7 and LEXLOAD), so a w2 close would have been refused for that too.
- **Criteria that pass without the work.** Milestone criterion 7 is green, but
  `print STDOUT $foo bar();` is still accepted. A FOREST commit silently broke
  w1's repaired `speed.go` check.
- **Shared-machine incidents.**
  - A worker's `pkill -f "make test"` killed another runner's test.
  - A worker used bare `git stash` across shared worktrees (its own report,
    2026-11-12).
  - The Go build cache filled the disk at 19 GB and stopped a test runner
    (orchestrator session, 2026-11-12).

  The worker brief now forbids pattern kills and bare stash.

## What Puzzles Us

- **The cold key changed meaning.** After CTORONCE the aggregate cold ratio fell
  from 4.99 to 2.72, but long-lived barely moved. "Cold" now reuses tables an
  earlier `New` built.
- **Garbage collection costs more than the forest.** In the forest-timed
  profile, forest build took 2.71 s and GC mark workers 4.66 s of 13.32 s of
  samples. The profile cannot split GC by source; by inference a large part of
  it traces the forest's pointer graph.
- **The perl.git corpus is heavy.** It is 38% of the speed aggregate in the one
  non-quiet run that included it. It raised the aggregate before any forest
  existed.

## Where A Human Had To Act

**Collaboration, as designed.** The milestone's Records hold 24 rulings by
perigrin during the waypoint. Among them:

- the library-operator design;
- three w1 criterion repairs;
- the `isa` preamble;
- the toy Leo guard;
- the `pat_psycho.t` exclusion;
- "merge 566 when all checks clear";
- "Measure properly first".

Each of these was his call.

**Friction to engineer away** (from Records and the orchestrator's session):

- Long closes outlived the 2-hour background limit and container restarts. They
  ran detached and were re-run by hand.
- The disk-full stop.
- The environment rewrites the PR body footer on every PR (ORACLEFACTS gate F3,
  PRs #561 to #565).
- The Windows `internal/updater` test timed out on PRs #564 and #565 and was
  rerun by hand.
- Stop-hook prompts to commit `test/e2e/.perl-version` while runs held it
  modified.

## What Will We Change

- **Spend the speed budget at refinement.** Give each stage a share of the
  remaining ratio, and give each stage an issue a criterion that the timed path
  including that stage stays within its share. Here that would have been forest
  ≤ 0.3 against 2.2. This would have stopped the waypoint at FOREST's gate
  rather than after its merge.
- **Put each stage into the timed path the moment it exists.** The ratchet's
  speed keys must time everything built so far, not wait for a late issue.
- **Lint criteria at refinement:**
  - Refuse any criterion with no paren-wrapped command.
  - Refuse any criterion that passes on the unchanged tree.
  - Refuse any criterion pinned to a file a later issue may edit. FOREST broke
    one such pin.
- **Make a GC budget part of a stage's design review.** Before refining another
  attempt, the forest's issue should state its memory shape and be measured
  for collector cost as well as build time; the review found index-based,
  pointer-free nodes (ponytail S1) the only lever on the GC cost here.
- **Check whether a successor can fit at all.** With lex plus recognize at 2.2,
  a successor has to answer whether precedence, resolve and readings can fit in
  0.8, before anyone refines a third attempt.

---
Generated by crochet:postmortem on 2026-11-13.

Data sources:
- `git zhi milestone show chart-perly-w2` (Records)
- crochet:review's three lens reports
- the forest-timed reading report
- `git log 03e4097b7..3487d258`
- gocyclo

Not gathered: `git zhi verify` (pending issues make it fail by construction),
`docs health`, and sanbao.
