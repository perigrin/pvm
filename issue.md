---
title: "[S8] Name tests disagree, and the ASCII word regex drops names."
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.776498165Z
updated: 2026-10-16T01:16:08.776498165Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S8**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

Name tests disagree, and the ASCII `word` regex drops names.
  - `sub ::foo`, `sub foo'bar` and `use utf8; sub café` lower to an `AnonSub`, so the name is lost, while `anchorKey` (`id.go:99`) still opens an anchor for them.
  - `$h{café}` gives `Key: Hole`, and `::Foo->new` gives `Invocant: Hole`.
  - Where: `decl.go:72` vs `id.go:99`, plus `literals.go:336`.
  - High. NEW (the lowering). Found by: contract-A #3, ponytail-A #11.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Name tests disagree, and the ASCII `word` regex drops names is fixed (`go test ./internal/lower/ -run '^TestLowerUnicodeAndPackageNames$' -count=1 -v | grep -q '^--- PASS: TestLowerUnicodeAndPackageNames'`)
