---
title: "[I1] A sub body's my $x overwrites the type of a same-named outer $x"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.126803459Z
updated: 2026-10-16T01:17:17.175321042Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:17.175321042Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.175321042Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I1**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/infer.go:246-258` walks the DeclSub body in a fresh scope that has no definitions. `ops.go:168-186` (`narrow` → `st.UpdateType`) then climbs to the file-level symbol.
- **Repro:** `logicB/sh1.pl`:
  ```perl
  my $x = 5; sub f { my $x = undef; ... } print $x + 1
  ```
  - New psc: `sh1.pl:4:7: error: left operand of "+": expected Num, got Undef`.
  - origin/pu: silent on line 4.
  - perl -w prints 6, with no warning.
- **Fix:** in pass 2, have `decl` `Define` each declared var in the current scope before narrowing, or re-enter pass 1's scope as the ponytail comment at infer.go:250 says.
- **Confidence:** High. **NEW.**
- **Issue:** 01a132ab-7cf7-7af0-b284-c350971557ab ("Sub bodies and bare blocks are scoped as legacy did") covers the scope. This finding adds something that issue does not describe: the inner `my` leaks into the outer variable and produces an error-severity false positive, not just a lost narrowing.
- **Found by:** logic-B #1.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A sub body's `my $x` overwrites the type of a same-named outer `$x` is fixed (`go test ./internal/infer/ -run '^TestSubBodyMyDoesNotLeakOutward$' -count=1 -v | grep -q '^--- PASS: TestSubBodyMyDoesNotLeakOutward'`)
