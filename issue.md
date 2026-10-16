---
title: "[S3] overrides misses a module's default @EXPORT."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.573592952Z
updated: 2026-10-16T01:17:18.601506393Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:18.601506393Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:18.601506393Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S3**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

`overrides` misses a module's default `@EXPORT`.
  - Where: `internal/infer/calls.go:88`.
  - Repro: `probe/p7.pl` (`use Foo2; sleep("abc")`) gives `warning: call to sleep: argument 1 expects Int, got Str` on both builds, while perl prints `slept abc`.
  - High. PRE-EXISTING. Found by: contract-B #3.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `overrides` misses a module's default `@EXPORT` is fixed (`go test ./internal/infer/ -run '^TestDefaultExportOverridesBuiltin$' -count=1 -v | grep -q '^--- PASS: TestDefaultExportOverridesBuiltin'`)
