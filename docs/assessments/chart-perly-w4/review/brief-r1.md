# crochet:review: chart-perly-w4 (chart precedence V2), round brief

You are one independent participant in crochet:review, the gate between execute and postmortem. You did not write this code, and you hold no view on it yet. A round airs views rather than deciding. Raise anything you see; classification comes next.

## Subject (it holds still for the round)

- **Repository:** /home/user/pvm. Work read-only there, or in a scratch worktree you create under `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/wt-<you>` and remove at the end. Do not edit, commit or push anything in the repository.
- **The delivery:** `git diff e3432c0f 8d8fff38`. That is 210 files, about +17.5k/−1.6k lines and 189 commits, merged to pu as PRs #655–#691.
  - e3432c0f is pu with V1, after the baseline retake.
  - 8d8fff38 is pu with V2 complete.
  - `git log e3432c0f..8d8fff38 --format='%(trailers:key=Implements,valueonly)'` lists the 28 issues.
- **The authorising decision:** `docs/plans/2026-11-25-chart-precedence-v2-design.md` (accepted revision 9).
  - Assessment minute: `docs/assessments/chart-precedence-v2.md`.
  - perigrin's rulings: `docs/assessments/chart-precedence-v2/rulings.md`. Only perigrin's words are rulings.
- **The milestone's record:** `git zhi milestone show chart-perly-w4 --format json | jq -r .body`. Its Records list every issue's gate, reopens and out-of-scope residuals.
- **The verdict reading:** `docs/assessments/chart-perly-w4/verdict/README.md` records `exit: met`. The early readings are kept beside it.
- **Project rules:** `/home/user/pvm/CLAUDE.md`, and `internal/chart/README.md`, which is the chart's live document and is test data under CLAUDE.md.

The shell's default `T1_CORPUS` is the wrong directory. Set `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit` explicitly on any test command you run.

## Design §7 stops

Any of these is a STOP finding for perigrin:
- a forest differs from V1's;
- an answer key moves, or a file's accept/refuse changes;
- a §5 feasibility test fails, or set 1626 is reached after `indirob ::= scalar`;
- something relies on an item fact other than the canonical spine top;
- the checks pass but the exit is unmet.

## What a round produces

Write a report file at the path your dispatch names. For every finding, give:
- **id:** your prefix and a number;
- **classification:** BLOCKING (the delivery should not be accepted until it is fixed) or RESIDUAL (real, but it holds no round open and goes to the postmortem or a follow-up issue);
- **where:** file and symbol;
- **evidence:** the commands you ran and what they showed;
- **grounds**, and a suggested remedy.

If you have no blocking findings, say so plainly. "No blocking findings" is a normal outcome, and it is worth recording.

Already known, so do not re-raise these as new; you may promote one to BLOCKING with grounds:
- four stale V1 code comments:
  - the ABOUTME in `internal/chart/precedence_test.go`;
  - `pipelineSource` in `waypoint2_refusals_test.go`;
  - `TestItemIsStateDistanceContext`;
  - `CandidateForest` and `CandidateToyForest` in `forest/equality.go`;
- the residuals listed in the milestone Records.

Your final reply is only the report path, plus one line: the number of findings by class.
