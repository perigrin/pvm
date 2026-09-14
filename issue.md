---
title: "Lexer feedback: term-vs-operator position"
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:44.556728897Z
updated: 2026-09-14T15:42:56.331157141Z
---

`/` is division or a pattern; `<FH>` is a readline or a glob; `x` is an
operator or the start of an identifier. Perl decides with one bit of lexer
state — is a TERM expected, or an OPERATOR — and a static lexer has to carry
the same bit.

Spec ch3 owns this and documents perl's `PL_expect` machine at
`perl.h:5972-5985`: 11 states. This issue implements the subset M0's corpus
needs, not all 11.

## The measured cases

§0.13 rank 2, **13 files**: `x` glued to its operand. `(1) x 3` parses;
`((1)x3, 2)` fails; and `(1)x3` alone is SILENTLY degenerate — a clean tree
that is not a parse of its source, which is worse than an error.

§0.13 rank 5, **6 files**: regex modifiers on the closing-delimiter line.
`s!a!b!x` is degenerate while `s!a!b!g` parses — which is rank 2 again,
wearing a different hat: after the final delimiter the lexer is in operator
position, and `x` there is the repetition operator rather than a modifier.

`t/base/lex.t` line 1 is the canonical trap: `$x = $#[0];` — not `$#` of an
anonymous array, but `$#` then `[0]`.

And from the same file: `eval '$foo{1} / 1;'` — is `/` division or a
pattern, immediately after a hash subscript? Perl says division, because a
subscript closes a term.

## Acceptance

- [ ] Slash after a term is division, after an operator is a pattern (`go test ./internal/lexer/ -run TestExpectSlash -count=1`)
- [ ] `x` in operator position is repetition even when glued to its operand (`go test ./internal/lexer/ -run TestExpectRepetition -count=1`)
- [ ] Regex modifiers on the closing-delimiter line lex as modifiers, not operators (`go test ./internal/lexer/ -run TestExpectRegexModifiers -count=1`)
- [ ] `$#[0]` lexes as `$#` then a subscript, per lex.t line 1 (`go test ./internal/lexer/ -run TestExpectArrayLastIndex -count=1`)
- [ ] Angle brackets after a term are comparison, after an operator a readline (`go test ./internal/lexer/ -run TestExpectAngle -count=1`)

Findings: §0.13 ranks 2 and 5 (19 files). Spec: ch3, perl.h:5972-5985.
