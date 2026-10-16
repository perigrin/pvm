---
title: "[I25] A CRLF heredoc terminator is never matched, and the rest of the file becomes a silent Hole"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.145225908Z
updated: 2026-10-16T01:17:16.135912797Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.135912797Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.135912797Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I25**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lexer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lexer/heredoc.go:173`, which compares the line with the `\r` still on it.
- **Repro:** `v/crlf2.pl` is `my $s = <<EOT;\r\nabc\r\nEOT\r\nmy $x = undef;\r\nmy $y = $x + 1;\r\n`.
  - New psc is silent, and the lowered tree is `Init:(Hole)` then `Stmts:(Hole)`.
  - origin/pu reports `crlf2.pl:5:9: error: ... got Undef`, and so does the LF version on both builds.
  - `perl -c` says syntax OK.
- **Fix:** `bytes.TrimSuffix(line, "\r")` before comparing; trim `\r` in scanFormatBody as well.
- **Confidence:** High. **NEW in reach** (lexer code pre-existing).
- **Issue:** the format half is 01a11f82-eabd-73f8-a20a-240468c80927 ("A format terminator in a CRLF file may not end the format where perl's does"). The heredoc half is not covered; this finding adds it.
- **Found by:** errors-A #2.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] A CRLF heredoc terminator is never matched, and the rest of the file becomes a silent Hole is fixed (`go test ./internal/lexer/ -run '^TestCRLFHeredocTerminator$' -count=1 -v | grep -q '^--- PASS: TestCRLFHeredocTerminator'`)
