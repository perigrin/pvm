---
title: 4-arg `select` with an undef timeout
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:07.594236261Z
updated: 2026-10-14T20:56:08.769345244Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: 3d565dc4b5aab40dd9abbfef2f37b4dc67957d68
  commits: 3
  started_at: 2026-10-14T17:13:39.54051216Z
  ended_at: 2026-10-14T20:56:08.769345244Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:39.54051216Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T20:56:08.769345244Z
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

- found by: cutover 1 (01a0c115-a44b) bucket `N12-SELECT-UNDEF-TIMEOUT`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 1 (1 added)  
**Likely fix site:** internal/parse/declarations/CORE.pmt `select` (the timeout admits undef: block indefinitely)

Net::FTP:1287. Legacy fires too.

**Repro `n12_select_undef_timeout.pl`:**
```perl
my ($rin, $rout) = ("", "");
select($rout = $rin, undef, undef, undef) if 0;
```
perl -w (rc=0): 
old: 2:36: error: call to select: argument 4 expects Num, got Undef [type-mismatch];
new: 2:36: error: call to select: argument 4 expects Num, got Undef [type-mismatch];

Examples from the diff:

- `privlib/Net/FTP.pm:1287` (added) `select($rout = $rin, undef, undef, undef);` -- call to select: argument 4 expects Num, got Undef

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestSelectAcceptsUndefTimeout$' -count=1 -v | grep -q '^--- PASS: TestSelectAcceptsUndefTimeout'`)
