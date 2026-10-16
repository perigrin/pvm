---
title: "[S4] psc parse and psc analyze parse without a loader or @INC, so they show a different tree from the one check ..."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.624725964Z
updated: 2026-10-16T01:16:08.624725964Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S4**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

`psc parse` and `psc analyze` parse without a loader or @INC, so they show a different tree from the one check analysed.
  - Where: `parse_command.go:35`, `analyze_command.go:69`.
  - Repro: `probe/p5.pl` gives a `length` diagnostic in check, while `psc parse --tree lowered` prints `Hole` for that statement.
  - High. NEW (the lowered printer). Found by: contract-B #4.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `psc parse` and `psc analyze` parse without a loader or @INC, so they show a different tree from the one check analysed is fixed (`go test ./internal/psc/ -run '^TestParseCommandResolvesImports$' -count=1 -v | grep -q '^--- PASS: TestParseCommandResolvesImports'`)
