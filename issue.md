---
title: "[S12] Minor lowering issues."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.920490757Z
updated: 2026-10-16T01:16:08.920490757Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S12**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

Minor lowering issues.
  - `my $x = &=;` gives `Var Sigil="&" Name="="`, not a call. This relates to 01a128ad-18f3-7c27-9544-5a077bd91ea8 ("&2 and &== 2 ...").
  - `Foo::->new` keeps a trailing `::`.
  - Medium. NEW. Found by: contract-A #10.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Minor lowering issues is fixed (`go test ./internal/lower/ -run '^TestLowerMinorShapes$' -count=1 -v | grep -q '^--- PASS: TestLowerMinorShapes'`)
