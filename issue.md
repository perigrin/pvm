---
title: "Cutover part 2: flip the LSP, routing position queries through the position map"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-09-20T23:10:23.450787225Z
updated: 2026-09-20T23:13:33.139565669Z
---

## Context

- paths: `internal/psc/lsp.go` (lines 22-23, 53, 109, 122, 196), `internal/psc/lsp_test.go`
- commands: `go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v`
- entrypoints: `internal/psc/lsp.go:53`

The second and last caller. `check_command.go` flipped first because it is a
batch path with no position queries; `lsp.go` is harder for exactly the reason
that made it second — it does not only call `Analyze`, it READS the annotation
map by byte offset:

	lsp.go:22-23   diagnostics []infer.Diagnostic; symbols *infer.SymbolTable
	lsp.go:53      annotations, diagnostics, symbols := infer.Analyze(tree, source, nil)
	lsp.go:109     Diagnostics(uri string) []infer.Diagnostic
	lsp.go:122     SymbolTable(uri string) *infer.SymbolTable
	lsp.go:196     DefinitionAtByte(uri string, offset uint32) (*infer.Symbol, bool)

`DefinitionAtByte` is the query that has to change shape: with annotations keyed
by node identity, a byte offset no longer indexes the map directly. **It goes
offset -> position map -> node id -> annotation**, which is what the position
map issue built, and it is why that issue's source-order and Hole requirements
are load-bearing here rather than theoretical.

Two behaviours must survive the flip and both have tests to write:

**Hover on an erased node still resolves.** §6.1.6 item 3's canonical case:
hover inside `($x)` resolves the type of `$x`, because the paren's CST node maps
forward to the inner lowered node (`lower.rs:1630`). Today it resolves because
the paren left no node behind and the byte key happened to land on `$x`. After
the flip it must resolve deliberately.

**Hover inside a Hole reports no information, not a wrong type.** §4.14.2's Hole
row: a position query inside a Hole's span returns the Hole, so the LSP reports
"no information" rather than a type. T1 has 4,944 Unknowns covering 29.1% of
corpus bytes, so this is the common case, not an edge case.

§6.1.6 item 1 stays true after this flip and is worth re-checking in review:
**the LSP's STRUCTURAL features still read the CST.** Folding, rename, format,
round-trip and any future extract-variable read `internal/parse`. Only type
queries go through the lowered tree. A flip that routes a structural feature
through the lowered tree has broken the boundary Sorbet's `preserveConcreteSyntax`
marks, and the fix is to route it back.

Do not build incrementality (§6.4.3): re-run analysis on the whole tree after
each parse. Measured, that is 0.5-2 ms against a 6-500 ms parse.

## Steps

- [ ] Write `TestLSPHoverThroughPositionMap`: a hover at an offset inside `($x)`
      resolves the type of `$x` via offset -> position map -> id -> annotation.
- [ ] Run to verify failure: `go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v`
- [ ] Route `DefinitionAtByte` and the hover path through the position map.
      Run to green. Commit.
- [ ] Write `TestLSPHoverInsideHoleReportsNothing`: a hover inside an `Unknown`
      span returns no type rather than the enclosing node's type.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLSPStructuralFeaturesStillReadCST`: the structural queries the
      server exposes resolve against `internal/parse` and not the lowered tree —
      §6.1.6 item 1's boundary, asserted rather than assumed.
      Run to verify failure, implement, run to green, commit.
- [ ] Flip `lsp.go:53` to `AnalyzeLowered`. Run to green. Commit.
- [ ] Run the full suite: `make test`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Hover on a parenthesised expression resolves through the position map (`go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v | grep -q "^--- PASS: TestLSPHoverThroughPositionMap"`)
- [ ] Hover inside a Hole reports no information rather than a wrong type (`go test ./internal/psc/ -run "^TestLSPHoverInsideHoleReportsNothing$" -count=1 -v | grep -q "^--- PASS: TestLSPHoverInsideHoleReportsNothing"`)
- [ ] Structural features still read the CST (`go test ./internal/psc/ -run "^TestLSPStructuralFeaturesStillReadCST$" -count=1 -v | grep -q "^--- PASS: TestLSPStructuralFeaturesStillReadCST"`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)
- [ ] The whole repository is green (`make test`)
