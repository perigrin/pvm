---
title: "[S19] When incRoots resolves no perl, @INC silently becomes PERL5LIB alone, and the parse changes with the enviro..."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.190755468Z
updated: 2026-10-16T01:16:09.190755468Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S19**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

When incRoots resolves no perl, @INC silently becomes PERL5LIB alone, and the parse changes with the environment.
  - Where: `check_command.go:106-110`.
  - Repro: `v/c6a.pl` reports `got Undef` without PERL5LIB and `got Undef|Int` with it.
  - Without resolution, `croak "x" }` becomes a Hole that also swallows the block's `}`, so the following `print` lowers *inside* the unless block. That is a parser-recovery defect worth its own issue.
  - Medium. NEW. Found by: errors-B #11.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] When incRoots resolves no perl, @INC silently becomes PERL5LIB alone, and the parse changes with the environment is fixed (`go test ./internal/psc/ -run '^TestIncRootsReportsUnresolvedPerl$' -count=1 -v | grep -q '^--- PASS: TestIncRootsReportsUnresolvedPerl'`)
