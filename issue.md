---
title: psc parse prints the internal/parse CST or the lowered tree, chosen by a flag
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-2fa3-770b-97d3-8d13ad9ed95d
blocks:
- 01a0c115-df67-7d69-8883-73d40b0ae15d
created: 2026-10-11T10:32:15.461832212Z
updated: 2026-10-12T22:10:46.302235363Z
sessions:
- start_sha: 797bed26b3602121b0e30cc887a4cd02148f76e7
  end_sha: ""
  commits: 0
  started_at: 2026-10-12T22:10:46.302235363Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-12T22:10:46.302235363Z
---

## Context

- paths: `internal/psc/parse_command.go`, `internal/psc/parse_command_test.go`, `internal/lower/` (a printer for the lowered tree)
- commands: `go test ./internal/psc/ -run "^TestParseCommandLayers$" -count=1 -v`
- entrypoints: `internal/psc/parse_command.go:32` (`parser.New()`), `:62` (`printTree(w, n *parser.Node, ...)`)

Goal (5) says no non-test file in `internal/psc` imports `internal/parser`.
`psc parse` does: it prints the tree-sitter tree, with `--format tree`
(default) or `--format sexpr` (`parse_command.go:56`). No issue owned it.

perigrin's ruling: once on `internal/parse`, `psc parse` prints BOTH trees,
the `internal/parse` CST and the lowered §4.14 tree, chosen by a flag. The
CST is what the parser produced and what the LSP's structural features
read; the lowered tree is what inference reads. Seeing both is how a
lowering bug gets told apart from a parser bug.

The flag is `--tree cst|lowered`, defaulting to `cst` (perigrin,
2026-10-11: decided), because the CST is the one tree every byte of the
file appears in. `--format tree|sexpr` stays and applies to both trees.

The CST printer prints each node's kind (`parse.Kind.String()`), its `Text`,
and the flags that are set (`resolved`, `arrow`, `paren`, `fat`, `handle`,
`indirect`, ...), and the refusal code on an Unknown. The lowered printer
prints each node's kind and fields, and marks each `Hole`. Put the lowered
printer in `internal/lower` so the lowering tests can use it too.

Rollback: one-commit revert while `internal/parser` still exists.

## Steps

- [ ] Write `TestParseCommandLayers`: on a fixture holding
      `my ($x) = f(); $h->{k}; print $fh "x";`, `--tree cst` prints
      `internal/parse` kind names (`declaration`, `index`, `call`) with the
      `paren`, `arrow` and `handle` flags, and `--tree lowered` prints
      `Decl`, `Key`, `Call` with `ListAssign`, `Arrow` and `Handle`.
      Neither prints a tree-sitter name such as `variable_declaration`.
- [ ] Run to verify failure: `go test ./internal/psc/ -run "^TestParseCommandLayers$" -count=1 -v`
- [ ] Add the lowered printer to `internal/lower`. Flip `parse_command.go` to
      `parse.Parse`, add `--tree`, and print the chosen tree in either
      format. Update `TestParseCommandSExpr`, which asserts the tree-sitter
      `source_file` root. Run to green. Commit.
- [ ] Write `TestParseCommandRejectsUnknownTree`: `--tree hir` is an error
      naming the two accepted values. Run to verify failure, implement, run
      to green, commit.
- [ ] Write `TestParseCommandOnBrokenFile`: on `my $x = ;` (perl rejects it)
      the CST shows the Unknown with its refusal code, the lowered tree
      shows a `Hole`, and the command exits 0. Run to verify failure,
      implement, run to green, commit.
- [ ] Run the psc suite: `go test ./internal/psc/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] psc parse prints the internal/parse CST or the lowered tree by flag, with no tree-sitter names in either (`go test ./internal/psc/ -run '^TestParseCommandLayers$' -count=1 -v | grep -q '^--- PASS: TestParseCommandLayers'`)
- [ ] psc parse no longer imports the tree-sitter path (`test -f internal/psc/parse_command.go && ! grep -q 'pvm/internal/parser"' internal/psc/parse_command.go`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)

### Negative Scenarios
- [ ] An unknown tree name is an error that names the two accepted values (`go test ./internal/psc/ -run '^TestParseCommandRejectsUnknownTree$' -count=1 -v | grep -q '^--- PASS: TestParseCommandRejectsUnknownTree'`)
- [ ] A file perl rejects prints an Unknown in the CST and a Hole in the lowered tree, and exits 0 (`go test ./internal/psc/ -run '^TestParseCommandOnBrokenFile$' -count=1 -v | grep -q '^--- PASS: TestParseCommandOnBrokenFile'`)
