---
title: "Lower special forms: MapGrepSort.Op, the eval boundary, and the renames"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
created: 2026-09-20T23:08:13.106524969Z
updated: 2026-09-20T23:11:47.05571659Z
---

## Context

- paths: `internal/lower/special.go` (new), `internal/lower/special_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Special forms"; §4.14.1 rule 4

This slice retires the third and worst of the measured consumer defects.

**`callElementType` (`infer.go:1313`) tells map from grep by
`strings.HasPrefix(strings.TrimSpace(call.Text(source)), "grep")`** — two
operations with opposite element-type rules, separated by a string match on
source. `MapGrepSort.Op` fixes exactly that: one field, three values, no text.

**§4.14.2's block-form row is STALE and this is the largest correction in the
milestone.** It says "block forms all `Unknown`", marks them PARSER and blocks
them on issue 01a0ac52 (the §4.9.2 brace classification). Verified at HEAD, all
three block forms parse clean:

	map { $_ } @a;          call "map"  resolved=true, block child, term "@a"
	grep { $_ } @a;         call "grep" resolved=true, block child, term "@a"
	sort { $a <=> $b } @a;  call "sort" resolved=true, block child, term "@a"

So `MapGrepSort` is LOWERING for BOTH forms. Re-verify before implementing; if
it holds, §4.14.1 rule 4's `BlockGuessed` also needs revisiting — the flag exists
to record when the brace heuristic fired, and if the lexer's brace stack is now
making the call deterministically there may be nothing to record. Settle that
here by looking at how the block child is produced, and update rule 4.

Remaining rows, all verified at HEAD:

- `DoBlock` LOWERING (rename) — `do { 1 }` gives `[call "do" [block ...]]`.
  `blockResultType` (`infer.go:1417`) already computes a block's value for
  `map`, so the consumer exists.
- `DoFile` PARSER — `do 'file.pl';` is `Unknown`. Consumer SPECULATIVE. Hole.
- `EvalBlock`, `EvalStr` LOWERING — `eval { 1 }` gives `[call "eval" [block ...]]`;
  `eval "1"` and `eval $s` give `[call "eval" [term ...]]`. No consumer today,
  and §4.14.2 says these are needed anyway: **`EvalStr` is the soundness
  boundary where the result is `Any`.** Lowering them is what lets inference
  stop being wrong there rather than silently right by accident.
- `AnonList`, `AnonHash` SHIPPED — `anon_array`, `anon_hash`; `+{a=>1}` gives
  `[unary "pos" [anon_hash ...]]`. Statement-start `{a=>1};` is still a
  wrong-tree PARSER case (it parses as a `block`; perl's `B::Deparse` gives
  `+{'a', 1}`). Do not lower the wrong tree — emit a `Hole` for that case or
  lower the `block` faithfully as a block, whichever the executor can justify.
  The naming (`AnonList`/`AnonHash` vs `ArrayLiteral`/`HashLiteral`) was settled
  by the vocabulary issue per §4.14.6.
- `Ref` LOWERING — `\@a` gives `[unary "ref" [term "@a"]]`; `\($a,$b)` gives
  `[unary "ref" [list ...]]`.
- `FileTest` LOWERING (rename) — `-e $f` gives `[call "-e" resolved=true [term "$f"]]`.
  §4.14.2 marks the distinct kind SPECULATIVE: the old grammar folds `-e` into
  `func1op_call_expression`, the builtin table returns `Bool`, and that works.
  Lower it as a rename because it is free; write no file-test-specific inference
  rule. `-r _` takes a `Bareword` operand per §4.14.1 rule 3, owned by the
  literals issue. A nil operand means `$_`.
- `Readline{Magic}` LOWERING for `<STDIN>`, `<$fh>`, `<>` — each one `term`.
  `<<>>` is PARSER (the lexer does not know the 5.22 form). Note this kind's
  consumer value: `infer.go:78-84` uses `my $line = <$T>` as the canonical
  irreducible `Unknown`, and the kind is what lets it NAME that rather than
  merely produce it.
- `Glob` PARSER — `<*.c>;` is `Unknown`; `glob("*.c")` is a `call "glob"`.
  Consumer SPECULATIVE.
- `LoopEx` SHIPPED — `last FOO` gives `[loop_control "last" [label "FOO"]]`;
  `return` is also `loop_control` (`return (1,2)` gives a `list` child).
  `goto &foo;` is PARSER. `internal/infer` reads `return_expression` (3), and
  `blockAlwaysExits` (`infer.go:3312-3342`) detects `return`/`die`/`exit` by
  TEXT and will need `last`/`next` for correct branch joins — `LoopEx.Op` is
  what makes that a field read.
- `Wantarray` LOWERING (rename) — `wantarray;` gives
  `call "wantarray" resolved=true`. This is the one exact name agreement with
  B::SoN. `NarrowByContext` (`infer.go:1291`) exists and has no context source;
  this kind is the first half of giving it one.
- `Match{Op}` SHIPPED as an opaque leaf — `s/a/b/g`, `m/x/`, `qr/x/`, `/a/`,
  `tr/a/b/`, `y/a/b/` each one `term` with the raw operator as `Text`.
  `Match{Target,Negated}` LOWERING — `$x =~ /a/` is `binary "=~"`, `$x !~ /a/`
  is `binary "!~"`; a nil `Target` means `$_`.
  `Match{Pattern,Replace,Flags}` PARSER — inside the one `Quote` token by
  design (`internal/lexer/quote.go`). Do not split them here. B::SoN's
  `Transliterate.pm` records why the `tr` case is not a regex: reading `tr`'s
  operands as a pattern compiles something the source never wrote.
  `nearestPatternBefore` (`infer.go:1724-1751`) resolves `$1` by "the last match
  before this point", so **the side table must preserve total source ORDER** —
  that is the position-map issue's problem, noted here because this is the kind
  that depends on it.

## Steps

- [ ] Write `TestLowerMapGrepSortOp`: block and expression forms of all three
      lower to `MapGrepSort` with the right `Op`, `Block`/`FirstExpr` and
      `List` — with no source-text inspection anywhere in the path.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerEvalBoundary`: `eval { 1 }` lowers to `EvalBlock` and
      `eval "1"` / `eval $s` to `EvalStr` — the soundness boundary is a distinct
      kind. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerSpecialRenames`: `do { 1 }`, `wantarray`, `-e $f`, `\@a`,
      `<STDIN>`, `<>`, `[1,2]`, `{a=>1}` in expression position, `last FOO`,
      `next`, `return (1,2)` each lower to their §4.14 kind. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLowerMatchTargetAndNegated`: `$x =~ /a/` sets `Target`,
      `$x !~ /a/` sets `Negated` too, and a bare `/a/` leaves `Target` nil
      (meaning `$_`). Run to verify failure, implement, run to green, commit.
- [ ] Update §4.14.2's MapGrepSort row and §4.14.1 rule 4 to what HEAD measures.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] map, grep and sort are separated by a field, retiring the HasPrefix test (`go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v | grep -q "^--- PASS: TestLowerMapGrepSortOp"`)
- [ ] String eval is a distinct kind from block eval (`go test ./internal/lower/ -run "^TestLowerEvalBoundary$" -count=1 -v | grep -q "^--- PASS: TestLowerEvalBoundary"`)
- [ ] The special-form renames all produce their §4.14 kind (`go test ./internal/lower/ -run "^TestLowerSpecialRenames$" -count=1 -v | grep -q "^--- PASS: TestLowerSpecialRenames"`)
- [ ] Match carries Target and Negated, with nil meaning the topic variable (`go test ./internal/lower/ -run "^TestLowerMatchTargetAndNegated$" -count=1 -v | grep -q "^--- PASS: TestLowerMatchTargetAndNegated"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)
