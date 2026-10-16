---
title: "[I16] An array in a builtin's $ slot is checked as Array, not as its count"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.721727634Z
updated: 2026-10-16T01:16:07.721727634Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I16**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/calls.go:196-210`.
- **Repro:** `v/eb5.pl`: `int(rand(@m))` gives `error: rand ... expects Num, got Array` on both builds; perl is silent.
- **Confidence:** High. **PRE-EXISTING.**
- **Issue:** 01a119df-a205-7015-a828-3f9a3d9db331 ("An array written in a $ slot is checked as Array, not as its count"). Covered; this finding adds nothing.
- **Found by:** errors-B #5.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] An array in a builtin's `$` slot is checked as Array, not as its count is fixed (`go test ./internal/infer/ -run '^TestArrayInScalarSlotIsCount$' -count=1 -v | grep -q '^--- PASS: TestArrayInScalarSlotIsCount'`)
