# chart-perly-w4 chain export

`milestone.json` holds the output of `git zhi milestone show chart-perly-w4 --format json`. `issues.json` holds the output of `git zhi issue list --milestone chart-perly-w4 --all --format json`.

This export was taken at close, on 2026-11-28.

## What the export holds

- The milestone's state is `completed`. The `--state complete` verify gate passed 210 of 210 acceptance criteria at pu `7aa30362`.
- There are 28 issues, all `done`. The 19 that chain-review left were joined in execution by nine more, each recorded in the milestone's Records:
  - follow-ups that perigrin ruled on (for example `01a21945-d1e6`, `01a21945-d35d`, `01a21a56-1251`);
  - `01a21db4-4792`, the `forest.Build` walk fix that the early speed reading made necessary.
- Each issue's body carries its gate reopens and any amendments that perigrin ruled on.
- The milestone's Records list:
  - every gate, by runner directory and commit;
  - every reopen and ruling;
  - the two early speed readings, kept in `../early-reading/` and `../early-rereading/`;
  - the formal verdict, kept in `../verdict/`, which reads `exit: met`;
  - the crochet:review minute.

## The review record

The crochet:review record is in `../review/`. It holds:
- the round briefs;
- each participant's report for rounds 1 and 2;
- the `git zhi verify` log at 8d8fff38 (209 of 210; the one regression was fixed by #692);
- the close gate's log.

## Related documents

- The milestone delivers the accepted decision `docs/plans/2026-11-25-chart-precedence-v2-design.md`.
- Its assessment minute is `docs/assessments/chart-precedence-v2.md`.
- perigrin's rulings are in `docs/assessments/chart-precedence-v2/rulings.md`.

## Earlier exports

This file and its two JSON files replace the exports taken after refinement (`c7157c2e`) and after chain-review (`73f94605`). Those versions remain in git history.

The chain itself, `refs/zhi/*`, cannot be pushed from these sessions, because the git proxy refuses the push with a policy 403. These files are the durable copy.
