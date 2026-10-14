---
title: "`use subs` overriding a builtin is still checked as the builtin"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:12.578551649Z
updated: 2026-10-14T17:13:39.705051721Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T17:13:39.705051721Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:39.705051721Z
---

## Context

- found by: cutover 1 (01a0c115-a44b) bucket `N13-USE-SUBS`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 1 (1 added)  
**Likely fix site:** internal/infer/lowered_calls.go `call` (a name the file imports or predeclares via `use subs` is a user sub)

operator_overrides.t `rename()`. Legacy fires too (arity 1; lowered says 0).

**Repro `n13_use_subs.pl`:**
```perl
use subs "rename";
sub rename { "fake" }
print rename(), "\n";
```
perl -w (rc=0): fake
old: 3:7: error: call to rename: expected at least 2 argument(s), got 1 [arity-mismatch];
new: 3:7: error: call to rename: expected at least 2 argument(s), got 0 [arity-mismatch];

Examples from the diff:

- `T1:unit/operator_overrides.t:96` (added) `is(rename(), "fake_rename", 'rename overridden with use subs');` -- call to rename: expected at least 2 argument(s), got 0

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestUseSubsOverridesBuiltinCheck$' -count=1 -v | grep -q '^--- PASS: TestUseSubsOverridesBuiltinCheck'`)
