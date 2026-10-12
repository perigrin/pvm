---
title: psc analyze reads its dependencies from the lowered tree
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-2fa3-770b-97d3-8d13ad9ed95d
blocks:
- 01a0c115-df67-7d69-8883-73d40b0ae15d
created: 2026-10-11T10:32:15.295165792Z
updated: 2026-10-12T22:10:44.022773092Z
sessions:
- start_sha: 797bed26b3602121b0e30cc887a4cd02148f76e7
  end_sha: ""
  commits: 0
  started_at: 2026-10-12T22:10:44.022773092Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-12T22:10:44.022773092Z
---

## Context

- paths: `internal/psc/analyze_command.go`, `internal/psc/analyze_command_test.go`, `internal/psc/testdata/analyze/` (new)
- commands: `go test ./internal/psc/ -run "^TestAnalyzeCommandMatchesGolden$" -count=1 -v`
- entrypoints: `internal/psc/analyze_command.go:32` (`parser.New()`), `:63` (`analyzeFile(..., p *parser.Parser, ...)`), `:92` (`collectDependencies(n *parser.Node, ...)`)

Goal (5) says no non-test file in `internal/psc` imports `internal/parser`.
`psc analyze` does, and no issue owned it. `collectDependencies` switches on
two tree-sitter kinds, `use_statement` and `require_expression`, and keeps a
child of kind `package` or `bareword` as the module name.

The statements issue lowers `use` and `require` to `UseStmt` with the module
as a field. Measured at 24bfac4a, `use Foo::Bar qw(a b);` parses as
`[use "use" [term "Foo::Bar"] [term "qw(a b)"]]` and `require Foo;` as
`[use "require" [term "Foo"]]`. So `collectDependencies` becomes a walk for
`UseStmt` that reads the module name.

`psc analyze` reads the lowered `UseStmt`, not the CST's `use` node
(perigrin, 2026-10-11: decided): goal (5) says psc reads `internal/parse`
THROUGH `internal/lower`, and the lowered `UseStmt` already names the
module.

**What must not change.** The output. Before the flip, capture what
`psc analyze` prints over a fixture tree into a golden and hold the flipped
command to it. Cases the fixture tree needs: `use strict;`, `use warnings;`,
`use Foo::Bar qw(a b);`, `require Foo;`, `require "foo.pl";`, `use v5.36;`,
`use parent -norequire, 'Base';`, and a nested `lib/` directory. Where the
tree-sitter version's output is wrong (for example a version number listed
as a dependency), fix it in its own commit with a `Diagnostics-changed:`
trailer, as for an inference golden.

Rollback: one-commit revert while `internal/parser` still exists.

## Steps

- [ ] Build the fixture tree under `internal/psc/testdata/analyze/` and
      capture today's `psc analyze` output on it as `analyze.golden`. Commit.
- [ ] Write `TestAnalyzeCommandMatchesGolden`: run the command on the tree
      and compare with the golden. It passes on the legacy path.
- [ ] Flip `analyze_command.go`: `parse.Parse` and `lower.Lower` in
      `analyzeFile`, which takes no parser; `collectDependencies` walks the
      lowered file for `UseStmt`. Run to green. Commit.
- [ ] Write `TestAnalyzeCommandOnBrokenFile`: a file holding
      `use Foo;` then `my $x = ;` (perl rejects the second line) still lists
      `Foo`, without a panic or an error exit. Run to verify failure,
      implement, run to green, commit.
- [ ] Run the psc suite: `go test ./internal/psc/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] psc analyze prints the same dependencies as before over the fixture tree (`go test ./internal/psc/ -run '^TestAnalyzeCommandMatchesGolden$' -count=1 -v | grep -q '^--- PASS: TestAnalyzeCommandMatchesGolden'`)
- [ ] psc analyze no longer imports the tree-sitter path or names its grammar (`test -f internal/psc/analyze_command.go && ! grep -qE 'pvm/internal/parser"|use_statement|require_expression' internal/psc/analyze_command.go`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)

### Negative Scenarios
- [ ] A file with a line perl rejects still lists the dependencies it declares, without a panic (`go test ./internal/psc/ -run '^TestAnalyzeCommandOnBrokenFile$' -count=1 -v | grep -q '^--- PASS: TestAnalyzeCommandOnBrokenFile'`)
