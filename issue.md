---
title: "`defined` given an assignment or `caller ++$l` is typed List"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:01:03.294489331Z
updated: 2026-10-14T20:56:04.577086617Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: 3d565dc4b5aab40dd9abbfef2f37b4dc67957d68
  commits: 3
  started_at: 2026-10-14T17:13:39.382878401Z
  ended_at: 2026-10-14T20:56:04.577086617Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:39.382878401Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T20:56:04.577086617Z
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

- found by: cutover 1 (01a0c115-a44b) bucket `N11-DEFINED-LIST`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 3 (3 added)  
**Likely fix site:** internal/infer/lowered_calls.go `args` (defined's operand is in scalar context: a scalar assignment, `caller EXPR`)

ExtUtils::ParseXS `defined($self->{lastline} = readline(...))`, base.pm `defined caller ++$lvl`. Legacy fires on the assignment repro.

**Repro `n11_defined_assign.pl`:**
```perl
open my $fh, "<", $0 or die;
my %self;
print "l\n" if defined($self{lastline} = readline($fh));
```
perl -w (rc=0): l
old: 3:51: error: call to readline: argument 1 expects Glob, got Scalar [type-mismatch];3:24: error: call to defined: argument 1 expects Scalar, got List [type-mismatch];
new: 3:51: error: call to readline: argument 1 expects Glob, got Scalar [type-mismatch];3:24: error: call to defined: argument 1 expects Scalar, got List [type-mismatch];

**Repro `n11_defined_list.pl`:**
```perl
my $lvl = 0;
sub depth { 1 while defined caller ++$lvl; $lvl }
open my $fh, "<", $0; print "l\n" if defined(my $x = readline($fh)); print depth(), "\n";
```
perl -w (rc=0): l 1
old: 
new: 2:29: error: call to defined: argument 1 expects Scalar, got List [type-mismatch];3:63: error: call to readline: argument 1 expects Glob, got Scalar [type-mismatch];

Examples from the diff:

- `privlib/ExtUtils/ParseXS.pm:3108` (added) `last unless defined($self->{lastline} = readline($self->{in_fh}));` -- call to defined: argument 1 expects Scalar, got List
- `privlib/base.pm:133` (added) `unshift @INC,        sub { return if defined $lvl; 1 while defined caller ++$lvl; () };` -- call to defined: argument 1 expects Scalar, got List
- `privlib/base.pm:134` (added) `splice  @INC, -1, 0, sub { return if defined caller $lvl; ++$dot_hidden, &base::__inc::unhook; () };` -- call to defined: argument 1 expects Scalar, got List

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestDefinedOfAssignmentIsScalar$' -count=1 -v | grep -q '^--- PASS: TestDefinedOfAssignmentIsScalar'`)
