---
title: "Compound assignment: thirteen operators, and the corpus writes none"
state: pending
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.345262752Z
updated: 2026-09-23T02:07:02.345262752Z
---

Thirteen compound assignment operators, and the corpus writes none of
them. Measured across all 179 files' `--- source` sections: zero
occurrences of `+=`, `-=`, `*=`, `/=`, `.=`, `%=`, `**=`, `x=`, `||=`,
`&&=`, `//=`, `|=`, `&=`, `^=`, `<<=`, `>>=`.

perlop's level 20 is the second-largest level in the table. The corpus
makes no claim about any of it. Plain `=` appears in ~155 files binding a
literal, which is scaffolding rather than a claim by the corpus's own
rule.

## What a parser gets wrong here, and what catches it

`$x += 1` and `$x = $x + 1` compute the same value, so OUTPUT cannot
separate them. They are different trees, but the corpus does not assert
trees. What separates them is the TOKEN STREAM, measured:

    $x += 1        ->  Variable($x) Operator(+=) Number(1)
    $x = $x + 1    ->  Variable($x) Operator(=) Variable($x)
                       Operator(+) Number(1)

One token against three. A lexer that split `+=` into `+` and `=` would
produce a stream the parser could still make sense of -- and a wrong tree
with correct output. That is precisely the case the token layer exists
for.

## Not one family, measured

`//=` and `||=` are NOT `+=` with a different operator character. They
short-circuit, and they emit dedicated ops where the arithmetic forms
reuse the binary op:

    $x += 2      ->  add
    $u //= 5     ->  dorassign
    $u ||= 5     ->  orassign

The ops matter only for TIER PLACEMENT -- `dorassign` and `orassign` are
unclaimed, `add` is tier 04's -- but the split tells us these are two
claims rather than one file's worth.

The short-circuit is also the only one with observable behaviour: `$u //=
5` leaves a defined `$u` untouched, which output can see. The arithmetic
forms cannot be caught by output at all.

## Files

Five, in `04_operators`: the arithmetic/string group (`+= -= *= /= .=`),
the short-circuit group (`//= ||= &&=`), `**=` (right-associative, and
the tier already pins `**` itself), `x=` (whose right operand is a count,
not a value), and the bitwise/shift compound forms -- which depend on the
bitwise issue landing first, since `|=` cannot be claimed before `|`.

## Acceptance criteria

- [ ] Tier 04 holds files for the compound assignment operators, each claiming its spelling as ONE operator token (`test 5 -le "$(ls conformance/04_operators/*.t | wc -l)"`)
- [ ] The corpus source contains compound assignment, which it did not before (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/04_operators/*.t | grep -q '+='`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The dependency lint passes: no file uses an op introduced after its tier (`go test ./internal/conformance/ -run '^TestTierOperatorsLint$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsLint'`)
- [ ] The tier's adjacency file holds every construct the tier introduces (`go test ./internal/conformance/ -run '^TestTierOperatorsAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsAdjacency'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
