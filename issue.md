---
title: "[S24] The check and LSP golden tests use an unresolved-parse oracle, do not call isolatePerl, and accept either s..."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.369856955Z
updated: 2026-10-16T01:16:09.369856955Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S24**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

The check and LSP golden tests use an unresolved-parse oracle, do not call isolatePerl, and accept either severity.
  - Where: `golden_test.go:94` (`parse.Parse`), `check_command_test.go:206`, `lsp_inference_test.go:215`.
  - Test robustness. High. NEW. Found by: contract-B #9.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] The check and LSP golden tests use an unresolved-parse oracle, do not call isolatePerl, and accept either severity is fixed (`go test ./internal/psc/ -run '^TestCheckGoldensUseResolvedParse$' -count=1 -v | grep -q '^--- PASS: TestCheckGoldensUseResolvedParse'`)
