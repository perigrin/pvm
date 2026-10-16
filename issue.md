---
title: "[I23] $h{cb}(1) and $r->[0](2) lower to an array element Index, not a call; $c->(1)(2) lowers to a dynamic Method..."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.043376609Z
updated: 2026-10-16T01:17:16.515251901Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.515251901Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.515251901Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I23**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lower/vars.go:401-410` (`subscript` accepts `index "("`), after `indexCall` declines (`call.go:82-100`). Also `arrowCall` (call.go:120).
- **Repro:**
  - `$h{cb}(1);` gives `Index Base:(Key ...) Idx:(NumLit 1)`.
  - `$c->(1)(2);` gives `MethodCall HasParens Invocant:$c Dynamic:(NumLit 1) Args:2`.
  - perl calls the code ref in each case.
- **Impact:** a wrong tree on realistic code (TAP/Harness.pm `$self->{bail_summary}()`). IsDeref reports it as an array deref, and the call is invisible to the call walks. `psc check --strict` on `v/cb.pl` reports nothing.
- **Fix:** `subscript` declines `(`. `indexCall` lowers any other base to `Call{Fn: base, Args, Parenthesized}`. arrowCall requires a `$`-led base before it reads a dynamic method.
- **Confidence:** High. **NEW** (the lowering). **Issue:** none. **Found by:** contract-A #1.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `$h{cb}(1)` and `$r->[0](2)` lower to an array element `Index`, not a call; `$c->(1)(2)` lowers to a dynamic MethodCall is fixed (`go test ./internal/lower/ -run '^TestCodeRefCallThroughSubscript$' -count=1 -v | grep -q '^--- PASS: TestCodeRefCallThroughSubscript'`)
