---
title: "Expression parser: the 32 precedence levels as a Pratt table"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-15T21:52:57.173529962Z
updated: 2026-09-15T23:14:54.372074032Z
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
    $ grep -cE '^%nonassoc' perly.y
    11

Verified at the pinned revision. An earlier pass in this project wrote "33"
into a README from memory and it propagated; the number is cheap to check.

Eleven of the 32 are `%nonassoc`, and a Pratt table alone does not express
them — a binding power that stops recursion still ACCEPTS `$a <=> $b <=> $c`
rather than rejecting it. Measured:

    $ perl -e 'my $q = $a <=> $b <=> $c'
    syntax error at -e line 1, near "$b <=>"

Chained comparison (5.32+) is the other half of the same rule, and it is the
one case where the shape differs from nested binaries:

    $ perl -MO=Deparse,-p -e 'my $q = $a < $b < $c'
    (my($q) = ($a < $b < $c));      # one n-ary node, $b evaluated once

Spec §4.3 gives the four operator classes and the two `yyerror` productions
that make CHRELOP chain and NCEQOP not.

## Term forms are here too

§4.4 — literals, variables and sigils, anonymous constructors, `\`
references, every dereference form, slices and kv-slices. The Pratt loop
needs a `nud` for each of these or it has nothing to bind. §4.4.5's `\` is
the srefgen signal the fidelity harness measures, so it is not optional
scaffolding.

## What this issue does NOT do

Named unaries and list operators are their own issue: they are where
precedence stops being a table and starts consulting the keyword table.
Method calls (§4.5) and sort/map/grep (§4.9) are M2's.

## Acceptance

- [ ] All 32 precedence levels are present, in perly.y order (`go test ./internal/parse/ -run '^TestPrecedenceTableComplete$' -count=1 -v | grep -q '^--- PASS: TestPrecedenceTableComplete'`)
- [ ] Right associativity holds: 2**3**2 groups right (`go test ./internal/parse/ -run '^TestRightAssociative$' -count=1 -v | grep -q '^--- PASS: TestRightAssociative'`)
- [ ] Left associativity holds: 1-2-3 groups left (`go test ./internal/parse/ -run '^TestLeftAssociative$' -count=1 -v | grep -q '^--- PASS: TestLeftAssociative'`)
- [ ] The 11 nonassoc levels REJECT a second operator rather than accepting it (`go test ./internal/parse/ -run '^TestNonassocRejects$' -count=1 -v | grep -q '^--- PASS: TestNonassocRejects'`)
- [ ] `$a < $b < $c` builds one chained node, not nested binaries (`go test ./internal/parse/ -run '^TestComparisonChains$' -count=1 -v | grep -q '^--- PASS: TestComparisonChains'`)
- [ ] Ternary and assignment nest correctly at their own levels (`go test ./internal/parse/ -run '^TestTernaryAndAssignment$' -count=1 -v | grep -q '^--- PASS: TestTernaryAndAssignment'`)
- [ ] Every §4.4 term form parses: literals, sigils, anon constructors, `\`, derefs, slices (`go test ./internal/parse/ -run '^TestTermForms$' -count=1 -v | grep -q '^--- PASS: TestTermForms'`)
- [ ] An unparseable operand becomes Unknown rather than a guess (`go test ./internal/parse/ -run '^TestExpressionUnknownNotGuess$' -count=1 -v | grep -q '^--- PASS: TestExpressionUnknownNotGuess'`)

Plan: M1. Spec: ch4 §4.1, §4.2, §4.3, §4.4; §1.2. perly.y:150-182, 1462-1495.
