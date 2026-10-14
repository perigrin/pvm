---
title: "`defined $our_var` after a branch join is typed a union including Void"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:26.46310058Z
updated: 2026-10-14T17:13:41.891002019Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T17:13:41.891002019Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:41.891002019Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N22-DEFINED-VOID-UNION`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 1 (1 added)  
**Likely fix site:** internal/infer/lowered_flow.go branch join (`ifStmt`/`blockWithGuard`: an unassigned path keeps the variable's own type, not Void)

Math::Complex:83 `$Inf = $DBL_MAX unless defined $Inf`. Legacy is silent.

**Repro `n22_defined_our.pl`:**
```perl
our ($Inf);
my $i = eval "9**9**9";
if (defined $i && $i > 1e308) { $Inf = $i }
$Inf = 1.79e308 unless defined $Inf;
```
perl -w (rc=0): 
old: 
new: 4:32: error: call to defined: argument 1 expects Scalar, got Boolean|Int|Num|Str|DualVar|Regex|ScalarRef|ArrayRef|HashRef|CodeRef|GlobRef|Object|Array|Hash|Code|Glob|NaN|Inf|IO|Void [type-mismatch];

Examples from the diff:

- `privlib/Math/Complex.pm:83` (added) `$Inf = $DBL_MAX unless defined $Inf;  # Oh well, close enough.` -- call to defined: argument 1 expects Scalar, got Boolean|Int|Num|Str|DualVar|Regex|ScalarRef|ArrayRef|HashRef|CodeRef|GlobRef|Object|Array|Hash|Code|Glob|NaN|Inf|IO|Void

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestBranchJoinNeverHoldsVoid$' -count=1 -v | grep -q '^--- PASS: TestBranchJoinNeverHoldsVoid'`)
