---
title: "A reference in a Str slot is a valid coercion: drop the warning, add an opt-in lossy-coercion warning"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-13T23:22:39.523151838Z
updated: 2026-10-14T21:32:24.396173968Z
sessions:
- start_sha: 36ecbc1380ea6afd9166cd772cff0dbe4ccf36de
  end_sha: 2ee5f736a0918539132d8ff50f168380b9d922d8
  commits: 4
  started_at: 2026-10-14T17:13:41.298228279Z
  ended_at: 2026-10-14T21:32:24.396173968Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T17:13:41.298228279Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T21:32:24.396173968Z
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

- paths: `internal/types/CORE.pmt` (Str slots: print, length, ...), `internal/infer/lowered_calls.go` (argument checks), `internal/infer/infer.go` (legacy), `internal/infer/options.go` (`Options`)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

perigrin ruled 2026-10-13: "Perl coerces the hashref to a string … so we should drop the warning on the input and add an optional warning on a lossy coercion from hashref to string." Today PSC reports `call to print: argument 1 expects Str, got HashRef` for `print $h` (and, since the calls port 01a12a85-32cf, for `print STDOUT $h`; golden `calls/print_ref_arg`). perl 5.42 -w is silent and prints `HASH(0x…)`: a reference is a valid Str by coercion. The coercion is lossy (the referent is not recoverable from the string), which is what an opt-in warning should name.

Related: `$h + 1` (golden `ops/paren_operand`) reports `left operand of "+": expected Num, got HashRef`; perigrin kept that one on 2026-10-13. Whether ref->Num joins this opt-in lossy-coercion class is a question to raise when this lands, not to decide here.

Do after the cutover (01a0c115-df67), on the one remaining path.


**Widened and rescheduled 2026-10-14.** perigrin ruled the Str-to-Num coercion warning (`$1 + 0`, `"$]" < 5.016`, `substr($s,0,1) % 2`; golden `ops/str_arith`; bucket STRNUM-FP in `scratchpad/cmp-a44b/buckets.md`) "the same lossy coercion warning": a Str in a numeric slot whose content is not known is silent by default and reported under the opt-in. A Str KNOWN to be non-numeric (a non-numeric literal) stays reported by default, because perl 5.42 -w warns there too: `"abc" + 1` gives "Argument \"abc\" isn't numeric". He also ruled every false positive cutover 1 adds is fixed BEFORE cutover 1 (01a0c115-a44b) merges, so this issue now lands on the lowered path before the cutover, not after it.

## Steps

- [ ] Failing tests first: a ref in a Str slot is no diagnostic by default; under the new option it is a `lossy-coercion` warning naming the ref type
- [ ] Option in `Options` (e.g. `WarnLossyCoercion`), off by default; psc flag if psc exposes Options
- [ ] `calls/print_ref_arg` golden edited under a `Diagnostics-changed:` trailer; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] a ref in a Str slot is silent by default (`go test ./internal/infer/ -run '^TestRefInStrSlotIsSilent$' -count=1 -v | grep -q '^--- PASS: TestRefInStrSlotIsSilent'`)
- [ ] the opt-in warns on a lossy ref-to-string coercion (`go test ./internal/infer/ -run '^TestLossyCoercionWarningIsOptIn$' -count=1 -v | grep -q '^--- PASS: TestLossyCoercionWarningIsOptIn'`)
- [ ] a Str of unknown content in a numeric slot is silent by default, and a known non-numeric literal is still reported (`go test ./internal/infer/ -run '^TestStrInNumericSlotIsOptIn$' -count=1 -v | grep -q '^--- PASS: TestStrInNumericSlotIsOptIn'`)
