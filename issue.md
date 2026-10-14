---
title: "Cutover part 1: flip psc check and ProjectIndex to the lowered tree"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-3417-75eb-bf44-d47cf6546e82
- 01a135fc-cffe-7d59-a23f-8f463b6b44d5
- 01a138e3-9661-7a5d-b04a-7377e8881950
- 01a138e3-972f-7271-9d9b-b7b96113186b
- 01a13a80-d202-7048-98af-137c2601f22e
- 01a13b5b-fac7-7334-a6f3-ca09a23281eb
- 01a13b5c-0164-72a7-95ef-65964ddbd311
- 01a13b5c-05b2-7eb2-9b82-4a386be69484
- 01a13b5c-0bcc-7f39-9b52-3aaef8d3341d
- 01a13b5c-140e-7c89-a854-76f806858b3c
- 01a13b5c-1bc6-7122-8f5d-cac48c897010
- 01a13b5c-1ff2-791d-8a21-6135ba4c5103
- 01a13b5c-2758-7067-ad11-78f6e8030a05
- 01a13b5c-2d29-74a2-b5bb-9e6eae83f59b
- 01a13b5c-3610-75e0-be4b-19568623617b
- 01a13b5c-3dbe-713e-aaaa-4b4ea2448789
- 01a13b5c-4409-7838-bc38-145a0fa5d7af
- 01a13b5c-4e8a-7494-ae38-e7f1a75e2ac0
- 01a13b5c-5852-7d1b-8e54-7a796127c31b
- 01a13b5c-6202-7d45-b492-38c913ee4ce0
- 01a13b5c-699e-772e-b664-b7ac44a455b5
- 01a13b5c-7537-7263-9fb5-341ead6d428c
- 01a13b5c-7dff-7edc-a075-cb4d85303e76
- 01a13b5c-8847-7b52-866f-1771f53e1dc7
- 01a13b5c-908b-74d1-a83a-05260908e9c5
- 01a13b5c-983f-7f8e-9f7c-88441e5b7142
- 01a13b5c-a052-7376-86b0-87544c835892
- 01a13793-4023-74f9-9c60-071d7dea2fe0
- 01a13c6a-5dc5-733e-b772-59f86ddebb06
- 01a13c6a-64ef-7104-9f3d-006af699b7c3
- 01a13c6a-69ab-721c-8b39-228563d1e69a
- 01a13c6a-71a1-7b9a-9571-512cd183f09b
- 01a13c6a-78f5-73aa-a52c-0babef634bbf
- 01a13c6a-80c1-7523-81d2-6b32715b44c3
blocks:
- 01a0c115-c0db-7923-b276-984794affd39
created: 2026-09-20T23:10:16.13915176Z
updated: 2026-10-14T17:01:53.51309404Z
sessions:
- start_sha: 1feeef97410aa0897fde9a5c646b36484162cc2b
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T04:07:25.06627144Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T04:07:25.06627144Z
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

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance Criteria

### Positive Scenarios
- [ ] The batch checker reports exactly the golden diagnostics from the lowered tree (`go test ./internal/psc/ -run '^TestCheckCommandOnLoweredTree$' -count=1 -v | grep -q '^--- PASS: TestCheckCommandOnLoweredTree'`)
- [ ] Neither the checker nor ProjectIndex imports the tree-sitter path (`test -f internal/psc/check_command.go && test -f internal/infer/project.go && ! grep -q 'pvm/internal/parser"' internal/psc/check_command.go internal/infer/project.go`)
- [ ] ProjectIndex still resolves and caches modules on the lowered path (`go test ./internal/infer/ -run '^TestAnalyzeFileCached$' -count=1 -v | grep -q '^--- PASS: TestAnalyzeFileCached'`)
- [ ] The psc package is green (`go test ./internal/psc/ -count=1`)

### Negative Scenarios
- [ ] psc check on a file perl rejects does not panic and says nothing about the Hole (`go test ./internal/psc/ -run '^TestCheckCommandOnBrokenFile$' -count=1 -v | grep -q '^--- PASS: TestCheckCommandOnBrokenFile'`)
