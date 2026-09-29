---
title: "`qr/(?{<<END})/` lexes its heredoc body as code"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T21:53:09.550184601Z
updated: 2026-09-29T21:53:09.550184601Z
---

A heredoc opened in a pattern's code block, as in `qr/(?{<<END})/` on one line, needs its body taken from the lines after the statement, as perl does (toke.c:11735). The lexer harvested heredoc openers only from /e replacements, so the body lexed as code. perl.git t/base/lex.t:346:

    print "not " unless qr/(?{<<END})/ eq '(?^:(?{<<END}))';
    foo
    END

Measured with perl 5.42: that shape and `m/x(??{<<E2})/` both skip their bodies and run the statements after them. Only the code inside `(?{ })` and `(??{ })` is scanned, so `<<` in regex text stays regex text.

## Acceptance criteria

- [ ] A code block's heredoc takes the following lines (`go test ./internal/lexer/ -run '^TestCodeBlockHeredoc$' -count=1 -v | grep -q '^--- PASS: TestCodeBlockHeredoc'`)
