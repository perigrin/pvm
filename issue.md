---
title: Declare the forking operators .. and x as multis in CORE.pmt
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4ca3-7d55-82ac-ba125816ee71
- 01a0fd9b-53a9-703f-82ab-eeb38765f7f1
blocks:
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
created: 2026-10-02T17:13:30.472741835Z
updated: 2026-10-02T17:24:17.033148624Z
---

## Prerequisites

- [ ] Read :context(...) and select a multi candidate by calling context
- [ ] Parse operator declarations and type perl's plain operators in CORE.pmt

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coreops_test.go`, `internal/types/select.go`, `internal/types/select_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/types/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: RFC 0001 "Operator declarations", "Operators that fork"; `internal/types/select.go`

`..` is a range in list context and a flip-flop in scalar context: a `:context` multi (`...` likewise). `x` repeats a list only when its left operand is parenthesised -- measured on 5.42, `(1,2) x 2` and `(@a) x 2` give `1 2 1 2`, `"ab" x 2` gives `abab`, `@a x 2` gives `22` -- so `multi sub x :infix(MUL) (List @l, Int $n) List;` and `multi sub x :infix(MUL) (Str $s, Int $n) Str;`. The parenthesis is a parse fact, as the comma is for map: `List` and `Str` are subtypes of each other's family, so types cannot select, and the call site must say whether the left operand is a parenthesised list (the caller-view sigil `@` against `$`). Remeasure each case with perl. This finishes "Operator declarations".

## Steps

- [ ] Write failing test `TestCoreRangeIsContextMulti`: CORE.pmt declares `..` and `...` as `:context` multis, list and scalar candidates
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreRangeIsContextMulti$' -count=1 -v`
- [ ] Add the declarations; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectByOperandShape` in `internal/types/select_test.go`: x's candidates select the `List` one for a parenthesised-list left operand and the `Str` one otherwise
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreRepeatIsShapeMulti`: CORE.pmt declares both `x` candidates with class `MUL`
- [ ] Add the declarations; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "Operator declarations" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- a parenthesised left operand selects the list repetition only in list context: measured on 5.42, `my $x = (1,2) x 2` and `my $x = (@a) x 2` give `22` (a string), so in scalar context the `Str` candidate is selected (`go test ./internal/types/ -run '^TestSelectRepeatScalarContextIsStr$' -count=1 -v | grep -q '^--- PASS: TestSelectRepeatScalarContextIsStr'`)
- an unparenthesised array (`@a x 2`) selects `Str`, not `List` (measured `22`) (`go test ./internal/types/ -run '^TestSelectRepeatBareArrayIsStr$' -count=1 -v | grep -q '^--- PASS: TestSelectRepeatBareArrayIsStr'`)
- boolean context is scalar, so `if (/a/ .. /b/)` selects the flip-flop candidate, not the range (`go test ./internal/types/ -run '^TestSelectRangeBooleanIsFlipFlop$' -count=1 -v | grep -q '^--- PASS: TestSelectRangeBooleanIsFlipFlop'`)

## Acceptance criteria

- [ ] the range operators are declared as context multis (`go test ./internal/parse/ -run '^TestCoreRangeIsContextMulti$' -count=1 -v | grep -q '^--- PASS: TestCoreRangeIsContextMulti'`)
- [ ] a parenthesised left operand selects the list repetition (`go test ./internal/types/ -run '^TestSelectByOperandShape$' -count=1 -v | grep -q '^--- PASS: TestSelectByOperandShape'`)
- [ ] x is declared as a multi selected by its left operand (`go test ./internal/parse/ -run '^TestCoreRepeatIsShapeMulti$' -count=1 -v | grep -q '^--- PASS: TestCoreRepeatIsShapeMulti'`)
- [ ] the RFC section on operator declarations is implemented (`grep -q '^### Operator declarations.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a parenthesised x in scalar context selects the string repetition (`go test ./internal/types/ -run '^TestSelectRepeatScalarContextIsStr$' -count=1 -v | grep -q '^--- PASS: TestSelectRepeatScalarContextIsStr'`)
- [ ] an unparenthesised array left of x selects the string repetition (`go test ./internal/types/ -run '^TestSelectRepeatBareArrayIsStr$' -count=1 -v | grep -q '^--- PASS: TestSelectRepeatBareArrayIsStr'`)
- [ ] a range in boolean context selects the flip-flop (`go test ./internal/types/ -run '^TestSelectRangeBooleanIsFlipFlop$' -count=1 -v | grep -q '^--- PASS: TestSelectRangeBooleanIsFlipFlop'`)
