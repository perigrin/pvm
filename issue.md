---
title: "[I7] negateGuard applies De Morgan only one level deep"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.381486091Z
updated: 2026-10-16T03:03:07.888422322Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 9b435686d672051efee1ba667b9380fd8f320817
  commits: 2
  started_at: 2026-10-16T01:17:17.863502306Z
  ended_at: 2026-10-16T03:03:07.888422322Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.863502306Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:03:07.888422322Z
observed_paths:
- docs/plans/2026-09-05-parser-conformance-plan.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/05-statements.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/infer/calls.go
- internal/infer/flow.go
- internal/infer/flow_test.go
- internal/infer/infer.go
- internal/lower/node.go
- internal/lower/positions.go
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I7**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/flow.go:637-651`. It flips `Negated` on children that may themselves be compounds, and flattenGuards ignores `Negated` on a compound.
- **Repro:** `logicB/c4c.pl`:
  ```perl
  if (!((!defined $a1 && !defined $b1) || defined $c1)) { $b1 + 1 }
  ```
  psc reports `got Undef` on both builds; perl prints 4. ponytail-B's mirror case is a missed diagnostic.
- **Fix:** recurse (`Left: negateGuard(l), Right: negateGuard(r)`), or apply ponytail-B's collapse: negateGuard only flips `Negated`, and flattenGuards does `negate = negate != guard.Negated` before its De Morgan switch.
- **Confidence:** High. **PRE-EXISTING.** **Issue:** none. **Found by:** logic-B #5, ponytail-B #3.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `negateGuard` applies De Morgan only one level deep is fixed (`go test ./internal/infer/ -run '^TestNegateGuardIsDeep$' -count=1 -v | grep -q '^--- PASS: TestNegateGuardIsDeep'`)
