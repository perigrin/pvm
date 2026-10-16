---
title: "[S10] package Foo 1.2 { ... } and class Foo 1.0 {} drop the version."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.848421938Z
updated: 2026-10-16T11:28:13.7031001Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.941951308Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.941951308Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S10**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

`package Foo 1.2 { ... }` and `class Foo 1.0 {}` drop the version. The later child overwrites `Init`.
  - Where: `lower/decl.go:221-232` (60-61).
  - High. NEW. Found by: contract-A #7, plan deviation 7.


**Ruled 2026-10-16 (perigrin): add `Decl.Version`.** A new optional field: a package or class version always goes in `Decl.Version`, and `Init` holds only the block (`package Foo 1.2;` gets Init nil). This amends the 2026-10-12 ruling that fixed Decl's fields; update the §4.14 Decl row.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `package Foo 1.2 { ... }` and `class Foo 1.0 {}` drop the version is fixed (`go test ./internal/lower/ -run '^TestPackageVersionBlockKeepsVersion$' -count=1 -v | grep -q '^--- PASS: TestPackageVersionBlockKeepsVersion'`)
