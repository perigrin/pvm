---
title: A heredoc in a multi-line /e replacement takes its body from after the statement
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T21:43:01.370947605Z
updated: 2026-09-29T21:43:02.911664642Z
sessions:
- start_sha: e9e9ce1063f8e622f3cc98776f202011fc20c4de
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T21:43:02.911664642Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T21:43:02.911664642Z
---

queueHeredocsIn queues every heredoc opener in an `s///e` replacement for the lines after the statement. When the replacement spans lines, perl takes the body from the replacement's own next line: toke.c's scan_heredoc looks for a newline in the construct's buffer first and climbs to the parent's only when there is none (toke.c:11735-11738). perl.git t/base/lex.t:323:

    s|(?:)|"${\<<END}"
    ok $test - heredoc in "" in multiline s///e outside eval
    END
    |e;

Here the lexer took the next thirteen lines of the file as the body, which swallowed the `/(?{print <<END ... })/` at line 336.

Measured with perl 5.42: the same shape followed by `print $_;` prints the inner line.

## Acceptance criteria

- [ ] A multi-line /e replacement keeps its heredoc body (`go test ./internal/lexer/ -run '^TestReplacementHeredocBodyInside$' -count=1 -v | grep -q '^--- PASS: TestReplacementHeredocBodyInside'`)
