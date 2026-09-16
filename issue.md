---
title: "The intuit_curly lookahead: a brace after a list operator"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-16T22:24:52.875856958Z
updated: 2026-09-16T22:25:09.747758976Z
---

`map { ; a => 1 }` parses as `call(anon_hash)` instead of a call taking a
block, because the lexer classifies the `{` before anything inside it is seen.

A list operator leaves a TERM expected, so the brace stack reads `{` in term
position and calls it a hash constructor. That is correct for `$h = { a => 1 }`
and wrong here.

## perl's rule

`intuit_curly` peeks PAST the brace at the first token inside:

    map { ; a => 1 } (1,2)      the `;` forces a BLOCK
    map { a => 1 }, (1,2)       bareword-then-`=>` forces a HASHREF
    map { a => 1 } (1,2)        hashref guessed, then no comma -> syntax error

Measured, and the difference is observable:

    $ perl -e 'my @r = map { a => 1 }, (1,2); print scalar @r'    2
    $ perl -e 'my @r = map { ; a => 1 } (1,2); print scalar @r'   4

Two elements per input versus four, decided by one leading semicolon.

## Why it is its own issue

This is the ONE place in the grammar where a brace's meaning depends on what
is inside it rather than on what precedes it. Everywhere else the expect
state settles it, which is why the brace stack works for every other case --
`$h{a}`, `sub f {`, `if (..) {`, a bare block at statement start.

Adding the lookahead means the lexer peeking an unbounded distance past a
token it has not consumed, or the parser overriding a decision the lexer
already recorded on the token. Neither is obviously right, and picking one
without measuring the corpus cost would be guessing.

## Acceptance

- [ ] `map { ; a => 1 } (1,2)` parses the brace as a Block (`go test ./internal/parse/ -run '^TestBlockVersusHashrefAfterListOp$' -count=1 -v | grep -q '^--- PASS: TestBlockVersusHashrefAfterListOp'`)
- [ ] `map { a => 1 }, (1,2)` parses the brace as an AnonHash (same test)
- [ ] The skip is removed from TestBlockVersusHashrefAfterListOp
- [ ] 620/620 corpus files still round-trip (`go test ./internal/parse/ -run '^TestTreeRoundTripsCorpus$' -count=1 -v | grep -q '^--- PASS: TestTreeRoundTripsCorpus'`)

Spec: ch4 §4.9.2. toke.c `intuit_curly`.
