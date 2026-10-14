---
title: A user sub returning fileno's result keeps its Undef in a numeric slot
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T21:56:13.17283603Z
updated: 2026-10-14T22:36:30.014316095Z
sessions:
- start_sha: 85bda4f9b766ba76e39cccddbd73509b57462ebd
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T22:36:30.014316095Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T22:36:30.014316095Z
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-2 re-measure, bucket `R2-USER-FILENO` (2026-10-14); evidence in `scratchpad/cmp-a44b/ev-r2.md`, `ev-r2b.md`, repros in `cmp-a44b/repro/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges, on the lowered path. 1 added line (IPC::Open3). `sub xfileno { return fileno $_[0] }` then `xfileno(\*STDOUT) == fileno \*STDOUT`: `left operand got Undef|Int`; perl is silent. The calls group's rule checks a builtin's own result without its Undef (01a13b5c-1ff2), but the same result returned through a user sub's ReturnType keeps it.

#### r2_user_fileno_cmp

```perl
sub xfileno { return fileno $_[0] }
print "same\n" if xfileno(\*STDOUT) == fileno \*STDOUT;
```

```perl
sub xfileno { return fileno $_[0] }
print "same\n" if xfileno(\*STDOUT) == fileno \*STDOUT;
```
perl -w (rc=0): same
old: 1:29: error: call to fileno: argument 1 expects Glob, got Scalar [type-mismatch];2:47: error: call to fileno: argument 1 expects Glob, got GlobRef [type-mismatch];2:19: warning: left operand of "==": expected Num, got Undef|Int [coercion-mismatch];2:40: warning: right operand of "==": expected Num, got Undef|Int [coercion-mismatch];
new: 2:19: warning: left operand of "==": expected Num, got Undef|Int [coercion-mismatch];

```perl
sub xfileno { return fileno $_[0] }
print "same\n" if xfileno(\*STDOUT) == fileno \*STDOUT;
```
perl -w (rc=0): same
old: 1:29: error: call to fileno: argument 1 expects Glob, got Scalar [type-mismatch];2:47: error: call to fileno: argument 1 expects Glob, got GlobRef [type-mismatch];2:19: warning: left operand of "==": expected Num, got Undef|Int [coercion-mismatch];2:40: warning: right operand of "==": expected Num, got Undef|Int [coercion-mismatch];
new: 2:19: warning: left operand of "==": expected Num, got Undef|Int [coercion-mismatch];

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; goldens under `Diagnostics-changed:` trailers; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestUserSubReturningBuiltinResult$' -count=1 -v | grep -q '^--- PASS: TestUserSubReturningBuiltinResult'`)
