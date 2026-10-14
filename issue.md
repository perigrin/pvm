---
title: "`defined $x and $x eq ...` does not narrow $x in the right operand"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:52.622627243Z
updated: 2026-10-14T17:13:41.451868239Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T17:13:41.451868239Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:41.451868239Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N5-DEFINED-AND`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 20 (20 added)  
**Likely fix site:** internal/infer/lowered_ops.go `binary` for `and`/`&&` (walk the right operand under the left operand's guard, as lowered_flow.go `guard` does for if)

JSON::PP's `$ch` (25 lines): `if (defined $ch and $ch eq "}")`. perl short-circuits, so the right side never sees undef. Legacy fires on the repro too. n5_undef_narrow2 shows the second half: after `if (defined $ch)` the variable is narrowed to "every defined type", which then fails a Str slot (`Undef|0x80000000` and `Int|Num|Str|NaN|Inf|0x80000000` in PP.pm also print an unnamed type bit).

**Repro `n5_undef_narrow3.pl`:**
```perl
my $ch = undef;
if (defined $ch and $ch eq '}') { print "close\n" }
```
perl -w (rc=0): 
old: 2:21: error: left operand of "eq": expected Str, got Undef [coercion-mismatch];
new: 2:21: error: left operand of "eq": expected Str, got Undef [coercion-mismatch];

**Repro `n5_undef_narrow2.pl`:**
```perl
my $ch;
if (defined $ch) { if (defined $ch and $ch eq '}') { print "close\n" } }
```
perl -w (rc=0): 
old: 2:40: error: left operand of "eq": expected Str, got Boolean|Int|Num|Str|DualVar|Regex|ScalarRef|ArrayRef|HashRef|CodeRef|GlobRef|Object|NaN|Inf [coercion-mismatch];
new: 2:40: error: left operand of "eq": expected Str, got Boolean|Int|Num|Str|DualVar|Regex|ScalarRef|ArrayRef|HashRef|CodeRef|GlobRef|Object|NaN|Inf [coercion-mismatch];

Examples from the diff:

- `privlib/JSON/PP.pm:1009` (added) `if(defined $ch and $ch eq ']'){` -- left operand of "eq": expected Str, got Undef|Int|Num|Str|NaN|Inf
- `privlib/JSON/PP.pm:1046` (added) `$at-- if defined $ch and $ch ne '';` -- left operand of "ne": expected Str, got Undef
- `privlib/JSON/PP.pm:1062` (added) `if (!defined $ch or $ch ne ')') {` -- left operand of "ne": expected Str, got Undef

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestDefinedAndNarrowsRightOperand$' -count=1 -v | grep -q '^--- PASS: TestDefinedAndNarrowsRightOperand'`)
