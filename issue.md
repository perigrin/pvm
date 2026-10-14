---
title: "`<=>` typed Undef|Int in arithmetic"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:19.743231735Z
updated: 2026-10-14T17:01:19.743231735Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N19-SPACESHIP-UNDEF`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 2 (2 added)  
**Likely fix site:** operator result type for `<=>` (Undef only for a NaN operand)

Math::Complex `$sgn * ($re1 <=> $re2)`. Legacy fires too; 2 lines added.

**Repro `n19_spaceship_undef.pl`:**
```perl
my ($sgn, $re1, $re2) = (1, 2, 3);
my $c = $sgn * ($re1 <=> $re2);
```
perl -w (rc=0): 
old: 2:17: error: right operand of "*": expected Num, got Undef|Int [coercion-mismatch];
new: 2:16: error: right operand of "*": expected Num, got Undef|Int [coercion-mismatch];

Examples from the diff:

- `privlib/Math/Complex.pm:623` (added) `return $sgn * ($re1 <=> $re2) if $re1 != $re2;` -- right operand of "*": expected Num, got Undef|Int
- `privlib/Math/Complex.pm:624` (added) `return $sgn * ($im1 <=> $im2);` -- right operand of "*": expected Num, got Undef|Int

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestSpaceshipInArithmeticIsSilent$' -count=1 -v | grep -q '^--- PASS: TestSpaceshipInArithmeticIsSilent'`)
