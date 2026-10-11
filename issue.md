---
title: "Port the operator rules onto the lowered tree: Binary.Op, Repeat, Assign and Range from fields"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-31a5-769d-88c4-e2add3f8fefc
blocks:
- 01a12a85-32cf-7d76-96f7-b814935374e1
created: 2026-10-11T10:32:14.642799615Z
updated: 2026-10-11T10:33:26.805293485Z
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

## Acceptance Criteria

### Positive Scenarios
- [ ] The lowered path holds the golden ratchet in both directions (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
- [ ] Every ops fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/ops/*.pl)" && ! ls internal/infer/testdata/lowered/ops/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)
- [ ] The infer package is green (`go test ./internal/infer/ -count=1`)

### Negative Scenarios
- [ ] An operator with an operand perl rejects analyses without a panic and draws no mismatch against the Hole (`go test ./internal/infer/ -run '^TestLoweredOperatorsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredOperatorsOverBrokenInput'`)
