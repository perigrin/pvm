---
title: "[I9] A compound assignment narrows its target to the right side's type"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.456818206Z
updated: 2026-10-16T01:16:07.456818206Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I9**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/ops.go:119-139`.
- **Repro:**
  - `logicB/a1.pl`: `$k = 4; $k //= undef; $k + 1` gives `error: got Undef`.
  - `logicB/a2.pl`: `$s .= <STDIN>; $s . "x"` gives `error: got Undef`.
  - perl is silent on both, and origin/pu gives the same output.
  - `$s .= 1; $s + 1` is a false negative: perl warns `isn't numeric`.
- **Fix:** narrow `op=` to the CORE.pmt return type of `op`. For `//=` and `||=`, use `Join(old &^ Undef, R)`; for `&&=`, use `Join(old, R)`.
- **Confidence:** High. **PRE-EXISTING.**
- **Issue:** 01a1335d-97c8-7bde-940f-99faca2aaf34 ("A compound assignment narrows to its right side's type"). This finding adds the `//=`/`||=`/`&&=` rules and the Undef false positives; the issue only names `.=` and `x=`.
- **Found by:** logic-B #8.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A compound assignment narrows its target to the right side's type is fixed (`go test ./internal/infer/ -run '^TestCompoundAssignNarrowsToOperatorResult$' -count=1 -v | grep -q '^--- PASS: TestCompoundAssignNarrowsToOperatorResult'`)
