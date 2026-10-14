---
title: "`splice @t, 0, +@d` length typed Num"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:04.905240839Z
updated: 2026-10-14T17:01:04.905240839Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N12-SPLICE-UPLUS`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 1 (1 added)  
**Likely fix site:** internal/infer/lowered_ops.go `unary` (unary `+` is the identity: `+@d` is the count, Int)

File::Path:267. Legacy is silent.

**Repro `n12_splice_plus.pl`:**
```perl
sub f { my @d = @{$_[0]}; my @t = @{$_[1]}; return join('/', @d) eq join('/', splice @t, 0, +@d) }
print f([1], [1, 2]) ? "same\n" : "differ\n";
```
perl -w (rc=0): same
old: 
new: 1:93: warning: call to splice: argument 3 expects Int, got Num [type-mismatch];

Examples from the diff:

- `privlib/File/Path.pm:267` (added) `return join( '/', @d ) eq join( '/', splice @t, 0, +@d );` -- call to splice: argument 3 expects Int, got Num

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestUnaryPlusArrayIsItsCount$' -count=1 -v | grep -q '^--- PASS: TestUnaryPlusArrayIsItsCount'`)
