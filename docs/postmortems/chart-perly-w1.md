# Milestone chart-perly-w1 Postmortem

**Verdict: not killed at waypoint 1.** The chart recognizes the whole superset population (1,394 files) and every gated count is 0. Two quiet readings at 52b8900a:

| Key | Reading 1 | Reading 2 | Limit |
|---|---|---|---|
| lex-overhead | 1.155 | 1.149 | 1.3 |
| long-lived ratio, std5k | 1.22 | 1.15 | 3 |
| long-lived ratio, aggregate | 2.11 | 2.24 | 3 |

The cold aggregate ratio is 4.85 to 4.90. Its kill condition is not tested, and it is inflated by per-file recognizer construction (see below).

## What Worked Well

- **Decomposition held.** 50 refined issues landed without a split or a cancellation. The other 19 were added under rulings or review: LEXSPEED, LEXGATE, T1VENDOR, nine review fixes and others. Median commits per issue: 2. Mean: 3.86.
- **Gates caught real defects before merge.** 15 issues hit the third review reopen; perigrin authorised a fourth once (package, 01a18fa7-0536). Gate findings included over-acceptances against `perl -c` in nearly every family port, a false Leo example in REVFIX3, and a wrong commit count in the PerlOnJava README.
- **Mutation-checked tests.**
  - TestUnfittedRulesMatchPerlyY holds all 349 rules to perly.y and found no drift.
  - The cache tests count work instead of timing it.
  - Every new guard was shown to go red on its mutation.
- **The fresh-reader seat paid for itself.** Each round's cold reader found something the resumed readers could not:
  - round 4: untested claims (F1–F3);
  - round 5: commitlint violations and a test that skipped silently on CI;
  - round 6: the quadratic transition cache hidden by the 10% exclusion.

## What Didn't Work

- **The review took 8 rounds, the bound.** Resumed readers converged by round 3. Every later round reopened only because a new lens arrived. The lenses were serial, not planned.
- **Criteria that pass before the work.** I wrote at least five criteria that would have passed on the unchanged tree: REVFIX2 ×2, REVFIX3, T1CHART, VERDICTDOC. Each was caught by hand before filing, but nothing checks for this.
- **Environment assertions masquerading as rulings.**
  - The worker brief said the commit-msg hook was "deliberately not installed". That was my assumption, so 91 of 274 commit headers failed commitlint, and fixing them needed a history rewrite.
  - My first probe of the hook misread it as passing everything: `pre-commit run --commit-msg-filename` does not feed it.
- **Out-of-tree corpus.** `make test` failed on any machine without `T1_CORPUS` (C1). It was found only at milestone review, then fixed by copying PerlOnJava in-tree (PR #554).
- **Scratch evidence was lost.** The checkpoint-2 quiet readings and `quiet2.sh` disappeared from the scratchpad, and the final verdict had to be re-measured. Container restarts killed jobs 8 times.

## What Puzzles Us

- **The 10% rule hid the worst case.** It removes the largest files, and those are where an O(n²) term shows. T1/code_too_large.t read 8.5× before CACHEFIX and 2.86× after; no key had reported it.
- **The allocation key measures construction, not per-token work.**
  - The cold key says 25.7 per token.
  - Warm, the chart allocates 2.68 per token against RD's 3.62.
  - recognize.New rebuilds grammar-only tables for every file, at 1,672 allocations.
- **A large "chart beats RD" count is mostly a measurement artifact.** On 165 files the chart accepts what RD refuses, but the ratchet runs RD without import resolution. Clean grammar wins: 2.

## Where A Human Had To Act

**Collaboration, as designed.** 64 rulings, including:
- three amendments to lex-overhead's definition, the third on the LSP buffers;
- the `x` narrowing deferred to waypoint 2;
- the in-tree corpus;
- both cache fixes.

These were perigrin's calls.

**Friction to engineer away:**
- A disk-full stop (the Go build cache had reached 21 GB).
- GraphQL blocked for `gh`. PR bodies get their attribution line rewritten by the environment.
- Stop-hook prompts to push during legitimate pre-push `make test` waits.
- Questions phrased in session jargon. perigrin asked for plain explanations.
- One gate reviewer used `-c core.hooksPath` for a read-only status, which breached the goal.

## What Will We Change

1. **Plan review lenses up front.** Run sceptic, tests-honest, PR-maintainer and cache/state-correctness as round-1 participants, not one new lens per round. Expected effect: convergence in 3–4 rounds.
2. **Mechanically reject vacuous criteria.** Refinement and the orchestrator run each new criterion against the base tree and require it to fail, except `unverifiable:` and negative-invariant lines, before `git zhi issue add`.
3. **Probe the environment, don't assert it.** The orchestrator's first step checks:
   - the installed hooks, including commit-msg;
   - that corpora are reachable with no environment variables set;
   - free disk.

   Worker briefs state only what was probed.
4. **Report what an exclusion hides.** Add a reported key for the maximum long-lived ratio over the files the 10% rule excludes. Waypoint 2 inherits this as a refinement input.
5. **Split construction from per-token cost.** Report recognize.New's allocations separately, or measure allocations warm. Build grammar-only recognizer tables once.
6. **Keep evidence in the Records, not the scratchpad.** Quiet readings record their commit and are pasted into Records when taken.
7. **Ask in plain language.** Every AskUserQuestion states the problem and its consequence before any finding ID.

---
Generated by crochet:postmortem on 2026-11-08
Data sources: milestone telemetry (`git zhi milestone show`: 69/69 done, mpg 3.86, buffer 28.5/34.5, fever GREEN), issue sessions (266 commits), milestone Records (963 bullets), milestone-review minutes (rounds 1–8). Not used: `git zhi sanbao`, which did not return within 2 minutes; `git zhi docs health`, which observes no chart document; and a full `git zhi verify` at the final head (per-issue verifies ran at each close).
