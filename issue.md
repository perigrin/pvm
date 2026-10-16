---
title: The LSP's parse Session never re-reads a module edited on disk
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-10-15T06:38:40.037697818Z
updated: 2026-10-16T12:06:33.89746069Z
sessions:
- start_sha: 77107dfbeb77e7b3f1d94437583eae6d28eb41c3
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T12:06:33.89746069Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T12:06:33.89746069Z
---

## Context

- paths: `internal/psc/lsp.go` (one `parse.Session` for the server's lifetime; `incRoots` per open), `internal/parse/` (Session module cache)
- commands: `go test ./internal/psc/`

Found 2026-10-15 by cutover 2 (01a0c115-c0db). The LSP keeps one `parse.Session` for its lifetime, and the Session caches each module's facts (exports, prototypes) the first time it reads them. A module edited on disk -- e.g. a sub gaining a prototype, or an `@EXPORT` change -- is not re-read until the server restarts, so a buffer that uses it is parsed and checked against stale facts. Also: each open re-resolves @INC (`incRoots`, ~35ms per open, marked `ponytail:`).

## Steps

- [ ] Failing test first: open a buffer that uses a module, edit the module on disk (or open it in the editor and change it), re-open or change the buffer, and see the new facts; invalidate a cached module by mtime/size, or when the editor reports a change to it (didChange/didSave/didChangeWatchedFiles); cache @INC per directory; `make test`; commit

## Acceptance criteria

- [ ] an edited module's facts reach the next analysis of a buffer that uses it (`go test ./internal/psc/ -run '^TestLSPSeesEditedModule$' -count=1 -v | grep -q '^--- PASS: TestLSPSeesEditedModule'`)
