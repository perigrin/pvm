---
title: "[I13] A negated ref guard on an Any-typed variable yields a union that fails every slot"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.607868947Z
updated: 2026-10-16T01:16:07.607868947Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I13**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/types/narrowing.go:245-263` and `internal/infer/flow.go:338-349`.
- **Repro:** `logicB/n1.pl`:
  ```perl
  my $v = eval "1"; unless (ref $v) { $v + 1 }
  ```
  New psc gives `error: ... got Undef|Boolean|...|Array|Hash|Code|Glob|IO|Void`; perl prints 2.
  - origin/pu also gave an error on n1, with a narrower union.
  - On n2 (`defined $v && !ref $v`), origin/pu gave a *warning* and new psc gives an *error*, so the severity has risen.
- **Fix:** narrow a `$` variable typed Any or Unknown from `Scalar`, not `Any`.
- **Confidence:** High. **PRE-EXISTING**; the n2 severity rise is NEW. **Issue:** none. **Found by:** logic-B #6.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A negated `ref` guard on an Any-typed variable yields a union that fails every slot is fixed (`go test ./internal/infer/ -run '^TestNegatedRefGuardOnAny$' -count=1 -v | grep -q '^--- PASS: TestNegatedRefGuardOnAny'`)
