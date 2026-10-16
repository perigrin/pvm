---
title: "[I3] for my $i ($n) never defines $i in the loop, writes back to the wrong variable, and carries a false comment"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.220835857Z
updated: 2026-10-16T01:17:17.290568193Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:17.290568193Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.290568193Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I3**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/flow.go:177-182` sets `iter = nil` and scopes `$_`; lines 214-219 write back to `$i`, not `$n`.
  - The comment at :178-181 says `for my $i ($n)` "iterates nothing and scopes `$_`". That is false: perl iterates once with `$i` aliased to `$n`.
- **Repro:** `logicB/f1.pl`:
  ```perl
  my $i = undef; for my $i ($n) { print $i + 1 }
  ```
  - New psc: `error: ... got Undef`.
  - origin/pu: silent.
  - perl prints 4.
- **Fix:**
  - When `loopVar` comes from `n.Var`, define it and write back to the List's scalar.
  - Scope `$_` only for the var-less `foreach ($n)`.
  - Correct the comment.
- **Confidence:** High. **NEW.**
- **Issue:** 01a132ab-7cf7-7af0-b284-c350971557ab, criterion `TestForMyScopesItsVariable`. This finding adds the false-positive repro, the write-back to the wrong variable, and the actively false comment.
- **Found by:** logic-B #3.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `for my $i ($n)` never defines `$i` in the loop, writes back to the wrong variable, and carries a false comment is fixed (`go test ./internal/infer/ -run '^TestForMyDefinesItsVariable$' -count=1 -v | grep -q '^--- PASS: TestForMyDefinesItsVariable'`)
