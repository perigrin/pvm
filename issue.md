---
title: "A `{` after a format body opens a hash: the format leaves a term expected"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:48:14.412286047Z
updated: 2026-09-29T20:48:16.194054104Z
sessions:
- start_sha: 1dfb3e29036b6b71055c50b1f716583f58b70fdc
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T20:48:16.194054104Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T20:48:16.194054104Z
---

After a format body, the lexer keeps the expect state from the format's `=` (a term). So a bare `{` on the next statement is classified as an anonymous hash, and the block closes after its first statement. toke.c brackets a format like a block: its `=` goes through yyl_leftcurly, which saves XSTATE on the bracket stack (toke.c:6689), and its `.` goes through yyl_rightcurly, which pops it (toke.c:6862).

Measured with perl 5.42: a file with `format X =`, `T`, `.` followed by `SKIP: { print 1; $x = 1; }` deparses with both statements inside the block.

perl.git t/op/write.t:1683. The block there is labelled; the label was not the problem.

## Acceptance criteria

- [ ] The `{` after a format body opens a block (`go test ./internal/lexer/ -run '^TestStatementAfterFormat$' -count=1 -v | grep -q '^--- PASS: TestStatementAfterFormat'`)
