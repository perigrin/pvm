---
title: "[I14] my ($a) = @arr types $a as the array's count (Int)"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.645066301Z
updated: 2026-10-16T01:16:07.645066301Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I14**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/ops.go:168-176` (`narrow`) applies the scalar count rule even when `d.ListAssign` is set.
- **Repro:** `v/eb2.pl`:
  ```perl
  my ($fh) = @ARGV; ... flock $fh, 1
  ```
  psc gives `error: call to flock: argument 1 expects Glob, got Int` (also `open`) on both builds; perl is silent. CPAN.pm:514 hits the same pattern.
- **Fix:** skip the count rule when the declaration is ListAssign; use the element type, or Scalar.
- **Confidence:** High. **PRE-EXISTING.** **Issue:** none (01a13318 is the `@`-sigil sibling). **Found by:** errors-B #2.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `my ($a) = @arr` types `$a` as the array's count (Int) is fixed (`go test ./internal/infer/ -run '^TestListAssignFromArrayIsElement$' -count=1 -v | grep -q '^--- PASS: TestListAssignFromArrayIsElement'`)
