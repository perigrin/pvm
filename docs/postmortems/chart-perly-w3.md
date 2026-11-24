# Milestone chart-perly-w3 Postmortem

The milestone replaced the chart's precedence judgement with state-free tables
and checked the result against the old judgement (decision
`2026-11-23-chart-stateless-precedence.md`, revision 9). It delivered 12 issues
in 12 PRs (#629–#640) between 2026-11-23 and 2026-11-24.

**The exit was met.** The verdict reading (#637) took three quiet runs before
and three after, interleaved B A B A B A. `speed-ratio-aggregate-longlived` fell
from a median of 5.17 (range 5.12–5.21) to 3.53 (3.48–3.56), and the precedence
stage fell from 2.34 to 0.68. All 29 correctness and count keys were identical
across the six runs, and no Stop fired.

## What Worked Well

- **Results-first held, and it found real defects.** A frozen referee
  (`ConstrainReference`), bison 3.8.2 on toy grammars, and four differentials
  (toys, generated sentences, the in-tree corpus, and perl.git with 617 files)
  found three things no criterion had asked about:
  - the referee's own `shifts` defect (Q1);
  - a gap in the generation guard (round 2, F1);
  - a nullable-shift limit shared by both judgements (G1).

  Each was confirmed against bison before anyone acted on it.
- **Differentials before evaluation.** Chain-review ruled that FOREST, CHART and
  CORPUS land before EVAL (PB1). Because of that, the central change (#636) was
  checked against three green guards on its first run.
- **Gates caught what workers missed.** Mutation probes in the gates found gaps
  in tests that the workers' own suites had passed:
  - GEN: the shift-side guard was untested;
  - EVAL: the parent-item memo key was untested;
  - PERLGITREF: the CI time was missing from the PR.

  Every one was fixed before merge. EVAL's sanbao difficulty was 0.088, so the
  sizing held.
- **Parallelism worked.** GEN and the three PRECDIFF issues ran together. They
  needed four rebases, all clean apart from README conflicts. The range-diffs
  showed that every code commit was unchanged.

## What Didn't Work

- **The live document was assembled by rebase.** Nine workers wrote
  `internal/chart/README.md` "The precedence constraint" in nine PRs, and
  conflicts were resolved by stacking paragraphs. That section drew most of the
  review's findings:
  - round 1: C1–C5;
  - round 2: C6–C9 and F2–F3;
  - round 3: C10, Q13–Q14 and N1.

  Most were claims that held for one package and were stated for all of them.
- **Review grew the milestone by three issues.** REVIEWFIX (#638), REVIEWFIX2
  (#639) and REVIEWFIX3 (#640) took about 40% of the elapsed time after
  execution. Two of them changed production-adjacent text or code.
- **Rule breaks by agents.** All were recorded and none stopped the milestone:
  - Worker 5 skipped `--state start` and `/simplify`.
  - Runner 3 ran `git checkout` in the shared checkout.
  - The orchestrator closed PRECDIFF-CORPUS with the shell's wrong `T1_CORPUS`, so 2 of its 7 criteria went unverified.
  - The orchestrator removed worker 10's live worktree, judging it "merged" because its HEAD was an ancestor of pu.
- **Infrastructure cost runs.** The losses were:
  - Two container restarts killed long ratchet runs (EVAL criteria, run 10).
  - One Haiku runner stalled for 47 minutes extracting commands. A driver script replaced extraction from then on.
  - An `internal/updater` network timeout failed Windows CI on #632, #633 and #634, and each needed a re-run.
  - Free disk fell to 2.6 GB because the Go build cache reached 18 GB.

## What Puzzles Us

- The reviewers agreed with each gate on every code finding, yet each review
  round produced new documentation findings at the same rate. That suggests the
  README section is too long for its subject: three harnesses, two judgements,
  two pinned limits and their caveats.
- `NewPrecedence` now costs about 0.8 s per process, roughly double the base,
  and #639 added another 10–15%. Nothing gates it, and the verdict does not see
  it, because the cost sits outside the timed span.

## Where A Human Had To Act

- **Rulings that were perigrin's:**
  - the CI wall time for STAGEKEYS (local timing accepted);
  - Q1 ("Pin and document"), after asking "Define production here?";
  - the scope of round 1;
  - F1 ("Fix it now");
  - G1 ("Check perly.y, then pin").

  These were collaboration, not friction. Each one changed what the referee
  means or what V1 covers.
- **Friction: three Bash permission prompts**, each of which stopped a subagent
  until perigrin answered (worker 6, gate 3, the round-3 quality lens). These
  were read-only or scratch-worktree commands. An allowlist for
  `go test`/`git worktree`/`bison` in scratch paths would have let them proceed.
- **Friction: the proxy-rewritten PR footer.** It could not be fixed from the
  session. Recorded once and then left.

## What Will We Change

- **One owner for a shared live-document section.** When three or more issues
  edit the same README section, refinement adds a final docs issue that
  rewrites the section whole. The other issues append bullets to a scratch file
  instead of editing the section.
- **Runners execute a prepared driver, never extract commands themselves.** Use
  `run-10-driver.sh`: one command per line, resumable from criterion *n*. Make
  it part of the runner brief. It survived two container restarts.
- **Export the corpus variables in every orchestrator close.** Run each
  `--state done` and `--state complete` with `T1_CORPUS` exported, and with
  `--timeout 1800` where the subcommand takes it. The shell's preset is wrong
  for this repository.
- **Never judge a worktree disposable by ancestry alone.** A worktree that has
  just started is an ancestor of pu too. Check the chain for an in-progress
  issue on that branch before removing it.
- **Gate the generation cost.** Add a `NewPrecedence` per-process timing key
  before V2 adds generation work.
- **Carry the follow-ups into the next refinement.** These have not been built:
  - the referee's `shifts` defect, and the G1 limit (decision §6);
  - pruning counts (A1, Q11);
  - test dedup (P1–P3, P8, P10–P14);
  - the baseline retake (Q6);
  - the stage pin's dependence on the referee's defect (Q12).

---
Generated by crochet:postmortem on 2026-11-24.

Data sources:
- issue list (12 done);
- milestone Records;
- PR commit counts (#629–#640: 5, 2, 3, 3, 2, 4, 5, 4, 1, 4, 4, 3);
- sanbao difficulty (0.037–0.161, all below 0.4);
- the full verify at 2a29efd2e (84 passed, plus one make test timeout under the default limit);
- docs health (0 docs observed, 41 coverage gaps: it watches no live document here);
- the review reports from rounds r1 to r3.
