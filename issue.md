---
title: "Lexer feedback M0 deferred: the keyword table and the brace stack"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a70f-3c3d-7ff7-aa8b-e59b91d899ac
- 01a0a75b-5e51-79f1-8575-3f30abd05ac9
- 01a0a75b-5e51-79ef-a59a-3a997f738cca
created: 2026-09-15T23:16:25.851305082Z
updated: 2026-09-15T23:19:25.269707085Z
---

Two lexer gaps that M0 left and that every parser issue after this one
assumes are closed. Both are `toke.c` lexer feedback — spec ch3's subject —
and both are in `internal/lexer`, not the parser.

## Gap 1: `after(Word)` has no keyword table

`expect.go:105-126` returns XTerm for every Word, with the reason recorded:

> A bareword is the one case the lexer genuinely cannot settle... perl decides
> with the symbol table, which a lexer does not have... The full fix is a
> keyword table, which belongs with the keyword-versus-identifier work in M1.

Measured against the shipped lexer:

    "my $t = time / 2;"   ->   Word "time"   UnknownRest "/ 2;"

`time` is a named unary taking no arguments, so the `/` is division. The
lexer guesses term, opens a pattern, and swallows the rest of the file.

## Gap 2: `after(CloseBracket)` has no brace stack

`expect.go:100` returns XOperator for every CloseBracket. Perl distinguishes
a block-closing `}` from a subscript-closing one with `PL_lex_brackstack`;
this lexer has no such stack. Measured:

    "if ($x) { 1 }\n%h = ();"        ->   CloseBracket "}"  Operator "%"  Word "h"
    "if ($x) { 1 }\n/foo/ and die;"  ->   CloseBracket "}"  Operator "/"  ...

A statement after a block starts in TERM position. The lexer thinks operator,
so `%h` is modulus and `/foo/` is division.

Neither gap shows on T2 — grepped, no `%`/`/`/`<` line follows a `}` line
outside string bodies in the 56 files — which is exactly why it needs its own
issue rather than waiting to be discovered. A committed misread of an rv2hv
site is WRONG under the harness, and T1 is 1,070 files nobody has measured
for it.

## Where the table lives

In `internal/lexer`, or a package the lexer imports. A table in the parser
cannot change tokens the lexer already emitted, which is the mistake the
first draft of the named-unary issue made: it put all five of its acceptance
criteria in the parser package while quoting a lexer comment about a lexer
fix.

## Re-baselining

Both fixes move tokens, so `internal/lexer/testdata/corpus.ratchet` and the
lex.t golden stream will move. That is a deliberate re-baseline recorded in
the same commit, which is what the ratchet's "an improvement fails the build
too" rule exists to force.

## Acceptance

- [ ] A keyword table classifies each builtin as named unary, list operator or term (`go test ./internal/lexer/ -run '^TestKeywordTableClassifies$' -count=1 -v | grep -q '^--- PASS: TestKeywordTableClassifies'`)
- [ ] `time / 2` lexes the slash as division, not a pattern (`go test ./internal/lexer/ -run '^TestKeywordAfterNiladicIsOperator$' -count=1 -v | grep -q '^--- PASS: TestKeywordAfterNiladicIsOperator'`)
- [ ] A brace stack distinguishes a block-closing `}` from a subscript-closing one (`go test ./internal/lexer/ -run '^TestBraceStackTracksBlocks$' -count=1 -v | grep -q '^--- PASS: TestBraceStackTracksBlocks'`)
- [ ] A statement after a block starts in term position: `%h` is a hash, `/re/` a pattern (`go test ./internal/lexer/ -run '^TestTermAfterBlockClose$' -count=1 -v | grep -q '^--- PASS: TestTermAfterBlockClose'`)
- [ ] The corpus ratchet and lex.t golden stream are re-baselined in this commit (`go test ./internal/lexer/ -run '^TestLexerRatchet$' -count=1 -v | grep -q '^--- PASS: TestLexerRatchet'`)

Spec: ch3 §3.2, §3.8.1. toke.c `PL_lex_brackstack`, `yyl_keylookup`.
