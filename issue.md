---
title: A capture returned after its own successful match keeps Undef through a user sub
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-15T00:27:10.793370616Z
updated: 2026-10-15T00:27:10.793370616Z
---

## Context

- found by: cutover 1 (01a0c115-a44b), round-3 re-measure, bucket `R3-USER-FILENO-CAPTURE` (2026-10-15); evidence in `scratchpad/cmp-a44b/ev-r3.md`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed before it merges. IPC/Open3.pm:335: `sub xfileno { return $1 if $_[0] =~ /\A=?(\d+)\z/; return fileno $_[0] }` then `xfileno(\*STDOUT) == fileno \*STDOUT` -- the lowered path reports `left operand of "==": expected Num, got Undef|Int|Num|Str|NaN|Inf`; perl -w is silent and prints `same`. `$1` inside the true branch of its own match (here the `if` modifier guarding the return) is defined; and the round-2 fix (01a13c6a-78f5) hands on a builtin result through a user sub but not a capture.

```perl
sub xfileno { return $1 if $_[0] =~ /\A=?(\d+)\z/; return fileno $_[0]; }
print "same\n" if xfileno(\*STDOUT) == fileno \*STDOUT;
```

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42 -w's verdict holds, keeping a control perl warns on; precision must stay >= 64/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the repro is silent as perl is (`go test ./internal/infer/ -run '^TestCaptureReturnedAfterMatchIsDefined$' -count=1 -v | grep -q '^--- PASS: TestCaptureReturnedAfterMatchIsDefined'`)
