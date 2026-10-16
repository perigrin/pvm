---
title: "[S23] Control characters in file names reach the terminal raw."
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.334467853Z
updated: 2026-10-16T03:25:28.697900105Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 29abad9cb87176fb5e3e8aa8e113ffd4e0fad78b
  commits: 3
  started_at: 2026-10-16T01:17:15.75352556Z
  ended_at: 2026-10-16T03:25:28.697900105Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:15.75352556Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:25:28.697900105Z
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
- internal/parse/loader_regular_test.go
- internal/parse/module.go
- internal/parse/staysinroots_test.go
- internal/perl/inc.go
- internal/perl/inc_test.go
- internal/perl/static.go
- internal/perl/static_test.go
- internal/psc/check_command.go
- internal/psc/check_command_test.go
- internal/psc/check_escape_test.go
- internal/psc/inc_roots_test.go
- internal/psc/lsp.go
- internal/psc/resolve_guard_test.go
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S23**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

Control characters in file names reach the terminal raw.
  - Repro: `v/sec7`. The output starts `x 033 ] 0 ; P W N E D \a .pl:1:24: error...`.
  - Where: `infer/diagnostics.go:65`.
  - High. PRE-EXISTING. Found by: security (S7).

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Control characters in file names reach the terminal raw is fixed (`go test ./internal/psc/ -run '^TestCheckEscapesControlCharacters$' -count=1 -v | grep -q '^--- PASS: TestCheckEscapesControlCharacters'`)
