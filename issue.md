---
title: "Lower special forms: MapGrepSort.Op, the eval boundary, and the renames"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
- 01a0c10f-9671-733f-8b98-30a78f634b0f
blocks:
- 01a0c113-dfd3-76a4-8b61-ae61b324bcb4
created: 2026-09-20T23:08:13.106524969Z
updated: 2026-10-12T17:09:29.810854417Z
sessions:
- start_sha: 9895b3086efb0ed7feb8aeb6b994c6b245ce47a5
  end_sha: 34b03d318ec7428fe87316e7752b9ef1f1f6164e
  commits: 3
  started_at: 2026-10-11T16:34:02.567653802Z
  ended_at: 2026-10-12T17:09:29.810854417Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T16:34:02.567653802Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-12T17:09:29.810854417Z
observed_paths:
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/lower/export_test.go
- internal/lower/expr.go
- internal/lower/literals.go
- internal/lower/literals_test.go
- internal/lower/node.go
- internal/lower/node_test.go
- internal/lower/special.go
- internal/lower/special_test.go
- internal/lower/vars.go
- internal/lower/vars_test.go
---

## Context

- paths: `internal/lower/special.go` (new), `internal/lower/special_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2 "Special forms"; §4.14.1 rule 4

This slice retires the third and worst of the measured consumer defects.
Shapes below were measured at 24bfac4a; re-measure at execution time. Line
numbers in `internal/infer` drift, so citations also name the function.

**`callElementType` (`infer.go:1307`, the test at `:1315`) tells map from grep by
`strings.HasPrefix(strings.TrimSpace(call.Text(source)), "grep")`** — two
operations with opposite element-type rules, separated by a string match on
source. `MapGrepSort.Op` fixes exactly that: one field, three values, no text.

**§4.14.2's block-form row is STALE and this is the largest correction in the
milestone.** It says "block forms all `Unknown`", marks them PARSER and blocks
them on issue 01a0ac52 (the §4.9.2 brace classification). All three block
forms parse clean:

	map { $_ } @a;          call "map"  resolved=true, block child, term "@a"
	grep { $_ } @a;         call "grep" resolved=true, block child, term "@a"
	sort { $a <=> $b } @a;  call "sort" resolved=true, block child, term "@a"

So `MapGrepSort` is LOWERING for BOTH forms, and §4.14.1 rule 4's
`BlockGuessed` needs revisiting — the flag exists
to record when the brace heuristic fired, and if the lexer's brace stack is now
making the call deterministically there may be nothing to record. Settle that
here by looking at how the block child is produced, and update rule 4.

Remaining rows:

- `DoBlock` LOWERING (rename) — `do { 1 }` gives `[call "do" [block ...]]`.
  `blockResultType` (`infer.go:1417`) already computes a block's value for
  `map`, so the consumer exists.
- `DoFile` LOWERING now — `do 'file.pl';` gives `[call "do" resolved [term "'file.pl'"]]`,
  told from `DoBlock` by the child's kind. Consumer SPECULATIVE: lower it,
  write no rule for it.
- `EvalBlock`, `EvalStr` LOWERING — `eval { 1 }` gives `[call "eval" [block ...]]`;
  `eval "1"` and `eval $s` give `[call "eval" [term ...]]`. No consumer today,
  and §4.14.2 says these are needed anyway: **`EvalStr` is the soundness
  boundary where the result is `Any`.** Lowering them is what lets inference
  stop being wrong there rather than silently right by accident.
- `ArrayLiteral`, `HashLiteral` SHIPPED — the parser's `anon_array` and
  `anon_hash` (`parse.AnonArray`, `parse.AnonHash`, unchanged) lower to
  `ArrayLiteral` and `HashLiteral`; `+{a=>1}` gives
  `[unary "pos" [anon_hash ...]]`. **Statement-start `{a=>1};` is fixed**: it
  gives `[anon_hash [call "a" fat] [term "1"]]`, which is what perl's
  `B::Deparse` reads (`+{'a', 1}`). Lower it as a `HashLiteral`.
  The lowered names are perigrin's ruling of 2026-10-11, recorded by the
  vocabulary issue (01a0c10f-37f8) in §4.14.6.
- `Ref` LOWERING — `\@a` gives `[unary "ref" [term "@a"]]`; `\($a,$b)` gives
  `[unary "ref" [list ...]]`.
- `FileTest` LOWERING (rename) — `-e $f` gives `[call "-e" resolved [term "$f"]]`,
  and `-e;` gives `[call "-e" resolved]` with no operand.
  §4.14.2 marks the distinct kind SPECULATIVE: the old grammar folds `-e` into
  `func1op_call_expression`, the builtin table returns `Bool`, and that works.
  Lower it as a rename because it is free; write no file-test-specific inference
  rule. `-r _` takes a `Bareword` operand per §4.14.1 rule 3, owned by the
  literals issue. A nil operand means `$_`.
- `Readline{Magic}` LOWERING for `<STDIN>`, `<$fh>`, `<>` — each one `term`.
  `<<>>` now lexes too: one `term "<<>>"`. Note this kind's consumer value:
  the `Options.Strict` comment at `infer.go:84` uses `my $line = <$T>` as the canonical
  irreducible `Unknown`, and the kind is what lets it NAME that rather than
  merely produce it.
- `Glob` LOWERING now — `<*.c>;` gives one `term "<*.c>"`, told from a
  readline by its text (§4.11: `<$x>` and `<$x/*>` differ by one
  character); `glob("*.c")` is a `call "glob"`. Consumer SPECULATIVE.
- `LoopEx` SHIPPED — `last FOO` gives `[loop_control "last" [label "FOO"]]`;
  `return` is also `loop_control` (`return (1,2)` gives a `list` child).
  `goto &foo;` now parses as `[loop_control "goto" [term "&foo"]]`.
  `internal/infer` reads `return_expression` (3), and `blockAlwaysExits`
  (`infer.go:3495`) detects `return`/`die`/`exit` by
  TEXT and will need `last`/`next` for correct branch joins — `LoopEx.Op` is
  what makes that a field read.
- `Wantarray` LOWERING (rename) — `wantarray;` gives
  `call "wantarray" resolved=true`. This is the one exact name agreement with
  B::SoN. `types.NarrowByContext` is called with a fixed `ScalarCtx` at
  `infer.go:535` and `:982`; this kind is the first half of giving it a
  sub's caller context.
- `Match{Op}` SHIPPED as an opaque leaf — `s/a/b/g`, `m/x/`, `qr/x/`, `/a/`,
  `tr/a/b/`, `y/a/b/` each one `term` with the raw operator as `Text`.
  `Match{Target,Negated}` LOWERING — `$x =~ /a/` is `binary "=~"`, `$x !~ /a/`
  is `binary "!~"`; a nil `Target` means `$_`.
  `Match{Pattern,Replace,Flags}` PARSER — inside the one `Quote` token by
  design (`internal/lexer/quote.go`). Do not split them here. B::SoN's
  `Transliterate.pm` records why the `tr` case is not a regex: reading `tr`'s
  operands as a pattern compiles something the source never wrote.
  `nearestPatternBefore` (`infer.go:1726`) resolves `$1` by "the last match
  before this point", so **the side table must preserve total source ORDER** —
  that is the position-map issue's problem, noted here because this is the kind
  that depends on it.

**Hole tests use input perl rejects** (confirmed with `perl -c` on 5.42):
`map {;` gives `[call "map" resolved [unknown refusal=unclosed_brace]]`,
`eval {;` gives `[call "eval" resolved [block [unknown refusal=unclosed_opener [statement]]]]`,
and `$x =~ ;` gives `[binary "=~" [term "$x"] [unknown refusal=not_a_term]]`.

**This issue is not split** (perigrin, 2026-10-11: decided). It covers
about a dozen kinds; commit per test, as the steps say, so progress stays
visible inside it.

## Steps

- [ ] Write `TestLowerMapGrepSortOp`: block and expression forms of all three
      lower to `MapGrepSort` with the right `Op`, `Block`/`FirstExpr` and
      `List` — with no source-text inspection anywhere in the path.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerMapGrepSortOp$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestLowerEvalBoundary`: `eval { 1 }` lowers to `EvalBlock` and
      `eval "1"` / `eval $s` to `EvalStr` — the soundness boundary is a distinct
      kind. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerSpecialRenames`: `do { 1 }`, `do 'file.pl'`,
      `wantarray`, `-e $f`, `-e`, `\@a`, `<STDIN>`, `<>`, `<<>>`, `<*.c>`,
      `[1,2]` (to `ArrayLiteral`), `{a=>1}` in expression position and at
      statement start (to `HashLiteral`),
      `last FOO`, `next`, `goto &foo`, `return (1,2)` each lower to their
      §4.14 kind. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLowerMatchTargetAndNegated`: `$x =~ /a/` sets `Target`,
      `$x !~ /a/` sets `Negated` too, and a bare `/a/` leaves `Target` nil
      (meaning `$_`). Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerSpecialFormsOverBrokenInput`: `map {;` lowers to a
      `MapGrepSort` with `Op` "map" and a `Hole` block; `eval {;` to an
      `EvalBlock` whose body holds a `Hole`; `$x =~ ;` to a `Binary "=~"`
      over `Var $x` and a `Hole` (not a `Match`: there is no pattern to
      carry). No panic. Run to verify
      failure, implement, run to green, commit.
- [ ] Update §4.14.2's MapGrepSort row and §4.14.1 rule 4 to what HEAD measures.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] map, grep and sort are separated by a field, retiring the HasPrefix test (`go test ./internal/lower/ -run '^TestLowerMapGrepSortOp$' -count=1 -v | grep -q '^--- PASS: TestLowerMapGrepSortOp'`)
- [ ] String eval is a distinct kind from block eval (`go test ./internal/lower/ -run '^TestLowerEvalBoundary$' -count=1 -v | grep -q '^--- PASS: TestLowerEvalBoundary'`)
- [ ] The special-form renames all produce their §4.14 kind (`go test ./internal/lower/ -run '^TestLowerSpecialRenames$' -count=1 -v | grep -q '^--- PASS: TestLowerSpecialRenames'`)
- [ ] Match carries Target and Negated, with nil meaning the topic variable (`go test ./internal/lower/ -run '^TestLowerMatchTargetAndNegated$' -count=1 -v | grep -q '^--- PASS: TestLowerMatchTargetAndNegated'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] A special form perl rejects lowers to its kind with a Hole where the parse failed, and nothing panics (`go test ./internal/lower/ -run '^TestLowerSpecialFormsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLowerSpecialFormsOverBrokenInput'`)
