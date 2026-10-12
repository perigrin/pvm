---
title: "Lower operators: Binary.Op from the field, Repeat and Assign from the Paren flag"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
- 01a0c10f-9671-733f-8b98-30a78f634b0f
blocks:
- 01a0c112-a8bb-7d7f-b156-d4ad99474e87
- 01a12a85-3233-7ec8-9428-747be4ccef1c
created: 2026-09-20T23:05:47.837817466Z
updated: 2026-10-12T19:03:21.807538222Z
sessions:
- start_sha: 9895b3086efb0ed7feb8aeb6b994c6b245ce47a5
  end_sha: 9895b3086efb0ed7feb8aeb6b994c6b245ce47a5
  commits: 0
  started_at: 2026-10-11T15:51:32.545872199Z
  ended_at: 2026-10-11T15:54:48.397191597Z
- start_sha: 9895b3086efb0ed7feb8aeb6b994c6b245ce47a5
  end_sha: 93874fbfe9d9e6c38b8b145da0b41e96452b674a
  commits: 7
  started_at: 2026-10-11T16:34:01.758261289Z
  ended_at: 2026-10-12T19:03:21.807538222Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T15:51:32.545872199Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T15:54:48.397191597Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T16:34:01.758261289Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-12T19:03:21.807538222Z
observed_paths:
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/lower/decl.go
- internal/lower/decl_test.go
- internal/lower/export_test.go
- internal/lower/expr.go
- internal/lower/id.go
- internal/lower/id_test.go
- internal/lower/literals.go
- internal/lower/literals_test.go
- internal/lower/node.go
- internal/lower/node_test.go
- internal/lower/ops.go
- internal/lower/ops_test.go
- internal/lower/slice.go
- internal/lower/slice_test.go
- internal/lower/special.go
- internal/lower/special_test.go
- internal/lower/vars.go
- internal/lower/vars_test.go
---

## Context

- paths: `internal/lower/ops.go` (new), `internal/lower/ops_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerBinaryOpFromField$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Operators"; §4.14.1 rule 1

This slice kills one of the three measured consumer defects the lowering exists
for: **`findOperatorText` (`infer.go:656`) re-derives the operator by
scanning a node's unnamed children for text that matches an operator table.**
`parse.Node.Text` already holds the operator, so the lowering SETS `Binary.Op`
from a field read and that function has no successor.

Status per §4.14.2 as the status-table issue re-measured it. Shapes below were
measured at 24bfac4a; re-measure at execution time. Line numbers in
`internal/infer` drift, so citations also name the function.

- `Unary{Op,Postfix}` SHIPPED — `unary "!"`, `"neg"`, `"~"`, `"++"`, `"ref"`
  (for backslash), `"pos"` (for unary plus); `postfix "++"`. `Postfix` is a KIND
  split in the CST (`Kind == Postfix` vs `Kind == Unary`), so the flag is a
  rename. `extractNegatedGuard` (`infer.go:2469`) currently tests text `== "!"`;
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
  **The measured defect in §4.14.2 is gone**: at 24bfac4a `my @r = (1..5);`
  parses as `declaration "my"` over `term "@r"` and `binary ".." paren` with
  `term "1"` and `term "5"` children. The lexer no longer takes `1..5` as one
  Number token, so `Range` is LOWERING for both variable and literal
  endpoints.
- `Repeat{ListRepeat}` LOWERING. **`parse.Node.Paren` ships**, so §4.14.1 rule
  1's fact is a field read rather than the unsound span heuristic the rule
  describes at length. Measured at 24bfac4a: `("a") x 3` gives
  `binary "x"` over `term "\"a\"" paren` and `term "3"`. `qw(a b) x 2` is the
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
  `parse.Node.Fat` ships and is set. At 24bfac4a `(a => 1, b => 2)` gives
  a `list paren` whose two `call` children carry `Fat`. §4.5.4: the fat comma quotes
  the word to its left, so without this the lowering could not tell `(a, 1)`
  (a call) from `(a => 1)` (a string).
- **Comma chains outside a `list` node.** Only a parenthesised list is a
  `list` node. Call arguments, unparenthesised lists and slice subscripts are
  left-nested `binary ","` trees, and inside them a fat comma is a
  `binary "=>"`, not a `Fat` flag. Measured at 24bfac4a:
  `f(1, 2, 3)` gives `[call "f" [binary "," [binary "," [term "1"] [term "2"]] [term "3"]]]`,
  and `f(a => 1, b => 2)` gives
  `[call "f" [binary "=>" [binary "," [binary "=>" [call "a"] [term "1"]] [call "b" fat]] [term "2"]]]`
  (note `Fat` is set on `b` but not on `a`). Lower a `,`/`=>` chain to ONE
  flat `List`, with `Fat[i]` set when element i was the left operand of a
  `=>` or carries the `Fat` flag. The calls slice reads `Call.Args` from it.
- `Paren` — §4.14.6 leaves node-vs-flag open; the vocabulary issue settles it.
  Whichever it chose, this issue is where the two rules that read it get
  written, which is what §4.14.6 says settles the question.

**Hole tests use input perl rejects** (each confirmed with `perl -c` on
5.42), so they cannot go stale when the parser improves: `$x = ;` gives
`[binary "=" [term "$x"] [unknown refusal=not_a_term]]`, `$x = 1 +;` puts the
Unknown under the `+`, and `1 x;` puts it in the count.

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
      `Fat` slice marks elements 0 and 2, and `(a, 1, b, 2)` marks none. The
      unparenthesised chains in `f(1, 2, 3)` and `f(a => 1, b => 2)` lower
      to the same flat `List` shapes as their parenthesised forms.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerAssignListness`: `($a,$b) = (1,2)` and `@a = (1,2)` set
      `ListAssign`, `$a = 1` does not. Run to verify failure, implement, run to
      green, commit.
- [ ] Write `TestLowerOperatorsOverBrokenInput`: `$x = ;` lowers to an
      `Assign` whose `R` is a `Hole`; `$x = 1 +;` to an `Assign` over a
      `Binary "+"` whose `R` is a `Hole`; `1 x;` to a `Repeat` whose `Count`
      is a `Hole`. No panic. Run to verify failure, implement, run to green,
      commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Binary.Op comes from the node's field, retiring findOperatorText (`go test ./internal/lower/ -run '^TestLowerBinaryOpFromField$' -count=1 -v | grep -q '^--- PASS: TestLowerBinaryOpFromField'`)
- [ ] List repeat reads the Paren flag, not a span heuristic (`go test ./internal/lower/ -run '^TestLowerListRepeatFromParenFlag$' -count=1 -v | grep -q '^--- PASS: TestLowerListRepeatFromParenFlag'`)
- [ ] Ranges keep both endpoints, literal ones included (`go test ./internal/lower/ -run '^TestLowerRangeEndpoints$' -count=1 -v | grep -q '^--- PASS: TestLowerRangeEndpoints'`)
- [ ] Fat commas survive in a list, and a comma chain lowers to the same flat List as its parenthesised form (`go test ./internal/lower/ -run '^TestLowerListFat$' -count=1 -v | grep -q '^--- PASS: TestLowerListFat'`)
- [ ] List assignment is distinguished from scalar assignment (`go test ./internal/lower/ -run '^TestLowerAssignListness$' -count=1 -v | grep -q '^--- PASS: TestLowerAssignListness'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] An operator with an operand perl rejects lowers with a Hole in that operand, and nothing panics (`go test ./internal/lower/ -run '^TestLowerOperatorsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLowerOperatorsOverBrokenInput'`)
