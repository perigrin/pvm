---
title: "[I18] A statement-modifier or or die guard does not narrow what follows it"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.800689132Z
updated: 2026-10-16T01:16:07.800689132Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I18**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/flow.go:25-37`.
- **Repro:** `v/eb7.pl`:
  ```perl
  die "x" unless defined $x; print $x + 1
  ```
  New psc gives `error: got Undef|Int`; origin/pu is silent.
- **Confidence:** High. **NEW in behaviour.**
- **Issue:** 01a13bb7-cbfa-786c-99e7-97bb229250fe ("A statement-modifier early exit does not narrow what follows it"). Covered; this finding adds the `defined $x or die;` spelling, which the issue does not name.
- **Found by:** errors-B #7.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A statement-modifier or `or die` guard does not narrow what follows it is fixed (`go test ./internal/infer/ -run '^TestModifierGuardNarrowsContinuation$' -count=1 -v | grep -q '^--- PASS: TestModifierGuardNarrowsContinuation'`)
