---
title: The lexer cannot know a bareword is a declared sub, so & and * after one are read as perl reads them only when it is
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T18:54:31.029217304Z
updated: 2026-10-16T18:48:35.621216918Z
---

## Context

- paths: `internal/lexer/expect.go` (after an unknown bareword the lexer assumes a term), `internal/parse/parse.go` (`newParser`, `TokenizeBarewords` -- passes only CORE.tpm's `:bareword` operators), `internal/parse/exports.go` (resolved imports)
- commands: `make test`; `make fuzz-parser`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-09 by the fuzz work on 01a121ac. Whether `&` or `*` after a bareword is an operator depends on whether the bareword is a declared sub, imports included. Measured on 5.42:

- `A &y;` -- "Transliteration replacement not terminated" (`&` is bitwise-and, `y` begins tr).
- `sub A {} A &y;` and `use POSIX; floor &y;` -- call `&y`.
- `A*y` -- "Transliteration pattern not terminated".
- `sub A {} A*y;` and `use POSIX; floor *y;` -- pass the glob `*y`.

The parser knows the declared and imported names; the lexer does not, and assumes a term. 01a121ac contains it by refusing the disagreement (Unknown when the parser reads such a sigil as infix and a quote-op name follows). The faithful fix is to give the lexer the parser's declared names -- local `sub` declarations plus resolved imports -- and treat an unknown bareword as a value, as perl does. Imports the parser cannot resolve stay a hedge.

## Steps

- [ ] Design how the lexer learns declared names (re-lex, or a known-subs set from exports.go); failing tests from the measurements above; drop 01a121ac's refusal where the lexer now decides; ratchets toward perl only; `make fuzz-parser`; `make test`; commit

## Acceptance criteria

- [ ] `&` and `*` after a bareword follow whether it is declared (`go test ./internal/parse/ -run '^TestSigilAfterBarewordFollowsDeclaration$' -count=1 -v | grep -q '^--- PASS: TestSigilAfterBarewordFollowsDeclaration'`)
