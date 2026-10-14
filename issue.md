---
title: A hash built from a key/value list has Str elements
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:28.530722829Z
updated: 2026-10-14T17:13:42.024024192Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T17:13:42.024024192Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:42.024024192Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N27-HASH-ELEM-STR`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 4 (4 added)  
**Likely fix site:** internal/infer/lowered_vars.go `recordElemType` (a hash list alternates keys and values; the element type is the values' type)

B::Deparse `$left{assoc_class($op)} == $left{...}` over `%left = (multiply => 19, ...)`. Legacy fires on the repro too.

**Repro `n27_hash_value_eq.pl`:**
```perl
my (%left, $op, $l) = ((), "multiply", "multiply");
%left = ('multiply' => 19, 'i_multiply' => 19);
print "same\n" if $left{$op} && $left{$l} and $left{$op} == $left{$l};
```
perl -w (rc=0): Use of uninitialized value $op in hash element at n27_hash_value_eq.pl line 3.
old: 3:47: warning: left operand of "==": expected Num, got Str [coercion-mismatch];3:61: warning: right operand of "==": expected Num, got Str [coercion-mismatch];
new: 3:47: warning: left operand of "==": expected Num, got Str [coercion-mismatch];3:61: warning: right operand of "==": expected Num, got Str [coercion-mismatch];

Examples from the diff:

- `privlib/B/Deparse.pm:3045` (added) `and $left{assoc_class($op)} == $left{assoc_class($left)})` -- right operand of "==": expected Num, got Str
- `privlib/B/Deparse.pm:3045` (added) `and $left{assoc_class($op)} == $left{assoc_class($left)})` -- left operand of "==": expected Num, got Str
- `privlib/B/Deparse.pm:3079` (added) `and $right{assoc_class($op)} == $right{assoc_class($right)})` -- right operand of "==": expected Num, got Str

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestKeyValueHashElementsKeepValueTypes$' -count=1 -v | grep -q '^--- PASS: TestKeyValueHashElementsKeepValueTypes'`)
