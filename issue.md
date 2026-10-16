---
title: "[S15] Config.pm values with \\' or \\\\ escapes are misread."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:09.034520475Z
updated: 2026-10-16T01:17:15.811735271Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:15.811735271Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:15.811735271Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **S15**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/perl/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

Config.pm values with `\'` or `\\` escapes are misread.
  - Where: `internal/perl/inc.go:21`, regex `'([^']*)'`.
  - From memory, perl's configpm escapes `'` and `\` inside single-quoted values. I could not confirm this against configpm source on this machine.
  - Only an apostrophe in the install path triggers it.
  - Medium. NEW. Found by: errors-A #5.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Config.pm values with `\'` or `\\` escapes are misread is fixed (`go test ./internal/perl/ -run '^TestIncConfigEscapes$' -count=1 -v | grep -q '^--- PASS: TestIncConfigEscapes'`)
