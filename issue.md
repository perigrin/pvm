---
title: "[I15] my @a = (1) narrows the aggregate to its one element's scalar type"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.684842992Z
updated: 2026-10-16T01:17:18.378077154Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:18.378077154Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:18.378077154Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I15**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/ops.go:180-182`.
- **Repro:** `v/eb3.pl`: `my @a = ("x"); push @a, 3` gives `error: push ... expects Array, got Str` on both builds; perl is silent.
- **Confidence:** High. **PRE-EXISTING.**
- **Issue:** 01a13318-c145-7143-807b-8f3b68ea1558 ("my @a = (1) types the array as Int"). Covered; this finding adds only the shift/push false-positive examples.
- **Found by:** errors-B #3.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `my @a = (1)` narrows the aggregate to its one element's scalar type is fixed (`go test ./internal/infer/ -run '^TestParenInitialiserIsListContext$' -count=1 -v | grep -q '^--- PASS: TestParenInitialiserIsListContext'`)
