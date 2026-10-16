---
title: "[C3] package or class with no name panics the parser, killing psc check and the LSP"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.040026283Z
updated: 2026-10-16T01:16:07.040026283Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **C3**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/parse/decl.go:1147`, in `finishPackage`: `n.Children[len(n.Children)-1]` runs with zero children. Nothing in `internal/psc` calls `recover()`.
- **Repro:** for each of `package`, `package;\nmy $x;`, `use v5.38;\nclass;` and `package Foo;\nmy $x = 1;\npackage `, `psc check pk.pl` prints `panic: runtime error: index out of range [-1]`. origin/pu's psc is silent on all of them.
- **Impact:**
  - The LSP uses full sync, so it reparses on every keystroke. Typing `package` at the end of a buffer kills `psc lsp`.
  - `psc check dir/` aborts the whole run.
  - OpenDocument can no longer return an error, so the whole-document error path in lsp_handler.go is dead.
- **Fix:** `if k := len(n.Children); k > 0 && n.Children[k-1].Kind == Block`. Add tests for `package`, `package;` and `class;`. Consider a `recover` in OpenDocument that turns a panic into its error return.
- **Confidence:** High. **NEW in reach:** the parser code is pre-existing, but this PR puts it under psc check and the LSP. **Issue:** none. **Found by:** errors-A #1, errors-B #1.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `package` or `class` with no name panics the parser, killing psc check and the LSP is fixed (`go test ./internal/parse/ -run '^TestNamelessPackageDoesNotPanic$' -count=1 -v | grep -q '^--- PASS: TestNamelessPackageDoesNotPanic'`)
