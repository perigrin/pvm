---
title: "[S17] Parse time is quadratic in statement nesting."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.116165922Z
updated: 2026-10-16T01:17:16.330265611Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.330265611Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.330265611Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S17**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

Parse time is quadratic in statement nesting.
  - Where: `parse/parse.go:891-925` (`innerHeredocBodies`).
  - Measured with `psc parse`: 40k nested blocks took 32 s, and 80k took 79 s (an 80 KB file).
  - Pathological input. High. Code PRE-EXISTING, newly reached. Found by: security (S4).

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Parse time is quadratic in statement nesting is fixed (`go test ./internal/parse/ -run '^TestNestedBlockParseIsNotQuadratic$' -count=1 -v | grep -q '^--- PASS: TestNestedBlockParseIsNotQuadratic'`)
