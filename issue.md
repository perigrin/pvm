---
title: "[I13] A negated ref guard on an Any-typed variable yields a union that fails every slot"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.607868947Z
updated: 2026-10-16T03:03:11.5623186Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 9b435686d672051efee1ba667b9380fd8f320817
  commits: 2
  started_at: 2026-10-16T01:17:17.947741843Z
  ended_at: 2026-10-16T03:03:11.5623186Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:17.947741843Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:03:11.5623186Z
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
