---
title: "Lower operators: Binary.Op from the field, Repeat and Assign from the Paren flag"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
- 01a0c10e-6a8d-721d-be45-a07ae1ee2033
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-09-20T23:05:47.837817466Z
updated: 2026-09-20T23:13:03.866632505Z
---

## Context

- paths: `internal/lower/ops.go` (new), `internal/lower/ops_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerBinaryOpFromField$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Operators"; §4.14.1 rule 1

This slice kills one of the three measured consumer defects the lowering exists
for: **`findOperatorText` (`infer.go:645-659`) re-derives the operator by
scanning a node's unnamed children for text that matches an operator table.**
`parse.Node.Text` already holds the operator, so the lowering SETS `Binary.Op`
from a field read and that function has no successor.

Status per §4.14.2, corrected at HEAD:

- `Unary{Op,Postfix}` SHIPPED — `unary "!"`, `"neg"`, `"~"`, `"++"`, `"ref"`
  (for backslash), `"pos"` (for unary plus); `postfix "++"`. `Postfix` is a KIND
  split in the CST (`Kind == Postfix` vs `Kind == Unary`), so the flag is a
  rename. `extractNegatedGuard` (`infer.go:2285`) currently tests text `== "!"`;
  the lowered `Op` replaces that comparison.
- `Binary{Op}` SHIPPED; **the lowering sets `Op` from `Text`.** The bridge to
  B::SoN's split family is a total table from `Op` and belongs in the consumer,
  not here: `==`→NumEq, `eq`→StrEq, `<`→NumLt, `lt`→StrLt, `.`→Concat,
  `x`→Repeat, `//`→DefinedOr. Keep one `Binary{Op}`.
- `Ternary` SHIPPED — `ternary "?:"`.
- `CmpChain` SHIPPED — `cmp_chain "<"` with flat operands. **No consumer**:
  `internal/infer` has no chained-comparison rule at all. §4.14.6 leaves the
  typing rule open and says it is "settled by writing the rule and seeing
  whether it wants a flat operand list or nested Binarys". Lower it faithfully
  (flat, as the CST has it) and leave the typing to the inference issue; do not
  invent a rule here.
- `Range{Exclusive}` — `$a..$b` is `binary ".."`, `$a...$b` is `binary "..."`.
  **The measured defect in §4.14.2 is GONE at HEAD**: `my @r = (1..5);` now
  parses as `declaration "my"` over `term "@r"` and `binary ".." paren` with
  `term "1"` and `term "5"` children. It no longer reads as `Str`, and the
  lexer no longer takes `1..5` as one Number token. Re-verify this before
  implementing — if it holds, `Range` is LOWERING for both variable and literal
  endpoints, which §4.14.2 marked PARSER.
- `Repeat{ListRepeat}` LOWERING. **`parse.Node.Paren` ships**, so §4.14.1 rule
  1's fact is a field read rather than the unsound span heuristic the rule
  describes at length. Verified at HEAD: `('a') x 3` gives
  `binary "x"` over `term "'a'" paren` and `term "3"`. `qw(a b) x 2` is the
  other list-repeat spelling and keys on the `qw` prefix of the term's `Text`.
  perl: `('a') x 3` gives three elements, `'a' x 3` gives one.
  **Do not reimplement the span heuristic.** §4.14.1 rule 1 spends a paragraph
  showing it is unsound (721 of 6,574 T1 nodes with a paren-leading span are
  not parenthesised operands, and `PrototypeNode` leaves carry their parens in
  `Text`). The flag exists now; read it.
- `Assign{ListAssign}` LOWERING — `binary "="`, `"+="`, ...; `($a,$b) = (1,2)`
  has a `list` LHS, `@a = (1,2)` an aggregate-sigil `term`, and a single
  parenthesised scalar sets `Paren`. `my (...) =` is under `Decl` and belongs to
  the declarations issue.
- `List{Fat}` SHIPPED as `List`, and **`Fat` is no longer PARSER**:
  `parse.Node.Fat` ships and is set. Verified at HEAD: `(a => 1, b => 2)` gives
  a `list` whose two `call` children carry `Fat`. §4.5.4: the fat comma quotes
  the word to its left, so without this the lowering could not tell `(a, 1)`
  (a call) from `(a => 1)` (a string).
- `Paren` — §4.14.6 leaves node-vs-flag open; the vocabulary issue settles it.
  Whichever it chose, this issue is where the two rules that read it get
  written, which is what §4.14.6 says settles the question.

## Steps

- [ ] Write `TestLowerBinaryOpFromField`: a table over `+ - . x == eq < lt // ..
      ... =~ !~ =` asserting `Binary.Op` matches `parse.Node.Text` exactly, and
      that no source rescan is needed. Include a case with an operator that is
      also a substring of an operand, which is what `findOperatorText` gets
      wrong.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerBinaryOpFromField$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerListRepeatFromParenFlag`: `('a') x 3` lowers to
      `Repeat{ListRepeat:true}`, `'a' x 3` to `ListRepeat:false`, and
      `qw(a b) x 2` to `true` via the `qw` prefix. Run to verify failure,
      implement, run to green, commit.
- [ ] Write `TestLowerRangeEndpoints`: `$a..$b`, `1..5` and `$a...$b` each lower
      to a `Range` with both endpoints present and `Exclusive` set only for the
      three-dot form. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerListFat`: `(a => 1, b => 2)` lowers to a `List` whose
      `Fat` slice marks elements 0 and 2, and `(a, 1, b, 2)` marks none.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerAssignListness`: `($a,$b) = (1,2)` and `@a = (1,2)` set
      `ListAssign`, `$a = 1` does not. Run to verify failure, implement, run to
      green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Binary.Op comes from the node's field, retiring findOperatorText (`go test ./internal/lower/ -run "^TestLowerBinaryOpFromField$" -count=1 -v | grep -q "^--- PASS: TestLowerBinaryOpFromField"`)
- [ ] List repeat reads the Paren flag, not a span heuristic (`go test ./internal/lower/ -run "^TestLowerListRepeatFromParenFlag$" -count=1 -v | grep -q "^--- PASS: TestLowerListRepeatFromParenFlag"`)
- [ ] Ranges keep both endpoints, literal ones included (`go test ./internal/lower/ -run "^TestLowerRangeEndpoints$" -count=1 -v | grep -q "^--- PASS: TestLowerRangeEndpoints"`)
- [ ] Fat commas survive inside a list (`go test ./internal/lower/ -run "^TestLowerListFat$" -count=1 -v | grep -q "^--- PASS: TestLowerListFat"`)
- [ ] List assignment is distinguished from scalar assignment (`go test ./internal/lower/ -run "^TestLowerAssignListness$" -count=1 -v | grep -q "^--- PASS: TestLowerAssignListness"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)
