---
title: "`qr/(?{<<END})/` lexes its heredoc body as code"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T21:53:09.550184601Z
updated: 2026-09-29T21:59:30.496120876Z
sessions:
- start_sha: 24fabeba92826a9f8d3b7cc6558cf7b853434e4b
  end_sha: 3ecd91947e517cd5d15b0e27aa9520246d64df7c
  commits: 1
  started_at: 2026-09-29T21:53:10.377968654Z
  ended_at: 2026-09-29T21:59:30.496120876Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T21:53:10.377968654Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T21:59:30.496120876Z
observed_paths:
- internal/lexer/codeblockheredoc_test.go
- internal/lexer/quote.go
- internal/lexer/scan.go
- internal/lexer/testdata/lex.t.tokens
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
---

A heredoc opened in a pattern's code block, as in `qr/(?{<<END})/` on one line, needs its body taken from the lines after the statement, as perl does (toke.c:11735). The lexer harvested heredoc openers only from /e replacements, so the body lexed as code. perl.git t/base/lex.t:346:

    print "not " unless qr/(?{<<END})/ eq '(?^:(?{<<END}))';
    foo
    END

Measured with perl 5.42: that shape and `m/x(??{<<E2})/` both skip their bodies and run the statements after them. Only the code inside `(?{ })` and `(??{ })` is scanned, so `<<` in regex text stays regex text.

## Acceptance criteria

- [ ] A code block's heredoc takes the following lines (`go test ./internal/lexer/ -run '^TestCodeBlockHeredoc$' -count=1 -v | grep -q '^--- PASS: TestCodeBlockHeredoc'`)
