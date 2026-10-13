---
title: "Port the operator rules onto the lowered tree: Binary.Op, Repeat, Assign and Range from fields"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-31a5-769d-88c4-e2add3f8fefc
- 01a0c111-8c3e-7451-97d9-10307441ba2d
blocks:
- 01a12a85-32cf-7d76-96f7-b814935374e1
created: 2026-10-11T10:32:14.642799615Z
updated: 2026-10-13T16:07:02.012667962Z
sessions:
- start_sha: 25ec742454544ea6db09a7d8ace24e430966243f
  end_sha: e61467531761799cbc9c52bf845a9f73a7ba2854
  commits: 1
  started_at: 2026-10-13T02:38:06.307745476Z
  ended_at: 2026-10-13T16:07:02.012667962Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-13T02:38:06.307745476Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-13T16:07:02.012667962Z
observed_paths:
- internal/infer/annotations_test.go
- internal/infer/golden_test.go
- internal/infer/infer.go
- internal/infer/lowered.go
- internal/infer/lowered_ops.go
- internal/infer/lowered_test.go
- internal/infer/lowered_vars.go
- internal/infer/testdata/lowered/matching.txt
- internal/infer/testdata/lowered/ops/add.golden
- internal/infer/testdata/lowered/ops/add.pl
- internal/infer/testdata/lowered/ops/aggregate_compare.golden
- internal/infer/testdata/lowered/ops/aggregate_compare.pl
- internal/infer/testdata/lowered/ops/arith_int.golden
- internal/infer/testdata/lowered/ops/arith_int.pl
- internal/infer/testdata/lowered/ops/arrayref_concat.golden
- internal/infer/testdata/lowered/ops/arrayref_concat.pl
- internal/infer/testdata/lowered/ops/assign_array.golden
- internal/infer/testdata/lowered/ops/assign_array.pl
- internal/infer/testdata/lowered/ops/assign_arrayref.golden
- internal/infer/testdata/lowered/ops/assign_arrayref.pl
- internal/infer/testdata/lowered/ops/assign_count.golden
- internal/infer/testdata/lowered/ops/assign_count.pl
- internal/infer/testdata/lowered/ops/assign_undeclared.golden
- internal/infer/testdata/lowered/ops/assign_undeclared.pl
- internal/infer/testdata/lowered/ops/bare_arith.golden
- internal/infer/testdata/lowered/ops/bare_arith.pl
- internal/infer/testdata/lowered/ops/bitwise.golden
- internal/infer/testdata/lowered/ops/bitwise.pl
- internal/infer/testdata/lowered/ops/compare.golden
- internal/infer/testdata/lowered/ops/compare.pl
- internal/infer/testdata/lowered/ops/compound_assign.golden
- internal/infer/testdata/lowered/ops/compound_assign.pl
- internal/infer/testdata/lowered/ops/concat.golden
- internal/infer/testdata/lowered/ops/concat.pl
- internal/infer/testdata/lowered/ops/concat_array.golden
- internal/infer/testdata/lowered/ops/concat_array.pl
- internal/infer/testdata/lowered/ops/count_of.golden
- internal/infer/testdata/lowered/ops/count_of.pl
- internal/infer/testdata/lowered/ops/defined_or.golden
- internal/infer/testdata/lowered/ops/defined_or.pl
- internal/infer/testdata/lowered/ops/foreach_alias.golden
- internal/infer/testdata/lowered/ops/foreach_alias.pl
- internal/infer/testdata/lowered/ops/foreach_named_alias.golden
- internal/infer/testdata/lowered/ops/foreach_named_alias.pl
- internal/infer/testdata/lowered/ops/hashref_arith.golden
- internal/infer/testdata/lowered/ops/hashref_arith.pl
- internal/infer/testdata/lowered/ops/incr.golden
- internal/infer/testdata/lowered/ops/incr.pl
- internal/infer/testdata/lowered/ops/isa.golden
- internal/infer/testdata/lowered/ops/isa.pl
- internal/infer/testdata/lowered/ops/list_assign.golden
- internal/infer/testdata/lowered/ops/list_assign.pl
- internal/infer/testdata/lowered/ops/logical.golden
- internal/infer/testdata/lowered/ops/logical.pl
- internal/infer/testdata/lowered/ops/neg.golden
- internal/infer/testdata/lowered/ops/neg.pl
- internal/infer/testdata/lowered/ops/not.golden
- internal/infer/testdata/lowered/ops/not.pl
- internal/infer/testdata/lowered/ops/paren_operand.golden
- internal/infer/testdata/lowered/ops/paren_operand.pl
- internal/infer/testdata/lowered/ops/range.golden
- internal/infer/testdata/lowered/ops/range.pl
- internal/infer/testdata/lowered/ops/range_num.golden
- internal/infer/testdata/lowered/ops/range_num.pl
- internal/infer/testdata/lowered/ops/reassign.golden
- internal/infer/testdata/lowered/ops/reassign.pl
- internal/infer/testdata/lowered/ops/repeat.golden
- internal/infer/testdata/lowered/ops/repeat.pl
- internal/infer/testdata/lowered/ops/repeat_div.golden
- internal/infer/testdata/lowered/ops/repeat_div.pl
- internal/infer/testdata/lowered/ops/repeat_num.golden
- internal/infer/testdata/lowered/ops/repeat_num.pl
- internal/infer/testdata/lowered/ops/repeat_paren_one.golden
- internal/infer/testdata/lowered/ops/repeat_paren_one.pl
- internal/infer/testdata/lowered/ops/repeat_qw.golden
- internal/infer/testdata/lowered/ops/repeat_qw.pl
- internal/infer/testdata/lowered/ops/str_arith.golden
- internal/infer/testdata/lowered/ops/str_arith.pl
- internal/infer/testdata/lowered/ops/str_compare_num.golden
- internal/infer/testdata/lowered/ops/str_compare_num.pl
- internal/infer/testdata/lowered/ops/str_eq.golden
- internal/infer/testdata/lowered/ops/str_eq.pl
- internal/infer/testdata/lowered/ops/ternary.golden
- internal/infer/testdata/lowered/ops/ternary.pl
- internal/infer/testdata/lowered/ops/ternary_mixed.golden
- internal/infer/testdata/lowered/ops/ternary_mixed.pl
- internal/infer/testdata/lowered/ops/ternary_str.golden
- internal/infer/testdata/lowered/ops/ternary_str.pl
- internal/infer/testdata/lowered/ops/xor.golden
- internal/infer/testdata/lowered/ops/xor.pl
---

