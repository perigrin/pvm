---
title: A `)` looks for its `{` inside a queued heredoc body
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:58:55.595516005Z
updated: 2026-09-29T20:58:55.595516005Z
---

When a `)` is followed by a `{`, the `{` opens a block (toke.c yyl_rightparen). The lexer checks this by peeking at the next non-space byte. When a heredoc is queued on the line, that next byte is the start of the heredoc body, not the `{`. So in `foreach my $l (split /\n/, <<'EOF')`, with the body and `{` below it, the loop's closing `}` does not count as closing a block, and the next bare block is read as a subscript. perl's skipspace never sees the body, because it has already been cut out of the line buffer.

Measured with perl 5.42: the foreach, then `{ g(2); h(3); }`, deparse as a loop followed by a bare block holding both statements.

perl.git t/comp/parser.t:169, which orphans four braces further down.

## Acceptance criteria

- [ ] A `{` after a queued heredoc body opens a block (`go test ./internal/lexer/ -run '^TestBlockAfterHeredocBody$' -count=1 -v | grep -q '^--- PASS: TestBlockAfterHeredocBody'`)
