---
title: "Compound assignment: thirteen operators, and the corpus writes none"
state: done
urgency: normal
milestone: corpus-complete
blocked_by:
- 01a0cc04-3303-7103-8769-9edca646b56e
created: 2026-09-23T02:07:02.345262752Z
updated: 2026-09-24T20:47:47.833130476Z
sessions:
- start_sha: 8a2ffb94624374191b8f6f4266bc8a20b0daf957
  end_sha: cbaacf10d93150fdf2a15e13ddca50cfce26104a
  commits: 1
  started_at: 2026-09-23T12:53:54.982200237Z
  ended_at: 2026-09-23T13:08:06.479465135Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T12:53:54.982200237Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T13:08:06.479465135Z
observed_paths:
- conformance/04_operators/00_adjacency.t
- conformance/04_operators/22_compound_arithmetic.t
- conformance/04_operators/23_compound_string.t
- conformance/04_operators/24_compound_repeat.t
- conformance/04_operators/25_compound_shortcircuit.t
- conformance/04_operators/26_compound_bitwise.t
- conformance/04_operators/README.md
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier04_test.go
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

- [ ] Tier 04 holds files for the compound assignment operators, each claiming its spelling as ONE operator token (`test 5 -le "$(grep -c '^## ' conformance/mdtest/compound-assignment.md)"`)
- [ ] The corpus source contains compound assignment, which it did not before (`python3 tools/mdtest/tiersrc.py 04_operators | grep -q '+='`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The dependency lint passes: no file uses an op introduced after its tier (`go test ./internal/conformance/ -run '^TestTierOperatorsLint$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsLint'`)
- [ ] The tier's adjacency file holds every construct the tier introduces (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/adjacency-'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
