---
title: "[S20] psc parse --format accepts any value."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.226023153Z
updated: 2026-10-16T01:17:19.15607547Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:19.15607547Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:19.15607547Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S20**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

`psc parse --format` accepts any value.
  - Where: `parse_command.go:37,60`. `--tree` is validated; `--format` is not.
  - High. PRE-EXISTING (origin/pu used the same switch). Found by: errors-B #13.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `psc parse --format` accepts any value is fixed (`go test ./internal/psc/ -run '^TestParseCommandRejectsUnknownFormat$' -count=1 -v | grep -q '^--- PASS: TestParseCommandRejectsUnknownFormat'`)
