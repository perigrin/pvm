---
title: A `)` does not see its `{` past a comment
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:03:03.485897094Z
updated: 2026-09-29T22:09:19.881602364Z
sessions:
- start_sha: 3ecd91947e517cd5d15b0e27aa9520246d64df7c
  end_sha: e5da762ddfc9da30d8dd6e16c6d2fe3d8f008058
  commits: 1
  started_at: 2026-09-29T22:03:05.121403472Z
  ended_at: 2026-09-29T22:09:19.881602364Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T22:03:05.121403472Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T22:09:19.881602364Z
observed_paths:
- internal/lexer/bracecomment_test.go
- internal/lexer/brackets.go
- internal/parse/sigcomment_test.go
- internal/parse/testdata/perlgitt.ratchet
---

peekIsOpenBrace skips whitespace but not comments. So for `if ($x) # c`, or a signature followed by a comment, with the `{` on the next line, the brace is not seen: the block's `}` doesn't count as closing a block, and in the signature case the sub refuses. toke.c's yyl_rightparen uses skipspace, which skips comments as well.

Measured with perl 5.42: a file with `sub f ($a) #c)))` then ` { $a }`, followed by `if ($x) # c` then `{ print 1; }` and `%h = ();`, deparses as a sub, a conditional block and a hash assignment.

perl.git t/op/signatures.t:1034 (t086) and :1059 (t087).

## Acceptance criteria

- [ ] A `{` after a comment opens a block (`go test ./internal/lexer/ -run '^TestBlockAfterComment$' -count=1 -v | grep -q '^--- PASS: TestBlockAfterComment'`)
- [ ] A signature followed by a comment parses (`go test ./internal/parse/ -run '^TestSignatureThenComment$' -count=1 -v | grep -q '^--- PASS: TestSignatureThenComment'`)
