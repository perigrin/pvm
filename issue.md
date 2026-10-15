---
title: "Retire tree-sitter: infer and psc off internal/parser, internal/parser and gotreesitter deleted, AnalyzeLowered renamed Analyze"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c115-c0db-7923-b276-984794affd39
- 01a12a85-34bf-7ca4-ac7f-1a9218918da1
- 01a12a85-3566-7ec5-9dcc-e864270f3415
blocks:
- 01a0c0e7-8234-7f04-80e5-1fab8682f9bd
- 01a0fa47-2737-7fdb-bd69-63e254d5a770
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
created: 2026-09-20T23:10:31.270515112Z
updated: 2026-10-15T06:46:46.076200488Z
sessions:
- start_sha: 1cf680fa57c3bae16c472d7d63cd7c9f940b741e
  end_sha: ""
  commits: 0
  started_at: 2026-10-15T06:46:46.076200488Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-15T06:46:46.076200488Z
---

## Context

- paths: `internal/infer/infer.go`, `internal/infer/symbols.go`, `internal/infer/lowered.go`, `internal/infer/legacy_capture_test.go`, `internal/infer/retired_test.go` (new), `internal/parser/` (deleted), `go.mod`, `go.sum`
- commands: `go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v`
- entrypoints: `internal/infer/infer.go:65` (`Analyze`), `:122` (`AnalyzeWithOptions`), `:156` (`walkNode`); `internal/infer/symbols.go:255` (`CollectDeclarations`)

This is the last 10-20% of the migration, filed as its own issue because that
is where the discipline usually fails: the old path stays live "just in
case", `internal/parser` keeps a consumer, and the milestone gets called done
at 85%.

**This is the point of no return.** Up to here, each cutover is a
one-commit revert because the legacy path still works. After this issue
deletes it, a regression can only be fixed forward. So the FIRST step,
before any deletion, re-runs both gates against both paths one last time and
records the result in this issue:

1. Re-capture every golden from the legacy path
   (`TestCaptureLegacyGolden -infer.capture`) into a scratch directory and
   diff it against the committed goldens. The only differences allowed are
   goldens that were edited with a `Diagnostics-changed:` trailer. Anything
   else means a golden drifted from what legacy says, and it is triaged
   before deletion.
2. Run `TestInferencePrecision` and record exact / wider / WRONG / no
   answer. Exact must be at least 62 and WRONG zero.
3. Run `TestLoweredGoldenRatchet` and confirm every fixture is in
   `matching.txt`.

Record the three results as a comment on this issue. If any fails, stop.

All four callers are flipped by the time this starts (`psc check`, the LSP,
`psc analyze`, `psc parse`). What remains:

- the legacy walkers in `internal/infer/infer.go` and `symbols.go` that
  switch on tree-sitter grammar names, and `Analyze` / `AnalyzeWithOptions`
  / `CollectDeclarations` on `*parser.Tree`, now with no caller;
- the text re-derivations the lowering replaced: `findOperatorText`
  (`infer.go:656`), `callElementType`'s
  `strings.HasPrefix(..., "grep")` (`:1315`), and the `@_`/`$_[0]`
  re-derivations;
- the `indirect_object` filehandle workaround (`infer.go:1144-1184`);
- the five duplicated builtin lookup sites;
- `legacy_capture_test.go`, the one test file that calls the legacy path.
  The goldens it produced stay; they are the guard from here on;
- whether `internal/parser/` has any consumer left at all.

**The deletion IS the acceptance criterion**, asserted by test: a test that
reads the non-test files of `internal/infer` and fails if a tree-sitter
grammar name remains in a `case` arm. Seed its forbidden list from the
status-table issue's count (01a0c10e-6a8d), so the two measure the same
thing.

