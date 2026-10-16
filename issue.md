---
title: "[I12] A nested sub's return leaks into the outer sub's return type"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.569468284Z
updated: 2026-10-16T01:17:17.562127546Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:17.562127546Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.562127546Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I12**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:**
  - `internal/infer/flow.go:655` (`explicitReturns`) recurses into AnonSub and nested DeclSub.
  - `flow.go:552` (`subReturnExpr`) and `calls.go:270` (`returnsBuiltinResult`) each skip a different subset.
- **Repro:** `probe/r2.pl`:
  ```perl
  sub outer { my $cb = sub { return undef }; return 1 } outer() + 1
  ```
  psc gives `error: got Undef|Int` on both builds; perl prints 2 silently.
- **Fix:** one "returns of this sub" walker that stops at AnonSub and DeclSub, used by all three.
- **Confidence:** High. **PRE-EXISTING.** **Issue:** none. **Found by:** contract-B #1.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A nested sub's `return` leaks into the outer sub's return type is fixed (`go test ./internal/infer/ -run '^TestNestedSubReturnDoesNotLeak$' -count=1 -v | grep -q '^--- PASS: TestNestedSubReturnDoesNotLeak'`)
