---
title: "Lower statements and the file: blocks, conditionals, loops, use, phasers, and lower.Lower"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-9671-733f-8b98-30a78f634b0f
blocks:
- 01a0c113-dfd3-76a4-8b61-ae61b324bcb4
- 01a12a85-34bf-7ca4-ac7f-1a9218918da1
- 01a12a85-3566-7ec5-9dcc-e864270f3415
created: 2026-10-11T10:32:13.987176039Z
updated: 2026-10-12T22:00:44.089570958Z
sessions:
- start_sha: 34b03d318ec7428fe87316e7752b9ef1f1f6164e
  end_sha: 1f1ba8482f96293cc81b90102002e4c610824d4d
  commits: 5
  started_at: 2026-10-12T17:10:10.881656312Z
  ended_at: 2026-10-12T22:00:44.089570958Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-12T17:10:10.881656312Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-12T22:00:44.089570958Z
observed_paths:
- docs/specs/perl-parser/04-expressions.md
- internal/lower/decl.go
- internal/lower/decl_test.go
- internal/lower/export_test.go
- internal/lower/expr.go
- internal/lower/id.go
- internal/lower/id_test.go
- internal/lower/lower.go
- internal/lower/node.go
- internal/lower/node_test.go
- internal/lower/ops.go
- internal/lower/ops_test.go
- internal/lower/slice.go
- internal/lower/slice_test.go
- internal/lower/special.go
- internal/lower/stmt.go
- internal/lower/stmt_test.go
- internal/lower/vars.go
- internal/lower/vars_test.go
---

## Context

