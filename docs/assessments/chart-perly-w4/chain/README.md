# chart-perly-w4 chain export

`milestone.json` and `issues.json` are the output of `git zhi milestone show chart-perly-w4 --format json` and `git zhi issue list --milestone chart-perly-w4 --all --format json`. They were exported on 2026-11-25, after crochet:refinement and before chain-review.

The milestone delivers the accepted decision `docs/plans/2026-11-25-chart-precedence-v2-design.md`. Its assessment minute is `docs/assessments/chart-precedence-v2.md`, and its rulings are in `docs/assessments/chart-precedence-v2/rulings.md`. The minute is not renamed to the milestone, because the decision and perigrin's goal cite these paths.

The chain itself (`refs/zhi/*`) cannot be pushed from these sessions: the git proxy refuses it with a policy 403. These files are the durable copy. They are exported again after chain-review and at close.
