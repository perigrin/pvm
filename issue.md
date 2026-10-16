---
title: "[I21] Niladic builtins called with (), such as wantarray() and time(), lower to a Hole"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.944191985Z
updated: 2026-10-16T01:17:16.455333157Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.455333157Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.455333157Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I21**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lower/expr.go:140-141` and `call.go:93-112`. `index "("` with one child matches no arm.
- **Repro:** `my $t = wantarray();` and `my $t = time();` each give `Decl Init:(Hole)`.
- **Impact:** `return wantarray() ? @r : \@r` loses its Wantarray context source. `time() + 1` is typed Any. contract-A counted 33 sites in perl's own lib.
- **Fix:** lower a one-child `index "("` over a childless `call` as that Call with Parenthesized set.
- **Confidence:** High. **NEW** (the lowering).
- **Issue:** none. 01a11317 is the `time(1)` refusal and does not cover this.
- **Found by:** logic-A #2, contract-A #2.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Niladic builtins called with `()`, such as `wantarray()` and `time()`, lower to a Hole is fixed (`go test ./internal/lower/ -run '^TestNiladicBuiltinWithParens$' -count=1 -v | grep -q '^--- PASS: TestNiladicBuiltinWithParens'`)
