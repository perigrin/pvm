---
title: "Lexer feedback: term-vs-operator position"
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:44.556728897Z
updated: 2026-09-14T22:45:18.755159683Z
---

`/` is division or a pattern; `<FH>` is a readline or a glob; `<<` is a left
shift or a heredoc; `x` is an operator or the start of an identifier. Perl
decides with one piece of lexer state — is a TERM expected, or an OPERATOR —
and a static lexer has to carry the same state.

This issue owns that state AND the dispatch of every construct whose
identity depends on it. The quote-scanner issue owns the keyword-introduced
forms, which need no such knowledge; the bare forms are here because they do.

## Define all eleven states now

Spec §2.15 item 3: "All eleven states (chapter 3 §3.1). Everything else
depends on it." §3.9.4: "Do not economise here." §3.12: "`Expect` enum with
all 11 states."

perl's enum is at `perl.h:5972-5985`. An earlier draft of this issue said to
implement "the subset M0's corpus needs"; that was wrong for the same reason
trivia-as-tokens is right — the states are load-bearing structure, and
retrofitting the missing seven later is a rewrite rather than an addition.
Define all eleven and the transition-by-token-class table; exercise the
subset the corpus reaches.

`XSTATE` in particular is not optional: POD recognition depends on it, and
so does `{` after a bare sigil (§3.1.4).

## The measured cases

§0.13 rank 2, **13 files**: `x` glued to its operand. `(1) x 3` parses;
`((1)x3, 2)` fails; and `(1)x3` alone is SILENTLY degenerate — a clean tree
that is not a parse of its source, which is worse than an error.

§0.13 rank 5, **6 files**: regex modifiers on the closing-delimiter line.
`s!a!b!x` degenerate, `s!a!b!g` fine — rank 2 wearing a different hat, since
after the final delimiter the lexer is in operator position and `x` there is
repetition rather than a modifier.

§0.13 rank 7, **3 files**: a bareword followed by `&&`. Spec §3.2 gives the
`&` row: operator position yields `&&`, term position yields `&name`. One
row in the table this issue is already building, so it lands here rather
than being deferred.

`t/base/lex.t:10` is the canonical trap: `$x = $#[0];` — not `$#` of an
anonymous array, but `$#` followed by a subscript. (Line 1 is `#!./perl`; an
earlier draft of this issue said line 1.)

`t/base/lex.t:20` has the other: `eval '$foo{1} / 1;'` — division or
pattern, immediately after a hash subscript? Perl says division, because a
subscript closes a term.

## Acceptance

- [ ] All eleven expect states exist with a transition table, XSTATE included (`go test ./internal/lexer/ -run '^TestExpectStatesComplete$' -count=1 -v | grep -q '^--- PASS: TestExpectStatesComplete'`)
- [ ] Slash after a term is division, after an operator a pattern (`go test ./internal/lexer/ -run '^TestExpectSlash$' -count=1 -v | grep -q '^--- PASS: TestExpectSlash'`)
- [ ] `x` in operator position is repetition even when glued to its operand (`go test ./internal/lexer/ -run '^TestExpectRepetition$' -count=1 -v | grep -q '^--- PASS: TestExpectRepetition'`)
- [ ] Regex modifiers on the closing-delimiter line lex as modifiers (`go test ./internal/lexer/ -run '^TestExpectRegexModifiers$' -count=1 -v | grep -q '^--- PASS: TestExpectRegexModifiers'`)
- [ ] `$#[0]` lexes as `$#` then a subscript, per lex.t line 10 (`go test ./internal/lexer/ -run '^TestExpectArrayLastIndex$' -count=1 -v | grep -q '^--- PASS: TestExpectArrayLastIndex'`)
- [ ] Angle brackets and `&` dispatch on position: comparison vs readline, `&&` vs `&name` (`go test ./internal/lexer/ -run '^TestExpectAngleAndAmp$' -count=1 -v | grep -q '^--- PASS: TestExpectAngleAndAmp'`)

Findings: §0.13 ranks 2, 5, 7 (22 files). Spec: §2.15 item 3, §3.1.4, §3.2, §3.9.4, §3.12. perl.h:5972-5985.