## Context

- paths: `internal/infer/lowered.go`, `internal/infer/testdata/lowered/ops/` (new)
- commands: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- entrypoints: `inferBinaryExprType` (`infer.go:428`), `inferUnaryExprType` (`:610`), `findOperatorText` (`:656`), `operandType` (`:1853`), `inferAssignmentNarrowing` (`:2105`), `inferConditionalType` (`:382`)

The fourth of seven port issues. It ports binary, unary, ternary, repeat,
range and assignment typing, and assignment narrowing.

What gets simpler, and what review should check:

- `findOperatorText` (`infer.go:656`) has no successor: `Binary.Op` is a
  field;
- `extractNegatedGuard`'s `== "!"` text test reads `Unary.Op`;
- the operator signatures come from `parse.CoreOperator` exactly as today
  (`infer.go:615`, `:1937`).

**Range changes a reported type, on purpose.** §4.14.2 measured legacy
inferring `my @r = (1..5)` as `Str`, because `findOperatorText` matched the
`.` before `..`. If that still holds on the legacy path, its golden says
`Str`; the lowered rule reads `Range` and should say `List` of `Int`. That
golden changes, with a `Diagnostics-changed:` trailer.

`CmpChain` has no rule today and gets none here: write none, type it as
legacy does.

The method is the one "AnalyzeLowered: the entry point, the golden
ratchet, and the declaration and literal rules" set up: fixtures
under `testdata/lowered/ops/`, goldens captured from the legacy path, and
`TestLoweredGoldenRatchet` holding `matching.txt` in both directions. Every
rule is REWRITTEN against the lowered vocabulary, not shimmed. No dataflow
IR (§6.1.6): if a rule wants flow, it keeps the approximation it has today.

Rollback: this issue changes only `AnalyzeLowered` and test data. Nothing in `internal/psc` calls it yet, so reverting it changes nothing PSC reports.

## Steps

- [ ] Write fixtures under `testdata/lowered/ops/`, drawn from the
      existing cases in `infer_test.go` where they exist, using only
      constructs whose rules are already ported plus this slice's.
- [ ] Capture their goldens from the legacy path:
      `go test ./internal/infer/ -run "^TestCaptureLegacyGolden$" -infer.capture -count=1`.
      Commit the fixtures and goldens.
- [ ] Run the ratchet and confirm the new fixtures do not match yet:
      `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- [ ] Port this slice's rules into `AnalyzeLowered`. Commit as each rule
      group goes green rather than once at the end.
- [ ] Update `matching.txt` with `-infer.ratchet.update` and re-run the
      ratchet to green. Every `ops/` fixture must match. One that cannot
      is a legacy defect or a regression: triage it before moving on. If it
      is a legacy defect, edit its golden in its own commit with a
      `Diagnostics-changed:` trailer naming the cause.
- [ ] Write `TestLoweredOperatorsOverBrokenInput`: `my $x = 1 +;` and
      `my $s = "a" x;` (perl rejects both) analyse without a panic and emit
      no type-mismatch diagnostic against the `Hole` operand. Run to verify
      failure, implement, run to green, commit.
- [ ] Run the infer suite: `go test ./internal/infer/ -count=1`
- [ ] Commit.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance Criteria

### Positive Scenarios
- [ ] The lowered path holds the golden ratchet in both directions (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
- [ ] Every ops fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/ops/*.pl)" && ! ls internal/infer/testdata/lowered/ops/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)

### Negative Scenarios
- [ ] An operator with an operand perl rejects analyses without a panic and draws no mismatch against the Hole (`go test ./internal/infer/ -run '^TestLoweredOperatorsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredOperatorsOverBrokenInput'`)
