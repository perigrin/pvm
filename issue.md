---
title: "[I2] A sub's assignment to an outer variable replaces that variable's type where the sub is declared"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.174295834Z
updated: 2026-10-16T03:37:16.429912515Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:17.233388527Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.233388527Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I2**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/infer.go:264-273`. `written` is applied with `UpdateType(name, typ)`, which overwrites rather than joins.
- **Repro:** `logicB/w2.pl`:
  ```perl
  my $p = 5; sub reset_p { $p = undef } print $p + 1
  ```
  - New psc: `error: ... got Undef`.
  - origin/pu: silent.
  - perl prints 6.
- **Fix:** `UpdateType(name, types.Join(outer, typ))`, or leave the outer type to widenWritten.
- **Confidence:** High. **NEW.** **Issue:** none covers it directly; 01a132ab is related. **Found by:** logic-B #2.


**Decided 2026-10-16 (option c, widen).** Three options were measured (scratchpad flowA report): (a) join outer with what the sub writes adds a false positive for `$p=undef; sub s{$p=5} s(); $p+1`; (b) no write-back adds one too; (c) widening a variable any sub writes to its sigil type -- as `widenWritten` already does for a sub that reads it -- adds no false positive and fixes the repro, at the cost of the "reset sub was called" true positive, which needs call-site effects (filed separately). perl decides each case by whether the sub ran, which PSC cannot see today; (c) is the sound approximation.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A sub's assignment to an outer variable replaces that variable's type where the sub is declared is fixed (`go test ./internal/infer/ -run '^TestSubAssignmentWidensOuterType$' -count=1 -v | grep -q '^--- PASS: TestSubAssignmentWidensOuterType'`)