**Delete tree-sitter** (perigrin, 2026-10-11). After the infer and psc
cutovers, `internal/parser/` has no consumer. At e18b59c8 it is imported
outside its own directory by `internal/infer` (`infer.go`, `symbols.go`,
`project.go` and three test files), `internal/psc` (`analyze_command.go`,
`check_command.go`, `lsp.go`, `parse_command.go`) and
`test/e2e/psc_check_test.go`; the earlier issues move every one of them,
the e2e test in the test-suite move. Measure again with
`grep -rl 'pvm/internal/parser"' --include='*.go' .` before deleting:
anything left outside `internal/parser/` is a missed caller, fixed before
the deletion. Then delete the whole `internal/parser/` directory, drop
`github.com/odvcencio/gotreesitter` and its `replace` line from `go.mod`,
and run `go mod tidy` so `go.sum` loses it too. Comments that cite
gotreesitter as history (`internal/lexer/quote.go:481`, for one) may stay;
the dependency may not.

Then rename what is named for the transition. `AnalyzeLowered` becomes
`Analyze`, and `lowered.go` folds into the package's normal layout, so no
name records a migration that is over. The legacy `Analyze` is gone by
then, so the name is free. No earlier issue's criterion names
`AnalyzeLowered`, `lowered.go`, `legacy_capture_test.go` or
`internal/parser` in a way the deletion turns red, so this breaks no
earlier issue's verify.

If anything cannot be deleted, **do not leave it silently**: file a
follow-up issue naming what remains and why, and link it here.

## Steps

- [ ] Run the three pre-deletion checks above. Record the results as a
      comment on this issue. Stop if any fails.
- [ ] Write `TestNoTreeSitterVocabularyRemains`: read the non-test `.go`
      files of `internal/infer` and assert none has a tree-sitter grammar
      name in a `case` arm. Seed the list from the status-table count.
- [ ] Run to verify failure: `go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v`
- [ ] Delete the legacy walkers, the legacy entry points and
      `legacy_capture_test.go`. Run to green. Commit.
- [ ] Write `TestNoSourceTextRederivation`: assert `findOperatorText` and the
      `HasPrefix(..., "grep")` test are gone, and that no inference rule
      reaches for `source` to decide which operator or which operation it
      has. Run to verify failure, delete, run to green, commit.
- [ ] Rename `AnalyzeLowered` to `Analyze`, update every caller and test,
      and fold `lowered.go` in. Run the suite. Commit.
- [ ] Confirm nothing outside `internal/parser/` imports it. Delete
      `internal/parser/`, remove gotreesitter from `go.mod`, run
      `go mod tidy`, and run `go build ./...` and `make test`. Commit.
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
- [ ] No tree-sitter grammar name survives in inference (`go test ./internal/infer/ -run '^TestNoTreeSitterVocabularyRemains$' -count=1 -v | grep -q '^--- PASS: TestNoTreeSitterVocabularyRemains'`)
- [ ] No inference rule re-derives an operator or an operation from source text (`go test ./internal/infer/ -run '^TestNoSourceTextRederivation$' -count=1 -v | grep -q '^--- PASS: TestNoSourceTextRederivation'`)
- [ ] The goldens still hold with the legacy path gone (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
- [ ] internal/infer reaches internal/parse only through internal/lower (`go list -f '{{join .Imports " "}}' ./internal/infer | grep -q 'pvm/internal/lower' && ! go list -deps ./internal/infer | grep -q 'pvm/internal/parser$'`)
- [ ] The tree-sitter package is deleted (`test ! -d internal/parser`)
- [ ] The gotreesitter dependency is gone and the module still builds (`! grep -q gotreesitter go.mod && ! grep -q gotreesitter go.sum && go build ./...`)
- [ ] The lowered entry point is named Analyze and takes the lowered tree (`go doc ./internal/infer Analyze | grep -q '^func Analyze(.*lower\.' && ! go doc ./internal/infer AnalyzeLowered`)

### Negative Scenarios
- [ ] Nothing in internal/infer, tests included, still calls the legacy path (`! grep -rl 'pvm/internal/parser"' internal/infer`)
