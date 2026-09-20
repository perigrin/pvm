---
title: The §4.14 vocabulary types and Hole in internal/lower/node.go
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10e-c81e-7565-88fb-d973f18ed667
blocks:
- 01a0c10f-9671-733f-8b98-30a78f634b0f
- 01a0c110-839f-77fb-8a3d-30fa5a8d6e96
- 01a0c111-72d3-70fa-88a8-a9a387f04261
- 01a0c111-8c3e-7451-97d9-10307441ba2d
created: 2026-09-20T23:03:15.1924043Z
updated: 2026-09-20T23:11:29.655582925Z
---

## Context

- paths: `internal/lower/node.go` (new), `internal/lower/node_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14 (the Go listing), §4.14.6

Pure declaration: the node types §4.14 specifies, with `nodeBase` carrying the
identity from `internal/lower/id.go`. No lowering rules -- those are separate
issues and they each depend on this one.

Two binding constraints from §6.1.6 shape the declaration:

**Every lowered field is optional (§6.1.6 item 5).** The CST this pass reads is
routinely broken; a lowering that assumes well-formed input panics on the
half-typed code the LSP exists for. rust-analyzer makes every accessor optional
for exactly this reason.

**`Hole` may stand anywhere a `Node` may (§4.14.2, `Hole` row).** All 4,944 T1
`Unknown` nodes are childless leaves -- 0 with children, 0 nested. The census
shows `Unknown` as a `binary` operand (281), an `index` base or subscript (74),
and once each under `unary` and `declaration`. §4.14.6 leaves open whether
`Decl.Init`, `Index.Base` and `Match.Target` tolerate a `Hole`; this issue
SETTLES that as yes, by test, because the census already shows the CST producing
those positions.

§4.14.6 open questions this issue closes by choosing, with the choice recorded
in a comment at the declaration:

- **`Paren`: node or bit.** §4.14.1 rule 1 says the information survives and
  `parse.Node.Paren` now SHIPS. Whether the lowered tree carries a `Paren` node
  or a `Parenthesized` flag on `Repeat`/`Assign`/`Decl` is "settled by writing
  those two rules and the position-map entry". Only two rules read it. Make the
  call here and say why.
- **`MethodCall.Indirect` / `.Ambiguous`.** `new Foo(1)` is still `Unknown` at
  HEAD, so the honest-uncertainty channel carries no signal. §4.14.6: "If it is
  permanently out of M1 scope, drop both fields rather than ship fields that are
  always false." Decide; do not ship an always-false field.
- **`AnonList`/`AnonHash` vs `ArrayLiteral`/`HashLiteral`.** §4.14.6 marks this
  "perigrin's call". Ask rather than guess.

Out of scope, explicitly: no dataflow IR, no HIR, no PIR (§6.1.6). This file
declares one tree.

## Steps

- [ ] Write `TestHoleStandsAnywhereANodeMay`: construct a `Decl` with a `Hole`
      `Init`, an `Index` with a `Hole` `Base`, a `Match` with a `Hole` `Target`,
      and a `Binary` with `Hole` operands. Assert each constructs and that a
      generic child walk reaches the `Hole` without panicking.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v`
- [ ] Declare the §4.14 types in `internal/lower/node.go` with `nodeBase`
      embedding the identity from `id.go`. Run to green. Commit.
- [ ] Write `TestZeroValueNodeIsUsable`: every declared type's zero value can be
      walked and its optional fields read as nil without panicking (§6.1.6 item 5).
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Record the three §4.14.6 decisions in comments at the declarations, and
      update §4.14.6 to strike the ones now settled.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] A Hole is accepted in Decl.Init, Index.Base, Match.Target and as a Binary operand (`go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v | grep -q "^--- PASS: TestHoleStandsAnywhereANodeMay"`)
- [ ] Every node type is usable at its zero value (`go test ./internal/lower/ -run "^TestZeroValueNodeIsUsable$" -count=1 -v | grep -q "^--- PASS: TestZeroValueNodeIsUsable"`)
- [ ] The package compiles and is green (`go test ./internal/lower/ -count=1`)
