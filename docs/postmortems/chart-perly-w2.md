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
  from three quiet readings under the unchanged protocol, about an hour after
  FOREST merged. It came before precedence, resolve or readings were built on a
  forest that cannot fit.
- **Gates caught real defects before merge.** Examples:
  - CTORONCE: a `weak.Make` crash on non-heap tables.
  - FOREST: superlinear forest construction (if/elsif ×2000 took 5.2 s; after
    the fix, 10.6 ms) and exponential enumeration.
  - REFRENDER: two terminals falsely listed as unrenderable, which put 20
    resolutions back under the referee.
  - PERLGITLABELS: two lexer sites that were never listed.

  No issue went past three reopens.
- **Mutation checking stayed the norm.** Every new guard was shown red under a
  named mutation. FOREST's corpus forests were compared node for node against a
  reference built with no pruning, 2,678 files in all.
- **The chart stays simple.** Its parser core has mean cyclomatic complexity
  6.7, against 8.2 for recursive descent. 2 functions are above 30, against 19.

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
  - Bare `git stash` was used across shared worktrees.
  - The Go build cache filled the disk at 19 GB.

  The worker brief now forbids pattern kills and bare stash.

## What Puzzles Us

- **The cold key changed meaning.** After CTORONCE the aggregate cold ratio fell
  from 4.99 to 2.72, but long-lived barely moved. "Cold" now reuses tables an
  earlier `New` built.
- **Garbage collection costs as much as the forest.** Of the chart side's CPU,
  lex takes 19%, recognize 45% and forest build 35%. GC marking is a further 35%
  of all samples, and most of it traces the forest's pointer graph.
- **The perl.git corpus is heavy.** It is 38% of the speed aggregate in the one
  non-quiet run that included it. It raised the aggregate before any forest
  existed.

## Where A Human Had To Act

**Collaboration, as designed.** perigrin made 29 rulings. Among them:

- the library-operator design;
- three w1 criterion repairs;
- the `isa` preamble;
- the toy Leo guard;
- the `pat_psycho.t` exclusion;
- "merge 566 when all checks clear";
- "Measure properly first".

Each of these was his call.

**Friction to engineer away:**

- Long closes outlived the 2-hour background limit and container restarts. They
  ran detached and were re-run by hand.
- The disk-full stop.
- GitHub rewrites the PR body footer on every PR.
- The Windows updater test timed out repeatedly and was rerun by hand.
- Stop-hook prompts to commit `test/e2e/.perl-version` during runs.
- Questions on the speed outlook had to be asked twice before a measurement
  settled them.

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
- **Design the forest for the collector.** Before another attempt, the forest
  should be index-based and pointer-free, and building it should reuse memory
  across files. That is the only lever the review found on the GC cost.
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
