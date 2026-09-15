---
title: "Named unary and list operators: the keyword table M0 deferred"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-15T21:53:08.668833928Z
updated: 2026-09-15T21:53:33.991815256Z
---

Named unary operators and list operators: where perl's precedence stops being
a table and starts consulting the symbol table.

M0 deferred this deliberately. `expect.go`'s `after()` currently returns
XTerm for every Word, with the reasoning recorded there:

> A bareword is the one case the lexer genuinely cannot settle... perl
> decides with the symbol table, which a lexer does not have -- toke.c's
> yyl_keylookup consults the keyword table AND the stash... The full fix is a
> keyword table, which belongs with the keyword-versus-identifier work in M1.

This is that work.

## What a keyword table buys, and what it cannot

perl's named unaries bind tighter than comparison but looser than
arithmetic — `length $x + 1` is `length($x + 1)`, measured — and list
operators swallow everything to the closing paren or the end of statement.
`toke.c`'s `S_lop` (2174) and `UNI3` (285) are the reference; `UNI3` returns
FUNC1 when the next character is `(` and UNIOP otherwise, which is why
`length ($x) + 1` and `length $x + 1` differ.

A static table covers the builtins. It cannot cover a user sub whose
prototype changes its parse, which is the srefgen signal the whole fidelity
harness measures — and which stays an honest Unknown until M4.

§0.13 rank 6 is here too: `try`, `defer`, `method`, `true` used as ordinary
identifiers in 5 corpus files. A keyword table that does not know they are
feature-gated turns valid old code into a parse error.

## Acceptance

- [ ] Named unaries bind tighter than comparison, looser than arithmetic (`go test ./internal/ast/ -run '^TestNamedUnaryPrecedence$' -count=1 -v | grep -q '^--- PASS: TestNamedUnaryPrecedence'`)
- [ ] `length ($x) + 1` and `length $x + 1` parse differently, per UNI3 (`go test ./internal/ast/ -run '^TestNamedUnaryParenForm$' -count=1 -v | grep -q '^--- PASS: TestNamedUnaryParenForm'`)
- [ ] A list operator swallows to the statement end or closing paren (`go test ./internal/ast/ -run '^TestListOperatorExtent$' -count=1 -v | grep -q '^--- PASS: TestListOperatorExtent'`)
- [ ] Post-5.36 keywords parse as identifiers where the feature is off (`go test ./internal/ast/ -run '^TestFeatureGatedKeywords$' -count=1 -v | grep -q '^--- PASS: TestFeatureGatedKeywords'`)
- [ ] A call to an unknown sub is Unknown-shaped rather than guessed (`go test ./internal/ast/ -run '^TestUnknownSubIsNotGuessed$' -count=1 -v | grep -q '^--- PASS: TestUnknownSubIsNotGuessed'`)

Findings: §0.13 rank 6 (5 files). Spec: ch3 §3.8.1, ch4. toke.c:285, 2174.
