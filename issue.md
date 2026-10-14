---
title: Builtin results whose type includes Undef rejected in a Str slot
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:14.526266585Z
updated: 2026-10-14T17:13:39.866081563Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T17:13:39.866081563Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:39.866081563Z
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
