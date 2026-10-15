---
title: "Cutover part 2: flip the LSP, routing position queries through the position map"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c115-a44b-7393-a828-ad29844b20b4
blocks:
- 01a0c115-df67-7d69-8883-73d40b0ae15d
created: 2026-09-20T23:10:23.450787225Z
updated: 2026-10-15T06:09:44.004319907Z
sessions:
- start_sha: 165beaaf607cdf54bb12a37c0fa0baa7a26bc070
  end_sha: ""
  commits: 0
  started_at: 2026-10-15T06:09:44.004319907Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-15T06:09:44.004319907Z
---

## Context

- paths: `internal/psc/lsp.go`, `internal/psc/lsp_test.go`, `internal/psc/lsp_inference_test.go`
- commands: `go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v`
- entrypoints: `internal/psc/lsp.go:44` (`OpenDocument`), `:138` (`TypeAtByte`), `:196` (`DefinitionAtByte`)

The second and last caller. `check_command.go` flipped first (01a0c115-a44b)
because it is a batch path with no position queries. `lsp.go` is harder: it
does not only call `Analyze`, it walks the tree-sitter tree and reads the
annotation map by byte offset. Every one of these changes:

	lsp.go:20       document.tree        *parser.Tree
	lsp.go:21       document.annotations map[uint32]types.Type
	lsp.go:29,37    LSPServer.parser     *parser.Parser, parser.New()
	lsp.go:48,53    s.parser.Parse(source); infer.Analyze(tree, source, nil)
	lsp.go:96       Annotations(uri) map[uint32]types.Type
	lsp.go:138-166  TypeAtByte: deepestNodeAt, then walk Parent() looking up StartByte
	lsp.go:171      deepestNodeAt(node *parser.Node, offset uint32) *parser.Node
	lsp.go:196-235  DefinitionAtByte: deepestNodeAt, then symbolNameFromNode
	lsp.go:237      symbolNameFromNode switches on "scalar", "array", "hash", "varname", "bareword"
	lsp.go:263      varname looks for a "varname" child

After the flip the document holds the `parse.Parse` root, the lowered file,
its position map, and `map[lower.ID]types.Type`. `TypeAtByte` goes
offset -> position map -> node id -> annotation. `DefinitionAtByte` goes
offset -> position map -> lowered node, and takes the symbol name from a
`Var`'s `(Sigil, Name)` or a `Call`'s `Name`. `deepestNodeAt`,
`symbolNameFromNode` and `varname` have no successor: the position map
answers "which node is here", and the lowered node carries its name as
fields. Delete them.

Behaviours that must survive the flip, each with a test:

**Hover on an erased node still resolves.** §6.1.6 item 3's canonical case:
hover inside `($x)` resolves the type of `$x`, because the paren's bytes map
forward to the inner node (`lower.rs:1630`). Today it resolves because the
byte key happened to land on `$x`; after the flip it resolves deliberately.

**Hover inside a Hole reports no information, not a wrong type** (§4.14.2,
Hole row).

**The LSP reports what `psc check` reports.** Both callers now run
`AnalyzeLowered`. For every fixture under `internal/infer/testdata/lowered/`,
`LSPServer.Diagnostics(uri)` must equal the fixture's golden, which
`check_command` already matches. That is the parity check goal (5) needs.

**Structural features.** The original draft asked for a test that the LSP's
structural features still read the CST. At 24bfac4a this server has none:
`lsp_handler.go:75-93` handles hover, definition and documentSymbol, all
type or symbol queries. So there is nothing to assert. §6.1.6 item 1 still
holds for future features; the position map issue's
`TestLoweredTreeHasNoSyntaxPointers` is what keeps the lowered tree from
becoming one.

Do not build incrementality (§6.4.3): re-run analysis on the whole tree
after each parse.

**Rollback.** Until 01a0c115-df67 deletes the legacy path, this issue is a
one-commit revert: `git revert` of the flip commit restores the tree-sitter
document. Keep the flip, including the deletion of `deepestNodeAt`,
`symbolNameFromNode` and `varname`, in one commit so that stays true.

## Steps

- [ ] Write `TestLSPHoverThroughPositionMap`: hover at an offset inside
      `($x)` resolves the type of `$x`.
- [ ] Run to verify failure: `go test ./internal/psc/ -run "^TestLSPHoverThroughPositionMap$" -count=1 -v`
- [ ] Write `TestLSPDiagnosticsMatchGolden`: open each fixture under
      `internal/infer/testdata/lowered/` in an `LSPServer` and assert its
      diagnostics equal the fixture's golden. Run it on the legacy path
      first; any mismatch is a difference between the two callers, triaged
      before the flip.
- [ ] Write `TestLSPHoverInsideHoleReportsNothing`: hover inside the `Hole`
      of `my $x = (;` returns no type rather than the enclosing node's
      type.
- [ ] Write `TestLSPEditToBrokenSourceKeepsServing`: open a clean document,
      replace it with `if ($x) {` (perl rejects it), and assert hover,
      definition and diagnostics answer without a panic.
- [ ] Flip `lsp.go` in one commit: the document fields, `OpenDocument`,
      `Annotations`, `TypeAtByte`, `DefinitionAtByte`; delete
      `deepestNodeAt`, `symbolNameFromNode` and `varname`. Run to green.
- [ ] Run the existing LSP suites: `go test ./internal/psc/ -run "^TestLSP" -count=1`
- [ ] Run the full suite: `make test`
- [ ] Commit.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance Criteria

### Positive Scenarios
- [ ] Hover on a parenthesised expression resolves through the position map (`go test ./internal/psc/ -run '^TestLSPHoverThroughPositionMap$' -count=1 -v | grep -q '^--- PASS: TestLSPHoverThroughPositionMap'`)
- [ ] The LSP reports exactly the golden diagnostics psc check reports (`go test ./internal/psc/ -run '^TestLSPDiagnosticsMatchGolden$' -count=1 -v | grep -q '^--- PASS: TestLSPDiagnosticsMatchGolden'`)
- [ ] Go-to-definition still finds a declared variable (`go test ./internal/psc/ -run '^TestLSPDefinitionAtByte$' -count=1 -v | grep -q '^--- PASS: TestLSPDefinitionAtByte'`)
- [ ] The LSP no longer walks the tree-sitter tree (`test -f internal/psc/lsp.go && ! grep -q 'pvm/internal/parser"' internal/psc/lsp.go && ! grep -qE 'func (deepestNodeAt|symbolNameFromNode|varname)' internal/psc/lsp.go`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)

### Negative Scenarios
- [ ] Hover inside a Hole reports no information rather than a wrong type (`go test ./internal/psc/ -run '^TestLSPHoverInsideHoleReportsNothing$' -count=1 -v | grep -q '^--- PASS: TestLSPHoverInsideHoleReportsNothing'`)
- [ ] An edit to source perl rejects leaves the server answering, without a panic (`go test ./internal/psc/ -run '^TestLSPEditToBrokenSourceKeepsServing$' -count=1 -v | grep -q '^--- PASS: TestLSPEditToBrokenSourceKeepsServing'`)
