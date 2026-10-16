---
title: "[I5] Writes in an elsif branch to the variable its own guard narrows are dropped from the join"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.309787051Z
updated: 2026-10-16T01:16:07.309787051Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I5**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/infer.go:1109-1160` iterates `preIfTypes`, which only holds the outer `if` guard's variables.
- **Repro:** `logicB/e1.pl`:
  ```perl
  my $y = @ARGV ? 3 : {};
  if ($c) {} elsif (ref $y) { $y = undef }
  $y + 1
  ```
  - perl -w warns `uninitialized value $y`.
  - `psc check --warn-lossy-coercion` says `lossy coercion of Int|HashRef`: the Undef is lost, and by default psc is silent.
  - origin/pu also lacked Undef (`Int|HashRef`).
- **Fix:** collect pre-if types for every elsif guard's variables, and join over the union of keys.
- **Confidence:** High. **PRE-EXISTING.** **Issue:** none (01a132f4 is related). **Found by:** logic-B #9.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Writes in an `elsif` branch to the variable its own guard narrows are dropped from the join is fixed (`go test ./internal/infer/ -run '^TestElsifWriteJoins$' -count=1 -v | grep -q '^--- PASS: TestElsifWriteJoins'`)
