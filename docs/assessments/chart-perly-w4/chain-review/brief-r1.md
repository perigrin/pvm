# crochet:chain-review brief for chart-perly-w4, round 1

You are one lens in crochet:chain-review. The skill files are in `/root/.claude/plugins/cache/perigrin-marketplace/crochet/0.9.0/skills/`:

- read `chain-review/chain-review.md` first;
- then read your own lens: `alignment/alignment.md` or `pushback/pushback.md`;
- then read `discernment/discernment.md`, the sections "What a round produces" and the release rule in "The loop".

## The subject

The subject is the chain for milestone `chart-perly-w4`. Read it live from `/home/user/pvm/.claude/worktrees/v2-refine`:

- `git zhi milestone show chart-perly-w4 --format json`;
- `git zhi issue list --milestone chart-perly-w4 --format json`;
- `git zhi list`.

There are 17 issues. The milestone body's Records section holds the residual-to-issue mapping and the open items refinement passed on.

## The spec

- **Design:** `docs/plans/2026-11-25-chart-precedence-v2-design.md`, accepted at revision 9.
- **Minute:** `docs/assessments/chart-precedence-v2.md`. It has 66 residual rows that the chain must carry or explicitly decline.
- **Rulings:** `docs/assessments/chart-precedence-v2/rulings.md`. They bind.
- **perigrin's goal requires:**
  - work in three stages, in order: the six feasibility counters as real tests, then the forest-equality differentials, then the recognizer;
  - a model label on every issue;
  - criteria keyed on the PR set or commit range, never on issue ids;
  - criteria red before the work and green after;
  - one owner per README section, with a final docs issue;
  - live docs changed first;
  - the stops in design §7 reported, never tuned toward.

## Where the authors are likely blind

The authors were four refinement subagents. The orchestrator then answered the assessment, so it is not neutral here. Look hardest at four places:

1. **Criteria that pass vacuously.** The SQE already reported nine. They are in Records. Judge each one, and look for more.
2. **Issue sizing.** For example, `d199` has 18 criteria and covers the reduce side, refusal records, Leo and the rerun in one issue. Is that one session?
3. **Ordering.** Is any issue placed before something it depends on? Can each intermediate state close with `make test` green?
4. **Hidden dependencies on another repository or on absent corpora.** perl.git is absent in CI.

Mechanical check already done: I extracted all 162 criterion commands and syntax-checked them with `sh -n`, and none failed. The `git zhi verify --dry-run` check lists only done issues, so it is empty today.

## Rules

- Re-derive every claim from the live chain. Do not trust this brief's description of it.
- Classify each finding as blocking or residual, using discernment's definitions. A defect in what an issue asserts counts as blocking, for example a criterion that cannot pass, or one that passes before its work exists.
- Make no edits to the chain or to the repository.
- Do not push, and never run `git config`.

## Report

Write your report to `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/r1-<alignment|pushback>.md`. Include:

- your recommendation: reject, modify or accept;
- a findings table with columns id, blocking/residual, issue id, claim, consequence, evidence, and suggested action;
- what you did not examine.

Use id prefix `AL` for alignment and `PB` for pushback.

Your final reply is only the path, the recommendation and a one-line count.
