---
title: A `)` looks for its `{` inside a queued heredoc body
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:58:55.595516005Z
updated: 2026-09-29T21:05:17.989209859Z
sessions:
- start_sha: b9d79647940b9422bbc05149e4d9970876e4d404
  end_sha: e9260f1981687ce4acae6a901b5da5d8b651d8be
  commits: 1
  started_at: 2026-09-29T20:58:56.991067814Z
  ended_at: 2026-09-29T21:05:17.989209859Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T20:58:56.991067814Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T21:05:17.989209859Z
observed_paths:
- internal/lexer/brackets.go
- internal/lexer/heredoc.go
- internal/lexer/heredocbrace_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
---

When a `)` is followed by a `{`, the `{` opens a block (toke.c yyl_rightparen). The lexer checks this by peeking at the next non-space byte. When a heredoc is queued on the line, that next byte is the start of the heredoc body, not the `{`. So in `foreach my $l (split /\n/, <<'EOF')`, with the body and `{` below it, the loop's closing `}` does not count as closing a block, and the next bare block is read as a subscript. perl's skipspace never sees the body, because it has already been cut out of the line buffer.

Measured with perl 5.42: the foreach, then `{ g(2); h(3); }`, deparse as a loop followed by a bare block holding both statements.

perl.git t/comp/parser.t:169, which orphans four braces further down.

## Acceptance criteria

- [ ] A `{` after a queued heredoc body opens a block (`go test ./internal/lexer/ -run '^TestBlockAfterHeredocBody$' -count=1 -v | grep -q '^--- PASS: TestBlockAfterHeredocBody'`)
