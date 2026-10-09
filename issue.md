---
title: say is read as the builtin whether or not the say feature is on
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T10:00:10.862346609Z
updated: 2026-10-09T10:00:10.862346609Z
---

## Context

- paths: `internal/lexer/keyword.go` (`emptyBraceIsHash`), `internal/lexer/lexer.go` (how `signatures` is tracked), `internal/parse` (`noteFeatures`, the `features` map)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by 01a11fec (2026-10-09). Measured on 5.42: without the say feature, `say {} "x"` is the method call `{ () }->say('x')` and `say {}, "x"` compiles; with `use feature 'say'`, `-E`, `use v5.10` or `use v5.36`, `say {} "x"` and `say({} "x")` are syntax errors and `say {}, "x"` is an anonymous hash. Neither the lexer nor the parser tracks the say feature (`noteFeatures` neither records it nor turns it on for a version bundle), and the parser treats `say` as the builtin everywhere -- so with the feature off pvm builds a say-with-filehandle-block tree where perl builds a method call, and with it on pvm accepts what perl refuses. Options: (1) leave it; (2) the lexer tracks the say feature as it tracks `signatures` and reads the empty brace as a hash only when it is on (recommended by 01a11fec's agent); (3) the parser records `say` and refuses -- splitting the brace decision between lexer and parser, which 01a11fbd removed. Needs perigrin's decision.

## Steps

- [ ] perigrin picks an option; measure `say` as a method vs the builtin across feature states; failing test first; implement; ratchets toward perl only

## Acceptance criteria

- [ ] say's empty brace and bareword reading follow the say feature as perl does (`go test ./internal/lexer/ -run '^TestSayFollowsItsFeature$' -count=1 -v | grep -q '^--- PASS: TestSayFollowsItsFeature'`)
