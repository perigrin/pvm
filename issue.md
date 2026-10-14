---
title: After a defined guard, a substr result keeps an unnamed type bit and fails a Str slot
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T21:56:08.046532878Z
updated: 2026-10-14T22:36:29.391078565Z
sessions:
- start_sha: 85bda4f9b766ba76e39cccddbd73509b57462ebd
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T22:36:29.391078565Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T22:36:29.391078565Z
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-2 re-measure, bucket `R2-DEFINED-UNION` (2026-10-14); evidence in `scratchpad/cmp-a44b/ev-r2.md`, `ev-r2b.md`, repros in `cmp-a44b/repro/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges, on the lowered path. 14 added lines (JSON::PP). `$ch` from `substr`, then `defined $ch and ($ch eq 'e' or ...)`: both binaries report `got Int|Num|Str|NaN|Inf|0x80000000`. The defined guard's narrowing leaves a bit with no name -- a type that cannot exist.

#### r2_defined_and_paren

```perl
my $ch = substr("1.e", 0, 1);
if (defined $ch and $ch eq '.') { if (!defined $ch or $ch !~ /\d/) { } }
if (defined $ch and ($ch eq 'e' or $ch eq 'E')) { print "exp\n" }
```

```perl
my $ch;
sub next_chr { $ch = substr("e", 0, 1) } next_chr();
if (defined $ch and ($ch eq 'e' or $ch eq 'E')) { print "exp\n" }
```
perl -w (rc=0): exp
old: 
new:

```perl
my $ch = substr("1.e", 0, 1);
if (defined $ch and $ch eq '.') { if (!defined $ch or $ch !~ /\d/) { } }
if (defined $ch and ($ch eq 'e' or $ch eq 'E')) { print "exp\n" }
```
perl -w (rc=0): 
old: 3:13: error: call to defined: argument 1 expects Scalar, got Int|Num|Str|NaN|Inf|0x80000000 [type-mismatch];3:22: error: left operand of "eq": expected Str, got Int|Num|Str|NaN|Inf|0x80000000 [coercion-mismatch];3:36: error: left operand of "eq": expected Str, got Int|Num|Str|NaN|Inf|0x80000000 [coercion-mismatch];
new: 3:13: error: call to defined: argument 1 expects Scalar, got Int|Num|Str|NaN|Inf|0x80000000 [type-mismatch];3:22: error: left operand of "eq": expected Str, got Int|Num|Str|NaN|Inf|0x80000000 [coercion-mismatch];3:36: error: left operand of "eq": expected Str, got Int|Num|Str|NaN|Inf|0x80000000 [coercion-mismatch];

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; goldens under `Diagnostics-changed:` trailers; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestDefinedGuardLeavesANamedType$' -count=1 -v | grep -q '^--- PASS: TestDefinedGuardLeavesANamedType'`)
