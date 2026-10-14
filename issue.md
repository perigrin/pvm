---
title: A hash element is typed the union of every value in the literal
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:17.495103906Z
updated: 2026-10-14T17:13:41.982144912Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T17:13:41.982144912Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:41.982144912Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N18-HASH-VALUE-UNION`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 1 (1 added)  
**Likely fix site:** internal/infer/lowered_vars.go `recordElemType`/`elemType` (a keyed element of a literal is that key's value type, or unchecked)

`{count => 1, values => [2]}->{count} == 1` draws "got Int|ArrayRef". Legacy is silent.

**Repro `n18_hash_value_union.pl`:**
```perl
my $source = { count => 1, values => [2] };
print "ok\n" if $source->{count} == 1;
```
perl -w (rc=0): ok
old: 
new: 2:17: warning: left operand of "==": expected Num, got Int|ArrayRef [coercion-mismatch];

Examples from the diff:

- `T1:unit/threads_shared_graphs.t:30` (added) `check($source->{count} == 1 && join(',', @{$source->{values}}) eq '2',` -- left operand of "==": expected Num, got Int|ArrayRef

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestHashElementTypeIsPerKey$' -count=1 -v | grep -q '^--- PASS: TestHashElementTypeIsPerKey'`)
