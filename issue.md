---
title: An untyped scalar fails a union parameter
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-14T17:00:57.560063819Z
updated: 2026-10-14T21:32:01.996417855Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: 2ee5f736a0918539132d8ff50f168380b9d922d8
  commits: 4
  started_at: 2026-10-14T17:13:40.365938363Z
  ended_at: 2026-10-14T21:32:01.996417855Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:40.365938363Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T21:32:01.996417855Z
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

- found by: cutover 1 (01a0c115-a44b) bucket `N7-SCALAR-UNION`, from `scratchpad/cmp-a44b/buckets.md` (2026-10-14)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-14 that every false positive cutover 1 adds to psc check is fixed BEFORE cutover 1 merges, on the lowered path (`internal/infer/lowered_*.go`; legacy is deleted by 01a0c115-df67). The bucket as measured:

**Class:** FALSE-POSITIVE  
**Lines in the cutover diff:** 18 (4 moved (message), 14 added)  
**Likely fix site:** types.IsCoercible / lowered_calls.go `checkArg` (Scalar, "some scalar", should be accepted by a parameter whose union covers the scalar types)

`chdir $dir`, `prototype($name)`, `truncate $fh, 0`. Legacy fires on the repro; 4 lines are moved (legacy said `Boolean|0x80000000`).

**Repro `n7_scalar_union.pl`:**
```perl
my $dir = shift // ".";
chdir $dir or die;
print defined prototype($dir) ? "p\n" : "none\n";
```
perl -w (rc=0): none
old: 2:7: error: call to chdir: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Scalar [type-mismatch];3:25: error: call to prototype: argument 1 expects Int|Num|Str|Code|NaN|Inf, got Scalar [type-mismatch];
new: 2:7: error: call to chdir: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Scalar [type-mismatch];3:25: error: call to prototype: argument 1 expects Int|Num|Str|Code|NaN|Inf, got Scalar [type-mismatch];

**Repro `n37_truncate_fh.pl`:**
```perl
open my $fh, "+>", "/tmp/claude-1001/-home-perigrin-dev-pvm/6880e770-79d8-446d-b973-79e4840cbcd9/scratchpad/cmp-a44b/repro/t.out" or die;
truncate $fh, 0;
```
perl -w (rc=0): 
old: 2:10: error: call to truncate: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Scalar [type-mismatch];
new: 2:10: error: call to truncate: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Scalar [type-mismatch];

Examples from the diff:

- `privlib/File/Find.pm:567` (msg) `$ok = chdir($updir_loc) unless ($p_dir eq $File::Find::current_dir);` -- call to chdir: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Boolean|0x80000000
- `privlib/File/Find.pm:581` (msg) `unless (chdir $updir_loc) {` -- call to chdir: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Boolean|0x80000000
- `privlib/File/Find.pm:612` (msg) `unless (chdir $updir_loc) {` -- call to chdir: argument 1 expects Int|Num|Str|GlobRef|Glob|NaN|Inf|IO, got Boolean|0x80000000

## Steps

- [ ] Failing test first from the repro above; fix at the named site so perl 5.42 -w's verdict holds; any golden edit in its own commit with a `Diagnostics-changed:` trailer; precision must stay >= 62/76 with WRONG 0; `make test`; commit

## Acceptance criteria

- [ ] the bucket's repro is silent as perl is (`go test ./internal/infer/ -run '^TestUntypedScalarSatisfiesUnionParameter$' -count=1 -v | grep -q '^--- PASS: TestUntypedScalarSatisfiesUnionParameter'`)
