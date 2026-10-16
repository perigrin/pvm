---
title: "[I6] A partial compound guard drops its operator, so negating it is unsound"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.345399321Z
updated: 2026-10-16T01:17:17.795812Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:17.795812Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.795812Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I6**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/flow.go:621-631` (`compound`). Consumed by narrowContinuation (:67-69), :80 and logicalRight (:303-308).
- **Repro:**
  - `logicB/c1b.pl`: `if (defined $x && $y) { exit 0 } print $x + 1` gives a new psc error `got Undef`; origin/pu is silent; perl prints 6.
  - `c2.pl` and `c3b.pl` (the `||` and positive forms) are false positives on both builds.
- **Fix:** keep the operator, as `Compound{Op, l, nil}`. flattenGuards then uses a nil side as "no information": `&&` keeps the known side only un-negated, and `||` only negated.
- **Confidence:** High. **NEW** through narrowContinuation (c1b); PRE-EXISTING for c2/c3b.
- **Issue:** 01a13c84-fb04-73c9-adc5-22efcd234f57 ("An or with one non-guard side narrows its then-branch") covers the `||` then-branch. This finding adds the negated `&&` continuation case (c1b), which is new in this PR.
- **Found by:** logic-B #4.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A partial compound guard drops its operator, so negating it is unsound is fixed (`go test ./internal/infer/ -run '^TestPartialCompoundGuardKeepsOperator$' -count=1 -v | grep -q '^--- PASS: TestPartialCompoundGuardKeepsOperator'`)
