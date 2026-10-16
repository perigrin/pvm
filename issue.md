---
title: "[S14] The lexer and parser read use 5.04 differently, so the signature disappears."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.998618527Z
updated: 2026-10-16T01:16:08.998618527Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S14**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lexer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

The lexer and parser read `use 5.04` differently, so the signature disappears.
  - Repro: `v/v504.pl` lowers `sub f ($x)` with no Sig, while perl Deparse shows `use feature ':5.40'; sub f ($x)`.
  - High. PRE-EXISTING.
  - Issue: 01a128af-b7d2-7c2a-81af-068cfa6a08a9. This finding adds the silent signature-loss consequence.
  - Found by: errors-A #4.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] The lexer and parser read `use 5.04` differently, so the signature disappears is fixed (`go test ./internal/lexer/ -run '^TestLexerVersionPadding$' -count=1 -v | grep -q '^--- PASS: TestLexerVersionPadding'`)