- paths: `internal/lower/stmt.go` (new), `internal/lower/stmt_test.go` (new), `internal/lower/lower.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerConditionals$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.5 (the `unless`, `until` and modifier rows); `docs/specs/perl-parser/05-statements.md` §5.17; `internal/parse/parse.go` (the `Conditional`, `Loop`, `Label`, `Use`, `Phaser`, `Block`, `Statement` kinds and the `Modifier`, `LoopVar`, `CStyle` flags)

§4.14 lists expression kinds only. It has no statement kinds, and no issue in
this milestone lowers statements. But `internal/infer` reads them:
`conditional_statement`, `elsif`, `else`, `loop_statement`, `for_statement`,
`block`, `expression_statement`, `use_statement` and `package_statement` all
appear in a `case` arm in `infer.go` or `symbols.go`. Flow narrowing
(`extractNegatedGuard`, `blockAlwaysExits`) lives on these shapes. Without
this slice the inference port has nothing to walk above an expression.

This issue also owns the file-level entry point, `lower.Lower`, which takes a
`parse.Parse` root, allocates ids (`internal/lower/id.go`), and returns the
lowered file. The expression slices each lower expression nodes; this one
ties them together.

**Names.** §4.14.5 says chapter 5 owns the statement kinds, and §5.17 lists
them: `ExprStmt`, `BlockStmt`, `IfStmt`, `WhileStmt`, `CForStmt`,
`ForeachStmt`, `UseStmt`, `PhaserStmt`. §5.17's shapes carry a `Span` on
every node. §6.1.6 item 4 forbids that on the lowered tree: spans live in the
position map, keyed by id. Use §5.17's names, drop `Span`, embed `nodeBase`
(perigrin, 2026-10-11: decided).

Measured at 24bfac4a with `parse.Parse`; re-measure at execution time:

	if ($x) { 1 }                 conditional "if" [term "$x"] [block ...]
	$y = 1 if $x;                 conditional "if" [binary "=" ...] [term "$x"]   Modifier set
	unless ($x) { 1 } else { 2 }  conditional "unless" [term "$x"] [block] [conditional "else" [block]]
	while ($x) { 1 }              loop "while" [term "$x"] [block]
	foreach my $i (@a) { 1 }      loop "foreach" [declaration "my" [term "$i"]] [term "@a"] [block]
	LABEL: for (;;) { last LABEL }  label "LABEL" then loop "for" [block]
	do { 1 } until $x;            loop "until" [call "do" resolved [block]] [term "$x"]
	BEGIN { 1 }                   phaser "BEGIN" [block]
	use Foo::Bar qw(a b);         use "use" [term "Foo::Bar"] [term "qw(a b)"]
	require Foo;                  use "require" [term "Foo"]

The §4.14.5 rules this slice must honour:

- **`unless` is a branch swap, not a synthesised `Not`.** `perl -MO=Deparse`
  gives `if ($c) {B} else {A}` for `unless ($c) {A} else {B}`. A synthesised
  `Not` is a node flow narrowing would have to see through.
- **The modifier form has its children in the opposite order** (body, then
  condition). Dispatch on the `Modifier` flag (`parse.go:396`), not on child
  position guesses.
- **`do {} until` is STRUCTURE.** perl: `$m++ until 1` leaves 0, and
  `do { $n++ } until 1` leaves 1. The do-form needs a bit (`PostCond`) on
  `WhileStmt`. The plain `until` is VARIATION: a `WhileStmt` with `Negated`.

Out of scope: the package `declaration` is the declarations slice's
(01a0c112-8a65). `given`/`when`, `try`, `defer` and `format` lower to a
`Hole` until a consumer needs them; name each in a comment.

**Hole tests use input perl rejects** (confirmed with `perl -c` on 5.42):
`if (;) { 1 }` gives `[conditional "if" [unknown refusal=not_a_term] [block ...]]`,
and `if ($x) {` gives `[conditional "if" [term "$x"] [unknown refusal=unclosed_opener [block]]]`,
an Unknown with a child.

## Steps

- [ ] Write `TestLowerConditionals`: the block form, the modifier form,
      `unless`/`else` as a branch swap, and an `if`/`elsif`/`else` chain each
      lower to one `IfStmt` shape.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerConditionals$" -count=1 -v`
- [ ] Declare the statement types in `internal/lower/stmt.go`. Implement. Run
      to green. Commit.
- [ ] Write `TestLowerLoops`: `while`, `until`, `foreach my $i (@a)`, the
      C-style `for (;;)`, a labelled loop, and `do { } until` with
      `PostCond` set. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerUseAndPhaser`: `use Foo::Bar qw(a b)`, `require Foo`,
      `require "foo.pl"` and `BEGIN { 1 }` lower to `UseStmt` /
      `PhaserStmt` with the module name and import list as fields. Run to
      verify failure, implement, run to green, commit.
- [ ] Write `TestLowerFileEntry`: `lower.Lower` over a whole file returns its
      statements in source order, gives every lowered node an id, and
      lowers `Trivia` to nothing. Run to verify failure, implement, run to
      green, commit.
- [ ] Write `TestLowerStatementsOverBrokenInput`: `if (;) { 1 }` lowers to an
      `IfStmt` with a `Hole` condition and an intact block; `if ($x) {`
      lowers to an `IfStmt` with an intact condition and a `Hole` body. No
      panic. Run to verify failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Block, modifier, unless and elsif forms lower to one conditional shape, unless as a branch swap (`go test ./internal/lower/ -run '^TestLowerConditionals$' -count=1 -v | grep -q '^--- PASS: TestLowerConditionals'`)
- [ ] Every loop form lowers, and do-until keeps the bit that makes it run once (`go test ./internal/lower/ -run '^TestLowerLoops$' -count=1 -v | grep -q '^--- PASS: TestLowerLoops'`)
- [ ] use, require and phaser blocks lower with their module and list as fields (`go test ./internal/lower/ -run '^TestLowerUseAndPhaser$' -count=1 -v | grep -q '^--- PASS: TestLowerUseAndPhaser'`)
- [ ] lower.Lower returns a whole file in source order with an id on every node (`go test ./internal/lower/ -run '^TestLowerFileEntry$' -count=1 -v | grep -q '^--- PASS: TestLowerFileEntry'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] A conditional perl rejects lowers with a Hole where the parse failed and the rest intact, and nothing panics (`go test ./internal/lower/ -run '^TestLowerStatementsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLowerStatementsOverBrokenInput'`)
