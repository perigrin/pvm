---
title: A string pattern on the right of `=~` expects Regex
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:59.04925911Z
updated: 2026-10-14T21:32:05.756144995Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: 2ee5f736a0918539132d8ff50f168380b9d922d8
  commits: 4
  started_at: 2026-10-14T17:13:40.5896481Z
  ended_at: 2026-10-14T21:32:05.756144995Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:40.5896481Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T21:32:05.756144995Z
observed_paths:
- internal/infer/core_test.go
- internal/infer/diagnostics.go
- internal/infer/export_test.go
- internal/infer/infer.go
- internal/infer/infer_test.go
- internal/infer/lowered.go
- internal/infer/lowered_calls.go
- internal/infer/lowered_calls_test.go
- internal/infer/lowered_flow.go
- internal/infer/lowered_flow_test.go
- internal/infer/lowered_ops.go
- internal/infer/lowered_ops_test.go
- internal/infer/lowered_test.go
- internal/infer/lowered_vars.go
- internal/infer/lowered_vars_test.go
- internal/infer/mutation_test.go
- internal/infer/symbols.go
- internal/infer/testdata/lowered/calls/method_new.golden
- internal/infer/testdata/lowered/calls/print_ref_arg.golden
- internal/infer/testdata/lowered/flow/and.golden
- internal/infer/testdata/lowered/flow/and_early_return.golden
- internal/infer/testdata/lowered/flow/and_keyword.golden
- internal/infer/testdata/lowered/flow/and_partial.golden
- internal/infer/testdata/lowered/flow/and_two_vars.golden
- internal/infer/testdata/lowered/flow/blessed_and_defined.golden
- internal/infer/testdata/lowered/flow/not_and.golden
- internal/infer/testdata/lowered/flow/or.golden
- internal/infer/testdata/lowered/flow/or_else.golden
- internal/infer/testdata/lowered/ops/arrayref_concat.golden
- internal/infer/testdata/lowered/ops/compare.golden
- internal/infer/testdata/lowered/ops/foreach_alias.golden
- internal/infer/testdata/lowered/ops/hashref_arith.golden
- internal/infer/testdata/lowered/ops/paren_operand.golden
- internal/infer/testdata/lowered/ops/str_arith.golden
- internal/infer/testdata/lowered/ops/str_compare_num.golden
- internal/infer/testdata/lowered/special/capture_arith.golden
- internal/infer/testdata/lowered/special/ref_arith.golden
- internal/infer/testdata/lowered/special/ref_paren.golden
- internal/parse/corenoproto_test.go
- internal/parse/coretypes_test.go
- internal/parse/declarations/CORE.pmt
- internal/psc/check_command.go
- internal/psc/check_command_test.go
- internal/types/types.go
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
