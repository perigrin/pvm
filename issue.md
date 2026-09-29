---
title: "`$ {^XY}` lexes as the variable `$ ` and a brace group"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T13:46:21.291785208Z
updated: 2026-09-29T13:46:21.822402521Z
sessions:
- start_sha: b9356616d0498b89062d8c2e0666ec39e089aef8
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T13:46:21.822402521Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T13:46:21.822402521Z
---

A braced variable name after a spaced sigil -- `$ {^XY}`, `$ { foo }` -- lexes as the one-byte punctuation variable `$ ` followed by a stray brace group. startsDerefExpression skips spaces to see the brace and correctly declines it as an expression, but the name path then read the space as the name.

Measured with perl 5.42:

    $ perl -e '$ {^XY} = 23; print $ {^XY}, " ", ${^XY}, "\n"; ${ foo } = 4; print $ { foo }, "\n";'
    23 23
    4

perl.git t/base/lex.t:129-139 and later, where it refuses 12 times.

## Acceptance criteria

- [ ] A braced name after spaces is one variable (`go test ./internal/lexer/ -run '^TestSpacedBracedName$' -count=1 -v | grep -q '^--- PASS: TestSpacedBracedName'`)
- [ ] REGRESSION GUARD: the base/lex.t golden stream holds (`go test ./internal/lexer/ -run '^TestLexDotTGoldenStream$' -count=1 -v | grep -q '^--- PASS: TestLexDotTGoldenStream'`)
