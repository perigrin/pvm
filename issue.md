---
title: "[S23] Control characters in file names reach the terminal raw."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.334467853Z
updated: 2026-10-16T01:17:15.75352556Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:15.75352556Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:15.75352556Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S23**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

Control characters in file names reach the terminal raw.
  - Repro: `v/sec7`. The output starts `x 033 ] 0 ; P W N E D \a .pl:1:24: error...`.
  - Where: `infer/diagnostics.go:65`.
  - High. PRE-EXISTING. Found by: security (S7).

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Control characters in file names reach the terminal raw is fixed (`go test ./internal/psc/ -run '^TestCheckEscapesControlCharacters$' -count=1 -v | grep -q '^--- PASS: TestCheckEscapesControlCharacters'`)
