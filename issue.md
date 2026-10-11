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
- 01a0c111-8c3e-7451-97d9-10307441ba2d
- 01a0c112-8a65-7d5c-94ec-69859b2928d3
- 01a0c113-c3b2-75d8-ae0e-4f146d610c49
- 01a0c113-dfd3-76a4-8b61-ae61b324bcb4
created: 2026-09-20T23:03:15.1924043Z
updated: 2026-10-11T10:32:49.875051223Z
---

## Context

- paths: `internal/lower/node.go` (new), `internal/lower/node_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14 (the Go listing), §4.14.6

Pure declaration: the node types §4.14 specifies, with `nodeBase` carrying the
identity from `internal/lower/id.go`. No lowering rules. Those are separate
issues and each depends on this one.

This issue runs after the status-table issue (01a0c10e-6a8d), because two of
the §4.14.6 decisions below rest on facts that issue re-measured.

Two binding constraints from §6.1.6 shape the declaration:

**Every lowered field is optional (§6.1.6 item 5).** The CST this pass reads is
routinely broken. A lowering that assumes well-formed input panics on the
half-typed code the LSP exists for. rust-analyzer makes every accessor
optional for exactly this reason.

**`Hole` may stand anywhere a `Node` may (§4.14.2, `Hole` row).** The census
shows `Unknown` as a `binary` operand, an `index` base or subscript, and
under `unary` and `declaration`. §4.14.6 leaves open whether `Decl.Init`,
`Index.Base` and `Match.Target` tolerate a `Hole`. This issue SETTLES that as
yes, by test, because the CST already produces those positions.

**A `Hole` is opaque, even over an Unknown with children.** The census said
every Unknown was a childless leaf. At 24bfac4a that is no longer true:
`if ($x) {` gives `[unknown refusal=unclosed_opener [block]]`. The `Hole`
does not lower the children. The position map still covers their bytes,
because the `Hole`'s span covers them. Record this in the `Hole` comment.

§4.14.6 open questions this issue closes by choosing. Record each choice in a
comment at the declaration:

- **`Paren`: node or bit.** §4.14.1 rule 1 says the information survives, and
  `parse.Node.Paren` ships (`parse.go:341`). Whether the lowered tree carries
  a `Paren` node or a `Parenthesized` flag on `Repeat`/`Assign`/`Decl` is
  "settled by writing those two rules and the position-map entry". Only two
  rules read it. Make the call here and say why.
- **`MethodCall.Indirect` / `.Ambiguous`.** The old premise was that
  `new Foo(1)` is always `Unknown`, so the field would always be false. That
  premise is gone: `parse.Node.Indirect` ships (`parse.go:373`). Measured at
  24bfac4a, `package Foo; sub new {1} package main; new Foo(1);` gives
  `call "new" resolved indirect` over `term "Foo"`, `term "1"`. An
  undeclared `new Foo(1);` is still one Unknown (`trailing_tokens`). So
  `Indirect` carries signal: keep it. `Ambiguous` still has no source;
  §4.14.6 says to drop a field that is always false, so drop it unless the
  executor finds a source.
- **`AnonList`/`AnonHash` vs `ArrayLiteral`/`HashLiteral`: DECIDED,
  `ArrayLiteral`/`HashLiteral`** (perigrin, 2026-10-11). §4.14.6 marked this
  "perigrin's call", and §4.14.5 leaned to keeping the local names; the
  ruling takes B::SoN's. Declare the lowered kinds as `ArrayLiteral` (for
  `[ ... ]`) and `HashLiteral` (for `{ ... }`). The PARSER's node kinds,
  `parse.AnonArray` and `parse.AnonHash`, do not change; only the lowered
  names do. Update the §4.14 Go listing (`type AnonList`, `type AnonHash`),
  the §4.14.2 row, §4.14.5's "keep ours" paragraph and §4.14.6 to match.

Out of scope: no dataflow IR, no HIR, no PIR (§6.1.6). This file declares one
tree. Statement kinds are declared by the statements issue, not here.

## Steps

- [ ] Write `TestHoleStandsAnywhereANodeMay`: construct a `Decl` with a `Hole`
      `Init`, an `Index` with a `Hole` `Base`, a `Match` with a `Hole`
      `Target`, and a `Binary` with `Hole` operands. Assert each constructs
      and that a generic child walk reaches the `Hole` without panicking.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestHoleStandsAnywhereANodeMay$" -count=1 -v`
- [ ] Declare the §4.14 types in `internal/lower/node.go` with `nodeBase`
      embedding the identity from `id.go`. Run to green. Commit.
- [ ] Write `TestZeroValueNodeIsUsable`: every declared type's zero value can
      be walked and its optional fields read as nil without panicking
      (§6.1.6 item 5). Run to verify failure, implement, run to green,
      commit.
- [ ] Record the three §4.14.6 decisions in comments at the declarations,
      and update §4.14.6 to strike the ones now settled. Rename
      `AnonList`/`AnonHash` to `ArrayLiteral`/`HashLiteral` in §4.14's
      listing, the §4.14.2 row and §4.14.5.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] A Hole is accepted in Decl.Init, Index.Base, Match.Target and as a Binary operand (`go test ./internal/lower/ -run '^TestHoleStandsAnywhereANodeMay$' -count=1 -v | grep -q '^--- PASS: TestHoleStandsAnywhereANodeMay'`)
- [ ] The package compiles and is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] Every node type is usable at its zero value, with every optional field nil (`go test ./internal/lower/ -run '^TestZeroValueNodeIsUsable$' -count=1 -v | grep -q '^--- PASS: TestZeroValueNodeIsUsable'`)
