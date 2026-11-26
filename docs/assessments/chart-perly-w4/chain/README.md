# chart-perly-w4 chain export

`milestone.json` and `issues.json` hold the output of `git zhi milestone show chart-perly-w4 --format json` and `git zhi issue list --milestone chart-perly-w4 --all --format json`.

This export was taken on 2026-11-26, after crochet:chain-review. Chain-review split two issues, so the chain now has 19, and it edited criteria across the chain. The milestone's Records list every change and its cause. The full chain-review record is in `../chain-review/`:

- the briefs;
- every participant's round report;
- the repair log for each round.

Chain-review reached its bound of eight rounds without a fixed point. It closed by perigrin's ruling "Apply fixes, accept (Recommended)", recorded in `docs/assessments/chart-precedence-v2/rulings.md`.

The milestone delivers the accepted decision `docs/plans/2026-11-25-chart-precedence-v2-design.md`. Its assessment minute is `docs/assessments/chart-precedence-v2.md`.

The chain itself, `refs/zhi/*`, cannot be pushed from these sessions: the git proxy refuses the push with a policy 403. These files are the durable copy. They are exported again at close.
