---
title: "[S21] checkFile reads each file twice."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.259895693Z
updated: 2026-10-16T01:16:09.259895693Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S21**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

checkFile reads each file twice.
  - Where: `check_command.go:118,123`. Line/col come from one read and spans from the other.
  - Fix: `sess.ParseSourceFrom(path, source, roots...)`.
  - High. NEW. Found by: concurrency #2, errors-B #14, ponytail-B #4.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] checkFile reads each file twice is fixed (`go test ./internal/psc/ -run '^TestCheckReadsFileOnce$' -count=1 -v | grep -q '^--- PASS: TestCheckReadsFileOnce'`)
