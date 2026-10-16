---
title: "[I11] A sub's return type drops arms of Unknown type, so it claims a concrete type"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.533106671Z
updated: 2026-10-16T01:17:17.480825701Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:17.480825701Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.480825701Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I11**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/infer.go:262` (`explicitReturns | implicitReturn`) and `flow.go:662-665`.
- **Repro:** `logicB/r1.pl`:
  ```perl
  sub k { ...; return undef unless $x; eval { 4 } } $u = k(3); $u + 1
  ```
  New psc gives `r1.pl:11:7: error: ... got Undef`; origin/pu is silent; perl prints 5.
- **Fix:** if any return path, including the fall-through, is Unknown, the return type is Unknown.
- **Confidence:** High. **NEW in behaviour:** origin/pu is silent. **Issue:** none. **Found by:** logic-B #10.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A sub's return type drops arms of Unknown type, so it claims a concrete type is fixed (`go test ./internal/infer/ -run '^TestReturnTypeKeepsUnknownArms$' -count=1 -v | grep -q '^--- PASS: TestReturnTypeKeepsUnknownArms'`)
