---
title: "Expression parser: the 32 precedence levels as a Pratt table"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-15T21:52:57.173529962Z
updated: 2026-09-15T21:53:22.185719545Z
---

The expression grammar, as a Pratt parser over the lexer's tokens.

Spec ch4 gives the precedence table in exactly the form a Pratt parser needs,
and §1.2 explains why this shape rather than a generated one:

> `goyacc` ships with the Go toolchain but is not in the standard library,
> and generating an LALR parser from a transcription of `perly.y` would
> inherit Perl's grammar conflicts without inheriting `toke.c`'s lexer
> feedback — the part that actually resolves them.

## Thirty-two levels, counted not remembered

    $ grep -cE '^%(left|right|nonassoc)' perly.y
    32

Verified at the pinned revision. An earlier pass in this project wrote "33"
into a README from memory and it propagated; the number is cheap to check and
worth checking.

Associativity is part of the table, not an afterthought: `2**3**2` is 512 and
not 64, `1-2-3` is -4 and not 2. Both are one-line measurements against real
perl and both belong in the tests.

## What this issue does NOT do

Named unary operators, list operators and their argument-boundary rules are
their own issue. They are where perl's precedence stops being a table and
starts consulting the symbol table, and mixing them in here makes both
harder to test.

## Acceptance

- [ ] All 32 precedence levels are present, in perly.y order (`go test ./internal/ast/ -run '^TestPrecedenceTableComplete$' -count=1 -v | grep -q '^--- PASS: TestPrecedenceTableComplete'`)
- [ ] Right associativity holds: 2**3**2 groups right (`go test ./internal/ast/ -run '^TestRightAssociative$' -count=1 -v | grep -q '^--- PASS: TestRightAssociative'`)
- [ ] Left associativity holds: 1-2-3 groups left (`go test ./internal/ast/ -run '^TestLeftAssociative$' -count=1 -v | grep -q '^--- PASS: TestLeftAssociative'`)
- [ ] Ternary and assignment nest correctly at their own levels (`go test ./internal/ast/ -run '^TestTernaryAndAssignment$' -count=1 -v | grep -q '^--- PASS: TestTernaryAndAssignment'`)
- [ ] An unparseable operand becomes Unknown rather than a guess (`go test ./internal/ast/ -run '^TestExpressionUnknownNotGuess$' -count=1 -v | grep -q '^--- PASS: TestExpressionUnknownNotGuess'`)

Plan: M1. Spec: ch4 §4.x, §1.2. perly.y:150-182.
