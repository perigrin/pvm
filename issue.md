---
title: "[I26] The LSP's long-lived parse.Session never re-reads a module, and caches \"not found\" forever"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.193695538Z
updated: 2026-10-16T01:17:18.730330071Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:18.730330071Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:18.730330071Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I26**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/psc/lsp.go:36,58` keeps one Session per server. `internal/parse/module.go:228-236` shares `seen` and `facts`. `module.go:302,312-313` sets `seen[module]` before loading and never clears it on not-found; `:385-389` does the same for a required file.
- **Confirmed by reading:**
  - A module missing at first open hits `seen` and takes the cycle branch (:295-298) on every later parse.
  - The concurrency specialist reproduced this with a `go test -overlay` (`scratchpad/review/conc/neg_test.go`).
- **Impact:**
  - Edits to a used module never reach the editor, so check and the LSP disagree.
  - A module created after the first open never resolves.
  - The map grows with every distinct directory list.
- **Fix:** use a fresh Session per OpenDocument, or invalidate `facts` and `seen` together. Also delete `seen[module]` on not-found.
- **Confidence:** High. **NEW.**
- **Issue:** 01a13e48-c9a5-79c4-b80b-636c1c9ec5be ("The LSP's parse Session never re-reads a module edited on disk"). This finding adds two things the issue's planned mtime fix cannot catch: the not-found case (there is no file to stat), and the trap where deleting `facts` but not `seen` turns a stale module into a missing one. It also adds the unbounded growth.
- **Found by:** concurrency #1, errors-A #6, errors-B #9, contract-A #9, contract-B #6.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] The LSP's long-lived parse.Session never re-reads a module, and caches "not found" forever is fixed (`go test ./internal/psc/ -run '^TestLSPSessionRereadsModules$' -count=1 -v | grep -q '^--- PASS: TestLSPSessionRereadsModules'`)
