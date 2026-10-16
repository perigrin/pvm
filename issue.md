---
title: "[I24] Heredocs lower to a Hole"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.095434453Z
updated: 2026-10-16T01:17:16.583702873Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.583702873Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.583702873Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I24**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lower/literals.go:290-329`. No arm matches `<<`.
- **Repro:** `my $s = <<"E";\na $x\nE` gives `Decl Init:(Hole)`.
- **Impact:** contract-A counted about 1,000 corpus sites, the largest class of non-Unknown CST nodes that lower to a Hole. Every heredoc is typed Any.
- **Fix:** produce `StrLit{Interpolated: delimiter != '\''}`, and add a §4.14.2 row.
- **Confidence:** High. **NEW** (the lowering). **Issue:** none. **Found by:** contract-A #4.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Heredocs lower to a Hole is fixed (`go test ./internal/lower/ -run '^TestHeredocLowers$' -count=1 -v | grep -q '^--- PASS: TestHeredocLowers'`)
