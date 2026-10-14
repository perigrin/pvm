---
title: Builtin results whose type includes Undef rejected in a Str slot
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:14.526266585Z
updated: 2026-10-14T20:56:17.983620517Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: 3d565dc4b5aab40dd9abbfef2f37b4dc67957d68
  commits: 3
  started_at: 2026-10-14T17:13:39.866081563Z
  ended_at: 2026-10-14T20:56:17.983620517Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:39.866081563Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T20:56:17.983620517Z
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

- found by: cutover 1 (01a0c115-a44b) bucket `N14-UNDEF-UNION-RET`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 3 (3 added)  
**Likely fix site:** CORE.pmt (`getsockopt` `Str|Undef`, `prototype`) and lowered_ops.go `checkOperand`/`checkArg` (same root as N6)

`unpack "i", getsockopt(...)`, `prototype($x) =~ ...` guarded by `defined prototype($x)`, Cmd.pm `system(ref $cmd ? @$cmd : $cmd)`. Legacy fires on n36, the lowered path adds n14.

**Repro `n14_unpack_getsockopt.pl`:**
```perl
use Socket;
socket(my $s, PF_INET, SOCK_STREAM, 0) or die;
my $sndbuf = unpack 'i', getsockopt($s, SOL_SOCKET, SO_SNDBUF);
print $sndbuf > 0 ? "ok\n" : "no\n";
```
perl -w (rc=0): ok
old: 3:37: error: call to getsockopt: argument 1 expects Glob, got Scalar [type-mismatch];
new: 3:37: error: call to getsockopt: argument 1 expects Glob, got Scalar [type-mismatch];3:26: error: call to unpack: argument 2 expects Str, got Undef|Int|Num|Str|NaN|Inf [type-mismatch];

**Repro `n36_defined_prototype.pl`:**
```perl
my $builtinname = "CORE::push";
print "p\n" if defined prototype($builtinname) && prototype($builtinname) =~ /^;?\*/;
```
perl -w (rc=0): 
old: 2:51: error: left operand of "=~": expected Str, got Undef|Int|Num|Str|NaN|Inf [coercion-mismatch];
new: 2:51: error: left operand of "=~": expected Str, got Undef|Int|Num|Str|NaN|Inf [coercion-mismatch];

Examples from the diff:

- `privlib/B/Deparse.pm:2512` (added) `&& prototype($builtinname) =~ /^;?\*/` -- left operand of "=~": expected Str, got Undef|Int|Num|Str|NaN|Inf
- `privlib/IPC/Cmd.pm:1737` (added) `system( ref $cmd ? @$cmd : $cmd ) == 0 or do {` -- call to system: argument 1 expects Str, got Undef|Boolean|Int|Num|Str|DualVar|Regex|ScalarRef|ArrayRef|HashRef|CodeRef|GlobRef|Object|Array|NaN|Inf
- `T1:unit/nonregex_baseline_closure.t:26` (added) `my $sndbuf = unpack 'i', getsockopt($server, SOL_SOCKET, SO_SNDBUF);` -- call to unpack: argument 2 expects Str, got Undef|Int|Num|Str|NaN|Inf

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestUndefUnionBuiltinResultInStrSlot$' -count=1 -v | grep -q '^--- PASS: TestUndefUnionBuiltinResultInStrSlot'`)
