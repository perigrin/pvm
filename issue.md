---
title: "[I4] An assignment in one if-branch leaks into the else branch and the continuation"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.268819188Z
updated: 2026-10-16T01:16:07.268819188Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I4**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/flow.go:25-84` (ifStmt) and 260-280 (blockWithGuard). An unguarded branch is walked with no snapshot, so `UpdateType` writes through to the outer symbol.
- **Repro:** `v/eb4.pl`:
  ```perl
  my $n = 1; if (@ARGV) { $n = undef } else { print $n + 1 }
  ```
  - New psc: `eb4.pl:2:51: error: ... got Undef`.
  - origin/pu: silent.
  - perl prints 2.
  - `v/e3.pl` shows the elsif form: after `elsif ($c > 1) { $y = undef }`, `$y + 1` is reported as `got Undef`, not `Undef|Int`.
- **Fix:** snapshot the types before the `if`, restore them before each branch, and join the branch-end types afterwards.
- **Confidence:** High. **NEW.**
- **Issue:** none. 01a132f4-f274-7586-91d8-040f30879fe3 (branch join drops an else) is adjacent but different.
- **Found by:** errors-B #4.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] An assignment in one if-branch leaks into the else branch and the continuation is fixed (`go test ./internal/infer/ -run '^TestIfBranchWriteDoesNotLeakIntoElse$' -count=1 -v | grep -q '^--- PASS: TestIfBranchWriteDoesNotLeakIntoElse'`)
