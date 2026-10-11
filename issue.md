---
title: "Cutover part 1: AnalyzeLowered alongside Analyze, and flip the batch checker"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10f-9671-733f-8b98-30a78f634b0f
- 01a0c110-839f-77fb-8a3d-30fa5a8d6e96
- 01a0c111-72d3-70fa-88a8-a9a387f04261
- 01a0c111-8c3e-7451-97d9-10307441ba2d
- 01a0c112-8a65-7d5c-94ec-69859b2928d3
- 01a0c112-a8bb-7d7f-b156-d4ad99474e87
- 01a0c113-c3b2-75d8-ae0e-4f146d610c49
- 01a0c115-88d1-73f0-8d38-a4182644108b
blocks:
- 01a0c115-c0db-7923-b276-984794affd39
created: 2026-09-20T23:10:16.13915176Z
updated: 2026-10-11T10:32:01.839951715Z
---

## Context

- paths: `internal/psc/check_command.go` (`:34`, `:83`, `:94`), `internal/psc/check_command_test.go`, `internal/infer/project.go` (`:78-84`), `internal/infer/project_test.go`
- commands: `go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v`
- entrypoints: `internal/psc/check_command.go:83` (`checkFile`); `internal/infer/project.go:62` (`ProjectIndex.AnalyzeFile`)

The strangler-fig step for the batch caller. By the time this starts, every
inference rule has been ported onto the lowered tree by the seven port
issues, every fixture under `internal/infer/testdata/lowered/` matches its
legacy golden, and `internal/infer`'s own tests and the precision gate run
on the lowered path. This issue only moves callers.

`internal/infer` has two non-test callers in `internal/psc`
(`check_command.go:94` and `lsp.go:53`). This one flips the batch caller
first: no hover, no position queries, no incremental path. The LSP is
01a0c115-c0db.

Three places in `check_command.go` name the tree-sitter path, and all three
change:

	check_command.go:34   p := parser.New()
	check_command.go:83   func checkFile(w ..., p *parser.Parser, path string, opts infer.Options) (bool, error)
	check_command.go:94   _, diags, _ := infer.AnalyzeWithOptions(tree, source, nil, opts)

`checkFile` takes no parser after the flip: it calls `parse.Parse`,
`lower.Lower` and `AnalyzeLowered`. `ProjectIndex.AnalyzeFile`
(`project.go:78-84`) makes the same three calls on the legacy path and
flips the same way. `NewProjectIndex` has no non-test caller at 24bfac4a,
so this changes no output; it moves so `project.go` stops importing
`internal/parser`. Its cache stays keyed by file path and is never
invalidated (`project.go:62-97`); this issue does not add invalidation.

**What the flip must not change.** The goldens already say what PSC reports
on the fixtures. Over a real corpus they say less, so before flipping, run
`psc check` over T1 (PerlOnJava `unit/*.t` at the revision in
`internal/parse/testdata/t1corpus.pin`) on both paths and compare. Measured
on the legacy path at 24bfac4a, one `psc check` per file: 243 of 986 files
report something, 6,945 diagnostics in all (5,938 `coercion-mismatch`, 1,006
`type-mismatch`, 1 `arity-mismatch`), and the run took 29.6 minutes of wall
time (about 19 minutes of CPU). So the comparison is a
one-time measurement, not a criterion (perigrin, 2026-10-11: decided; no
T1 golden is committed): record counts per code before and
after in the commit message, and explain each change by cause under a
`Diagnostics-changed:` trailer. Per-diagnostic explanation of thousands of
lines is not useful; per-cause explanation is.

**Rollback.** Until 01a0c115-df67 deletes the legacy path, this issue is a
one-commit revert: `git revert` of the flip commit restores
`parser.New()` and `AnalyzeWithOptions` in `check_command.go` and
`project.go`, and nothing else depends on it. Keep the flip in one commit
so that stays true.

## Steps

- [ ] Write `TestCheckCommandOnLoweredTree`: run `checkFile` over every
      fixture under `internal/infer/testdata/lowered/` and assert its output
      lines are exactly the fixture's golden diagnostics, formatted by
      `infer.FormatDiagnostic`.
- [ ] Run to verify failure: `go test ./internal/psc/ -run "^TestCheckCommandOnLoweredTree$" -count=1 -v`
- [ ] Run `psc check` over T1 on the legacy path and save the per-code
      counts.
- [ ] Flip `check_command.go:34`, `:83` and `:94`, and `project.go:78-84`,
      in one commit. Move `project_test.go` with it. Run to green.
- [ ] Run `psc check` over T1 on the lowered path. Put both sets of counts
      and the per-cause explanation in the flip commit's message under
      `Diagnostics-changed:`.
- [ ] Write `TestCheckCommandOnBrokenFile`: `psc check` on a file holding
      `my $x = ;` (perl rejects it) exits without a panic and reports no
      type diagnostic inside the `Hole`. Run to verify failure, implement,
      run to green, commit.
- [ ] Run the full suite: `make test`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] The batch checker reports exactly the golden diagnostics from the lowered tree (`go test ./internal/psc/ -run '^TestCheckCommandOnLoweredTree$' -count=1 -v | grep -q '^--- PASS: TestCheckCommandOnLoweredTree'`)
- [ ] Neither the checker nor ProjectIndex imports the tree-sitter path (`test -f internal/psc/check_command.go && test -f internal/infer/project.go && ! grep -q 'pvm/internal/parser"' internal/psc/check_command.go internal/infer/project.go`)
- [ ] ProjectIndex still resolves and caches modules on the lowered path (`go test ./internal/infer/ -run '^TestAnalyzeFileCached$' -count=1 -v | grep -q '^--- PASS: TestAnalyzeFileCached'`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)
- [ ] The whole repository is green (`make test`)

### Negative Scenarios
- [ ] psc check on a file perl rejects does not panic and says nothing about the Hole (`go test ./internal/psc/ -run '^TestCheckCommandOnBrokenFile$' -count=1 -v | grep -q '^--- PASS: TestCheckCommandOnBrokenFile'`)
