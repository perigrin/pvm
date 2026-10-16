---
title: "[S7] .5 and v-strings lower to a Hole, and v65 lowers to a Bareword."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.742417465Z
updated: 2026-10-16T01:16:08.742417465Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S7**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

`.5` and v-strings lower to a Hole, and `v65` lowers to a Bareword.
  - Where: `lower/literals.go:296,305`.
  - Repro: `my $x = .5;` and `my $v = v5.36.0;` each give `Init:(Hole)`.
  - 01a0c13f-97f5-7f98-b32d-07245ec6ddfe (done) fixed the parser side only.
  - High. NEW (the lowering). Found by: logic-A #9, contract-A #5.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `.5` and v-strings lower to a Hole, and `v65` lowers to a Bareword is fixed (`go test ./internal/lower/ -run '^TestLowerLeadingDotAndVStrings$' -count=1 -v | grep -q '^--- PASS: TestLowerLeadingDotAndVStrings'`)
