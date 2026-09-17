---
title: "The intuit_curly lookahead: a brace after a list operator"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
blocks:
- 01a0ad52-f1e4-7e19-80c7-a99f859951eb
created: 2026-09-16T22:24:52.875856958Z
updated: 2026-09-17T19:56:35.802547891Z
---

A one-line change here is correct and makes T1 worse. That is the finding, and it is why this issue is bigger than it looks.

## The defect

`expect.go`'s CloseBracket case applies perl's right-paren lookahead to
EVERY closing bracket:

    if t.nextIsOpenBrace {
        return XBlock
    }

Its own comment says "A `)` with a `{` next" and cites `yyl_rightparen`
(toke.c:7156), which handles the right PAREN alone. The code checks no byte.
So a subscript chain opens a block:

    $a[0]{k}     the `{` follows a `]`
    $h{a}{b}     the `{` follows a `}`
    $x->[0]{k}

All three are subscripts of the value the previous bracket produced. Nothing
in them opens a block. Verified 2026-09-17 at 3c997a6b: all three set
`OpensBlock` on their second brace.

## The fix, and why it is not committed

    if t.nextIsOpenBrace && len(t.text) == 1 && t.text[0] == ')' {

Correct by toke.c, and it turns the failing test green. Measured on T1
graded (`unit/*.t`, 986 files):

                      before    after
    T1 clean          408       406
    T1 Unknown        4,944     4,972

Worse on both axes. Something downstream depends on the wrong
classification, and until that is found, fixing the lexer in isolation moves
the number the wrong way. The likely shape is a parser path that reads
`OpensBlock` and gets a correct answer for the first time, having been
written against the incorrect one -- but that is a hypothesis, not a
measurement, and finding out is this issue's work rather than a side errand.

Two attempts have now stopped here from different directions: this one, and
the dereference leaf of 01a0ad52, which cannot proceed until brace
classification is right.

## What this issue actually covers

Not just the `intuit_curly` lookahead. The term-position brace decision as a
whole:

- the right-paren lookahead applying to every closer (above)
- `{` after a list operator, this issue's original subject
- `${EXPR}` vs `${name}`, needed by 01a0ad52
- whatever downstream code the first item's fix disturbs

Doing them together is the point. Each one alone moves the T1 number the
wrong way, because they are one decision spread across the lexer.

## Acceptance

- [ ] `$a[0]{k}`, `$h{a}{b}`, `$x->[0]{k}` classify both brackets as terms (`go test ./internal/lexer/ -run '^TestRightParenLookaheadIsParenOnly$' -count=1 -v | grep -q '^--- PASS: TestRightParenLookaheadIsParenOnly'`)
- [ ] `if (...) {`, `while (...) {`, `for (;;) {` still open blocks (`go test ./internal/lexer/ -run '^TestRightParenStillOpensABlock$' -count=1 -v | grep -q '^--- PASS: TestRightParenStillOpensABlock'`)
- [ ] A brace after a list operator is decided by lookahead, not by position alone (`go test ./internal/lexer/ -run '^TestIntuitCurlyLookahead$' -count=1 -v | grep -q '^--- PASS: TestIntuitCurlyLookahead'`)
- [ ] `map { $_ => 1 } @a` and `map { ; $_ } @a` both parse (`go test ./internal/parse/ -run '^TestMapBlockForms$' -count=1 -v | grep -q '^--- PASS: TestMapBlockForms'`)
- [ ] Whatever regressed T1 from 408/4,944 when the right-paren rule was narrowed is identified and named in a test comment (`go test ./internal/lexer/ -run '^TestRightParenRegressionIsNamed$' -count=1 -v | grep -q '^--- PASS: TestRightParenRegressionIsNamed'`)
- [ ] T1 clean does not fall below 408 and T1 Unknown does not rise above 4,944 (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

Note on the ratchet AC: TestParseRatchet does not exist yet. It belongs to the
M1 gate (01a0a70f-3ccc), which owns the parse ratchet and has none of its
eight AC tests written.
