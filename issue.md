---
title: A string pattern on the right of `=~` expects Regex
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:59.04925911Z
updated: 2026-10-14T17:00:59.04925911Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N9-STR-PATTERN`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 4 (4 added)  
**Likely fix site:** internal/infer/lowered_ops.go `binary`/`infix` shapes for `=~` (the right operand accepts Str: perl compiles it as a pattern)

Module::CoreList `$_[1] =~ PKG_PATTERN` with `sub PKG_PATTERN () { q#...# }`. perl matches. Legacy is silent on the repro.

**Repro `n9_const_sub_pattern.pl`:**
```perl
sub PKG_PATTERN () { q#\A[a-zA-Z_]\w*\z# }
my @a = ("Foo");
print "m\n" if defined $a[0] and $a[0] =~ PKG_PATTERN;
```
perl -w (rc=0): m
old: 
new: 3:43: error: right operand of "=~": expected Regex, got Str [coercion-mismatch];

Examples from the diff:

- `privlib/Module/CoreList.pm:105` (added) `shift if defined $_[1] and $_[1] =~ PKG_PATTERN and _looks_like_invocant $_[0];` -- right operand of "=~": expected Regex, got Str
- `privlib/Module/CoreList.pm:124` (added) `shift if defined $_[1] and $_[1] =~ PKG_PATTERN and _looks_like_invocant $_[0];` -- right operand of "=~": expected Regex, got Str
- `privlib/Module/CoreList.pm:51` (added) `shift if defined $_[1] and $_[1] =~ PKG_PATTERN and _looks_like_invocant $_[0];` -- right operand of "=~": expected Regex, got Str

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestStringPatternRightOfBindIsSilent$' -count=1 -v | grep -q '^--- PASS: TestStringPatternRightOfBindIsSilent'`)
