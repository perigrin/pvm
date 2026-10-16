---
title: "[I8] $x // d and $x || d keep the left operand's Undef"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.419094347Z
updated: 2026-10-16T01:16:07.419094347Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I8**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/ops.go:57-58`: `Join(valueOf(l), valueOf(r))` for every logical op.
- **Repro:**
  - `logicB/o1.pl`: `my $v = $x // 0; $v + 1` and `my $w = $x || 0; $w + 1` give new psc errors `got Undef|Int` on lines 4 and 6. origin/pu is silent on both, and perl prints 4 and 4.
  - `v/o2.pl`: `my $v = $o || 0; length($v)` gives `error: call to length: argument 1 expects Str, got Undef|Int` plus a "hint: Add guard" on the most common default idiom in Perl.
- **Fix:** for `//`, `||` and `or`, use `Join(valueOf(l) &^ Undef, valueOf(r))`.
- **Confidence:** High. **NEW** in the variable case. **Issue:** none. **Found by:** logic-B #7.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `$x // d` and `$x || d` keep the left operand's Undef is fixed (`go test ./internal/infer/ -run '^TestDefinedOrDropsLeftUndef$' -count=1 -v | grep -q '^--- PASS: TestDefinedOrDropsLeftUndef'`)
