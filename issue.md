---
title: "The position map: a side table keyed by id, with erased nodes mapping forward"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-2fa3-770b-97d3-8d13ad9ed95d
- 01a0c110-839f-77fb-8a3d-30fa5a8d6e96
- 01a0c113-c3b2-75d8-ae0e-4f146d610c49
blocks:
- 01a12a85-304b-77f1-9bda-e6b73c317305
created: 2026-09-20T23:08:20.306792879Z
updated: 2026-10-12T17:11:42.83804746Z
---

## Context

- paths: `internal/lower/positions.go` (new), `internal/lower/positions_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestPositionMapErasedNodeMapsForward$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/06-incremental-lsp.md` §6.1.6 items 1 and 3

This issue builds the map from `lower.Lower`'s output, so it runs after the
statements issue (which owns `lower.Lower`), the variables issue (its
`($x)` case) and the special-forms issue (its `Match` source-order case).

§6.1.6 item 1 is the boundary this issue defends: **`internal/parse/` stays the
only tree the parser produces and the only tree the LSP's structural features
read.** Positions for every byte, trees for broken code, round-trip, folding,
rename, format and any future extract-variable refactor read the CST. Sorbet
marks exactly that boundary — its `preserveConcreteSyntax` mode
(`ast/desugar/Desugar.cc:32`) exists for one caller, extract-variable
(`main/lsp/ExtractVariable.cc:314`) — because source-rewriting features must
see what the user typed. Type queries do not.

§6.1.6 item 3 specifies the connection: **a side table in both directions, and
an erased CST node maps FORWARD to its survivor.** rust-analyzer's
`expr_map`/`expr_map_back` (`expr_store.rs:168-186`) and Roslyn's
`BoundNode.Syntax` (`BoundNode.cs:24`) are the two shapes; the side table is
chosen here so **the lowered tree carries no syntax pointers**. The map is
many-CST-to-one-lowered, which is what trivia needs too (43,250 T1 nodes,
15.9%, all erased).

The canonical case is `lower.rs:1630` verbatim: the paren's CST node maps to the
inner expression's lowered node, so **hover on the paren resolves a type**. Note
this is independent of how §4.14.6's Paren node-or-flag question was settled by
the vocabulary issue: either way, the paren's BYTES must resolve.

**Total source ORDER is a hard requirement, not a nicety.** §4.14's preamble
states it and names the consumer: `nearestPatternBefore` (`infer.go:1726`)
resolves `$1` by "the last match before this point". A table that can answer
"which node is at this offset" but not "which match precedes this offset" does
not meet the contract. Spans live in this table, keyed by node id — NOT on the
lowered nodes (§6.1.6 item 4: "never stored on CST nodes", and the lowered tree
is a value type for the same reason).

**Do not build incrementality here.** §6.4.3 is explicit: re-run `Analyze` on
the whole tree after each parse. Measured, `Analyze` is 0.5-2 ms against a parse
of 6-500 ms on the same files, so it is nowhere near the tightest constraint. If
it stops fitting the fix is to make `Analyze` incremental per-sub — a much later
project.

## Steps

- [ ] Write `TestPositionMapErasedNodeMapsForward`: lower `($x);` and assert
      that an offset inside the parens resolves to the lowered node for `$x` —
      the `lower.rs:1630` case. Add a trivia case: an offset inside a comment
      between two statements resolves to the enclosing lowered node rather than
      nothing.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestPositionMapErasedNodeMapsForward$" -count=1 -v`
- [ ] Implement, run to green, commit.
- [ ] Write `TestPositionMapPreservesSourceOrder`: lower a file with two matches
      and a `$1` between them, and assert the table answers "the last match
      before this offset" with the first match, not the second — the
      `nearestPatternBefore` contract.
- [ ] Run to verify failure, implement, run to green, commit.
- [ ] Write `TestPositionMapInsideHole`: lower `$x = ;` and
      `if ($x) {` (perl rejects both). An offset inside each `Hole`'s span,
      including the bytes of the second one's Unknown child, resolves to the
      `Hole`, so the LSP can report "no information" rather than a type
      (§4.14.2, Hole row). Run to verify failure, implement, run to
      green, commit.
- [ ] Write `TestLoweredTreeHasNoSyntaxPointers`: a structural assertion that no
      lowered node type holds a `*parse.Node` field (§6.1.6 item 3). Reflection
      over the declared types is enough. Run to verify failure, implement, run
      to green, commit.
- [ ] Write `TestPositionMapOutOfRange`: an offset past the end of the
      source, an offset in trailing whitespace, and a query on an empty
      file each return "no node" without a panic. Run to verify failure,
      implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

### Identity for a lowered node with no parse node of its own (2026-10-11)

Three lowering slices met this and left it here, where ids are used. Decide
one rule, apply it, and test it:

- **01a0c112-8a65 (declarations):** a named sub lowers to `Decl` over an
  `AnonSub` built from the same parse node, and the AnonSub reuses the Decl's
  id. Two lowered nodes then share one key.
- **01a0c110-839f (variables):** `$#{$x}`, `$#$x` and `&$x` are childless
  terms, so their `Deref.Expr` is a Hole with no parse node, and its id is
  the zero value.
- **01a0c113-c3b2 (special forms):** `Readline.Handle` and `Glob.Pattern`
  from `<...>` are bytes inside one token, so they are left nil. `<STDIN>` and
  `<$fh>` therefore lower alike, as do `<>` and `<<>>`.

Options the slices named: share the parent's id (collides in id-keyed
tables), the zero id (collides with the root), nil (loses the fact), or a
derived id such as the parent's id plus a role. A derived id keeps every key
distinct and still maps to the parent's span.

## Acceptance Criteria

### Positive Scenarios
- [ ] An erased paren's bytes resolve forward to the surviving lowered node (`go test ./internal/lower/ -run '^TestPositionMapErasedNodeMapsForward$' -count=1 -v | grep -q '^--- PASS: TestPositionMapErasedNodeMapsForward'`)
- [ ] The table answers last-match-before-offset in source order (`go test ./internal/lower/ -run '^TestPositionMapPreservesSourceOrder$' -count=1 -v | grep -q '^--- PASS: TestPositionMapPreservesSourceOrder'`)
- [ ] No lowered node type carries a syntax pointer (`go test ./internal/lower/ -run '^TestLoweredTreeHasNoSyntaxPointers$' -count=1 -v | grep -q '^--- PASS: TestLoweredTreeHasNoSyntaxPointers'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] An offset inside a Hole, an Unknown's children included, resolves to the Hole and not to a type (`go test ./internal/lower/ -run '^TestPositionMapInsideHole$' -count=1 -v | grep -q '^--- PASS: TestPositionMapInsideHole'`)
- [ ] An offset outside the source resolves to no node, without a panic (`go test ./internal/lower/ -run '^TestPositionMapOutOfRange$' -count=1 -v | grep -q '^--- PASS: TestPositionMapOutOfRange'`)
