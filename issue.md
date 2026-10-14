---
title: "`fileno`/`read` results typed Int|Undef in every slot"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:55.666144365Z
updated: 2026-10-14T20:56:00.85582515Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: 3d565dc4b5aab40dd9abbfef2f37b4dc67957d68
  commits: 3
  started_at: 2026-10-14T17:13:39.227410558Z
  ended_at: 2026-10-14T20:56:00.85582515Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:39.227410558Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T20:56:00.85582515Z
observed_paths:
- internal/infer/lowered.go
- internal/infer/lowered_calls.go
- internal/infer/lowered_calls_test.go
- internal/infer/lowered_flow.go
- internal/infer/lowered_flow_test.go
- internal/infer/lowered_ops.go
- internal/infer/lowered_vars.go
- internal/infer/lowered_vars_test.go
- internal/infer/symbols.go
- internal/infer/testdata/lowered/flow/and.golden
- internal/infer/testdata/lowered/flow/and_early_return.golden
- internal/infer/testdata/lowered/flow/and_keyword.golden
- internal/infer/testdata/lowered/flow/and_partial.golden
- internal/infer/testdata/lowered/flow/and_two_vars.golden
- internal/infer/testdata/lowered/flow/blessed_and_defined.golden
- internal/infer/testdata/lowered/flow/not_and.golden
- internal/infer/testdata/lowered/flow/or.golden
- internal/infer/testdata/lowered/flow/or_else.golden
- internal/parse/corenoproto_test.go
- internal/parse/coretypes_test.go
- internal/parse/declarations/CORE.pmt
- internal/types/types.go
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N6-UNDEF-INT-RET`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 11 (11 added)  
**Likely fix site:** internal/parse/declarations/CORE.pmt (`fileno` `Int|Undef`, `read` `Int|Undef`) and lowered_ops.go `checkOperand` (Undef from a builtin's failure return is not a coercion of the success value)

`fileno(STDOUT).">"`, `vec($rin, fileno($fh), 1) = 1`, `fileno STDERR != fileno STDOUT`, `read(...) > 0`. perl -w is silent on an open handle. Legacy fires on the repros.

**Repro `n6_undef_int_ret.pl`:**
```perl
my $fd = fileno(STDOUT) . ">";
my $rin = ""; vec($rin, fileno(STDIN), 1) = 1;
print "$fd\n";
```
perl -w (rc=0): 1>
old: 1:10: error: left operand of ".": expected Str, got Undef|Int [coercion-mismatch];2:25: error: call to vec: argument 2 expects Int, got Undef|Int [type-mismatch];
new: 1:10: error: left operand of ".": expected Str, got Undef|Int [coercion-mismatch];2:25: error: call to vec: argument 2 expects Int, got Undef|Int [type-mismatch];

**Repro `n30_fileno_compare.pl`:**
```perl
print "differ\n" if defined fileno STDERR && fileno STDERR != fileno STDOUT;
```
perl -w (rc=0): differ
old: 1:46: warning: left operand of "!=": expected Num, got Undef|Int [coercion-mismatch];1:63: warning: right operand of "!=": expected Num, got Undef|Int [coercion-mismatch];
new: 1:46: warning: left operand of "!=": expected Num, got Undef|Int [coercion-mismatch];1:63: warning: right operand of "!=": expected Num, got Undef|Int [coercion-mismatch];

**Repro `n31_read_gt_zero.pl`:**
```perl
open my $fh, "<", $0 or die; my $buf;
1 while read($fh, $buf, 4096) > 0;
```
perl -w (rc=0): 
old: 2:14: error: call to read: argument 1 expects Glob, got Scalar [type-mismatch];
new: 2:14: error: call to read: argument 1 expects Glob, got Scalar [type-mismatch];2:9: warning: left operand of ">": expected Num, got Undef|Int [coercion-mismatch];

Examples from the diff:

- `privlib/CPAN/Tarzip.pm:83` (added) `while read($fhw,$buffer,4096) > 0 ;` -- left operand of ">": expected Num, got Undef|Int
- `privlib/IPC/Cmd.pm:1688` (added) `fileno(STDOUT).'>',` -- left operand of ".": expected Str, got Undef|Int
- `privlib/IPC/Cmd.pm:1690` (added) `fileno(STDERR).'>',` -- left operand of ".": expected Str, got Undef|Int

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestFilenoAndReadResultsInStrAndNumSlots$' -count=1 -v | grep -q '^--- PASS: TestFilenoAndReadResultsInStrAndNumSlots'`)
