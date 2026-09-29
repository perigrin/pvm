---
title: A `)` does not see its `{` past a comment
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:03:03.485897094Z
updated: 2026-09-29T22:03:05.121403472Z
sessions:
- start_sha: 3ecd91947e517cd5d15b0e27aa9520246d64df7c
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T22:03:05.121403472Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T22:03:05.121403472Z
---

peekIsOpenBrace skips whitespace but not comments. So for `if ($x) # c`, or a signature followed by a comment, with the `{` on the next line, the brace is not seen: the block's `}` doesn't count as closing a block, and in the signature case the sub refuses. toke.c's yyl_rightparen uses skipspace, which skips comments as well.

Measured with perl 5.42: a file with `sub f ($a) #c)))` then ` { $a }`, followed by `if ($x) # c` then `{ print 1; }` and `%h = ();`, deparses as a sub, a conditional block and a hash assignment.

perl.git t/op/signatures.t:1034 (t086) and :1059 (t087).

## Acceptance criteria

- [ ] A `{` after a comment opens a block (`go test ./internal/lexer/ -run '^TestBlockAfterComment$' -count=1 -v | grep -q '^--- PASS: TestBlockAfterComment'`)
- [ ] A signature followed by a comment parses (`go test ./internal/parse/ -run '^TestSignatureThenComment$' -count=1 -v | grep -q '^--- PASS: TestSignatureThenComment'`)
