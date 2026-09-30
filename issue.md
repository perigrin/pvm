---
title: "comp/final_line_num.t is invalid by design: the ratchet counts its correct refusal as dirty"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-30T00:03:04.44330232Z
updated: 2026-09-30T00:03:06.227778388Z
sessions:
- start_sha: e0b0b839dc99de8f4a391cb2caee59b020124999
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T00:03:06.227778388Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T00:03:06.227778388Z
---

comp/final_line_num.t tests perl's own error reporting. It ends in `print 1+`, and its BEGIN __DIE__ handler prints "ok 1" only when the syntax error is reported on line 13. Without the handler, perl 5.42 dies with "syntax error at ... at EOF" and status 255. The parser refuses it with one missing_operand refusal on line 13, which is the right answer. But the ratchet counts every refusal as dirty, so the file can never be clean, and accepting it would be the bug.

perigrin decided (2026-09-30) to remove it from the harness. It leaves the Unknown ratchet and is held to a stricter assertion instead: it must be refused, first on the line perl reports. A `perl -c` sweep of all 620 files finds no other file 5.42 rejects, apart from blead-only syntax that blead accepts.

## Acceptance criteria

- [ ] A file perl rejects is refused on perl's line (`go test ./internal/parse/ -run '^TestPerlGitTRejects$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRejects'`)
