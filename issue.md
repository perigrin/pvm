---
title: Retire the tree-sitter vocabulary from internal/infer
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c115-c0db-7923-b276-984794affd39
blocks:
- 01a0c0e7-8234-7f04-80e5-1fab8682f9bd
- 01a0fa47-2737-7fdb-bd69-63e254d5a770
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
created: 2026-09-20T23:10:31.270515112Z
updated: 2026-10-02T17:14:28.287408763Z
---

## Context

- paths: `internal/infer/infer.go`, `internal/infer/lowered.go`, `internal/parser/`, `internal/infer/retired_test.go` (new)
- commands: `go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v`
- entrypoints: `internal/infer/infer.go`; `internal/parser/parser.go`

This is the last 10-20% of the migration, filed as its own issue because that is
where the discipline usually fails: the old path stays live "just in case",
`internal/parser` keeps a consumer, and the milestone gets called done at 85%.

Both callers are flipped by the time this starts. What remains:

- the old walkers in `internal/infer/infer.go` that switch on tree-sitter
  grammar names, now with no caller
- the three text-re-derivation sites the lowering replaced:
  `findOperatorText` (`infer.go:645-659`), `callElementType`'s
  `strings.HasPrefix(..., "grep")` (`infer.go:1313`), and the `@_`/`$_[0]`
  re-derivations (`infer.go:3659`, `:3753`, `:4193`)
- the `infer.go:1155-1165` filehandle workaround, whose comment documents the
  false `Str` mismatch it exists to avoid — `Call.Handle` removes the cause
- the five duplicated `GetBuiltin` sites (`infer.go:693`, `:881`, `:1009`,
  `:4004`, `:4048`)
- whether `internal/parser/` still has a non-test consumer at all

**The deletion IS the acceptance criterion**, and it is asserted by test rather
than by inspection: a test that greps the non-test files of `internal/infer` for
tree-sitter grammar names and fails if any remain. That is the executable form
of "the vocabulary migrated", and it is the only one that cannot be satisfied by
a comment.

Use the number the status-table issue established by its stated method (grammar
names in a `case` arm of a `Kind()` switch) as the starting inventory, so this
test and that count measure the same thing.

**`internal/parser/` may still have legitimate consumers** — it is a thin façade
over `gotreesitter` and the LSP's structural features were never its only
possible user. Do not delete the package on the assumption that it is dead;
measure. If it has no non-test consumer, propose removing it and let perigrin
decide, since deleting a package is not a refactor.

If anything cannot be deleted, **do not leave it silently**: file a follow-up
issue naming what remains and why, and link it here. CLAUDE.md: "If you find
yourself saying 'let''s defer this to a future project', write down where it is
deferred to. Unlabeled deferrals become drift."

## Steps

- [ ] Write `TestNoTreeSitterVocabularyRemains`: read the non-test `.go` files of
      `internal/infer` and assert none contains a tree-sitter grammar name in a
      `case` arm. Seed the forbidden list from the status-table issue's count.
- [ ] Run to verify failure: `go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v`
- [ ] Delete the dead walkers. Run to green. Commit.
- [ ] Write `TestNoSourceTextRederivation`: assert `findOperatorText` and the
      `strings.HasPrefix(..., "grep")` test are gone, and that no remaining
      inference rule reaches for `source` to decide which operator or which
      operation it has. Run to verify failure, delete, run to green, commit.
- [ ] Measure whether `internal/parser/` has a non-test consumer. Record the
      answer in the issue; if none, propose removal and ask rather than delete.
- [ ] Fold `internal/infer/lowered.go` back into the package's normal layout if
      the split was only a migration scaffold, so no file is named for a
      transition that is over.
- [ ] Run the full suite: `make test`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] No tree-sitter grammar name survives in inference (`go test ./internal/infer/ -run "^TestNoTreeSitterVocabularyRemains$" -count=1 -v | grep -q "^--- PASS: TestNoTreeSitterVocabularyRemains"`)
- [ ] No inference rule re-derives an operator or an operation from source text (`go test ./internal/infer/ -run "^TestNoSourceTextRederivation$" -count=1 -v | grep -q "^--- PASS: TestNoSourceTextRederivation"`)
- [ ] The infer package is green (`go test ./internal/infer/ -count=1`)
- [ ] The whole repository is green (`make test`)
